package apidetect

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
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
	litellmDeploymentCaps(ctx, dep, ev)
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

	// The stack, served id/architecture, reasoning and capability facts cross
	// the hop. Ollama's thinking capability crosses as gateway evidence so
	// Ollama's own knobs are never advertised.
	switch {
	case identified != "":
		ev.Stack = identified
		ev.Via = "litellm"
		tracef(ctx, "upstream %s identified: it becomes the stack (via litellm); its own params are forwarded by litellm, standard ones only when litellm supports them for this model", identified)
		if len(up.ReasoningEfforts) > 0 && up.EffortsSource != "" {
			ev.ReasoningEfforts = append([]string(nil), up.ReasoningEfforts...)
			ev.EffortsSource = up.EffortsSource
			tracef(ctx, "upstream %s listing offers reasoning efforts %s", identified, tracedKeys(up.ReasoningEfforts))
		}
		if up.NoEffort {
			ev.NoEffort = true
			tracef(ctx, "upstream %s listing says the model takes no effort level", identified)
		}
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
	if identified != "" || kind == "ollama" {
		crossUpstreamFacts(ctx, kind, &up, ev)
	}
}

// Capability facts that need a standard request param LiteLLM validates:
// when supported_openai_params is known and does not list it, LiteLLM never
// lets the field reach the upstream, so a positive fact stays behind (a
// negative one is true whatever LiteLLM forwards).
var hopGatedCaps = map[string]string{
	apispec.CapFieldTools:         "tools",
	apispec.CapFieldToolsParallel: "parallel_tool_calls",
	apispec.CapFieldStructured:    "response_format",
}

// crossUpstreamFacts copies the upstream's capability facts into the fields
// hop 1 left unset: LiteLLM's own group and deployment facts describe what
// it serves and keep precedence, except for a window limit, where the
// upstream's smaller figure wins (LiteLLM's comes from its cost map or the
// operator's model_info and is not enforced). The thinking default and
// mandatory reasoning describe the model, not the transport, so they cross
// too.
func crossUpstreamFacts(ctx context.Context, kind string, up, ev *apispec.Evidence) {
	crossing := apispec.CapabilityFacts{ModelCapabilities: up.Caps.ModelCapabilities, Sources: map[string]string{}}
	fields := make([]string, 0, len(up.Caps.Sources))
	for field := range up.Caps.Sources {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		source := up.Caps.Sources[field]
		if own, set := ev.Caps.Sources[field]; set {
			if !crossSmallerWindow(ctx, kind, field, up, ev) {
				tracef(ctx, "upstream %s capabilities.%s not crossed: litellm reports it (%s)", kind, field, own)
			}
			continue
		}
		if param, gated := hopGatedCaps[field]; gated && claimsSupport(&up.Caps, field) && ev.LiteLLMSupportedParams != nil && !containsString(ev.LiteLLMSupportedParams, param) {
			tracef(ctx, "upstream %s capabilities.%s not crossed: litellm does not forward %s for this model", kind, field, param)
			continue
		}
		// LiteLLM's Ollama providers carry text and image_url parts to
		// Ollama; an audio part is not known to survive the translation.
		if kind == "ollama" && field == apispec.CapFieldModalitiesInput && containsString(up.Caps.Modalities.Input, "audio") {
			var kept []string
			for _, m := range up.Caps.Modalities.Input {
				if m != "audio" {
					kept = append(kept, m)
				}
			}
			tracef(ctx, "upstream ollama capabilities.modalities.input: audio not crossed (litellm's ollama providers translate text and image parts only)")
			if len(kept) == 0 {
				continue
			}
			crossing.Modalities = &system.Modalities{Input: kept, Output: up.Caps.Modalities.Output}
		}
		crossing.Sources[field] = source
		tracef(ctx, "upstream %s capabilities.%s crosses the hop (%s)", kind, field, source)
	}
	ev.Caps.FillFrom(&crossing)

	if ev.ThinkingDefault == "" && up.ThinkingDefault != "" {
		ev.ThinkingDefault = up.ThinkingDefault
		ev.ThinkingDefaultSource = up.ThinkingDefaultSource
		tracef(ctx, "upstream %s thinking default %q crosses the hop (%s)", kind, up.ThinkingDefault, up.ThinkingDefaultSource)
	}
	if up.MandatoryReasoning && !ev.MandatoryReasoning {
		ev.MandatoryReasoning = true
		tracef(ctx, "upstream %s reports reasoning as mandatory for this model", kind)
	}
}

