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
//
// The mapping of a kind to what is reported (litellmUpstream): venice,
// openrouter, vllm and the engines the fingerprint finds become the spec's
// stack once the hop confirmed them; openai is taken on the hostname alone;
// ollama and anthropic stay evidence-only (their LiteLLM providers speak
// the native API — Ollama's /api, Anthropic Messages on the claudeai
// transport — not the surface our tables describe through LiteLLM's
// OpenAI-compatible endpoint).
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
// that upstream once, with no credential at all, to learn what it serves and
// whether the model reasons. The upstream's answers land in a scratch
// Evidence and only its stack, served id / architecture and reasoning facts
// are copied out of it (never a chat template, a registry's own bindings or
// parameter list).
//
// When the hop identifies the upstream — venice / openrouter by hostname or
// listing shape, openai by hostname, an engine confirmed at the api_base —
// that upstream becomes the reported stack, with Via = litellm: LiteLLM
// forwards params it does not recognise verbatim as provider kwargs and
// validates the standard OpenAI ones per model, so the upstream's own
// vocabulary is what actually works through it (the composer narrows the
// standard params to LiteLLM's supported_openai_params). Ollama and
// Anthropic upstreams stay evidence-only with stack litellm: LiteLLM's
// ollama/ and ollama_chat/ providers speak Ollama's native API rather than
// the /v1 surface the ollama table describes, and the anthropic preset
// describes Anthropic Messages on the claudeai transport, which is not what
// a model talking to LiteLLM's OpenAI-compatible endpoint speaks. An
// unidentified upstream (unknown prefix, no api_base, hop refused, failed
// or timed out) leaves the stack litellm as well.
// https://docs.litellm.ai/docs/proxy/model_management
// https://docs.litellm.ai/docs/completion/provider_specific_params
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

	// Venice (and possibly others) append inline "key=value" parameters to
	// the model id (e.g. "deepseek-v4-pro:include_venice_system_prompt=false");
	// only the bare id names a real family or a listing entry, so everything
	// from here on uses it — the raw, possibly-suffixed form traced above is
	// the only place it is ever seen.
	upstreamID = bareModelID(upstreamID)

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

	// (c) one anonymous hop, under its own budget and through the hop
	// client (gated dialer), into a scratch Evidence.
	if !hopAllowed(apiBase) {
		tracef(ctx, "litellm api_base %s refused for the second hop (scheme/address)", RedactURL(apiBase))
		return
	}
	hopCtx, cancel := context.WithTimeout(context.WithValue(ctx, hopCtxKey{}, true), hopTimeout)
	defer cancel()
	bases := baseCandidates(apiBase)
	// up is the scratch evidence; identified is the upstream stack the hop
	// confirmed ("" when it could not, or when the kind stays evidence-only).
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
			// evidence-only, traced below like the ollama/ prefix case
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
		// The listing shape decides (a venice host may serve an
		// openrouter-shaped listing); an unreadable listing leaves the
		// upstream unidentified.
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
		tracef(ctx, "upstream anthropic stays evidence-only (the anthropic preset speaks Anthropic Messages on the claudeai transport, not litellm's OpenAI-compatible endpoint); not read, the spec's stack stays litellm")
		return
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

	// (d) What crosses the hop: the identified stack (with Via), the served
	// id / architecture and the reasoning facts. The scratch evidence's chat
	// template, registry bindings, parameters and reasoning efforts never
	// do. Ollama's thinking capability arrives as gateway evidence, never as
	// the Ollama capability path (that would yield Ollama's own knobs).
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

// bareModelID strips inline "key=value" parameters some vendors append to a
// model id as extra ":"-separated segments (Venice: e.g.
// "deepseek-v4-pro:include_venice_system_prompt=false", possibly several
// chained). Only segments containing "=" are dropped: an Ollama-style tag
// (e.g. "gpt-oss:120b", "qwen3:8b") carries no "=" and is left intact, since
// it names the model itself rather than an inline parameter.
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

// hopAllowed reports whether apiBase is safe to dial for the second hop.
// The URL must parse with an http or https scheme and a non-empty hostname;
// when the hostname is a literal IP (zoned IPv6 and v4-mapped forms
// included) it must pass hopAddrAllowed. A hostname passes this gate: what
// it resolves to is checked again at dial time by the hop client's dialer
// (hopDialControl), so a name pointing at a refused range — DNS rebinding,
// 169.254.169.254.nip.io and the like — is stopped there.
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

// hopAddrAllowed is the one address predicate behind both gates (the URL
// gate on api_base and the dial-time gate on what a hostname resolves to):
// the unspecified, link-local (unicast or multicast) and multicast ranges
// are refused — classic SSRF targets (cloud metadata, mDNS/link-local
// discovery, ...) that a third-party api_base should never be able to point
// this hop at. Loopback and private (RFC1918 and equivalent) addresses stay
// allowed: self-hosted vLLM/Ollama upstreams legitimately live there.
// v4-mapped IPv6 addresses are judged as the IPv4 address they carry.
func hopAddrAllowed(addr netip.Addr) bool {
	addr = addr.Unmap()
	return !(addr.IsUnspecified() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast())
}

// errHopAddrRefused is the dial-time gate's refusal, wrapped into the dial
// error the probe sees.
var errHopAddrRefused = errors.New("second hop refused: address is unspecified, link-local or multicast")

// hopDialControl is the hop client's net.Dialer.Control: it sees the
// resolved ip:port every connection is about to be made to and refuses what
// hopAddrAllowed refuses, so DNS answers are covered, not just literals.
// Anything that is not an ip:port is refused too (fail closed).
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

// hopCtxKey marks a context as the LiteLLM second hop: requests under it go
// through Detector.hopClient (clientFor) rather than the probe client.
type hopCtxKey struct{}

// newHopClient builds the second hop's own client: the probe client's
// per-request timeout and same-host redirect policy, over a transport whose
// dialer runs hopDialControl on every resolved address. Everything else
// mirrors the default transport (proxy from the environment, TLS, pool
// limits), so the hop behaves like the first one except for the gate.
func newHopClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   hopDialControl,
	}).DialContext
	return &http.Client{Timeout: timeout, CheckRedirect: refuseCrossHostRedirect, Transport: transport}
}
