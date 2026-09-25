// Package modelhealth periodically verifies that the models from
// models-config.json that have an active on-chain bid from this provider
// actually respond to inference requests. Results are cached in memory and
// exposed through the public /healthcheck endpoint, deliberately excluding
// private config fields (the models-config modelName, apiUrl, apiKey). The
// report's modelName is the public name registered on-chain, which anyone
// can already derive from the model ID.
package modelhealth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apiprobe"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/attestation"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
)

const (
	bidsPageLimit = 100
	// maxBidPages caps pagination (10k bids) as a fail-safe against a
	// misbehaving RPC that keeps returning full pages.
	maxBidPages = 100
)

type AdapterProvider interface {
	GetAdapter(ctx context.Context, chatID, modelID, sessionID common.Hash, storeChatContext, forwardChatContext bool) (aiengine.AIEngineStream, error)
}

type BidProvider interface {
	GetActiveBidsByProvider(ctx context.Context, provider common.Address, offset *big.Int, limit uint8, order registries.Order) ([]*structs.Bid, error)
}

type ModelMetaProvider interface {
	GetModelNameAndTags(ctx context.Context, modelID common.Hash) (string, []string, error)
}

type ModelConfigProvider interface {
	GetAll() ([]common.Hash, []config.ModelConfig)
}

// TeeStatusProvider exposes the backend TEE attestation state and lets the
// checker re-run attestation on every sweep (implemented by
// attestation.BackendVerifier). Optional dependency.
type TeeStatusProvider interface {
	GetStatus(modelID string) *attestation.BackendAttestationSnapshot
	// ReattestBackend re-runs full backend attestation for the model;
	// endpoint is the model's apiUrl used to resolve the attestation URL
	// when no snapshot exists yet.
	ReattestBackend(ctx context.Context, modelID string, endpoint string) error
}

type ApiDetector interface {
	Detect(ctx context.Context, cfg config.ModelConfig) *system.ModelApiSpec
}

// ProbeAwareDetector is the optional side of an ApiDetector that can say
// beforehand whether Detect will reach the backend (a cache miss) rather
// than answer from its cache; the sweep paces those calls by probeDelay.
// A detector without it is paced on every call.
type ProbeAwareDetector interface {
	WouldProbe(cfg config.ModelConfig) bool
}

// DefaultMaxConsecutiveErrors is the number of consecutive real-session
// prompt failures after which a model is flipped to unhealthy (or degraded,
// when the whole streak is upstream rate limiting) without waiting for the
// next scheduled probe sweep.
const DefaultMaxConsecutiveErrors = 3

// A cold detector cache must not consume the whole model-health timeout.
const detectBudget = 10 * time.Second

// modelMeta caches the public on-chain facts about a model so each sweep
// only pays one registry call per previously unseen model.
type modelMeta struct {
	name      string
	modelType structs.ModelType
	isTee     bool
	// mediaOnly marks a model whose tags name a media family this checker has
	// no probe for. DetectModelType does not recognise these, so they arrive as
	// Unknown and would otherwise inherit the LLM fallback below and be sent a
	// text prompt.
	mediaOnly bool
}

// unprobeableMediaTags are tag families that clearly are not text models but
// that DetectModelType does not classify, so they resolve to Unknown. They are
// deliberately NOT part of DetectModelType: this is checker policy about what
// can be probed, not a claim about the model type the rest of the router uses.
//
// "chat" and "code" are absent on purpose. Both are Unknown today - the
// create-model API even tells you to tag "chat" while DetectModelType looks
// only for llm/textgeneration/t2t - and both should take the LLM probe.
var unprobeableMediaTags = map[string]struct{}{
	"image":         {},
	"images":        {},
	"text2image":    {},
	"text-to-image": {},
	"t2i":           {},
	"video":         {},
	"text2video":    {},
	"text-to-video": {},
	"t2v":           {},
	"audio":         {},
	"music":         {},
}

func hasUnprobeableMediaTag(tags []string) bool {
	for _, raw := range tags {
		if _, ok := unprobeableMediaTags[strings.ToLower(strings.TrimSpace(raw))]; ok {
			return true
		}
	}
	return false
}

type Checker struct {
	deps       Deps
	interval   time.Duration
	timeout    time.Duration
	probeDelay time.Duration
	// maxConsecutiveErrors is the threshold of consecutive session prompt
	// failures that dynamically flips a model to unhealthy or degraded.
	maxConsecutiveErrors int
	log                  lib.ILogger

	mu        sync.RWMutex
	reports   map[string]system.ModelHealthReport
	modelMeta map[string]modelMeta
	// failures tracks the consecutive real-session prompt failure streak per
	// model, including whether the whole streak was upstream rate limiting.
	failures map[string]failureStreak
	// runtime holds the evidence the health probe's own response established
	// per model (design contract §5, harvest) and what the verifier
	// established; merged onto the detection result after every probe.
	runtime *runtimeStore
	// sleepFn paces backend requests; nil means a timer. Tests observe the
	// pacing through it.
	sleepFn func(ctx context.Context, d time.Duration) bool

	// triggerCh carries manual re-check requests into the Run loop; buffered
	// at 1 so triggers queue at most one extra sweep and can never overlap.
	triggerCh chan struct{}

	// firstTeeSweepDone flips after the first sweep. Backend TEE attestation
	// already runs at process startup (cmd/main.go), and the first sweep
	// fires right after it, so that sweep only reads the cached snapshot;
	// re-attestation starts from the second sweep onward.
	firstTeeSweepDone bool
}

