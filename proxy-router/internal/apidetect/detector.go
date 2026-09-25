// Package apidetect identifies, without sending inference requests, the API
// serving a configured model backend and hands the evidence to apispec.
package apidetect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec/registry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

const (
	cacheTTL      = 30 * time.Minute
	probeTimeout  = 2 * time.Second
	detectTimeout = 10 * time.Second
	cacheSweepAt  = 256
)

// Variables so tests can map an httptest host to a vendor and keep a
// vendor-probe fixture free of the offline registry's facts.
var (
	stackFromHost = apispec.StackFromHost
	registryMatch = registry.Match
)

type Options struct {
	// Diagnostics only. Not applied to the LiteLLM second hop, where
	// recognising api_base's host is the point.
	IgnoreHostVendors bool
	ProbeTimeout      time.Duration
	TwoHop            bool
	// CacheTTL bounds how long a detection result and a registry listing are
	// reused (MODEL_API_DETECT_TTL); <= 0 keeps cacheTTL.
	CacheTTL time.Duration
}

func DefaultOptions() Options { return Options{TwoHop: true} }

type cacheEntry struct {
	api *system.ModelApiSpec
	at  time.Time
}

type Detector struct {
	log       lib.ILogger
	client    *http.Client
	hopClient *http.Client
	opts      Options
	ttl       time.Duration

	mu       sync.Mutex
	cache    map[string]cacheEntry
	listings map[string]listingEntry
}

func NewDetector(log lib.ILogger, opts Options) *Detector {
	timeout := opts.ProbeTimeout
	if timeout <= 0 {
		timeout = probeTimeout
	}
	ttl := opts.CacheTTL
	if ttl <= 0 {
		ttl = cacheTTL
	}
	return &Detector{
		log:       log.Named("API_DETECT"),
		client:    &http.Client{Timeout: timeout, CheckRedirect: refuseCrossHostRedirect},
		hopClient: newHopClient(timeout),
		opts:      opts,
		ttl:       ttl,
		cache:     make(map[string]cacheEntry),
		listings:  make(map[string]listingEntry),
	}
}

func (d *Detector) clientFor(ctx context.Context) *http.Client {
	if ctx.Value(hopCtxKey{}) != nil {
		return d.hopClient
	}
	return d.client
}

// Probes may carry the configured bearer key: a redirect to another
// host:port, or from https down to http, is not followed.
func refuseCrossHostRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	prev := via[len(via)-1].URL
	if !strings.EqualFold(req.URL.Host, prev.Host) || (prev.Scheme == "https" && req.URL.Scheme != "https") {
		return http.ErrUseLastResponse
	}
	return nil
}

func (d *Detector) Detect(ctx context.Context, cfg config.ModelConfig) *system.ModelApiSpec {
	api, _ := d.DetectProbed(ctx, cfg)
	return api
}

// WouldProbe reports whether Detect for cfg would reach the backend rather
// than answer from the cache or the static tables; the checker paces such
// calls.
func (d *Detector) WouldProbe(cfg config.ModelConfig) bool {
	if apispec.StackFor(cfg.ApiStack) != "" || (cfg.ApiURL == "" && cfg.ModelName == "") {
		return false
	}
	d.mu.Lock()
	entry, ok := d.cache[cacheKey(cfg)]
	d.mu.Unlock()
	return !ok || time.Since(entry.at) >= d.ttl
}

