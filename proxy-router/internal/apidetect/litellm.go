package apidetect

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// Bounds the second hop on its own so a slow upstream never costs the first
// hop's evidence. Variable so tests can exercise the timeout.
var hopTimeout = 4 * time.Second

// Provider whose upstream kind is decided by the api_base host: how Venice,
// which has no native LiteLLM provider, is reached.
const upstreamByHost = "host"

// Only prefixes documented at docs.litellm.ai/docs/providers/* are listed.
// "" marks a hosted vendor or cloud: the bare model id is taken, nothing is
// probed.
var litellmProviderStacks = map[string]string{
	"openai":       upstreamByHost, // https://docs.litellm.ai/docs/providers/openai_compatible
	"hosted_vllm":  "vllm",         // https://docs.litellm.ai/docs/providers/vllm
	"vllm":         "vllm",         // https://docs.litellm.ai/docs/providers/vllm (deprecated in-process form)
	"ollama":       "ollama",       // https://docs.litellm.ai/docs/providers/ollama
	"ollama_chat":  "ollama",       // https://docs.litellm.ai/docs/providers/ollama
	"openrouter":   "openrouter",   // https://docs.litellm.ai/docs/providers/openrouter
	"anthropic":    "anthropic",    // https://docs.litellm.ai/docs/providers/anthropic
	"bedrock":      "",             // https://docs.litellm.ai/docs/providers/bedrock (also bedrock/converse/, bedrock/invoke/)
	"vertex_ai":    "",             // https://docs.litellm.ai/docs/providers/vertex
	"gemini":       "",             // https://docs.litellm.ai/docs/providers/gemini
	"azure":        "",             // https://docs.litellm.ai/docs/providers/azure (also azure/o_series/, azure/gpt5_series/)
	"groq":         "",             // https://docs.litellm.ai/docs/providers/groq
	"together_ai":  "",             // https://docs.litellm.ai/docs/providers/togetherai
	"fireworks_ai": "",             // https://docs.litellm.ai/docs/providers/fireworks_ai
	"deepinfra":    "",             // https://docs.litellm.ai/docs/providers/deepinfra
	"xai":          "",             // https://docs.litellm.ai/docs/providers/xai
	"mistral":      "",             // https://docs.litellm.ai/docs/providers/mistral
	"deepseek":     "",             // https://docs.litellm.ai/docs/providers/deepseek
}

