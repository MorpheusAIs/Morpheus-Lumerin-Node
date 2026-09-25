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
)

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
}

type ThinkingSpec struct {
	Mode string `json:"mode"` // always_on (cannot be disabled) | controllable (reasoning.disable/enable bindings) | tunable (effort/budget bindings only)
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
	return &c
}

func (b *ParamBinding) Clone() *ParamBinding {
	if b == nil {
		return nil
	}
	c := *b
	if b.EnumValues != nil {
		c.EnumValues = append([]string(nil), b.EnumValues...)
	}
	if m, ok := b.Value.(map[string]any); ok {
		mc := make(map[string]any, len(m))
		for k, v := range m {
			mc[k] = v
		}
		c.Value = mc
	}
	return &c
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