// Deps groups the external dependencies of the checker.
type Deps struct {
	Adapters     AdapterProvider
	Bids         BidProvider
	Models       ModelMetaProvider
	ModelConfigs ModelConfigProvider
	// TeeStatus is optional; when set, backend attestation of TEE-tagged
	// models is re-run on every sweep after the first (pass or fail; the
	// first sweep relies on the startup attestation), and models whose
	// attestation does not pass are reported as tee_unverified.
	TeeStatus TeeStatusProvider
	ApiDetect ApiDetector
	// ApiDetectTTL is MODEL_API_DETECT_TTL, the TTL the detector caches by;
	// runtime evidence harvested from the health probe is re-merged for the
	// same span after a probe stops producing it, and the verifier runs once
	// per span. <= 0 means DefaultRuntimeTTL.
	ApiDetectTTL time.Duration
	// VerifyBindings is MODEL_API_VERIFY_BINDINGS: after a healthy LLM
	// probe, once per TTL, resend the probe with the reasoning.disable
	// binding applied and with a dummy tool (design contract §5, verifier).
	VerifyBindings bool
	// ProbeCapabilities is MODEL_API_PROBE_CAPABILITIES: the json_schema,
	// image and vLLM render probes on the same schedule.
	ProbeCapabilities bool
	// ProbeHosted is MODEL_API_PROBE_HOSTED: let those requests reach hosted
	// vendors and gateways, which bill per request.
	ProbeHosted bool
}

func NewChecker(deps Deps, interval, timeout, probeDelay time.Duration, maxConsecutiveErrors int, log lib.ILogger) *Checker {
	if maxConsecutiveErrors <= 0 {
		maxConsecutiveErrors = DefaultMaxConsecutiveErrors
	}
	return &Checker{
		deps:                 deps,
		interval:             interval,
		timeout:              timeout,
		probeDelay:           probeDelay,
		maxConsecutiveErrors: maxConsecutiveErrors,
		log:                  log.Named("MODEL_HEALTH"),
		reports:              make(map[string]system.ModelHealthReport),
		modelMeta:            make(map[string]modelMeta),
		failures:             make(map[string]failureStreak),
		runtime:              newRuntimeStore(deps.ApiDetectTTL),
		triggerCh:            make(chan struct{}, 1),
	}
}

// Run probes all configured models immediately and then on every interval
// tick until the context is cancelled. It never returns a non-context error
// so a probe failure cannot bring down the provider errgroup.
func (c *Checker) Run(ctx context.Context, walletAddr common.Address) error {
	c.checkAll(ctx, walletAddr)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			c.checkAll(ctx, walletAddr)
		case <-c.triggerCh:
			c.log.Info("manual model health re-check triggered")
			c.checkAll(ctx, walletAddr)
			// restart the scheduled cadence relative to the manual sweep so
			// a trigger isn't immediately followed by a scheduled sweep
			ticker.Reset(c.interval)
		}
	}
}

// TriggerNow queues an immediate re-check sweep on the Run loop and reports
// whether it was accepted. It returns false when a manual sweep is already
// queued; the sweep itself runs asynchronously — callers observe progress
// through the lastChecked timestamps in the reports.
func (c *Checker) TriggerNow() bool {
	select {
	case c.triggerCh <- struct{}{}:
		return true
	default:
		return false
	}
}

// GetReports returns a snapshot of the latest per-model reports sorted by model ID.
func (c *Checker) GetReports() []system.ModelHealthReport {
	c.mu.RLock()
	defer c.mu.RUnlock()

	reports := make([]system.ModelHealthReport, 0, len(c.reports))
	for _, r := range c.reports {
		reports = append(reports, r)
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].ModelID < reports[j].ModelID })
	return reports
}

