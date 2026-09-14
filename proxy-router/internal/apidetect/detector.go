// Package apidetect identifies, at runtime, what API actually serves a
// configured model backend: the serving stack or hosted vendor (vLLM, SGLang,
// llama.cpp, Ollama, TGI, LM Studio, KoboldCpp, LiteLLM, OpenRouter, Venice,
// OpenAI, Anthropic, ...), the model family and — where derivable — the knob
// that controls thinking. Detection is non-inference introspection: hostname
// recognition, distinctive service endpoints (/props, /api/tags,
// /get_model_info, /info, /version, /health/liveliness, ...), chat-template
// source parsing and registry model listings. No inference request is ever
// sent. The package only gathers apispec.Evidence; apispec.Compose turns it
// into the spec, so the stack/family/binding knowledge lives in one place.
//
// Results feed system.ModelHealthReport.Api (source: detected) and therefore
// surface on the public /healthcheck endpoint and in the morrpc pong models
// list. Everything reported is self-observed and unverified.
package apidetect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

const (
	// cacheTTL bounds how long a detection result is reused before the
	// backend is re-inspected, so config/backend changes surface without a
	// restart while sweeps stay cheap.
	cacheTTL = 30 * time.Minute
	// probeTimeout caps each individual probe request.
	probeTimeout = 2 * time.Second
	// detectTimeout caps one whole detection pass when the caller's context
	// has no earlier deadline.
	detectTimeout = 10 * time.Second
	// cacheSweepAt is the cache size at which expired entries are evicted on
	// insert, bounding growth from key rotation or config churn.
	cacheSweepAt = 256
)

// stackFromHost is apispec.StackFromHost behind a variable so tests can map
// an httptest host to a hosted vendor.
var stackFromHost = apispec.StackFromHost

// Options tunes a Detector. Production wiring uses DefaultOptions.
type Options struct {
	// IgnoreHostVendors disables hostname recognition so hosted vendors are
	// identified purely by their endpoints and model-listing shape, as an
	// unknown custom domain would be. Diagnostics only (cmd/apidetect
	// -ignore-host); leave false in the sweep path.
	IgnoreHostVendors bool
	// ProbeTimeout overrides the 2 s per-probe timeout (tests only; 0 keeps
	// the default).
	ProbeTimeout time.Duration
}

// DefaultOptions is the production configuration.
func DefaultOptions() Options { return Options{} }

type cacheEntry struct {
	api *system.ModelApiSpec
	at  time.Time
}

// Detector performs cached runtime API detection for model backends.
type Detector struct {
	log    lib.ILogger
	client *http.Client
	opts   Options
	ttl    time.Duration

	mu    sync.Mutex
	cache map[string]cacheEntry
}

func NewDetector(log lib.ILogger, opts Options) *Detector {
	timeout := opts.ProbeTimeout
	if timeout <= 0 {
		timeout = probeTimeout
	}
	return &Detector{
		log:    log.Named("API_DETECT"),
		client: &http.Client{Timeout: timeout},
		opts:   opts,
		ttl:    cacheTTL,
		cache:  make(map[string]cacheEntry),
	}
}

// Detect returns the API spec of cfg's backend, or nil when nothing could be
// determined. A declared apiStack short-circuits to the static apispec.Build
// (explicit declaration wins; the backend is not probed). Detection results
// are cached per (apiUrl, modelName, apiStack, modelFamily, apiKey) for the
// cache TTL; a pass cut short by the caller's deadline is returned but not
// cached. The returned spec is the caller's own copy. Safe for concurrent use.
func (d *Detector) Detect(ctx context.Context, cfg config.ModelConfig) *system.ModelApiSpec {
	if cfg.ApiURL == "" && cfg.ModelName == "" {
		return nil
	}
	if apispec.StackFor(cfg.ApiStack) != "" {
		return apispec.Build(cfg)
	}

	key := cacheKey(cfg)
	d.mu.Lock()
	entry, ok := d.cache[key]
	d.mu.Unlock()
	if ok && time.Since(entry.at) < d.ttl {
		return entry.api.Clone()
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, detectTimeout)
		defer cancel()
	}

	api := d.detect(ctx, cfg)
	if ctx.Err() != nil {
		d.log.Debugf("model %q: detection cut short (%v); result not cached", cfg.ModelName, ctx.Err())
		return api
	}
	if api != nil {
		d.log.Debugf("model %q: detected api stack=%q family=%q", cfg.ModelName, api.Stack, api.ModelFamily)
	}

	d.mu.Lock()
	if len(d.cache) >= cacheSweepAt {
		for k, e := range d.cache {
			if time.Since(e.at) >= d.ttl {
				delete(d.cache, k)
			}
		}
	}
	d.cache[key] = cacheEntry{api: api, at: time.Now()}
	d.mu.Unlock()
	return api.Clone()
}