// DetectProbed is Detect that also reports whether the backend was reached.
func (d *Detector) DetectProbed(ctx context.Context, cfg config.ModelConfig) (*system.ModelApiSpec, bool) {
	if apispec.StackFor(cfg.ApiStack) != "" {
		return apispec.Build(cfg), false
	}
	if cfg.ApiURL == "" && cfg.ModelName == "" {
		return nil, false
	}

	key := cacheKey(cfg)
	d.mu.Lock()
	entry, ok := d.cache[key]
	d.mu.Unlock()
	if ok && time.Since(entry.at) < d.ttl {
		return entry.api.Clone(), false
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, detectTimeout)
		defer cancel()
	}

	api := d.detect(ctx, cfg)
	if ctx.Err() != nil {
		d.log.Debugf("backend %s: detection cut short (%v); result not cached", RedactURL(cfg.ApiURL), ctx.Err())
		return api, true
	}
	if api != nil {
		d.log.Debugf("backend %s: detected api stack=%q family=%q", RedactURL(cfg.ApiURL), api.Stack, api.ModelFamily)
	}

	d.mu.Lock()
	if len(d.cache) >= cacheSweepAt {
		sweepCache(d.cache, cacheSweepAt, d.ttl, func(e cacheEntry) time.Time { return e.at })
	}
	d.cache[key] = cacheEntry{api: api, at: time.Now()}
	d.mu.Unlock()
	return api.Clone(), true
}

// sweepCache drops expired entries and, when the map is still at limit,
// the oldest half so a burst of fresh keys cannot grow it without bound.
func sweepCache[E any](m map[string]E, limit int, ttl time.Duration, at func(E) time.Time) {
	for k, e := range m {
		if time.Since(at(e)) >= ttl {
			delete(m, k)
		}
	}
	if len(m) < limit {
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return at(m[keys[i]]).Before(at(m[keys[j]])) })
	for _, k := range keys[:len(keys)-limit/2] {
		delete(m, k)
	}
}

func (d *Detector) DetectWithTrace(ctx context.Context, cfg config.ModelConfig) (*system.ModelApiSpec, []string) {
	if stack := apispec.StackFor(cfg.ApiStack); stack != "" {
		return apispec.Build(cfg), []string{fmt.Sprintf("apiStack %q declared in models-config: static spec, no probing", stack)}
	}
	if cfg.ApiURL == "" && cfg.ModelName == "" {
		return nil, []string{"nothing to detect: config has neither apiUrl nor modelName"}
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, detectTimeout)
		defer cancel()
	}

	ctx, trace := withTrace(ctx)
	api := d.detect(ctx, cfg)
	return api, trace.lines
}

// Runs for a model whose host resolved to the hosted vendor `ev.Stack`; the
// probe owns its own trace lines. Keyed by the stack name StackFromHost
// returns; vendors without an entry are not probed.
type hostedProbe func(ctx context.Context, d *Detector, bases []string, cfg config.ModelConfig, ev *apispec.Evidence)

var hostedProbes = map[string]hostedProbe{}

// Called from init() in the probe file that owns the vendor.
func registerHostedProbe(stack string, probe hostedProbe) {
	if _, dup := hostedProbes[stack]; dup {
		panic("apidetect: hosted probe for " + stack + " registered twice")
	}
	hostedProbes[stack] = probe
}

