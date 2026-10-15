package system

type FD struct {
	ID   string
	Path string
}

type SetEthNodeURLReq struct {
	URLs []string `json:"urls" binding:"required" validate:"required,url"`
}

type ConfigResponse struct {
	Version             string
	Commit              string
	DerivedConfig       interface{}
	Config              interface{}
	GatewayCapabilities []string
}

type HealthCheckResponse struct {
	Status     string              `json:"status"`
	Version    string              `json:"version"`
	Uptime     string              `json:"uptime"`
	Components map[string]string   `json:"components,omitempty"`
	Models     []ModelHealthReport `json:"models,omitempty"`
}

const (
	ModelHealthStatusHealthy   = "healthy"
	ModelHealthStatusUnhealthy = "unhealthy"
	ModelHealthStatusNoBid     = "no_bid"
	ModelHealthStatusNoModel   = "no_model_configured"
	ModelHealthStatusSkipped   = "skipped"
	// ModelHealthStatusTeeUnverified marks a TEE-tagged model whose backend
	// TEE self-attestation failed (or never succeeded) on this provider.
	// Session requests for the model would be rejected, so it is reported
	// separately from a plain backend probe failure.
	ModelHealthStatusTeeUnverified = "tee_unverified"
	// ModelHealthStatusDegraded marks a model that is serving but at reduced
	// capacity: emitted when a health probe fails with HTTP 429 (upstream
	// throttled). It signals capacity risk, not failure: consumers treat it
	// as serviceable under the permissive and preferred health policies;
	// only the strict policy excludes it.
	ModelHealthStatusDegraded = "degraded"

	ModelHealthErrorTimeout     = "timeout"
	ModelHealthErrorConnection  = "connection"
	ModelHealthErrorBadResponse = "bad_response"
	ModelHealthErrorRateLimited = "rate_limited"
	ModelHealthErrorTypeLookup  = "model_type_lookup"
	// ModelHealthErrorSessionErrors marks a model flipped to unhealthy — or
	// to degraded, when the whole failure streak was upstream rate limiting
	// (429) — by consecutive real-session prompt failures rather than a
	// scheduled probe.
	ModelHealthErrorSessionErrors = "session_errors"
	// ModelHealthErrorTeeAttestation accompanies ModelHealthStatusTeeUnverified.
	ModelHealthErrorTeeAttestation = "tee_attestation"
)

// ModelHealthReport is the sanitized per-model self-report exposed on the
// public /healthcheck endpoint. It covers the union of configured models and
// this provider's active bids, and intentionally carries only on-chain data
// and derived status — never the private modelName, apiUrl or apiKey from
// models-config.json. The ModelName here is the public name registered
// on-chain for the model ID, not the private backend model string.
type ModelHealthReport struct {
	ModelID       string `json:"modelId"`
	ModelName     string `json:"modelName,omitempty"`
	ModelType     string `json:"modelType,omitempty"`
	HasActiveBid  bool   `json:"hasActiveBid"`
	BidID         string `json:"bidId,omitempty"`
	Status        string `json:"status"`
	LastHealthy   int64  `json:"lastHealthy,omitempty"`
	LastChecked   int64  `json:"lastChecked"`
	LatencyMs     int64  `json:"latencyMs,omitempty"`
	PromptCorrect *bool  `json:"promptCorrect,omitempty"`
	ErrorKind     string `json:"errorKind,omitempty"`
	// HttpStatus is the HTTP status code returned by the upstream backend
	// when the probe failed with a non-200 response (e.g. 402, 429).
	// Zero when the probe succeeded or never got an HTTP response.
	HttpStatus int `json:"httpStatus,omitempty"`
	// Api describes the API serving this model: stack, model family, thinking
	// control and request-param bindings. Never carries the backend URL, key or
	// hostname. Absent when nothing is known.
	Api *ModelApiSpec `json:"api,omitempty"`
}

const (
	ThinkingModeAlwaysOn     = "always_on"
	ThinkingModeControllable = "controllable"
	ThinkingModeTunable      = "tunable"
)

const (
	BindingKindBodyParam = "body_param"
	// Param starts with "chat_template_kwargs.".
	BindingKindTemplateKwarg = "template_kwarg"
	// Param is empty; Hint carries the text.
	BindingKindSystemPrompt = "system_prompt"
	// Informational only, unreachable on the OpenAI-compatible endpoint; Hint
	// names the endpoint.
	BindingKindNativeBodyParam = "native_body_param"
	// Param is "tools"; Value is a tool object the consumer appends to the
	// request's tools[] (Anthropic {"type":"web_search_20250305","name":"web_search"},
	// Gemini {"googleSearch":{}}).
	BindingKindToolEntry = "tool_entry"
)