func (d *Detector) litellmUpstream(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) {
	info := d.requestJSON(ctx, http.MethodGet, base+"/model/info", apiKey, nil, maxProbeBody)
	if info == nil {
		return
	}
	data, _ := info["data"].([]any)
	var dep map[string]any
	for _, item := range data {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if name, _ := m["model_name"].(string); strings.EqualFold(name, modelName) {
			dep = m
			break
		}
	}
	if dep == nil {
		tracef(ctx, "litellm /model/info lists %d deployments but none is named %q; no upstream evidence", len(data), modelName)
		return
	}
	params, _ := dep["litellm_params"].(map[string]any)
	model, _ := params["model"].(string)
	provider, _ := params["custom_llm_provider"].(string)
	apiBase, _ := params["api_base"].(string)

	prefix, rest := "", model
	if i := strings.Index(model, "/"); i >= 0 {
		prefix, rest = model[:i], model[i+1:]
	}
	if provider == "" {
		provider = prefix
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	kind, known := litellmProviderStacks[provider]
	upstreamID := model
	if known && prefix != "" && strings.EqualFold(prefix, provider) {
		upstreamID = rest
	}
	// Nothing else from the deployment may reach the trace (keys, raw fields).
	apiBaseText := "(none)"
	if apiBase != "" {
		apiBaseText = RedactURL(apiBase)
	}
	tracef(ctx, "litellm /model/info: deployment %q -> provider %q, upstream model %q, api_base %s", modelName, provider, upstreamID, apiBaseText)

	upstreamID = bareModelID(upstreamID)

	if upstreamID != "" {
		ev.ServedModelID = upstreamID
	}
	switch {
	case !known:
		tracef(ctx, "litellm provider %q is not in the documented prefix table; upstream not probed", provider)
		return
	case kind == "":
		tracef(ctx, "litellm provider %q is a hosted vendor or cloud; upstream not probed (the spec's stack stays litellm)", provider)
		return
	case apiBase == "" && kind != "anthropic":
		tracef(ctx, "litellm deployment %q (provider %q) has no api_base; upstream not probed", modelName, provider)
		return
	}
	if kind == upstreamByHost {
		kind = stackFromHost(apiBase)
	}

	if apiBase != "" && !hopAllowed(apiBase) {
		tracef(ctx, "litellm api_base %s refused for the second hop (scheme/address)", RedactURL(apiBase))
		return
	}
	hopCtx, cancel := context.WithTimeout(context.WithValue(ctx, hopCtxKey{}, true), hopTimeout)
	defer cancel()
	bases := baseCandidates(apiBase)
	var up apispec.Evidence
	identified := ""
	switch kind {
	case "":
		tracef(ctx, "litellm openai-compatible upstream at an unrecognised host; fingerprinting engines at %v without credentials", bases)
		d.fingerprintEngines(hopCtx, bases, upstreamID, "", &up, false)
		switch up.Stack {
		case "":
			tracef(ctx, "no engine fingerprint matched at the upstream; upstream evidence limited to the bare model id")
		case "litellm":
			tracef(ctx, "upstream is itself a LiteLLM proxy; not followed (one hop only)")
		case "ollama":
		default:
			identified = up.Stack
		}
		kind = up.Stack
	case "venice", "openrouter":
		tracef(ctx, "upstream %s recognised by hostname; reading its model listing without credentials", kind)
		d.probeRegistryShape(hopCtx, bases, upstreamID, "", &up)
		if !up.GatewayReasoning {
			tracef(ctx, "upstream %s listing: no reasoning evidence for %q", kind, upstreamID)
		}
		// up.Stack, not kind: a venice host may serve an openrouter-shaped listing.
		identified = up.Stack
	case "vllm":
		for _, b := range bases {
			if d.probeVLLM(hopCtx, b, upstreamID, "", &up) {
				break
			}
		}
		identified = up.Stack
	case "ollama":
		for _, b := range bases {
			if d.probeOllama(hopCtx, b, upstreamID, "", &up) {
				break
			}
		}
	case "openai":
		tracef(ctx, "upstream openai recognised by hostname; nothing to read from it")
		identified = kind
	case "anthropic":
		tracef(ctx, "upstream anthropic recognised; nothing to read from it")
		identified = kind
	default:
		tracef(ctx, "upstream %q is a hosted vendor; no second hop (only registry vendors and engines are read)", kind)
		return
	}
	if hopCtx.Err() != nil {
		if ctx.Err() != nil {
			tracef(ctx, "second hop cut short by the detection deadline; hop-1 evidence stands")
		} else {
			tracef(ctx, "second hop timed out after %v; hop-1 evidence stands", hopTimeout)
		}
	}

	// Only the stack, served id/architecture and reasoning facts cross the hop.
	// Ollama's thinking capability crosses as gateway evidence so Ollama's own
	// knobs are never advertised.
	switch {
	case identified != "":
		ev.Stack = identified
		ev.Via = "litellm"
		tracef(ctx, "upstream %s identified: it becomes the stack (via litellm); its own params are forwarded by litellm, standard ones only when litellm supports them for this model", identified)
	case kind == "ollama":
		tracef(ctx, "upstream ollama stays evidence-only (litellm's ollama providers speak the native API, not the /v1 surface the ollama table describes); the spec's stack stays litellm")
	}
	if up.Architecture != "" {
		ev.Architecture = up.Architecture
		tracef(ctx, "upstream %s reports architecture %q", kind, up.Architecture)
	}
	if up.ServedModelID != "" {
		ev.ServedModelID = up.ServedModelID
		tracef(ctx, "upstream %s serves %q", kind, up.ServedModelID)
	}
	switch {
	case up.OllamaThinking:
		ev.GatewayReasoning = true
		tracef(ctx, "upstream ollama reports the thinking capability for %q (bindings stay in litellm's vocabulary)", upstreamID)
	case up.GatewayReasoning:
		ev.GatewayReasoning = true
		tracef(ctx, "upstream %s listing reports reasoning support for %q", kind, upstreamID)
	}
}

// Venice appends inline "key=value" parameters to a model id as extra
// ":"-separated segments; an Ollama-style tag ("qwen3:8b") carries no "="
// and must survive.
func bareModelID(id string) string {
	parts := strings.Split(id, ":")
	kept := make([]string, 1, len(parts))
	kept[0] = parts[0]
	for _, p := range parts[1:] {
		if !strings.Contains(p, "=") {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ":")
}

// A hostname passes here; what it resolves to is checked at dial time
// (hopDialControl), which covers DNS rebinding and *.nip.io names.
func hopAllowed(apiBase string) bool {
	u, err := url.Parse(apiBase)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := u.Hostname()
	if host == "" {
		return false
	}
	if addr, err := netip.ParseAddr(host); err == nil && !hopAddrAllowed(addr) {
		return false
	}
	return true
}

// Refuses classic SSRF targets (cloud metadata, link-local discovery).
// Loopback and private ranges stay allowed: self-hosted upstreams live there.
func hopAddrAllowed(addr netip.Addr) bool {
	addr = addr.Unmap()
	return !(addr.IsUnspecified() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast())
}

var errHopAddrRefused = errors.New("second hop refused: address is unspecified, link-local or multicast")

func hopDialControl(network, address string, _ syscall.RawConn) error {
	ap, err := netip.ParseAddrPort(address)
	if err != nil {
		return fmt.Errorf("second hop refused: %s address %q is not an ip:port", network, address)
	}
	if !hopAddrAllowed(ap.Addr()) {
		return fmt.Errorf("%w (%s)", errHopAddrRefused, address)
	}
	return nil
}

type hopCtxKey struct{}

func newHopClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   hopDialControl,
	}).DialContext
	return &http.Client{Timeout: timeout, CheckRedirect: refuseCrossHostRedirect, Transport: transport}
}
