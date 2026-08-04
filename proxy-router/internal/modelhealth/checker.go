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
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
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

// DefaultMaxConsecutiveErrors is the number of consecutive real-session
// prompt failures after which a model is flipped to unhealthy without
// waiting for the next scheduled probe sweep.
const DefaultMaxConsecutiveErrors = 3

// modelMeta caches the public on-chain facts about a model so each sweep
// only pays one registry call per previously unseen model.
type modelMeta struct {
	name      string
	modelType structs.ModelType
	isTee     bool
}

type Checker struct {
	deps       Deps
	interval   time.Duration
	timeout    time.Duration
	probeDelay time.Duration
	// maxConsecutiveErrors is the threshold of consecutive session prompt
	// failures that dynamically flips a model to unhealthy.
	maxConsecutiveErrors int
	log                  lib.ILogger

	mu        sync.RWMutex
	reports   map[string]system.ModelHealthReport
	modelMeta map[string]modelMeta
	// failures counts consecutive real-session prompt failures per model.
	failures map[string]int

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
		failures:             make(map[string]int),
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

	probed := false
	for i, modelID := range modelIDs {
		select {
		case <-ctx.Done():
			return
		default:
		}
		bidID, hasBid := bidsByModel[modelID]
		// Pace backend probes so a provider with many models doesn't burst
		// its upstream (often a single account behind all models) and
		// self-inflict rate-limit failures. Only probed models (those with
		// an active bid) hit the backend, so only they are paced.
		if hasBid && probed && !c.sleep(ctx, c.probeDelay) {
			return
		}
		c.checkModel(ctx, modelID, bidID, hasBid, modelCfgs[i].ApiURL, reattestTee)
		if hasBid {
			probed = true
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

// sleep blocks for d unless the context is cancelled first; it reports
// whether the full delay elapsed.
func (c *Checker) sleep(ctx context.Context, d time.Duration) bool {
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

func (c *Checker) checkModel(ctx context.Context, modelID common.Hash, bidID common.Hash, hasBid bool, apiURL string, reattestTee bool) {
	report := system.ModelHealthReport{
		ModelID:      modelID.Hex(),
		HasActiveBid: hasBid,
		LastChecked:  time.Now().Unix(),
	}

	if hasBid {
		report.BidID = bidID.Hex()
	}

	if prev, ok := c.getReport(modelID.Hex()); ok {
		report.LastHealthy = prev.LastHealthy
	}

	if !hasBid {
		report.Status = system.ModelHealthStatusNoBid
		c.setReport(report)
		return
	}

	meta, err := c.modelMetaFor(ctx, modelID)
	if err != nil {
		c.log.Warnf("model %s: cannot resolve model metadata: %s", lib.Short(modelID), err)
		report.Status = system.ModelHealthStatusSkipped
		report.ErrorKind = system.ModelHealthErrorTypeLookup
		c.setReport(report)
		return
	}
	report.ModelName = meta.name
	report.ModelType = string(meta.modelType)

	// Re-run backend attestation for TEE-tagged models on every sweep after
	// the first (the first sweep relies on the startup attestation),
	// regardless of the current snapshot status: a transient failure at
	// startup self-heals here instead of sticking until a restart, and a
	// stale pass is re-proven against the live backend. A model whose
	// attestation does not pass cannot serve sessions (session requests are
	// rejected by the receiver), so it is reported as tee_unverified and the
	// backend probe is skipped.
	if meta.isTee && c.deps.TeeStatus != nil {
		if reattestTee {
			attestCtx, cancel := context.WithTimeout(ctx, c.timeout)
			if err := c.deps.TeeStatus.ReattestBackend(attestCtx, modelID.Hex(), apiURL); err != nil {
				c.log.Warnf("model %s: backend TEE re-attestation failed: %s", lib.Short(modelID), err)
			}
			cancel()
		}

		snapshot := c.deps.TeeStatus.GetStatus(modelID.Hex())
		if snapshot == nil || snapshot.Status != attestation.StatusPassed {
			c.log.Warnf("model %s: backend TEE attestation is not passed, reporting tee_unverified", lib.Short(modelID))
			report.Status = system.ModelHealthStatusTeeUnverified
			report.ErrorKind = system.ModelHealthErrorTeeAttestation
			c.setReport(report)
			return
		}
	}

	var probe func(ctx context.Context, adapter aiengine.AIEngineStream, report *system.ModelHealthReport) error
	switch meta.modelType {
	case structs.ModelTypeLLM:
		probe = c.probeLLM
	case structs.ModelTypeEMBEDDING:
		probe = c.probeEmbeddings
	default:
		report.Status = system.ModelHealthStatusSkipped
		c.setReport(report)
		return
	}

	adapter, err := c.deps.Adapters.GetAdapter(ctx, common.Hash{}, modelID, common.Hash{}, false, false)
	if err != nil {
		c.log.Warnf("model %s: cannot get adapter: %s", lib.Short(modelID), err)
		report.Status = system.ModelHealthStatusUnhealthy
		report.ErrorKind = system.ModelHealthErrorBadResponse
		c.setReport(report)
		return
	}

	probeCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	start := time.Now()
	err = probe(probeCtx, adapter, &report)
	report.LatencyMs = time.Since(start).Milliseconds()

	if err != nil {
		c.log.Warnf("model %s: probe failed: %s", lib.Short(modelID), err)
		report.Status = system.ModelHealthStatusUnhealthy
		// 429 means the backend is up and correctly configured, just
		// throttled right now — a materially different signal than a
		// billing (402) or configuration (404) failure.
		if report.HttpStatus == http.StatusTooManyRequests {
			report.ErrorKind = system.ModelHealthErrorRateLimited
		} else {
			report.ErrorKind = classifyError(err)
		}
	} else {
		report.Status = system.ModelHealthStatusHealthy
		report.LastHealthy = time.Now().Unix()
		// A passing probe is direct evidence the backend works again, so the
		// consecutive session-failure streak (if any) is over.
		c.resetFailures(modelID)
	}

	c.setReport(report)
}

func (c *Checker) probeLLM(ctx context.Context, adapter aiengine.AIEngineStream, report *system.ModelHealthReport) error {
	a, b, prompt := generateArithmeticPrompt()

	req := &gcs.OpenAICompletionRequestExtra{
		ChatCompletionRequest: openai.ChatCompletionRequest{
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Stream: false,
		},
	}

	var response strings.Builder
	var engineErr error
	err := adapter.Prompt(ctx, req, func(ctx context.Context, chunk gcs.Chunk, aiEngineErr *gcs.AiEngineErrorResponse) error {
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
		return err
	}
	if engineErr != nil {
		return engineErr
	}
	if response.Len() == 0 {
		return errors.New("empty completion response")
	}

	correct := verifyArithmeticAnswer(response.String(), a+b)
	report.PromptCorrect = &correct
	return nil
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

	meta := modelMeta{name: name, modelType: blockchainapi.DetectModelType(tags), isTee: blockchainapi.IsTeeModel(tags)}

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

// ReportPromptFailure records a failed real-session prompt for the model.
// After maxConsecutiveErrors failures in a row the model's health report is
// flipped to unhealthy immediately, without waiting for the next scheduled
// probe sweep, so consumers pinging before session open skip this provider.
func (c *Checker) ReportPromptFailure(modelID common.Hash) {
	id := modelID.Hex()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.failures[id]++
	count := c.failures[id]
	if count < c.maxConsecutiveErrors {
		c.log.Warnf("model %s: session prompt failed (%d/%d consecutive)", lib.Short(modelID), count, c.maxConsecutiveErrors)
		return
	}

	report, ok := c.reports[id]
	if !ok {
		// A session prompt implies the model has (or had) an active bid even
		// if no sweep has reported on it yet.
		report = system.ModelHealthReport{ModelID: id, HasActiveBid: true}
	}
	if report.Status != system.ModelHealthStatusUnhealthy {
		c.log.Warnf("model %s: %d consecutive session prompt failures, marking unhealthy", lib.Short(modelID), count)
	}
	report.Status = system.ModelHealthStatusUnhealthy
	report.ErrorKind = system.ModelHealthErrorSessionErrors
	report.LastChecked = time.Now().Unix()
	c.reports[id] = report
}

// ReportPromptSuccess records a successful real-session prompt: the failure
// streak resets and, if the model was flipped to unhealthy by session errors,
// it is healed right away (a real prompt succeeding is stronger evidence than
// a probe).
func (c *Checker) ReportPromptSuccess(modelID common.Hash) {
	id := modelID.Hex()

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.failures, id)

	report, ok := c.reports[id]
	if !ok {
		return
	}
	if report.Status == system.ModelHealthStatusUnhealthy && report.ErrorKind == system.ModelHealthErrorSessionErrors {
		report.Status = system.ModelHealthStatusHealthy
		report.ErrorKind = ""
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