func claimsSupport(caps *apispec.CapabilityFacts, field string) bool {
	switch field {
	case apispec.CapFieldTools:
		return caps.Tools != nil && *caps.Tools
	case apispec.CapFieldToolsParallel:
		return caps.ToolsParallel != nil && *caps.ToolsParallel
	}
	return true
}

// crossSmallerWindow publishes the upstream's context or maxOutput when it
// is below the figure LiteLLM reports: the upstream's is what a request meets.
func crossSmallerWindow(ctx context.Context, kind, field string, up, ev *apispec.Evidence) bool {
	var upV, ownV *int64
	switch field {
	case apispec.CapFieldContext:
		upV, ownV = up.Caps.Context, ev.Caps.Context
	case apispec.CapFieldMaxOutput:
		upV, ownV = up.Caps.MaxOutput, ev.Caps.MaxOutput
	default:
		return false
	}
	if upV == nil || ownV == nil || *upV >= *ownV {
		return false
	}
	source := up.Caps.Sources[field]
	own := ev.Caps.Sources[field]
	if field == apispec.CapFieldContext {
		ev.Caps.SetContext(*upV, source)
	} else {
		ev.Caps.SetMaxOutput(*upV, source)
	}
	tracef(ctx, "upstream %s capabilities.%s %d is below litellm's %d (%s); the smaller crosses the hop (%s)", kind, field, *upV, *ownV, own, source)
	return true
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

func (d *Detector) probeLiteLLM(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence, introspect bool) bool {
	body := d.getText(ctx, base+"/health/liveliness", apiKey)
	if body == "" {
		return false
	}
	alive := strings.Trim(body, `"`)
	if !strings.EqualFold(alive, "I'm alive!") {
		return false
	}
	ev.Stack = "litellm"
	ev.LiteLLMSeen = true
	tracef(ctx, "identified litellm by /health/liveliness")

	if !introspect || modelName == "" {
		return true
	}
	if info := d.getJSON(ctx, base+"/model_group/info?model_group="+url.QueryEscape(modelName), apiKey); info != nil {
		group := info
		if data, ok := info["data"].([]any); ok {
			if len(data) == 0 {
				group = nil
			} else if group, ok = data[0].(map[string]any); !ok {
				group = nil
			}
		}
		if group != nil {
			if params, ok := group["supported_openai_params"].([]any); ok {
				// Non-nil even when empty: nil means the list could not be read.
				supported := append(make([]string, 0, len(params)), jsonStrings(params)...)
				sort.Strings(supported)
				ev.LiteLLMSupportedParams = supported
				tracef(ctx, "litellm /model_group/info lists %d supported_openai_params", len(supported))
			}
			if reasons, ok := group["supports_reasoning"].(bool); ok && reasons {
				ev.GatewayReasoning = true
				tracef(ctx, "litellm /model_group/info reports supports_reasoning")
			}
			if efforts := jsonStrings(group["supported_reasoning_efforts"]); efforts != nil {
				ev.ReasoningEfforts = append(ev.ReasoningEfforts, efforts...)
				if len(ev.ReasoningEfforts) > 0 {
					ev.EffortsSource = litellmGroupSource + "supported_reasoning_efforts"
					tracef(ctx, "litellm /model_group/info lists supported_reasoning_efforts %v", ev.ReasoningEfforts)
				}
			}
			litellmGroupCaps(ctx, group, ev)
		}
	}
	if d.opts.TwoHop {
		d.litellmUpstream(ctx, base, modelName, apiKey, ev)
	}
	return true
}

// Evidence source prefixes: the facts below describe what LiteLLM serves and
// stay attributed to litellm even when an identified upstream becomes the
// stack (Via = litellm).
const (
	litellmGroupSource      = "litellm:/model_group/info."
	litellmDeploymentSource = "litellm:/model/info.model_info."
)

// ModelGroupInfo (https://github.com/BerriAI/litellm/blob/main/litellm/types/router.py)
// aggregates the group's deployments: a limit is the group's, and every
// supports_* flag is Field(default=False), flipped to true only when a
// deployment's cost-map entry says so. A false therefore also covers a model
// the cost map does not know, so only true crosses as a fact; the one
// exception is supports_reasoning false (contract rule 3), which the
// composer publishes only when no reasoning toggle or always_on mode
// contradicts it. The group carries no supports_audio_input; audio is a
// deployment-level field.
// supports_url_context has no wire field.
func litellmGroupCaps(ctx context.Context, group map[string]any, ev *apispec.Evidence) {
	src := func(field string) string { return litellmGroupSource + field }
	if v, ok := jsonWholeNumber(group["max_input_tokens"]); ok {
		if ev.Caps.SetContext(v, src("max_input_tokens")) {
			tracef(ctx, "litellm /model_group/info max_input_tokens %d -> capabilities.context", v)
		} else {
			tracef(ctx, "litellm /model_group/info max_input_tokens %d out of bounds; ignored", v)
		}
	}
	if v, ok := jsonWholeNumber(group["max_output_tokens"]); ok {
		if ev.Caps.SetMaxOutput(v, src("max_output_tokens")) {
			tracef(ctx, "litellm /model_group/info max_output_tokens %d -> capabilities.maxOutput", v)
		} else {
			tracef(ctx, "litellm /model_group/info max_output_tokens %d out of bounds; ignored", v)
		}
	}
	if v, ok := group["supports_vision"].(bool); ok && v && ev.Caps.AddInputModality("image", src("supports_vision")) {
		tracef(ctx, "litellm /model_group/info supports_vision -> capabilities.modalities.input += image")
	}
	if v, ok := group["supports_function_calling"].(bool); ok && v {
		ev.Caps.SetTools(true, src("supports_function_calling"))
		tracef(ctx, "litellm /model_group/info supports_function_calling -> capabilities.tools")
	}
	if v, ok := group["supports_parallel_function_calling"].(bool); ok && v {
		ev.Caps.SetToolsParallel(true, src("supports_parallel_function_calling"))
		tracef(ctx, "litellm /model_group/info supports_parallel_function_calling -> capabilities.toolsParallel")
	}
	if v, ok := group["supports_web_search"].(bool); ok && v {
		ev.Caps.SetWebSearch(true, src("supports_web_search"))
		tracef(ctx, "litellm /model_group/info supports_web_search -> capabilities.webSearch")
	}
	// A true flag is GatewayReasoning (read above); only the false crosses as
	// a capability, since reasoning = true is derived from surviving bindings.
	if v, ok := group["supports_reasoning"].(bool); ok && !v {
		ev.Caps.SetReasoning(false, src("supports_reasoning"))
		tracef(ctx, "litellm /model_group/info supports_reasoning false -> capabilities.reasoning")
	}
	// JSON mode (response_format {type: json_object}) is available whenever
	// litellm lists response_format among the model's supported params; a
	// schema needs supports_response_schema on a deployment
	// (https://docs.litellm.ai/docs/completion/json_mode).
	if containsString(ev.LiteLLMSupportedParams, "response_format") && ev.Caps.SetStructured([]string{"json_object"}, src("supported_openai_params")) {
		tracef(ctx, "litellm /model_group/info supported_openai_params lists response_format -> capabilities.structured = json_object")
	}
}

// Per-deployment model_info (https://github.com/BerriAI/litellm/blob/main/litellm/types/utils.py
// ModelInfoBase); read from the deployment the hop already matched, so no
// second request. Only true flags are facts: an absent cost-map entry leaves
// them false.
func litellmDeploymentCaps(ctx context.Context, dep map[string]any, ev *apispec.Evidence) {
	info, _ := dep["model_info"].(map[string]any)
	if info == nil {
		return
	}
	src := func(field string) string { return litellmDeploymentSource + field }
	if v, ok := info["supports_prompt_caching"].(bool); ok && v {
		ev.Caps.SetPromptCache(true, src("supports_prompt_caching"))
		tracef(ctx, "litellm /model/info model_info.supports_prompt_caching -> capabilities.promptCache")
	}
	if v, ok := info["supports_response_schema"].(bool); ok && v && ev.Caps.SetStructured([]string{"json_object", "json_schema"}, src("supports_response_schema")) {
		tracef(ctx, "litellm /model/info model_info.supports_response_schema -> capabilities.structured = json_object, json_schema")
	}
	if v, ok := info["supports_pdf_input"].(bool); ok && v && ev.Caps.AddInputModality("pdf", src("supports_pdf_input")) {
		tracef(ctx, "litellm /model/info model_info.supports_pdf_input -> capabilities.modalities.input += pdf")
	}
	if v, ok := info["supports_audio_input"].(bool); ok && v && ev.Caps.AddInputModality("audio", src("supports_audio_input")) {
		tracef(ctx, "litellm /model/info model_info.supports_audio_input -> capabilities.modalities.input += audio")
	}
}
