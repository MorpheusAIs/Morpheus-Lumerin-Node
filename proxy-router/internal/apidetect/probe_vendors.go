package apidetect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
)

// Hosted vendors' per-model metadata endpoints. Each probe talks to the
// model's own host only, carries the model's key only where the vendor
// documents one, and names its facts "<vendor>:<endpoint>.<field>" with the
// endpoint spelled with a literal {id}: the configured model id never enters
// a source string. OpenAI registers nothing: its listing carries no
// capability field (https://developers.openai.com/api/reference/resources/models/methods/retrieve).
func init() {
	registerHostedProbe("anthropic", probeAnthropic)
	registerHostedProbe("gemini", probeGemini)
	registerHostedProbe("mistral", probeMistral)
	registerHostedProbe("groq", probeGroq)
	registerHostedProbe("xai", probeXAI)
	registerHostedProbe("together", probeTogether)
	registerHostedProbe("cerebras", probeCerebras)
	registerHostedProbe("deepinfra", probeDeepInfra)
	registerHostedProbe("hf-router", probeHFRouter)
}

const (
	anthropicModelSource = "anthropic:/v1/models/{id}"
	geminiModelSource    = "gemini:/v1beta/models/{id}"
	mistralModelSource   = "mistral:/v1/models/{id}"
	groqModelSource      = "groq:/openai/v1/models/{id}"
	xaiModelSource       = "xai:/v1/language-models/{id}"
	togetherListSource   = "together:/v1/models"
	cerebrasModelSource  = "cerebras:/public/v1/models/{id}"
	deepinfraListSource  = "deepinfra:/v1/openai/models"
	hfRouterListSource   = "hf-router:/v1/models"

	// https://platform.claude.com/docs/en/api/versioning
	anthropicVersion = "2023-06-01"
)

// The order Anthropic documents the levels in; the composer's enum keeps it.
var anthropicEffortLevels = []string{"low", "medium", "high", "xhigh", "max"}