const (
	IntentReasoningDisable = "reasoning.disable"
	IntentReasoningEnable  = "reasoning.enable"
	IntentReasoningEffort  = "reasoning.effort"
	IntentReasoningBudget  = "reasoning.budget"
	IntentReasoningFormat  = "reasoning.format"

	IntentResponseFormatJSON    = "response_format.json"
	IntentResponseFormatSchema  = "response_format.schema"
	IntentResponseFormatGrammar = "response_format.grammar"
	IntentResponseFormatChoice  = "response_format.choice"

	IntentSamplingTopK              = "sampling.top_k"
	IntentSamplingMinP              = "sampling.min_p"
	IntentSamplingRepetitionPenalty = "sampling.repetition_penalty"
	IntentSamplingTypicalP          = "sampling.typical_p"
	IntentSamplingTopA              = "sampling.top_a"

	IntentContextNumCtx      = "context.num_ctx"
	IntentStreamIncludeUsage = "stream.include_usage"
	IntentCachePrompt        = "cache.prompt"
	IntentToolsParallel      = "tools.parallel"

	IntentWebSearchEnable    = "web_search.enable"
	IntentWebSearchCitations = "web_search.citations"
)

// Evidence keys are wire field paths. Sources are "<stack>:<endpoint>.<field>"
// (vllm:/v1/models.max_model_len), "declared", "registry:<name>@<YYYY-MM-DD>",
// "derived:<what>" (the composer inferred it from the thinking mode or the
// surviving reasoning.* bindings; cleared together with detected facts),
// "health-probe:<what>" or "verify:<what>@<unix>".
const EvidenceKeyThinkingDefault = "thinking.default"

func EvidenceKeyCapability(field string) string { return "capabilities." + field }

func EvidenceKeyBinding(intent string) string { return "bindings." + intent }

const (
	ApiSpecSourceDeclared = "declared"
	ApiSpecSourceDetected = "detected"
)

type ModelApiSpec struct {
	// Stack is the serving stack or vendor: vllm | sglang | llamacpp | ollama |
	// venice | openrouter | litellm | anthropic | openai, or a detected engine or
	// hosted vendor without a bindings table (tgi, lmstudio, koboldcpp, groq,
	// together, ...). With Via set, Stack is the upstream behind the gateway and
	// Bindings use that upstream's vocabulary.
	Stack string `json:"stack,omitempty"`
	// Via is the gateway in front of Stack when detection identified the
	// upstream behind it; today only "litellm". Absent when the backend is
	// reached directly or the upstream is unknown (then Stack is the gateway).
	Via string `json:"via,omitempty"`
	// ModelFamily is the canonical model family (qwen3, deepseek-r1, claude, ...).
	ModelFamily string `json:"modelFamily,omitempty"`
	// Thinking summarizes reasoning controllability; the knobs are in Bindings.
	Thinking *ThinkingSpec `json:"thinking,omitempty"`
	// Bindings maps canonical request intents to how this backend expresses
	// them; an absent intent cannot be expressed.
	Bindings map[string]*ParamBinding `json:"bindings,omitempty"`
	// Parameters lists the request parameters the stack documents as accepted,
	// in its own vocabulary: OpenAI chat-completions names, or Anthropic
	// Messages names for anthropic. Behind LiteLLM, narrowed to what LiteLLM
	// supports for the model.
	Parameters []string `json:"parameters,omitempty"`
	// Source is declared (models-config apiStack preset) or detected (runtime
	// probing of the backend's service endpoints).
	Source     string `json:"source,omitempty"`
	DeclaredAt int64  `json:"declaredAt,omitempty"`
	// Capabilities are backend-verified or declared facts about the model as
	// served; an absent field is unknown, never false. Never derived from the
	// stack tables or the family name.
	Capabilities *ModelCapabilities `json:"capabilities,omitempty"`
	// Evidence maps a wire field path (capabilities.context, thinking.default,
	// bindings.reasoning.disable, ...) to the source it was set from: a
	// backend endpoint, declared, registry, derived, health-probe or verify
	// (see EvidenceKeyThinkingDefault).
	Evidence map[string]string `json:"evidence,omitempty"`
}

type ThinkingSpec struct {
	// always_on: cannot be disabled (a reasoning.disable binding may remain only with verified: false, the verifier's refutation) | controllable: reasoning.disable/enable bindings | tunable: effort/budget bindings only
	Mode string `json:"mode"`
	// Default is "on", "off" or an effort level name; absent when unknown.
	Default string `json:"default,omitempty"`
}

