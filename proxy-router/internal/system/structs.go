package system

type FD struct {
	ID   string
	Path string
}

type SetEthNodeURLReq struct {
	URLs []string `json:"urls" binding:"required" validate:"required,url"`
}

type ConfigResponse struct {
	Version       string
	Commit        string
	DerivedConfig interface{}
	Config        interface{}
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
	// Api is the provider-declared API spec for this model (see internal/apispec): serving stack preset, model family and the request-param bindings a consumer needs to translate canonical fields. Derived from models-config presets only — never the backend URL, key or private model string.
	Api *ModelApiSpec `json:"api,omitempty"`
}

const (
	ThinkingModeAlwaysOn = "always_on"
	// ThinkingModeControllable means thinking can be switched: a
	// reasoning.disable or reasoning.enable binding exists.
	ThinkingModeControllable = "controllable"
	// ThinkingModeTunable means thinking intensity can be adjusted but never
	// fully disabled (effort/budget bindings only, e.g. gpt-oss).
	ThinkingModeTunable = "tunable"
)

// Binding kinds: the mechanism a ParamBinding uses.
const (
	// BindingKindBodyParam is a request body param addressed by dotted path
	// on the OpenAI-compatible chat endpoint (e.g. "reasoning_effort",
	// "venice_parameters.disable_thinking").
	BindingKindBodyParam = "body_param"
	// BindingKindTemplateKwarg is a chat-template variable passed via
	// chat_template_kwargs (vLLM / SGLang / llama.cpp); Param starts with
	// "chat_template_kwargs.".
	BindingKindTemplateKwarg = "template_kwarg"
	// BindingKindSystemPrompt means the intent is realized by magic text in
	// the system prompt (Hint carries the text).
	BindingKindSystemPrompt = "system_prompt"
	// BindingKindNativeBodyParam exists only on the stack's native
	// (non-OpenAI-compatible) endpoint; informational, Hint names the endpoint.
	BindingKindNativeBodyParam = "native_body_param"
)

// Canonical request intents bindings are keyed by. Open-ended: new groups
// join without schema changes.
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

// ModelApiSpec describes the API serving a model as declared by the
// provider's models-config presets. Self-reported and unverified: consumers
// treat it as a routing/translation hint, not truth.
type ModelApiSpec struct {
	// Stack is the serving stack / vendor preset: vllm | sglang | llamacpp |
	// ollama | venice | openrouter | litellm | anthropic | openai.
	Stack string `json:"stack,omitempty"`
	// ModelFamily is the canonical model family (qwen3, deepseek-r1, claude…).
	ModelFamily string `json:"modelFamily,omitempty"`
	// Thinking summarizes reasoning controllability; knobs live in Bindings.
	Thinking *ThinkingSpec `json:"thinking,omitempty"`
	// Bindings maps canonical request intents to how this backend spells them.
	// An absent intent means the backend has no way to express it.
	Bindings map[string]*ParamBinding `json:"bindings,omitempty"`
	// Parameters lists the request parameters the stack documents as
	// accepted, in the stack's own request vocabulary — OpenAI
	// chat-completions names for every stack except anthropic, which uses
	// Anthropic Messages names (an upper bound: server-side flags are
	// invisible).
	Parameters []string `json:"parameters,omitempty"`
	DeclaredAt int64    `json:"declaredAt,omitempty"`
}

// ThinkingSpec summarizes whether reasoning output can be controlled.
type ThinkingSpec struct {
	Mode string `json:"mode"` // always_on | controllable | tunable
}

// ParamBinding says how one canonical intent maps onto this backend.
type ParamBinding struct {
	Kind      string `json:"kind"`
	Param     string `json:"param,omitempty"`     // dotted path within the mechanism
	ParamType string `json:"paramType,omitempty"` // boolean | number | enum | string | object | array
	// Value is the fixed value realizing the intent (e.g. false for
	// reasoning.disable via enable_thinking). Nil for caller-supplied values.
	Value      any      `json:"value,omitempty"`
	EnumValues []string `json:"enumValues,omitempty"`
	Hint       string   `json:"hint,omitempty"`
}

type StatusRes struct {
	Status string `json:"status"`
}

func OkRes() StatusRes {
	return StatusRes{Status: "ok"}
}