// checkAll reports on the provider's whole ecosystem: every configured model
// (probed when it has an active bid) plus every active bid whose model is
// missing from models-config.json (reported as no_model_configured — the
// provider is selling capacity it cannot serve).
func (c *Checker) checkAll(ctx context.Context, walletAddr common.Address) {
	modelIDs, modelCfgs := c.deps.ModelConfigs.GetAll()

	bidsByModel, err := c.activeBidModels(ctx, walletAddr)
	if err != nil {
		c.log.Warnf("cannot get active bids by provider %s: %s", walletAddr, err)
		return
	}

	seen := make(map[string]bool, len(modelIDs)+len(bidsByModel))

	// Skip TEE re-attestation on the very first sweep: it runs right after
	// the startup attestation in cmd/main.go and would just duplicate it.
	c.mu.Lock()
	reattestTee := c.firstTeeSweepDone
	c.firstTeeSweepDone = true
	c.mu.Unlock()

	backendHit := false
	for i, modelID := range modelIDs {
		select {
		case <-ctx.Done():
			return
		default:
		}
		bidID, hasBid := bidsByModel[modelID]
		var cfg config.ModelConfig
		if i < len(modelCfgs) {
			cfg = modelCfgs[i]
		}
		// Pace backend requests so a provider with many models doesn't burst
		// its upstream (often a single account behind all models) and
		// self-inflict rate-limit failures. Only requests that reach the
		// backend are paced: the health probe of a model with an active bid,
		// and an API detection the detector cannot answer from its cache.
		// Both kinds share one pace, since they share that account.
		willHit := hasBid || c.detectionWouldProbe(cfg)
		if willHit && backendHit && !c.sleep(ctx, c.probeDelay) {
			return
		}
		c.checkModel(ctx, modelID, bidID, hasBid, cfg, reattestTee)
		if willHit {
			backendHit = true
		}
		seen[modelID.Hex()] = true
	}

	for modelID, bidID := range bidsByModel {
		if seen[modelID.Hex()] {
			continue
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
		c.reportUnconfigured(ctx, modelID, bidID)
		seen[modelID.Hex()] = true
	}

	c.prune(seen)
}

// reportUnconfigured records a bid whose model has no entry in
// models-config.json. There is no backend to probe, so only the on-chain
// facts (bid ID, model type from tags) are reported.
func (c *Checker) reportUnconfigured(ctx context.Context, modelID common.Hash, bidID common.Hash) {
	report := system.ModelHealthReport{
		ModelID:      modelID.Hex(),
		HasActiveBid: true,
		BidID:        bidID.Hex(),
		Status:       system.ModelHealthStatusNoModel,
		LastChecked:  time.Now().Unix(),
	}

	c.log.Warnf("active bid %s for model %s has no entry in models config", lib.Short(bidID), lib.Short(modelID))

	if meta, err := c.modelMetaFor(ctx, modelID); err == nil {
		report.ModelName = meta.name
		report.ModelType = string(meta.modelType)
	}

	c.setReport(report)
}

// detectable is the checkModel branch that hands cfg to the detector: a
// chat-transport model without a declared apiStack.
func (c *Checker) detectable(cfg config.ModelConfig) bool {
	return c.deps.ApiDetect != nil && apispec.StackFor(cfg.ApiStack) == "" && config.IsChatTransport(cfg.ApiType) && (cfg.ApiURL != "" || cfg.ModelName != "")
}

// detectionWouldProbe reports whether checkModel's detection of cfg will
// reach the backend; unknown for a detector without ProbeAwareDetector,
// which is then treated as probing.
func (c *Checker) detectionWouldProbe(cfg config.ModelConfig) bool {
	if !c.detectable(cfg) {
		return false
	}
	if aware, ok := c.deps.ApiDetect.(ProbeAwareDetector); ok {
		return aware.WouldProbe(cfg)
	}
	return true
}

// sleep blocks for d unless the context is cancelled first; it reports
// whether the full delay elapsed.
func (c *Checker) sleep(ctx context.Context, d time.Duration) bool {
	if c.sleepFn != nil {
		return c.sleepFn(ctx, d)
	}
	if d <= 0 {
		return true
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// prune drops cached reports for models that are no longer configured and no
// longer have an active bid, so removed models don't linger in the report.
func (c *Checker) prune(seen map[string]bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id := range c.reports {
		if !seen[id] {
			delete(c.reports, id)
			c.runtime.forget(id)
		}
	}
}

// activeBidModels maps each model with an active bid from this provider to
// its most recent bid ID (bids are fetched newest-first). Pages through all
// active bids so providers with more than one page are fully covered.
func (c *Checker) activeBidModels(ctx context.Context, walletAddr common.Address) (map[common.Hash]common.Hash, error) {
	byModel := make(map[common.Hash]common.Hash)

	for page := 0; ; page++ {
		if page == maxBidPages {
			c.log.Warnf("provider %s has more than %d active bids, health report may be incomplete", walletAddr, maxBidPages*bidsPageLimit)
			break
		}

		offset := big.NewInt(int64(page) * bidsPageLimit)
		bids, err := c.deps.Bids.GetActiveBidsByProvider(ctx, walletAddr, offset, bidsPageLimit, registries.OrderDESC)
		if err != nil {
			return nil, err
		}

		for _, bid := range bids {
			if _, ok := byModel[bid.ModelAgentId]; !ok {
				byModel[bid.ModelAgentId] = bid.Id
			}
		}

		if len(bids) < bidsPageLimit {
			break
		}
	}

	return byModel, nil
}

func (c *Checker) checkModel(ctx context.Context, modelID common.Hash, bidID common.Hash, hasBid bool, cfg config.ModelConfig, reattestTee bool) {
	report := system.ModelHealthReport{
		ModelID:      modelID.Hex(),
		HasActiveBid: hasBid,
		LastChecked:  time.Now().Unix(),
	}

	if hasBid {
		report.BidID = bidID.Hex()
	}

	prev, hasPrev := c.getReport(modelID.Hex())
	if hasPrev {
		report.LastHealthy = prev.LastHealthy
	}

	if cfg.ModelFamily != "" && !apispec.IsKnownFamily(cfg.ModelFamily) {
		c.log.Warnf("model %s: unknown modelFamily %q — no family bindings will be advertised", lib.Short(modelID), cfg.ModelFamily)
	}

	// Image adapters (prodia-*, hyperbolic-sd) and decisions have no chat API
	// to probe; detecting them would cost ~16 requests per sweep for nothing.
	// A declared capabilities block is published even when nothing detects
	// (detection off, or a transport it does not cover).
	var api *system.ModelApiSpec
	if apispec.StackFor(cfg.ApiStack) != "" {
		api = apispec.Build(cfg)
	} else if c.detectable(cfg) {
		budget := c.timeout
		if budget <= 0 || budget > detectBudget {
			budget = detectBudget
		}
		detectCtx, cancel := context.WithTimeout(ctx, budget)
		api = c.deps.ApiDetect.Detect(detectCtx, cfg)
		cancel()
	} else if cfg.Capabilities != nil {
		api = apispec.Build(cfg)
	}

	// The api block is finalised after the probe on every exit: the probe's
	// own response is runtime evidence (design contract §3 rule 11), and a
	// sweep that never probes still re-merges what an earlier one harvested.
	var meta *aiengine.ResponseMeta
	var prevApi *system.ModelApiSpec
	if hasPrev {
		prevApi = prev.Api
	}
	defer func() {
		report.Api = c.publishedApi(modelID, cfg, api, meta, prevApi)
		c.setReport(report)
	}()

	if !hasBid {
		report.Status = system.ModelHealthStatusNoBid
		return
	}

	modelMeta, err := c.modelMetaFor(ctx, modelID)
	if err != nil {
		c.log.Warnf("model %s: cannot resolve model metadata: %s", lib.Short(modelID), err)
		report.Status = system.ModelHealthStatusSkipped
		report.ErrorKind = system.ModelHealthErrorTypeLookup
		return
	}
	report.ModelName = modelMeta.name
	report.ModelType = string(modelMeta.modelType)

	// Re-run backend attestation for TEE-tagged models on every sweep after
	// the first (the first sweep relies on the startup attestation),
	// regardless of the current snapshot status: a transient failure at
	// startup self-heals here instead of sticking until a restart, and a
	// stale pass is re-proven against the live backend. A model whose
	// attestation does not pass cannot serve sessions (session requests are
	// rejected by the receiver), so it is reported as tee_unverified and the
	// backend probe is skipped.
	if modelMeta.isTee && c.deps.TeeStatus != nil {
		if reattestTee {
			attestCtx, cancel := context.WithTimeout(ctx, c.timeout)
			if err := c.deps.TeeStatus.ReattestBackend(attestCtx, modelID.Hex(), cfg.ApiURL); err != nil {
				c.log.Warnf("model %s: backend TEE re-attestation failed: %s", lib.Short(modelID), err)
			}
			cancel()
		}

		snapshot := c.deps.TeeStatus.GetStatus(modelID.Hex())
		if snapshot == nil || snapshot.Status != attestation.StatusPassed {
			c.log.Warnf("model %s: backend TEE attestation is not passed, reporting tee_unverified", lib.Short(modelID))
			report.Status = system.ModelHealthStatusTeeUnverified
			report.ErrorKind = system.ModelHealthErrorTeeAttestation
			return
		}
	}

	// typeGuessed records that the model carries no recognised type tag and the
	// LLM probe below is an assumption rather than a known type. A failed
	// assumption must not be reported as unhealthy: see the error branch after
	// the probe.
	typeGuessed := false

	// Only the LLM probe harvests response metadata: the OpenAI adapter is
	// the one that fills it, and it serves chat completions.
	llmProbe := func(ctx context.Context, adapter aiengine.AIEngineStream, report *system.ModelHealthReport) error {
		var err error
		meta, err = c.probeLLM(ctx, adapter, report)
		return err
	}

	var probe func(ctx context.Context, adapter aiengine.AIEngineStream, report *system.ModelHealthReport) error
	switch modelMeta.modelType {
	case structs.ModelTypeLLM:
		probe = llmProbe
	case structs.ModelTypeEMBEDDING:
		probe = c.probeEmbeddings
	case structs.ModelTypeDECISIONS:
		probe = c.probeDecisions
	case structs.ModelTypeUnknown:
		// A model with no recognised tag still needs a health answer. Skipping
		// publishes no health for it at all, and consumers that filter on the
		// provider's self-report drop the bid entirely, so a working backend
		// looks unavailable. Only the model owner can add tags, so the provider
		// serving it cannot fix this at source. Probe as LLM, which is what an
		// untyped model on this network almost always is; if the guess is
		// wrong the probe fails and reports unhealthy, which is a more useful
		// answer than silence. Known non-LLM types below still skip.
		//
		// Except where the tags name a media family DetectModelType does not
		// classify. Those reach Unknown too, and firing a text prompt at an
		// image or video backend buys an upstream call and a misleading
		// unhealthy row rather than information.
		if modelMeta.mediaOnly {
			report.Status = system.ModelHealthStatusSkipped
			return
		}
		typeGuessed = true
		probe = llmProbe
	default:
		report.Status = system.ModelHealthStatusSkipped
		return
	}

	adapter, err := c.deps.Adapters.GetAdapter(ctx, common.Hash{}, modelID, common.Hash{}, false, false)
	if err != nil {
		c.log.Warnf("model %s: cannot get adapter: %s", lib.Short(modelID), err)
		report.Status = system.ModelHealthStatusUnhealthy
		report.ErrorKind = system.ModelHealthErrorBadResponse
		return
	}

	probeCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	start := time.Now()
	err = probe(probeCtx, adapter, &report)
	report.LatencyMs = time.Since(start).Milliseconds()

	if err != nil {
		c.log.Warnf("model %s: probe failed: %s", lib.Short(modelID), err)
		// 429 means the backend is up and correctly configured, just
		// throttled right now — capacity risk, not failure, and materially
		// different from a billing (402) or configuration (404) problem.
		// Reported as degraded so consumers still treat the model as
		// serviceable (only the strict session health policy excludes it).
		if report.HttpStatus == http.StatusTooManyRequests {
			report.Status = system.ModelHealthStatusDegraded
			report.ErrorKind = system.ModelHealthErrorRateLimited
		} else if typeGuessed {
			// The type was assumed, not known, so a failed probe is not
			// evidence the backend is broken - it is just as likely that this
			// is not a text model at all. That distinction matters because
			// unhealthy is excluded by EVERY session health policy, including
			// the default permissive one, while skipped is not. Reporting
			// unhealthy here would take an untagged image or video bid that was
			// merely unroutable and hard-block it from opening sessions, which
			// is strictly worse than the silence this patch set out to fix.
			// The error kind is still recorded so the failure is diagnosable.
			report.Status = system.ModelHealthStatusSkipped
			report.ErrorKind = classifyError(err)
		} else {
			report.Status = system.ModelHealthStatusUnhealthy
			report.ErrorKind = classifyError(err)
		}
		return
	}

	report.Status = system.ModelHealthStatusHealthy
	report.LastHealthy = time.Now().Unix()
	// A passing probe is direct evidence the backend works again, so the
	// consecutive session-failure streak (if any) is over.
	c.resetFailures(modelID)

	// The inference-traffic probes follow a healthy chat probe only: they
	// are chat requests, and a backend that failed the plain one answers
	// nothing about its features.
	if (modelMeta.modelType == structs.ModelTypeLLM || typeGuessed) && api != nil {
		c.verify(ctx, modelID, cfg, api, adapter)
	}
}

// verify runs the verifier and the capability probes (design contract §5)
// for a model whose probe just passed, when the flags allow: once per TTL
// per (model, spec), each request paced by probeDelay, and never against a
// hosted vendor or gateway unless MODEL_API_PROBE_HOSTED. The spec the
// probes are chosen from carries the runtime evidence already established,
// so a binding an earlier sweep saw rejected is not sent again; the
// once-per-TTL key hashes the detection result itself, which is stable
// across sweeps. A run cut short by the context records nothing, so the
// next sweep retries.
func (c *Checker) verify(ctx context.Context, modelID common.Hash, cfg config.ModelConfig, api *system.ModelApiSpec, adapter aiengine.AIEngineStream) {
	if !c.deps.VerifyBindings && !c.deps.ProbeCapabilities {
		return
	}
	if api.Stack == "" || api.Via != "" {
		return
	}
	if apispec.IsHostedOrGatewayStack(api.Stack) && !c.deps.ProbeHosted {
		return
	}
	id := modelID.Hex()
	key := verifyKey(api)
	if !c.runtime.verifyDue(id, key) {
		return
	}
	spec := api.Clone()
	if rt, ok := c.runtime.harvest(id, nil); ok {
		apispec.MergeRuntime(spec, rt)
	}

	_, _, prompt := generateArithmeticPrompt()
	v := &apiprobe.Verifier{
		Adapter: adapter, Stack: spec.Stack, ModelName: cfg.ModelName, ApiURL: cfg.ApiURL, ApiKey: cfg.ApiKey,
		Prompt: prompt,
		Pace:   func(ctx context.Context) bool { return c.sleep(ctx, c.probeDelay) },
		Now:    c.runtime.now,
	}
	var verdicts []apiprobe.Verdict
	run := func(probe func(context.Context) apiprobe.Verdict) bool {
		probeCtx, cancel := context.WithTimeout(ctx, c.timeout)
		verdict := probe(probeCtx)
		cancel()
		if ctx.Err() != nil {
			return false
		}
		verdicts = append(verdicts, verdict)
		line := fmt.Sprintf("model %s: verify %s: %s (HTTP %d): %s", lib.Short(modelID), verdict.Probe, verdict.Outcome, verdict.Status, verdict.Reason)
		switch verdict.Outcome {
		case apiprobe.OutcomeRefuted, apiprobe.OutcomeRejected, apiprobe.OutcomeRefused:
			c.log.Info(line)
		default:
			c.log.Debug(line)
		}
		return true
	}

	disableVerified := false
	if c.deps.VerifyBindings {
		if apiprobe.ReasoningDisableBinding(spec) != nil {
			if !run(func(ctx context.Context) apiprobe.Verdict { return v.VerifyReasoningDisable(ctx, spec) }) {
				return
			}
			disableVerified = verdicts[len(verdicts)-1].Outcome == apiprobe.OutcomeAccepted
		}
		if spec.Capabilities == nil || spec.Capabilities.Tools == nil {
			if !run(v.ProbeTools) {
				return
			}
		}
	}
	if c.deps.ProbeCapabilities {
		if !hasCapListEntry(spec, apispec.CapFieldStructured, "json_schema") {
			if !run(v.ProbeJSONSchema) {
				return
			}
		}
		if !hasCapListEntry(spec, apispec.CapFieldModalitiesInput, "image") {
			if !run(v.ProbeImage) {
				return
			}
		}
		// The render check proves the template honours the kwarg without a
		// completion; redundant once the completion with the kwarg applied
		// already came back without reasoning content.
		if _, _, ok := apiprobe.TemplateToggle(spec); ok && spec.Stack == "vllm" && !disableVerified {
			if !run(func(ctx context.Context) apiprobe.Verdict { return v.RenderThinkingToggle(ctx, spec) }) {
				return
			}
		}
	}

	var facts apispec.Runtime
	apiprobe.Fold(&facts, verdicts...)
	c.runtime.recordVerify(id, key, facts)
}

// verifyKey names the spec a verifier run answered for: the stack, the
// gateway and the bindings (intent, kind, parameter, value). A changed
// binding table is verified afresh before the TTL; runtime evidence is
// excluded so the key does not change with the verdicts themselves.
func verifyKey(api *system.ModelApiSpec) string {
	type entry struct {
		Kind  string `json:"kind"`
		Param string `json:"param"`
		Value any    `json:"value"`
	}
	bindings := make(map[string]entry, len(api.Bindings))
	for intent, b := range api.Bindings {
		if b != nil {
			bindings[intent] = entry{Kind: b.Kind, Param: b.Param, Value: b.Value}
		}
	}
	raw, _ := json.Marshal(struct {
		Stack    string           `json:"stack"`
		Via      string           `json:"via"`
		Bindings map[string]entry `json:"bindings"`
	}{api.Stack, api.Via, bindings})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func hasCapListEntry(api *system.ModelApiSpec, field, value string) bool {
	if api.Capabilities == nil {
		return false
	}
	var list []string
	switch field {
	case apispec.CapFieldStructured:
		list = api.Capabilities.Structured
	case apispec.CapFieldModalitiesInput:
		if api.Capabilities.Modalities != nil {
			list = api.Capabilities.Modalities.Input
		}
	}
	for _, have := range list {
		if have == value {
			return true
		}
	}
	return false
}

// publishedApi is the api block of one report: the declared or detected spec
// with the model's runtime evidence merged onto a clone (the detector's cache
// stays free of it), or nil when nothing is known. A LiteLLM proxy seen only
// in the probe's headers names the stack on its own when detection found
// nothing; the composer adds no reasoning knob to it, since no reasoning
// evidence exists. declaredAt survives a sweep whose merged spec is equal to
// the previous one.
func (c *Checker) publishedApi(modelID common.Hash, cfg config.ModelConfig, api *system.ModelApiSpec, meta *aiengine.ResponseMeta, prev *system.ModelApiSpec) *system.ModelApiSpec {
	rt, hasRuntime := c.runtime.harvest(modelID.Hex(), meta)
	if api == nil && hasRuntime && len(rt.StackHints) > 0 {
		if stack := apispec.StackFor(rt.StackHints[0]); stack != "" {
			api = apispec.Compose(apispec.Evidence{Stack: stack, ModelName: cfg.ModelName, ServedModelID: rt.ServedModelID})
			if api.Evidence == nil {
				api.Evidence = map[string]string{}
			}
			api.Evidence[evidenceKeyStack] = sourceStackHeaders
		}
	}
	if api == nil {
		return nil
	}
	live := api.Clone()
	if hasRuntime {
		if trace := apispec.MergeRuntime(live, rt); len(trace) > 0 {
			c.log.Debugf("model %s: runtime evidence: %s", lib.Short(modelID), strings.Join(trace, "; "))
		}
		// A refuted binding's evidence names what the probe saw
		// (verify:reasoning_content_still_present@<at>), so the CLI can show
		// why verified is false; MergeRuntime writes the intent's own name.
		for intent, verdict := range rt.Verified {
			if verdict.Verified || verdict.Reason == "" || live.Bindings[intent] == nil {
				continue
			}
			if slug := apiprobe.ReasonSlug(verdict.Reason); slug != "" {
				if live.Evidence == nil {
					live.Evidence = map[string]string{}
				}
				live.Evidence[system.EvidenceKeyBinding(intent)] = apiprobe.VerifySource(slug, verdict.At)
			}
		}
	}
	live.DeclaredAt = time.Now().Unix()
	if prev != nil && sameApiSpecIgnoringDeclaredAt(prev, live) {
		live.DeclaredAt = prev.DeclaredAt
	}
	return live
}

func sameApiSpecIgnoringDeclaredAt(a, b *system.ModelApiSpec) bool {
	ac, bc := *a, *b
	ac.DeclaredAt, bc.DeclaredAt = 0, 0
	return reflect.DeepEqual(ac, bc)
}

// probeLLM sends the arithmetic prompt and returns what the response revealed
// about the backend (aiengine.ResponseMeta); the meta is filled by the OpenAI
// adapter only and stays empty for the others.
func (c *Checker) probeLLM(ctx context.Context, adapter aiengine.AIEngineStream, report *system.ModelHealthReport) (*aiengine.ResponseMeta, error) {
	a, b, prompt := generateArithmeticPrompt()

	req := &gcs.OpenAICompletionRequestExtra{
		ChatCompletionRequest: openai.ChatCompletionRequest{
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Stream: false,
		},
	}

	meta := &aiengine.ResponseMeta{}
	var response strings.Builder
	var engineErr error
	err := adapter.Prompt(aiengine.WithResponseMeta(ctx, meta), req, func(ctx context.Context, chunk gcs.Chunk, aiEngineErr *gcs.AiEngineErrorResponse) error {
		if aiEngineErr != nil {
			report.HttpStatus = aiEngineErr.StatusCode
			engineErr = errors.New("model backend returned an error response")
			return nil
		}
		if chunk != nil {
			response.WriteString(chunk.String())
		}
		return nil
	})
	if err != nil {
		return meta, err
	}
	if engineErr != nil {
		return meta, engineErr
	}
	if response.Len() == 0 {
		return meta, errors.New("empty completion response")
	}

	correct := verifyArithmeticAnswer(response.String(), a+b)
	report.PromptCorrect = &correct
	return meta, nil
}

func (c *Checker) probeEmbeddings(ctx context.Context, adapter aiengine.AIEngineStream, report *system.ModelHealthReport) error {
	req := &gcs.EmbeddingsRequest{
		EmbeddingRequest: openai.EmbeddingRequest{Input: "health check"},
	}

	var gotVector bool
	err := adapter.Embeddings(ctx, req, func(ctx context.Context, chunk gcs.Chunk, aiEngineErr *gcs.AiEngineErrorResponse) error {
		if aiEngineErr != nil {
			report.HttpStatus = aiEngineErr.StatusCode
			return nil
		}
		resp, ok := chunk.Data().(gcs.EmbeddingsResponse)
		if !ok {
			return nil
		}
		if len(resp.Data) > 0 && len(resp.Data[0].Embedding) > 0 {
			gotVector = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !gotVector {
		return errors.New("empty embeddings response")
	}
	return nil
}

// probeDecisions sends a minimal valid TypeSafe Decisions request (tiny state +
// one typed noul question). Mirrors embeddings: success = non-empty answers.
func (c *Checker) probeDecisions(ctx context.Context, adapter aiengine.AIEngineStream, report *system.ModelHealthReport) error {
	req := &gcs.DecisionsRequest{
		State: json.RawMessage(`"health"`),
		Questions: map[string]json.RawMessage{
			"ok": json.RawMessage(`{"type":"noul","instructions":"Is the service healthy?"}`),
		},
	}

	var gotAnswers bool
	err := adapter.Decisions(ctx, req, func(ctx context.Context, chunk gcs.Chunk, aiEngineErr *gcs.AiEngineErrorResponse) error {
		if aiEngineErr != nil {
			report.HttpStatus = aiEngineErr.StatusCode
			return nil
		}
		resp, ok := chunk.Data().(gcs.DecisionsResponse)
		if !ok {
			return nil
		}
		if len(resp.Answers) > 0 {
			gotAnswers = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !gotAnswers {
		return errors.New("empty decisions response")
	}
	return nil
}

// modelMetaFor resolves the public on-chain name and type of a model,
// caching the result. Both come from a single registry read.
func (c *Checker) modelMetaFor(ctx context.Context, modelID common.Hash) (modelMeta, error) {
	c.mu.RLock()
	cached, ok := c.modelMeta[modelID.Hex()]
	c.mu.RUnlock()
	if ok {
		return cached, nil
	}

	name, tags, err := c.deps.Models.GetModelNameAndTags(ctx, modelID)
	if err != nil {
		return modelMeta{}, err
	}

	meta := modelMeta{
		name:      name,
		modelType: blockchainapi.DetectModelType(tags),
		isTee:     blockchainapi.IsTeeModel(tags),
		mediaOnly: hasUnprobeableMediaTag(tags),
	}

	c.mu.Lock()
	c.modelMeta[modelID.Hex()] = meta
	c.mu.Unlock()
	return meta, nil
}

func (c *Checker) getReport(modelID string) (system.ModelHealthReport, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.reports[modelID]
	return r, ok
}

func (c *Checker) setReport(report system.ModelHealthReport) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reports[report.ModelID] = report
}

// failureStreak tracks consecutive real-session prompt failures for a model.
type failureStreak struct {
	count int
	// rateLimitedOnly stays true while every failure in the current streak
	// was an upstream 429 — a pure throttling streak flips the model to
	// degraded (capacity risk) instead of unhealthy.
	rateLimitedOnly bool
}

// ReportPromptFailure records a failed real-session prompt for the model;
// upstreamStatus is the HTTP status returned by the backend (0 when the
// failure never got an HTTP response). After maxConsecutiveErrors failures in
// a row the model's health report is flipped immediately, without waiting for
// the next scheduled probe sweep: to degraded when the whole streak was
// upstream rate limiting (429), to unhealthy otherwise.
func (c *Checker) ReportPromptFailure(modelID common.Hash, upstreamStatus int) {
	id := modelID.Hex()

	c.mu.Lock()
	defer c.mu.Unlock()

	streak, ok := c.failures[id]
	if !ok {
		streak.rateLimitedOnly = true
	}
	streak.count++
	streak.rateLimitedOnly = streak.rateLimitedOnly && upstreamStatus == http.StatusTooManyRequests
	c.failures[id] = streak

	if streak.count < c.maxConsecutiveErrors {
		c.log.Warnf("model %s: session prompt failed (%d/%d consecutive)", lib.Short(modelID), streak.count, c.maxConsecutiveErrors)
		return
	}

	status := system.ModelHealthStatusUnhealthy
	if streak.rateLimitedOnly {
		status = system.ModelHealthStatusDegraded
	}

	report, ok := c.reports[id]
	if !ok {
		// A session prompt implies the model has (or had) an active bid even
		// if no sweep has reported on it yet.
		report = system.ModelHealthReport{ModelID: id, HasActiveBid: true}
	}
	if report.Status != status {
		c.log.Warnf("model %s: %d consecutive session prompt failures, marking %s", lib.Short(modelID), streak.count, status)
	}
	report.Status = status
	report.ErrorKind = system.ModelHealthErrorSessionErrors
	if streak.rateLimitedOnly {
		report.HttpStatus = http.StatusTooManyRequests
	}
	report.LastChecked = time.Now().Unix()
	c.reports[id] = report
}

// ReportPromptSuccess records a successful real-session prompt: the failure
// streak resets and, if the model was flipped to unhealthy or degraded by
// session errors, it is healed right away (a real prompt succeeding is
// stronger evidence than a probe).
func (c *Checker) ReportPromptSuccess(modelID common.Hash) {
	id := modelID.Hex()

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.failures, id)

	report, ok := c.reports[id]
	if !ok {
		return
	}
	sessionFlipped := report.Status == system.ModelHealthStatusUnhealthy || report.Status == system.ModelHealthStatusDegraded
	if sessionFlipped && report.ErrorKind == system.ModelHealthErrorSessionErrors {
		report.Status = system.ModelHealthStatusHealthy
		report.ErrorKind = ""
		report.HttpStatus = 0
		report.LastChecked = time.Now().Unix()
	}
	report.LastHealthy = time.Now().Unix()
	c.reports[id] = report
}

// ReportTeeFailure immediately marks the model as tee_unverified after a
// failed backend TEE self-verification, without waiting for the next
// scheduled probe sweep. The status heals on a later sweep once the backend
// attestation passes again.
func (c *Checker) ReportTeeFailure(modelID common.Hash) {
	id := modelID.Hex()

	c.mu.Lock()
	defer c.mu.Unlock()

	report, ok := c.reports[id]
	if !ok {
		report = system.ModelHealthReport{ModelID: id, HasActiveBid: true}
	}
	if report.Status != system.ModelHealthStatusTeeUnverified {
		c.log.Warnf("model %s: backend TEE self-verification failed, marking tee_unverified", lib.Short(modelID))
	}
	report.Status = system.ModelHealthStatusTeeUnverified
	report.ErrorKind = system.ModelHealthErrorTeeAttestation
	report.LastChecked = time.Now().Unix()
	c.reports[id] = report
}

// resetFailures clears the consecutive session-failure counter for a model.
func (c *Checker) resetFailures(modelID common.Hash) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.failures, modelID.Hex())
}

// generateArithmeticPrompt returns two random small numbers and a prompt
// asking for their sum, so cached/replayed answers cannot fake correctness.
func generateArithmeticPrompt() (int, int, string) {
	a := rand.Intn(50) + 1
	b := rand.Intn(50) + 1
	return a, b, fmt.Sprintf("What is %d+%d? Reply with only the number.", a, b)
}

func verifyArithmeticAnswer(response string, expected int) bool {
	re := regexp.MustCompile(fmt.Sprintf(`(^|[^\d])%d($|[^\d])`, expected))
	return re.MatchString(response)
}

// classifyError maps a probe error to a coarse category so error details
// (which may contain the private backend URL) never leak into the report.
func classifyError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return system.ModelHealthErrorTimeout
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return system.ModelHealthErrorTimeout
		}
		return system.ModelHealthErrorConnection
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return system.ModelHealthErrorConnection
	}
	msg := err.Error()
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "no such host") || strings.Contains(msg, "failed to send request") {
		return system.ModelHealthErrorConnection
	}
	if strings.Contains(msg, "context deadline exceeded") {
		return system.ModelHealthErrorTimeout
	}
	return system.ModelHealthErrorBadResponse
}