func (d *Detector) detect(ctx context.Context, cfg config.ModelConfig) *system.ModelApiSpec {
	ev := apispec.Evidence{ModelName: cfg.ModelName, ModelFamily: cfg.ModelFamily, DeclaredCaps: cfg.Capabilities}

	if d.opts.IgnoreHostVendors {
		tracef(ctx, "hostname recognition disabled; identifying purely by endpoints and listing shape")
	} else {
		ev.Stack = stackFromHost(cfg.ApiURL)
	}
	// Captured before probing: a stack a probe fingerprints later is a
	// self-hosted engine, which the registry does not describe.
	hostRecognised := ev.Stack != ""
	bases := baseCandidates(cfg.ApiURL)

	if ev.Stack == "" {
		apiURLText := "(no apiUrl)"
		if cfg.ApiURL != "" {
			apiURLText = RedactURL(cfg.ApiURL)
		}
		tracef(ctx, "host of %s is not a known hosted vendor; fingerprinting engines at %v", apiURLText, bases)
		d.fingerprintEngines(ctx, bases, cfg.ModelName, cfg.ApiKey, &ev, true)
		if ev.Stack == "" {
			tracef(ctx, "no engine fingerprint matched; checking for a registry-shaped model listing")
			d.probeRegistryShape(ctx, bases, cfg.ModelName, cfg.ApiKey, &ev)
		}
	} else if probe, ok := hostedProbes[ev.Stack]; ok {
		probe(ctx, d, bases, cfg, &ev)
	} else {
		tracef(ctx, "host recognized as hosted vendor %q by hostname; no probing needed", ev.Stack)
	}

	if ev.Registry == nil && (hostRecognised || ev.Stack == "openai") {
		lookupRegistry(ctx, cfg.ModelName, &ev)
	}

	api, lines := apispec.ComposeWithTrace(ev)
	for _, line := range lines {
		tracef(ctx, "%s", line)
	}
	if api == nil {
		return nil
	}

	// A stack whose transport is not cfg.ApiType cannot be what this model
	// speaks to, so it, its vocabulary and the capabilities read from its
	// endpoints go. Not applied behind a gateway (Via set): the wire is then
	// LiteLLM's OpenAI-compatible surface whatever the upstream's own
	// transport. Family, always_on and declared / registry capabilities
	// describe the model, not the wire, and stay.
	if transport, ok := config.StackTransport[api.Stack]; ok && transport != cfg.ApiType && api.Via == "" {
		tracef(ctx, "detected stack %q speaks %q but apiType is %q — stack dropped", api.Stack, transport, cfg.ApiType)
		api.Stack = ""
		api.Via = ""
		api.Bindings = nil
		api.Parameters = nil
		// Binding evidence goes with the bindings whatever its source: a
		// web_search.* entry carries the webSearch capability's source, and a
		// declared one would otherwise survive the clear below and name a
		// binding that no longer exists.
		for key := range api.Evidence {
			if strings.HasPrefix(key, "bindings.") {
				delete(api.Evidence, key)
			}
		}
		apispec.ClearDetectedCapabilities(api, nil)
		if api.Thinking != nil && api.Thinking.Mode != system.ThinkingModeAlwaysOn {
			api.Thinking = nil
		}
		if api.ModelFamily == "" && api.Thinking == nil && api.Capabilities == nil {
			tracef(ctx, "nothing known about this backend: no api block")
			return nil
		}
	}
	return api
}

// The offline registry (contract §3 rule 10, §6) describes hosted vendors'
// catalogues, so it is consulted for a hostname-recognised vendor and the
// openai stack only; a self-hosted engine serves whatever window and
// quantisation it was started with, and the snapshot would be a guess. The
// composer lets the entry fill only fields nothing else set. The trace names
// which candidate matched, never the id: a configured name may be anything.
func lookupRegistry(ctx context.Context, modelName string, ev *apispec.Evidence) {
	kinds := []string{"served model id", "bare model name", "configured model name"}
	facts, pos, ok := registryMatch(ev.ServedModelID, bareModelID(modelName), modelName)
	if !ok {
		tracef(ctx, "registry: %s@%s has no entry for the served id, bare name or configured name", registryName, registry.Date())
		return
	}
	ev.Registry = facts
	tracef(ctx, "registry: %s@%s matched %q", facts.Name, facts.Date, kinds[pos])
}

// The snapshot's sources, for the trace line when nothing matched.
const registryName = "models.dev+litellm"

// The key hash is part of the cache key so a rotated key re-inspects a
// backend whose cached result may have failed auth; the declared
// capabilities are part of it because they are composed into the result.
func cacheKey(cfg config.ModelConfig) string {
	sum := sha256.Sum256([]byte(cfg.ApiKey))
	return strings.Join([]string{cfg.ApiURL, cfg.ModelName, cfg.ApiStack, cfg.ModelFamily, hex.EncodeToString(sum[:8]), declaredCapsHash(cfg.Capabilities)}, "|")
}

// Empty for no block; otherwise a digest of its canonical JSON (struct field
// order, so two equal blocks hash alike).
func declaredCapsHash(caps *config.DeclaredCapabilities) string {
	if caps == nil {
		return ""
	}
	raw, err := json.Marshal(caps)
	if err != nil {
		return "?"
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:8])
}

const redactedURL = "<unparseable url>"

// Input that is not a URL with a host may be a pasted secret: never echoed.
func RedactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return redactedURL
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
