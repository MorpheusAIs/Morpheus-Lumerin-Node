package apidetect

import (
	"context"
	"net/http"
	"sort"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
)

// Registry listings grow with the vendor's catalogue (OpenRouter ~0.7 MB).
// Variable so tests can exercise truncation.
var maxRegistryBody int64 = 16 << 20

// OpenRouter GET /api/v1/models entry, field names per
// https://openrouter.ai/docs/api/api-reference/models/get-models
const (
	openrouterModelsSource    = "openrouter:/models"
	openrouterParamsSource    = openrouterModelsSource + ".supported_parameters"
	openrouterReasoningSource = openrouterModelsSource + ".reasoning"
	openrouterEffortsSource   = openrouterReasoningSource + ".supported_efforts"
	// The web plugin (plugins: [{id: "web"}], alias ":online") attaches
	// search results to any model, so the capability holds listing-wide on
	// the OpenRouter service itself (hostname-recognized), not on another
	// host whose listing merely shares the shape:
	// https://openrouter.ai/docs/guides/features/plugins/web-search
	openrouterPluginsSource = openrouterModelsSource + ".plugins"
)

// Venice GET /api/v1/models entry, field names per
// https://docs.venice.ai/api-reference/endpoint/models/list
const (
	veniceSpecSource = "venice:/models.model_spec"
	veniceCapsSource = veniceSpecSource + ".capabilities"
	// The placeholder Venice reports for models without a quantization label.
	veniceQuantizationUnknown = "not-available"
)

// OpenRouter's supported_parameters entries that state a capability, each
// with the capabilities field it proves.
var openrouterCapabilityParams = map[string]string{
	"tools":               apispec.CapFieldTools,
	"parallel_tool_calls": apispec.CapFieldToolsParallel,
	"response_format":     apispec.CapFieldStructured,
	"structured_outputs":  apispec.CapFieldStructured,
	"web_search_options":  apispec.CapFieldWebSearch,
}

// Venice capability flags naming an input modality.
var veniceInputModalities = []struct{ flag, modality string }{
	{"supportsVision", "image"},
	{"supportsVideoInput", "video"},
	{"supportsAudioInput", "audio"},
}

func init() {
	for _, vendor := range []string{"openrouter", "venice"} {
		registerHostedProbe(vendor, probeHostedRegistry)
	}
}

func probeHostedRegistry(ctx context.Context, d *Detector, bases []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as %s by hostname; importing its model listing", ev.Stack)
	d.probeRegistryShape(ctx, bases, cfg.ModelName, cfg.ApiKey, ev)
}

func (d *Detector) probeRegistryShape(ctx context.Context, bases []string, modelName, apiKey string, ev *apispec.Evidence) {
	// A stack set before the listing is read came from hostname recognition;
	// the fingerprint fallback arrives with none, and service-level facts are
	// then not asserted from the listing's shape alone.
	hosted := ev.Stack != ""
	for _, base := range bases {
		list := d.requestJSON(ctx, http.MethodGet, base+"/models", apiKey, nil, maxRegistryBody)
		if list == nil {
			continue
		}
		data, ok := list["data"].([]any)
		if !ok {
			continue
		}
		entry := matchModelEntry(data, modelName)
		if entry == nil {
			tracef(ctx, "model listing at %s/models has %d entries but none matches model %q (config drift?)", base, len(data), modelName)
			continue
		}

		if params, ok := entry["supported_parameters"].([]any); ok {
			if ev.Stack == "" {
				ev.Stack = "openrouter"
			}
			tracef(ctx, "model listing at %s/models has supported_parameters -> openrouter-style registry", base)
			readOpenRouterEntry(ctx, entry, params, modelName, hosted, ev)
			return
		}

		if spec, ok := entry["model_spec"].(map[string]any); ok {
			if ev.Stack == "" {
				ev.Stack = "venice"
			}
			tracef(ctx, "model listing at %s/models has model_spec -> venice-style registry", base)
			readVeniceSpec(ctx, spec, ev)
			return
		}
	}
}

