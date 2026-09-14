package apidetect

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// hopTimeout bounds the second hop (every request to the upstream behind a
// LiteLLM deployment) as a whole: the hop runs under a sub-context of
// min(remaining detection budget, hopTimeout), so a slow upstream can cost
// the hop its own evidence but never the first hop's. A var so tests can
// exercise the timeout path.
var hopTimeout = 4 * time.Second

// upstreamByHost marks a provider whose upstream is a custom OpenAI-compatible
// endpoint: the api_base host decides the kind. This is how Venice is reached
// through LiteLLM (it has no native provider).
const upstreamByHost = "host"

// litellmProviderStacks maps a LiteLLM provider — litellm_params.model's
// "<provider>/" prefix or custom_llm_provider — to the upstream kind the
// second hop may probe. Only prefixes documented in the research reference
// (docs.litellm.ai/docs/providers/*) are listed; anything else is an unknown
// upstream. "" marks a documented prefix whose upstream is a hosted vendor or
// cloud: recognised so the bare model id can be taken, never probed, never
// reported (the spec's stack stays litellm).
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

// litellmUpstream is the second hop: it reads the deployment behind modelName
// from LiteLLM's /model/info (provider key, same host), takes the bare
// upstream model id as family evidence and — when the upstream is a registry
// vendor (venice, openrouter), an engine (vllm, ollama) or an unrecognised
// OpenAI-compatible host (fingerprinted like a self-hosted engine) — reads
// that upstream once, with no credential at all, only to learn what it serves
// and whether the model reasons. The upstream's answers land in a scratch
// Evidence and only family / reasoning facts are copied out of it: nothing
// from the upstream is reported directly, the stack stays litellm and the
// bindings stay in LiteLLM's vocabulary (its param translation and
// drop_params rules are invisible from here).
// https://docs.litellm.ai/docs/proxy/model_management
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
	// Only these values ever reach the trace: no keys, no raw deployment
	// fields, api_base redacted.
	apiBaseText := "(none)"
	if apiBase != "" {
		apiBaseText = RedactURL(apiBase)
	}
	tracef(ctx, "litellm /model/info: deployment %q -> provider %q, upstream model %q, api_base %s", modelName, provider, upstreamID, apiBaseText)

	// (b) the bare upstream id is family / always-on evidence.
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
	case apiBase == "":
		tracef(ctx, "litellm deployment %q (provider %q) has no api_base; upstream not probed", modelName, provider)
		return
	}
	if kind == upstreamByHost {
		// "" = no recognised vendor at that host: fingerprint it below.
		kind = stackFromHost(apiBase)
	}

	// (c) one anonymous hop, under its own budget, into a scratch Evidence.
	if !hopAllowed(apiBase) {
		tracef(ctx, "litellm api_base %s refused for the second hop (scheme/address)", RedactURL(apiBase))
		return
	}
	hopCtx, cancel := context.WithTimeout(ctx, hopTimeout)
	defer cancel()
	bases := baseCandidates(apiBase)
	var up apispec.Evidence
	switch kind {
	case "":
		tracef(ctx, "litellm openai-compatible upstream at an unrecognised host; fingerprinting engines at %v without credentials", bases)
		d.fingerprintEngines(hopCtx, bases, upstreamID, "", &up, false)
		switch up.Stack {
		case "":
			tracef(ctx, "no engine fingerprint matched at the upstream; upstream evidence limited to the bare model id")
		case "litellm":
			tracef(ctx, "upstream is itself a LiteLLM proxy; not followed (one hop only)")
		}
		kind = up.Stack
	case "venice", "openrouter":
		tracef(ctx, "upstream %s recognised by hostname; reading its model listing without credentials", kind)
		d.probeRegistryShape(hopCtx, bases, upstreamID, "", &up)
		if !up.GatewayReasoning {
			tracef(ctx, "upstream %s listing: no reasoning evidence for %q", kind, upstreamID)
		}
	case "vllm":
		for _, b := range bases {
			if d.probeVLLM(hopCtx, b, upstreamID, "", &up) {
				break
			}
		}
	case "ollama":
		for _, b := range bases {
			if d.probeOllama(hopCtx, b, upstreamID, "", &up) {
				break
			}
		}
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

	// (d) Only family / reasoning facts cross the hop: the scratch
	// evidence's stack, chat template, registry bindings, parameters and
	// reasoning efforts never do. Ollama's thinking capability arrives as
	// gateway evidence, never as the Ollama capability path (that would
	// yield Ollama's own knobs).
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
		tracef(ctx, "upstream %s listing reports reasoning support for %q (bindings stay in litellm's vocabulary)", kind, upstreamID)
	}
}

// hopAllowed reports whether apiBase is safe to dial for the second hop.
// The URL must parse with an http or https scheme and a non-empty hostname;
// when the hostname is a literal IP, the unspecified, link-local (unicast
// or multicast) and multicast ranges are refused — classic SSRF targets
// (cloud metadata, mDNS/link-local discovery, ...) that a third-party
// api_base should never be able to point this hop at. Loopback and private
// (RFC1918 and equivalent) addresses stay allowed: self-hosted vLLM/Ollama
// upstreams legitimately live there.
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
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
			return false
		}
	}
	return true
}