// DetectWithTrace runs one fresh detection pass (bypassing and not touching
// the cache) and returns, alongside the result, the human-readable steps that
// produced it: every probe attempted, what it returned, and which evidence
// source decided the family and thinking knob. Diagnostics tooling only.
func (d *Detector) DetectWithTrace(ctx context.Context, cfg config.ModelConfig) (*system.ModelApiSpec, []string) {
	if cfg.ApiURL == "" && cfg.ModelName == "" {
		return nil, []string{"nothing to detect: config has neither apiUrl nor modelName"}
	}
	if stack := apispec.StackFor(cfg.ApiStack); stack != "" {
		return apispec.Build(cfg), []string{fmt.Sprintf("apiStack %q declared in models-config: static spec, no probing", stack)}
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

// detect gathers evidence over HTTP and hands it to the composer.
func (d *Detector) detect(ctx context.Context, cfg config.ModelConfig) *system.ModelApiSpec {
	ev := apispec.Evidence{ModelName: cfg.ModelName, ModelFamily: cfg.ModelFamily}

	if d.opts.IgnoreHostVendors {
		tracef(ctx, "hostname recognition disabled; identifying purely by endpoints and listing shape")
	} else {
		ev.Stack = stackFromHost(cfg.ApiURL)
	}
	bases := baseCandidates(cfg.ApiURL)

	switch ev.Stack {
	case "":
		tracef(ctx, "host of %s is not a known hosted vendor; fingerprinting engines at %v", RedactURL(cfg.ApiURL), bases)
		// Self-hosted or unrecognized host: fingerprint the engine, and when
		// that yields nothing, check for a registry-shaped model listing
		// (OpenRouter-compatible proxies on custom domains).
		d.fingerprintEngines(ctx, bases, cfg.ModelName, cfg.ApiKey, &ev)
		if ev.Stack == "" {
			tracef(ctx, "no engine fingerprint matched; checking for a registry-shaped model listing")
			d.probeRegistryShape(ctx, bases, cfg.ModelName, cfg.ApiKey, &ev)
		}
	case "openrouter", "venice":
		tracef(ctx, "host recognized as %s by hostname; importing its model listing", ev.Stack)
		d.probeRegistryShape(ctx, bases, cfg.ModelName, cfg.ApiKey, &ev)
	default:
		tracef(ctx, "host recognized as hosted vendor %q by hostname; no probing needed", ev.Stack)
	}

	api, lines := apispec.ComposeWithTrace(ev)
	for _, line := range lines {
		tracef(ctx, "%s", line)
	}
	return api
}

// cacheKey identifies a backend for result caching. The API key is folded
// in as a hash so a rotated key re-inspects the backend instead of reusing a
// result that may have failed auth; apiStack and modelFamily are part of the
// key because they change what is composed.
func cacheKey(cfg config.ModelConfig) string {
	sum := sha256.Sum256([]byte(cfg.ApiKey))
	return strings.Join([]string{cfg.ApiURL, cfg.ModelName, cfg.ApiStack, cfg.ModelFamily, hex.EncodeToString(sum[:8])}, "|")
}

// RedactURL strips userinfo and query from a URL for logs and traces, so
// credentials embedded in a configured endpoint never surface.
func RedactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