func readOpenRouterEntry(ctx context.Context, entry map[string]any, params []any, modelName string, hosted bool, ev *apispec.Evidence) {
	var structured []string
	for _, p := range params {
		s, ok := p.(string)
		if !ok {
			continue
		}
		ev.Parameters = append(ev.Parameters, s)
		switch openrouterCapabilityParams[s] {
		case apispec.CapFieldTools:
			ev.Caps.SetTools(true, openrouterParamsSource)
			tracef(ctx, "/models.supported_parameters lists tools -> tools")
		case apispec.CapFieldToolsParallel:
			ev.Caps.SetToolsParallel(true, openrouterParamsSource)
			tracef(ctx, "/models.supported_parameters lists parallel_tool_calls -> toolsParallel")
		case apispec.CapFieldStructured:
			kind := "json_object"
			if s == "structured_outputs" {
				kind = "json_schema"
			}
			structured = append(structured, kind)
			tracef(ctx, "/models.supported_parameters lists %s -> structured %s", s, kind)
		case apispec.CapFieldWebSearch:
			ev.Caps.SetWebSearch(true, openrouterParamsSource)
			tracef(ctx, "/models.supported_parameters lists web_search_options -> webSearch")
		}
	}
	sort.Strings(ev.Parameters)
	for _, p := range ev.Parameters {
		if p == "reasoning" {
			ev.GatewayReasoning = true
			tracef(ctx, "openrouter listing reports reasoning support for %q", modelName)
			break
		}
	}
	if structured != nil {
		ev.Caps.SetStructured(structured, openrouterParamsSource)
	}
	if ev.Caps.WebSearch == nil {
		if hosted {
			ev.Caps.SetWebSearch(true, openrouterPluginsSource)
			tracef(ctx, "/models: the web plugin searches for any openrouter model -> webSearch")
		} else {
			tracef(ctx, "/models: openrouter-shaped listing on an unrecognized host; the web plugin is not assumed, webSearch left unset")
		}
	}

	if n, ok := registryInt(entry["context_length"]); ok {
		if ev.Caps.SetContext(n, openrouterModelsSource+".context_length") {
			tracef(ctx, "/models.context_length %d -> context", n)
		} else {
			tracef(ctx, "/models.context_length %d is out of range; ignored", n)
		}
	}
	if top, ok := entry["top_provider"].(map[string]any); ok {
		if ev.Caps.Context == nil {
			if n, ok := registryInt(top["context_length"]); ok {
				if ev.Caps.SetContext(n, openrouterModelsSource+".top_provider.context_length") {
					tracef(ctx, "/models.top_provider.context_length %d -> context (context_length absent)", n)
				} else {
					tracef(ctx, "/models.top_provider.context_length %d is out of range; ignored", n)
				}
			}
		}
		if n, ok := registryInt(top["max_completion_tokens"]); ok {
			if ev.Caps.SetMaxOutput(n, openrouterModelsSource+".top_provider.max_completion_tokens") {
				tracef(ctx, "/models.top_provider.max_completion_tokens %d -> maxOutput", n)
			} else {
				tracef(ctx, "/models.top_provider.max_completion_tokens %d is out of range; ignored", n)
			}
		}
	}
	if arch, ok := entry["architecture"].(map[string]any); ok {
		if in := registryStrings(arch["input_modalities"]); in != nil {
			// OpenRouter names document input "file"; the wire vocabulary calls it pdf.
			for i, m := range in {
				if m == "file" {
					in[i] = "pdf"
				}
			}
			if ev.Caps.SetInputModalities(in, openrouterModelsSource+".architecture.input_modalities") {
				tracef(ctx, "/models.architecture.input_modalities %v -> modalities.input %v", in, ev.Caps.Modalities.Input)
			} else {
				tracef(ctx, "/models.architecture.input_modalities %v has no known entry; ignored", in)
			}
		}
		if out := registryStrings(arch["output_modalities"]); out != nil {
			if ev.Caps.SetOutputModalities(out, openrouterModelsSource+".architecture.output_modalities") {
				tracef(ctx, "/models.architecture.output_modalities %v -> modalities.output %v", out, ev.Caps.Modalities.Output)
			} else {
				tracef(ctx, "/models.architecture.output_modalities %v has no known entry; ignored", out)
			}
		}
	}

	reasoning, ok := entry["reasoning"].(map[string]any)
	if !ok {
		return
	}
	mandatory, _ := reasoning["mandatory"].(bool)
	if mandatory {
		ev.MandatoryReasoning = true
		tracef(ctx, "/models.reasoning.mandatory -> reasoning cannot be disabled")
	}
	if efforts := registryStrings(reasoning["supported_efforts"]); efforts != nil {
		ev.ReasoningEfforts = efforts
		ev.EffortsSource = openrouterEffortsSource
		tracef(ctx, "/models.reasoning.supported_efforts %v -> reasoning.effort levels", efforts)
	}
	// default_enabled is whether reasoning runs unasked; default_effort is the
	// level applied once it runs. The level is a default only when the toggle
	// is on or reasoning is mandatory (the live listing pairs
	// default_enabled: false with a default_effort on many models).
	enabled, hasEnabled := reasoning["default_enabled"].(bool)
	effort, _ := reasoning["default_effort"].(string)
	switch {
	case !mandatory && hasEnabled && !enabled:
		ev.ThinkingDefault = "off"
		ev.ThinkingDefaultSource = openrouterReasoningSource + ".default_enabled"
		tracef(ctx, "/models.reasoning.default_enabled false -> thinking.default off (default_effort %q applies only when reasoning is requested)", effort)
	case !mandatory && !hasEnabled:
		tracef(ctx, "/models.reasoning.default_enabled absent and reasoning not mandatory; thinking.default left unset (default_effort %q applies only when reasoning is requested)", effort)
	case effort != "":
		ev.ThinkingDefault = effort
		ev.ThinkingDefaultSource = openrouterReasoningSource + ".default_effort"
		tracef(ctx, "/models.reasoning.default_effort %q -> thinking.default", effort)
	case hasEnabled:
		ev.ThinkingDefault = "on"
		ev.ThinkingDefaultSource = openrouterReasoningSource + ".default_enabled"
		tracef(ctx, "/models.reasoning.default_enabled true -> thinking.default on")
	default:
		ev.ThinkingDefault = "on"
		ev.ThinkingDefaultSource = openrouterReasoningSource + ".mandatory"
		tracef(ctx, "/models.reasoning.mandatory -> thinking.default on")
	}
	if supportsMax, ok := reasoning["supports_max_tokens"].(bool); ok && !supportsMax {
		tracef(ctx, "/models.reasoning.supports_max_tokens false; the reasoning.budget binding stays as documented for the stack")
	}
}