// Vendors document their metadata paths from the host root, so the
// configured chat path is not a prefix to them.
func vendorOrigin(apiURL string) string {
	u, err := url.Parse(apiURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// The lookup id is the configured name without inline parameters, escaped
// so a "/" in it stays one path segment.
func vendorModelID(ctx context.Context, cfg config.ModelConfig) string {
	id := bareModelID(cfg.ModelName)
	if id == "" {
		tracef(ctx, "no modelName configured; nothing to look up")
		return ""
	}
	return url.PathEscape(id)
}

func bearerHeader(apiKey string) map[string]string {
	if apiKey == "" {
		return nil
	}
	return map[string]string{"Authorization": "Bearer " + apiKey}
}

// vendorGet fetches a JSON document with exactly the given headers: vendor
// auth schemes (x-api-key, x-goog-api-key) and top-level arrays (Together)
// are outside requestJSON. A non-2xx status, an oversized body and a
// non-JSON body are each one trace line and a nil document.
func (d *Detector) vendorGet(ctx context.Context, urlStr string, headers map[string]string, limit int64) any {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := d.clientFor(ctx).Do(req)
	if err != nil {
		tracef(ctx, "GET %s -> %v", urlStr, err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		tracef(ctx, "GET %s -> HTTP %d", urlStr, resp.StatusCode)
		return nil
	}
	limited := &countingReader{r: io.LimitReader(resp.Body, limit)}
	var out any
	if err := json.NewDecoder(limited).Decode(&out); err != nil {
		if limited.n >= limit {
			tracef(ctx, "GET %s -> HTTP %d, but the body exceeds the %d-byte limit; ignored", urlStr, resp.StatusCode, limit)
		} else {
			tracef(ctx, "GET %s -> HTTP %d, but not JSON: %v", urlStr, resp.StatusCode, err)
		}
		return nil
	}
	switch v := out.(type) {
	case map[string]any:
		tracef(ctx, "GET %s -> HTTP %d, JSON object with keys %s", urlStr, resp.StatusCode, tracedKeys(mapKeys(v)))
	case []any:
		tracef(ctx, "GET %s -> HTTP %d, JSON array of %d", urlStr, resp.StatusCode, len(v))
	default:
		tracef(ctx, "GET %s -> HTTP %d, but neither a JSON object nor an array; ignored", urlStr, resp.StatusCode)
		return nil
	}
	return out
}

func (d *Detector) vendorObject(ctx context.Context, urlStr string, headers map[string]string) map[string]any {
	m, _ := d.vendorGet(ctx, urlStr, headers, maxProbeBody).(map[string]any)
	return m
}

// A listing is either a bare array (Together) or OpenAI's {"data": [...]}.
func (d *Detector) vendorList(ctx context.Context, urlStr string, headers map[string]string) []any {
	switch v := d.vendorGet(ctx, urlStr, headers, maxRegistryBody).(type) {
	case []any:
		return v
	case map[string]any:
		list, _ := v["data"].([]any)
		return list
	}
	return nil
}

// A hosted listing is keyed by id: a lone entry whose id is not the
// configured model describes a different model (a key scoped to one
// dedicated endpoint, a filtered listing), so findModelEntry's
// single-entry rule for self-hosted engines does not apply here.
func vendorEntry(ctx context.Context, list []any, endpoint, modelName string) map[string]any {
	if list == nil {
		return nil
	}
	names := []string{modelName}
	if bare := bareModelID(modelName); bare != modelName {
		names = append(names, bare)
	}
	for _, name := range names {
		for _, item := range list {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if id, ok := m["id"].(string); ok && strings.EqualFold(id, name) {
				return m
			}
		}
	}
	tracef(ctx, "%s has %d entries but none matches the configured model (config drift?)", endpoint, len(list))
	return nil
}

func objectAt(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	sub, _ := m[key].(map[string]any)
	return sub
}

// The setters apply the capability bound; the reader only wants a whole number.
func integerAt(m map[string]any, key string) (int64, bool) {
	return jsonWholeNumber(m[key])
}

func boolAt(m map[string]any, key string) (bool, bool) {
	b, ok := m[key].(bool)
	return b, ok
}

func stringsAt(m map[string]any, key string) []string { return jsonStrings(m[key]) }

// Anthropic capability leaves are {"supported": bool}.
func supportedAt(m map[string]any, key string) bool {
	b, _ := objectAt(m, key)["supported"].(bool)
	return b
}

func traceFact(ctx context.Context, source string, value any, field string) {
	tracef(ctx, "%s = %v -> capabilities.%s", source, value, field)
}

func setContextAt(ctx context.Context, caps *apispec.CapabilityFacts, m map[string]any, key, source string) {
	if n, ok := integerAt(m, key); ok && caps.SetContext(n, source) {
		traceFact(ctx, source, n, apispec.CapFieldContext)
	}
}

func setMaxOutputAt(ctx context.Context, caps *apispec.CapabilityFacts, m map[string]any, key, source string) {
	if n, ok := integerAt(m, key); ok && caps.SetMaxOutput(n, source) {
		traceFact(ctx, source, n, apispec.CapFieldMaxOutput)
	}
}

func addInputModalityWhen(ctx context.Context, caps *apispec.CapabilityFacts, m map[string]any, key, modality, source string) {
	if on, ok := boolAt(m, key); ok && on && caps.AddInputModality(modality, source) {
		traceFact(ctx, source, true, apispec.CapFieldModalitiesInput+" "+modality)
	}
}

// https://platform.claude.com/docs/en/api/models/retrieve — x-api-key and
// anthropic-version, never a bearer.
func probeAnthropic(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as anthropic by hostname; asking /v1/models/{id} for the model's limits and capabilities")
	id := vendorModelID(ctx, cfg)
	if id == "" {
		return
	}
	headers := map[string]string{"anthropic-version": anthropicVersion}
	if cfg.ApiKey != "" {
		headers["x-api-key"] = cfg.ApiKey
	}
	body := d.vendorObject(ctx, vendorOrigin(cfg.ApiURL)+"/v1/models/"+id, headers)
	if body == nil {
		return
	}
	setContextAt(ctx, &ev.Caps, body, "max_input_tokens", anthropicModelSource+".max_input_tokens")
	setMaxOutputAt(ctx, &ev.Caps, body, "max_tokens", anthropicModelSource+".max_tokens")

	caps := objectAt(body, "capabilities")
	if caps == nil {
		return
	}
	if supportedAt(caps, "image_input") && ev.Caps.AddInputModality("image", anthropicModelSource+".capabilities.image_input") {
		traceFact(ctx, anthropicModelSource+".capabilities.image_input", true, apispec.CapFieldModalitiesInput+" image")
	}
	if supportedAt(caps, "pdf_input") && ev.Caps.AddInputModality("pdf", anthropicModelSource+".capabilities.pdf_input") {
		traceFact(ctx, anthropicModelSource+".capabilities.pdf_input", true, apispec.CapFieldModalitiesInput+" pdf")
	}

	// Composition treats present facts as decisive, so a capabilities object
	// without the thinking leaf must stay "not stated" rather than become
	// "thinking unsupported" and silence the name-based fallback.
	thinking := objectAt(caps, "thinking")
	if thinking == nil {
		tracef(ctx, "%s.capabilities has no thinking object; anthropic facts not stated", anthropicModelSource)
		return
	}
	facts := &apispec.AnthropicFacts{Source: anthropicModelSource + ".capabilities", Structured: supportedAt(caps, "structured_outputs")}
	facts.ThinkingSupported, _ = boolAt(thinking, "supported")
	types := objectAt(thinking, "types")
	for _, kind := range []string{"adaptive", "enabled"} {
		if supportedAt(types, kind) {
			facts.ThinkingTypes = append(facts.ThinkingTypes, kind)
		}
	}
	effort := objectAt(caps, "effort")
	for _, level := range anthropicEffortLevels {
		if supportedAt(effort, level) {
			facts.EffortLevels = append(facts.EffortLevels, level)
		}
	}
	ev.Anthropic = facts
	tracef(ctx, "%s.capabilities: thinking.supported=%t thinking.types=%v effort=%v structured_outputs=%t -> anthropic facts",
		anthropicModelSource, facts.ThinkingSupported, facts.ThinkingTypes, facts.EffortLevels, facts.Structured)
}

// https://ai.google.dev/api/models#method:-models.get — x-goog-api-key. The
// resource name is models/{model}; a configured "models/" prefix is the same
// model.
func probeGemini(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as gemini by hostname; asking /v1beta/models/{id} for the model's token limits")
	id := vendorModelID(ctx, config.ModelConfig{ModelName: strings.TrimPrefix(cfg.ModelName, "models/")})
	if id == "" {
		return
	}
	var headers map[string]string
	if cfg.ApiKey != "" {
		headers = map[string]string{"x-goog-api-key": cfg.ApiKey}
	}
	body := d.vendorObject(ctx, vendorOrigin(cfg.ApiURL)+"/v1beta/models/"+id, headers)
	if body == nil {
		return
	}
	setContextAt(ctx, &ev.Caps, body, "inputTokenLimit", geminiModelSource+".inputTokenLimit")
	setMaxOutputAt(ctx, &ev.Caps, body, "outputTokenLimit", geminiModelSource+".outputTokenLimit")
	if thinking, ok := boolAt(body, "thinking"); ok && thinking {
		ev.Caps.SetReasoning(true, geminiModelSource+".thinking")
		traceFact(ctx, geminiModelSource+".thinking", true, apispec.CapFieldReasoning)
	}
}

// https://docs.mistral.ai/api/#tag/models/operation/retrieve_model_v1_models__model_id__get
func probeMistral(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as mistral by hostname; asking /v1/models/{id} for the model's context and capabilities")
	id := vendorModelID(ctx, cfg)
	if id == "" {
		return
	}
	body := d.vendorObject(ctx, vendorOrigin(cfg.ApiURL)+"/v1/models/"+id, bearerHeader(cfg.ApiKey))
	if body == nil {
		return
	}
	setContextAt(ctx, &ev.Caps, body, "max_context_length", mistralModelSource+".max_context_length")
	caps := objectAt(body, "capabilities")
	if caps == nil {
		return
	}
	if tools, ok := boolAt(caps, "function_calling"); ok {
		ev.Caps.SetTools(tools, mistralModelSource+".capabilities.function_calling")
		traceFact(ctx, mistralModelSource+".capabilities.function_calling", tools, apispec.CapFieldTools)
	}
	if reasoning, ok := boolAt(caps, "reasoning"); ok {
		ev.Caps.SetReasoning(reasoning, mistralModelSource+".capabilities.reasoning")
		traceFact(ctx, mistralModelSource+".capabilities.reasoning", reasoning, apispec.CapFieldReasoning)
	}
	addInputModalityWhen(ctx, &ev.Caps, caps, "vision", "image", mistralModelSource+".capabilities.vision")
	addInputModalityWhen(ctx, &ev.Caps, caps, "audio", "audio", mistralModelSource+".capabilities.audio")
}

// https://console.groq.com/docs/api-reference#models-retrieve — the API is
// rooted at /openai/v1.
func probeGroq(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as groq by hostname; asking /openai/v1/models/{id} for the model's window")
	id := vendorModelID(ctx, cfg)
	if id == "" {
		return
	}
	body := d.vendorObject(ctx, vendorOrigin(cfg.ApiURL)+"/openai/v1/models/"+id, bearerHeader(cfg.ApiKey))
	if body == nil {
		return
	}
	setContextAt(ctx, &ev.Caps, body, "context_window", groqModelSource+".context_window")
	setMaxOutputAt(ctx, &ev.Caps, body, "max_completion_tokens", groqModelSource+".max_completion_tokens")
}

// https://docs.x.ai/developers/rest-api-reference/inference/models — no
// context field is exposed.
func probeXAI(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as xai by hostname; asking /v1/language-models/{id} for the model's modalities and efforts")
	id := vendorModelID(ctx, cfg)
	if id == "" {
		return
	}
	body := d.vendorObject(ctx, vendorOrigin(cfg.ApiURL)+"/v1/language-models/"+id, bearerHeader(cfg.ApiKey))
	if body == nil {
		return
	}
	if in := stringsAt(body, "input_modalities"); ev.Caps.SetInputModalities(in, xaiModelSource+".input_modalities") {
		traceFact(ctx, xaiModelSource+".input_modalities", in, apispec.CapFieldModalitiesInput)
	}
	if out := stringsAt(body, "output_modalities"); ev.Caps.SetOutputModalities(out, xaiModelSource+".output_modalities") {
		traceFact(ctx, xaiModelSource+".output_modalities", out, apispec.CapFieldModalitiesOutput)
	}
	caps := objectAt(body, "capabilities")
	if caps == nil {
		return
	}
	if efforts := stringsAt(caps, "reasoning_effort"); len(efforts) > 0 {
		ev.ReasoningEfforts = efforts
		ev.EffortsSource = xaiModelSource + ".capabilities.reasoning_effort"
		tracef(ctx, "%s = %v -> reasoning efforts", ev.EffortsSource, efforts)
	}
	if level, ok := caps["default_reasoning_effort"].(string); ok && level != "" {
		ev.ThinkingDefault = level
		ev.ThinkingDefaultSource = xaiModelSource + ".capabilities.default_reasoning_effort"
		tracef(ctx, "%s = %s -> thinking.default", ev.ThinkingDefaultSource, level)
	}
}

// https://docs.together.ai/reference/models-1 — a bare array.
func probeTogether(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as together by hostname; reading the model's entry in /v1/models")
	list := d.vendorList(ctx, vendorOrigin(cfg.ApiURL)+"/v1/models", bearerHeader(cfg.ApiKey))
	entry := vendorEntry(ctx, list, "/v1/models", cfg.ModelName)
	if entry == nil {
		return
	}
	setContextAt(ctx, &ev.Caps, entry, "context_length", togetherListSource+".context_length")
}

// https://inference-docs.cerebras.ai/api-reference/models/public-models —
// public, so the key stays home.
func probeCerebras(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as cerebras by hostname; asking the public /public/v1/models/{id} without the key")
	id := vendorModelID(ctx, cfg)
	if id == "" {
		return
	}
	body := d.vendorObject(ctx, vendorOrigin(cfg.ApiURL)+"/public/v1/models/"+id, nil)
	if body == nil {
		return
	}
	if limits := objectAt(body, "limits"); limits != nil {
		setContextAt(ctx, &ev.Caps, limits, "max_context_length", cerebrasModelSource+".limits.max_context_length")
		setMaxOutputAt(ctx, &ev.Caps, limits, "max_completion_tokens", cerebrasModelSource+".limits.max_completion_tokens")
	}
	if quant, ok := body["quantization"].(string); ok && ev.Caps.SetQuantization(quant, cerebrasModelSource+".quantization") {
		traceFact(ctx, cerebrasModelSource+".quantization", quant, apispec.CapFieldQuantization)
	}
	caps := objectAt(body, "capabilities")
	if caps == nil {
		return
	}
	if tools, ok := boolAt(caps, "function_calling"); ok {
		ev.Caps.SetTools(tools, cerebrasModelSource+".capabilities.function_calling")
		traceFact(ctx, cerebrasModelSource+".capabilities.function_calling", tools, apispec.CapFieldTools)
	}
	if parallel, ok := boolAt(caps, "parallel_tool_calls"); ok {
		ev.Caps.SetToolsParallel(parallel, cerebrasModelSource+".capabilities.parallel_tool_calls")
		traceFact(ctx, cerebrasModelSource+".capabilities.parallel_tool_calls", parallel, apispec.CapFieldToolsParallel)
	}
	addInputModalityWhen(ctx, &ev.Caps, caps, "vision", "image", cerebrasModelSource+".capabilities.vision")
	var structured []string
	source := cerebrasModelSource + ".capabilities.json_mode"
	if on, ok := boolAt(caps, "json_mode"); ok && on {
		structured = append(structured, "json_object")
	}
	if on, ok := boolAt(caps, "structured_outputs"); ok && on {
		structured = append(structured, "json_schema")
		source = cerebrasModelSource + ".capabilities.structured_outputs"
	}
	if ev.Caps.SetStructured(structured, source) {
		traceFact(ctx, source, structured, apispec.CapFieldStructured)
	}
	if reasoning, ok := boolAt(caps, "reasoning"); ok {
		ev.Caps.SetReasoning(reasoning, cerebrasModelSource+".capabilities.reasoning")
		traceFact(ctx, cerebrasModelSource+".capabilities.reasoning", reasoning, apispec.CapFieldReasoning)
	}
}

// Undocumented but open listing (live 2026-09-25); read without the key,
// best-effort.
func probeDeepInfra(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as deepinfra by hostname; reading the model's entry in the open /v1/openai/models without the key")
	list := d.vendorList(ctx, vendorOrigin(cfg.ApiURL)+"/v1/openai/models", nil)
	entry := vendorEntry(ctx, list, "/v1/openai/models", cfg.ModelName)
	if entry == nil {
		return
	}
	meta := objectAt(entry, "metadata")
	if meta == nil {
		return
	}
	setContextAt(ctx, &ev.Caps, meta, "context_length", deepinfraListSource+".metadata.context_length")
	setMaxOutputAt(ctx, &ev.Caps, meta, "max_tokens", deepinfraListSource+".metadata.max_tokens")
}

// https://huggingface.co/docs/inference-providers/index — the router's open
// listing; one entry aggregates several providers, so the facts are the
// best any live provider offers.
func probeHFRouter(ctx context.Context, d *Detector, _ []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as hf-router by hostname; reading the model's entry in the open /v1/models without the key")
	list := d.vendorList(ctx, vendorOrigin(cfg.ApiURL)+"/v1/models", nil)
	entry := vendorEntry(ctx, list, "/v1/models", cfg.ModelName)
	if entry == nil {
		return
	}
	if arch := objectAt(entry, "architecture"); arch != nil {
		if in := stringsAt(arch, "input_modalities"); ev.Caps.SetInputModalities(in, hfRouterListSource+".architecture.input_modalities") {
			traceFact(ctx, hfRouterListSource+".architecture.input_modalities", in, apispec.CapFieldModalitiesInput)
		}
		if out := stringsAt(arch, "output_modalities"); ev.Caps.SetOutputModalities(out, hfRouterListSource+".architecture.output_modalities") {
			traceFact(ctx, hfRouterListSource+".architecture.output_modalities", out, apispec.CapFieldModalitiesOutput)
		}
	}

	providers, _ := entry["providers"].([]any)
	var (
		maxContext            int64
		live                  int
		toolsSeen, tools      bool
		structSeen, structOut bool
	)
	for _, item := range providers {
		p, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if status, _ := p["status"].(string); status != "live" && status != "ok" {
			continue
		}
		live++
		if n, ok := integerAt(p, "context_length"); ok && n > maxContext {
			maxContext = n
		}
		if b, ok := boolAt(p, "supports_tools"); ok {
			toolsSeen = true
			tools = tools || b
		}
		if b, ok := boolAt(p, "supports_structured_output"); ok {
			structSeen = true
			structOut = structOut || b
		}
	}
	if len(providers) > 0 && live == 0 {
		tracef(ctx, "/v1/models providers[]: none of %d is live; no provider facts", len(providers))
		return
	}
	if maxContext > 0 && ev.Caps.SetContext(maxContext, hfRouterListSource+".providers.context_length") {
		traceFact(ctx, hfRouterListSource+".providers.context_length", maxContext, apispec.CapFieldContext+" (largest live provider)")
	}
	if toolsSeen {
		ev.Caps.SetTools(tools, hfRouterListSource+".providers.supports_tools")
		traceFact(ctx, hfRouterListSource+".providers.supports_tools", tools, apispec.CapFieldTools+" (any live provider)")
	}
	if structSeen && structOut && ev.Caps.SetStructured([]string{"json_schema"}, hfRouterListSource+".providers.supports_structured_output") {
		traceFact(ctx, hfRouterListSource+".providers.supports_structured_output", true, apispec.CapFieldStructured+" json_schema")
	}
}
