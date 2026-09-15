// Package apidetect identifies, without sending inference requests, the API
// serving a configured model backend and hands the evidence to apispec.
package apidetect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
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

// Variable so tests can map an httptest host to a vendor.
var stackFromHost = apispec.StackFromHost

type Options struct {
	// Diagnostics only. Not applied to the LiteLLM second hop, where
	// recognising api_base's host is the point.
	IgnoreHostVendors bool
	ProbeTimeout      time.Duration
	TwoHop            bool
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

	mu    sync.Mutex
	cache map[string]cacheEntry
}

func NewDetector(log lib.ILogger, opts Options) *Detector {
	timeout := opts.ProbeTimeout
	if timeout <= 0 {
		timeout = probeTimeout
	}
	return &Detector{
		log:       log.Named("API_DETECT"),
		client:    &http.Client{Timeout: timeout, CheckRedirect: refuseCrossHostRedirect},
		hopClient: newHopClient(timeout),
		opts:      opts,
		ttl:       cacheTTL,
		cache:     make(map[string]cacheEntry),
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
	if apispec.StackFor(cfg.ApiStack) != "" {
		return apispec.Build(cfg)
	}
	if cfg.ApiURL == "" && cfg.ModelName == "" {
		return nil
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
		d.log.Debugf("backend %s: detection cut short (%v); result not cached", RedactURL(cfg.ApiURL), ctx.Err())
		return api
	}
	if api != nil {
		d.log.Debugf("backend %s: detected api stack=%q family=%q", RedactURL(cfg.ApiURL), api.Stack, api.ModelFamily)
	}

	d.mu.Lock()
	if len(d.cache) >= cacheSweepAt {
		for k, e := range d.cache {
			if time.Since(e.at) >= d.ttl {
				delete(d.cache, k)
			}
		}
		if len(d.cache) >= cacheSweepAt {
			keys := make([]string, 0, len(d.cache))
			for k := range d.cache {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool { return d.cache[keys[i]].at.Before(d.cache[keys[j]].at) })
			for _, k := range keys[:len(keys)-cacheSweepAt/2] {
				delete(d.cache, k)
			}
		}
	}
	d.cache[key] = cacheEntry{api: api, at: time.Now()}
	d.mu.Unlock()
	return api.Clone()
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
	if api == nil {
		return nil
	}

	// A stack whose transport is not cfg.ApiType cannot be what this model
	// speaks to, so it and its vocabulary go. Not applied behind a gateway
	// (Via set): the wire is then LiteLLM's OpenAI-compatible surface whatever
	// the upstream's own transport. Family and always_on describe the model,
	// not the wire, and stay.
	if transport, ok := config.StackTransport[api.Stack]; ok && transport != cfg.ApiType && api.Via == "" {
		tracef(ctx, "detected stack %q speaks %q but apiType is %q — stack dropped", api.Stack, transport, cfg.ApiType)
		api.Stack = ""
		api.Via = ""
		api.Bindings = nil
		api.Parameters = nil
		if api.Thinking != nil && api.Thinking.Mode != system.ThinkingModeAlwaysOn {
			api.Thinking = nil
		}
		if api.ModelFamily == "" && api.Thinking == nil {
			tracef(ctx, "nothing known about this backend: no api block")
			return nil
		}
	}
	return api
}

// The key hash is part of the cache key so a rotated key re-inspects a
// backend whose cached result may have failed auth.
func cacheKey(cfg config.ModelConfig) string {
	sum := sha256.Sum256([]byte(cfg.ApiKey))
	return strings.Join([]string{cfg.ApiURL, cfg.ModelName, cfg.ApiStack, cfg.ModelFamily, hex.EncodeToString(sum[:8])}, "|")
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