func readVeniceSpec(ctx context.Context, spec map[string]any, ev *apispec.Evidence) {
	if n, ok := registryInt(spec["availableContextTokens"]); ok {
		if ev.Caps.SetContext(n, veniceSpecSource+".availableContextTokens") {
			tracef(ctx, "/models.model_spec.availableContextTokens %d -> context", n)
		} else {
			tracef(ctx, "/models.model_spec.availableContextTokens %d is out of range; ignored", n)
		}
	}
	if n, ok := registryInt(spec["maxCompletionTokens"]); ok {
		if ev.Caps.SetMaxOutput(n, veniceSpecSource+".maxCompletionTokens") {
			tracef(ctx, "/models.model_spec.maxCompletionTokens %d -> maxOutput", n)
		} else {
			tracef(ctx, "/models.model_spec.maxCompletionTokens %d is out of range; ignored", n)
		}
	}
	caps, ok := spec["capabilities"].(map[string]any)
	if !ok {
		return
	}
	if tools, ok := caps["supportsFunctionCalling"].(bool); ok {
		ev.Caps.SetTools(tools, veniceCapsSource+".supportsFunctionCalling")
		tracef(ctx, "/models.model_spec.capabilities.supportsFunctionCalling %t -> tools", tools)
	}
	for _, m := range veniceInputModalities {
		if v, ok := caps[m.flag].(bool); ok && v {
			ev.Caps.AddInputModality(m.modality, veniceCapsSource+"."+m.flag)
			tracef(ctx, "/models.model_spec.capabilities.%s -> modalities.input %s", m.flag, m.modality)
		}
	}
	if search, ok := caps["supportsWebSearch"].(bool); ok {
		ev.Caps.SetWebSearch(search, veniceCapsSource+".supportsWebSearch")
		tracef(ctx, "/models.model_spec.capabilities.supportsWebSearch %t -> webSearch", search)
	}
	if schema, ok := caps["supportsResponseSchema"].(bool); ok && schema {
		ev.Caps.SetStructured([]string{"json_schema"}, veniceCapsSource+".supportsResponseSchema")
		tracef(ctx, "/models.model_spec.capabilities.supportsResponseSchema -> structured json_schema")
	}
	if reasons, ok := caps["supportsReasoning"].(bool); ok {
		if reasons {
			ev.GatewayReasoning = true
			tracef(ctx, "venice listing reports supportsReasoning")
		} else {
			ev.Caps.SetReasoning(false, veniceCapsSource+".supportsReasoning")
			tracef(ctx, "/models.model_spec.capabilities.supportsReasoning false -> reasoning false")
		}
	}
	if effortful, ok := caps["supportsReasoningEffort"].(bool); ok && effortful {
		if efforts := registryStrings(caps["reasoningEffortOptions"]); efforts != nil {
			ev.ReasoningEfforts = efforts
			ev.EffortsSource = veniceCapsSource + ".reasoningEffortOptions"
			tracef(ctx, "/models.model_spec.capabilities.reasoningEffortOptions %v -> reasoning.effort levels", efforts)
		}
	}
	if effort, ok := caps["defaultReasoningEffort"].(string); ok && effort != "" {
		ev.ThinkingDefault = effort
		ev.ThinkingDefaultSource = veniceCapsSource + ".defaultReasoningEffort"
		tracef(ctx, "/models.model_spec.capabilities.defaultReasoningEffort %q -> thinking.default", effort)
	}
	if label, ok := caps["quantization"].(string); ok && label != "" && label != veniceQuantizationUnknown {
		if ev.Caps.SetQuantization(label, veniceCapsSource+".quantization") {
			tracef(ctx, "/models.model_spec.capabilities.quantization %q -> quantization", label)
		} else {
			tracef(ctx, "/models.model_spec.capabilities.quantization is not a plain label; ignored")
		}
	}
}

// JSON numbers decode as float64; a fractional or non-numeric value is not a
// token count.
func registryInt(v any) (int64, bool) {
	f, ok := v.(float64)
	if !ok || f != float64(int64(f)) {
		return 0, false
	}
	return int64(f), true
}

func registryStrings(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