// Pointers so that an explicit false or a reported value survives omitempty;
// 0 is never a valid capability value (setters reject it).
type ModelCapabilities struct {
	// Context is the served window (prompt + completion tokens): the number to
	// plan requests against.
	Context *int64 `json:"context,omitempty"`
	// ContextTrained is the model's trained window when the backend reports
	// both.
	ContextTrained *int64      `json:"contextTrained,omitempty"`
	MaxOutput      *int64      `json:"maxOutput,omitempty"`
	Modalities     *Modalities `json:"modalities,omitempty"`
	Tools          *bool       `json:"tools,omitempty"`
	ToolsParallel  *bool       `json:"toolsParallel,omitempty"`
	// Structured is a subset of json_object, json_schema, grammar, regex, choice.
	Structured []string `json:"structured,omitempty"`
	// Reasoning says the model can reason; the control is in Thinking and
	// Bindings.
	Reasoning *bool `json:"reasoning,omitempty"`
	// WebSearch says the backend can search the web for this model on request;
	// the knob is a web_search.* binding.
	WebSearch   *bool `json:"webSearch,omitempty"`
	Embeddings  *bool `json:"embeddings,omitempty"`
	PromptCache *bool `json:"promptCache,omitempty"`
	// Quantization is the backend's own label (fp8, Q4_K_M, bf16, ...).
	Quantization string `json:"quantization,omitempty"`
}

type Modalities struct {
	Input  []string `json:"input,omitempty"`  // text, image, audio, video, pdf
	Output []string `json:"output,omitempty"` // text, image, audio, embeddings
}

type ParamBinding struct {
	Kind      string `json:"kind"`
	Param     string `json:"param,omitempty"`     // dotted path within the mechanism
	ParamType string `json:"paramType,omitempty"` // boolean | number | enum | string | object | array
	// Value is the fixed value realizing the intent; nil when the caller
	// supplies it.
	Value      any      `json:"value,omitempty"`
	EnumValues []string `json:"enumValues,omitempty"`
	Hint       string   `json:"hint,omitempty"`
	// Verified is set by the runtime verifier; absent means never checked.
	Verified   *bool `json:"verified,omitempty"`
	VerifiedAt int64 `json:"verifiedAt,omitempty"`
}

func (s *ModelApiSpec) Clone() *ModelApiSpec {
	if s == nil {
		return nil
	}
	c := *s
	if s.Thinking != nil {
		t := *s.Thinking
		c.Thinking = &t
	}
	if s.Bindings != nil {
		c.Bindings = make(map[string]*ParamBinding, len(s.Bindings))
		for intent, b := range s.Bindings {
			c.Bindings[intent] = b.Clone()
		}
	}
	if s.Parameters != nil {
		c.Parameters = append([]string(nil), s.Parameters...)
	}
	c.Capabilities = s.Capabilities.Clone()
	if s.Evidence != nil {
		c.Evidence = make(map[string]string, len(s.Evidence))
		for k, v := range s.Evidence {
			c.Evidence[k] = v
		}
	}
	return &c
}

func (m *ModelCapabilities) Clone() *ModelCapabilities {
	if m == nil {
		return nil
	}
	c := *m
	c.Context = cloneInt64(m.Context)
	c.ContextTrained = cloneInt64(m.ContextTrained)
	c.MaxOutput = cloneInt64(m.MaxOutput)
	c.Tools = cloneBool(m.Tools)
	c.ToolsParallel = cloneBool(m.ToolsParallel)
	c.Reasoning = cloneBool(m.Reasoning)
	c.WebSearch = cloneBool(m.WebSearch)
	c.Embeddings = cloneBool(m.Embeddings)
	c.PromptCache = cloneBool(m.PromptCache)
	if m.Structured != nil {
		c.Structured = append([]string(nil), m.Structured...)
	}
	if m.Modalities != nil {
		mod := Modalities{}
		if m.Modalities.Input != nil {
			mod.Input = append([]string(nil), m.Modalities.Input...)
		}
		if m.Modalities.Output != nil {
			mod.Output = append([]string(nil), m.Modalities.Output...)
		}
		c.Modalities = &mod
	}
	return &c
}

func cloneInt64(p *int64) *int64 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func cloneBool(p *bool) *bool {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func (b *ParamBinding) Clone() *ParamBinding {
	if b == nil {
		return nil
	}
	c := *b
	if b.EnumValues != nil {
		c.EnumValues = append([]string(nil), b.EnumValues...)
	}
	c.Value = cloneValue(b.Value)
	c.Verified = cloneBool(b.Verified)
	return &c
}

// cloneValue copies the JSON-shaped containers a binding value may hold
// (objects and arrays, nested) so a clone never aliases the preset table it
// came from: the openrouter web_search.enable value is [{"id": "web"}], and
// a consumer editing its copy must not edit the table. Scalars are returned
// as they are.
func cloneValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, item := range t {
			out[k] = cloneValue(item)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = cloneValue(item)
		}
		return out
	}
	return v
}

type StatusRes struct {
	Status string `json:"status"`
}

func OkRes() StatusRes {
	return StatusRes{Status: "ok"}
}

// ModelsReloadRes is the response of POST /config/models/reload.
type ModelsReloadRes struct {
	Added             []string `json:"added"`
	Removed           []string `json:"removed"`
	HealthSweepQueued bool     `json:"healthSweepQueued"`
}

// ErrorResponse is a swagger-friendly error body for system endpoints.
type ErrorResponse struct {
	Error string `json:"error"`
}
