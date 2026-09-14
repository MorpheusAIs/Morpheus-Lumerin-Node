package apispec

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// Per-serving-stack request-param remap tables: what each stack documents
// beyond the standard OpenAI chat-completions surface, keyed by canonical
// intent. These are static tables, not introspected from a live backend;
// every entry cites the official documentation it was verified against in
// the doc-URL comment above its table. Keep entries literal to the docs — a
// wrong binding is worse than a missing one, since consumers translate
// requests with it.
//
// Reasoning intents are deliberately sparse here: whether a model can think,
// and via which template kwarg, is a model property handled by the family
// and template layers. Stacks only contribute model-independent reasoning
// knobs (output format, token budget, gateway-normalized effort), and those
// merge only when reasoning evidence already exists for the model.

const (
	bodyParam   = system.BindingKindBodyParam
	nativeParam = system.BindingKindNativeBodyParam
)

// bindingSet is the generic remap table under construction: canonical
// intent -> how this backend realizes it.
type bindingSet = map[string]*system.ParamBinding

var jsonObjectFormat = map[string]any{"type": "json_object"}

var stackBindings = map[string]bindingSet{
	// vLLM OpenAI-compatible server.
	// https://docs.vllm.ai/en/latest/features/reasoning_outputs/
	// https://docs.vllm.ai/en/latest/features/structured_outputs/
	// https://docs.vllm.ai/en/latest/api/vllm/sampling_params/
	// https://docs.vllm.ai/en/v0.18.0/serving/openai_compatible_server/
	"vllm": {
		system.IntentReasoningBudget:           {Kind: bodyParam, Param: "thinking_token_budget", ParamType: "number", Hint: "server must run with --reasoning-parser"},
		system.IntentReasoningFormat:           {Kind: bodyParam, Param: "include_reasoning", ParamType: "boolean", Hint: "false omits parsed reasoning from the response (tokens are still generated)"},
		system.IntentResponseFormatJSON:        {Kind: bodyParam, Param: "response_format", ParamType: "object", Value: jsonObjectFormat, Hint: "not combinable with tools"},
		system.IntentResponseFormatSchema:      {Kind: bodyParam, Param: "response_format", ParamType: "object", Hint: `{"type":"json_schema","json_schema":{"name":...,"schema":{...},"strict":bool}}; not combinable with tools`},
		system.IntentResponseFormatGrammar:     {Kind: bodyParam, Param: "structured_outputs.grammar", ParamType: "string", Hint: "vLLM >= 0.12 (older builds: guided_grammar / guided_choice); EBNF/Lark grammar; structured_outputs.regex for a regex; one constraint kind per request"},
		system.IntentResponseFormatChoice:      {Kind: bodyParam, Param: "structured_outputs.choice", ParamType: "array", Hint: "vLLM >= 0.12 (older builds: guided_grammar / guided_choice); array of allowed strings"},
		system.IntentSamplingTopK:              {Kind: bodyParam, Param: "top_k", ParamType: "number", Hint: "0 or -1 = all tokens"},
		system.IntentSamplingMinP:              {Kind: bodyParam, Param: "min_p", ParamType: "number", Hint: "[0,1]"},
		system.IntentSamplingRepetitionPenalty: {Kind: bodyParam, Param: "repetition_penalty", ParamType: "number", Hint: ">0, 1.0 = off"},
		system.IntentStreamIncludeUsage:        {Kind: bodyParam, Param: "stream_options.include_usage", ParamType: "boolean", Value: true},
		system.IntentToolsParallel:             {Kind: bodyParam, Param: "parallel_tool_calls", ParamType: "boolean", Hint: "tool_choice auto also needs --enable-auto-tool-choice --tool-call-parser"},
	},

	// SGLang OpenAI-compatible server.
	// https://docs.sglang.io/basic_usage/openai_api_completions.html
	// https://docs.sglang.io/advanced_features/separate_reasoning.html
	// https://docs.sglang.io/advanced_features/structured_outputs.html
	// https://docs.sglang.io/basic_usage/sampling_params.html
	"sglang": {
		system.IntentReasoningFormat:           {Kind: bodyParam, Param: "separate_reasoning", ParamType: "boolean", Hint: "true (default) splits CoT into message.reasoning_content; needs --reasoning-parser"},
		system.IntentResponseFormatJSON:        {Kind: bodyParam, Param: "response_format", ParamType: "object", Value: jsonObjectFormat, Hint: "enforced as json_schema {type: object}"},
		system.IntentResponseFormatSchema:      {Kind: bodyParam, Param: "response_format", ParamType: "object", Hint: `{"type":"json_schema","json_schema":{"name":...,"schema":{...}}}; exclusive with regex/ebnf`},
		system.IntentResponseFormatGrammar:     {Kind: bodyParam, Param: "ebnf", ParamType: "string", Hint: "EBNF (xgrammar/llguidance backends); `regex` for a regex; exclusive with json_schema"},
		system.IntentSamplingTopK:              {Kind: bodyParam, Param: "top_k", ParamType: "number", Hint: "-1 disables"},
		system.IntentSamplingMinP:              {Kind: bodyParam, Param: "min_p", ParamType: "number", Hint: "0.0 disables"},
		system.IntentSamplingRepetitionPenalty: {Kind: bodyParam, Param: "repetition_penalty", ParamType: "number", Hint: "(0,2], 1.0 = off"},
		system.IntentStreamIncludeUsage:        {Kind: bodyParam, Param: "stream_options.include_usage", ParamType: "boolean", Value: true},
		system.IntentToolsParallel:             {Kind: bodyParam, Param: "parallel_tool_calls", ParamType: "boolean", Hint: "shapes the tool_choice grammar constraint; needs --tool-call-parser"},
	},

	// llama.cpp llama-server.
	// https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md
	"llamacpp": {
		system.IntentReasoningBudget:           {Kind: bodyParam, Param: "reasoning_budget", ParamType: "number", Hint: "-1 unrestricted, 0 ends thinking immediately"},
		system.IntentReasoningDisable:          {Kind: bodyParam, Param: "reasoning_budget", ParamType: "number", Value: 0, Hint: "0 = immediate end of thinking"},
		system.IntentReasoningFormat:           {Kind: bodyParam, Param: "reasoning_format", ParamType: "enum", EnumValues: []string{"none", "auto", "deepseek", "deepseek-legacy"}, Hint: "how reasoning is split out of content"},
		system.IntentResponseFormatJSON:        {Kind: bodyParam, Param: "response_format", ParamType: "object", Value: jsonObjectFormat},
		system.IntentResponseFormatSchema:      {Kind: bodyParam, Param: "response_format", ParamType: "object", Hint: `{"type":"json_schema","json_schema":{"schema":{...}}}`},
		system.IntentResponseFormatGrammar:     {Kind: bodyParam, Param: "grammar", ParamType: "string", Hint: "GBNF grammar"},
		system.IntentSamplingTopK:              {Kind: bodyParam, Param: "top_k", ParamType: "number"},
		system.IntentSamplingMinP:              {Kind: bodyParam, Param: "min_p", ParamType: "number"},
		system.IntentSamplingRepetitionPenalty: {Kind: bodyParam, Param: "repeat_penalty", ParamType: "number"},
		system.IntentSamplingTypicalP:          {Kind: bodyParam, Param: "typical_p", ParamType: "number"},
		system.IntentStreamIncludeUsage:        {Kind: bodyParam, Param: "stream_options.include_usage", ParamType: "boolean", Value: true},
		system.IntentCachePrompt:               {Kind: bodyParam, Param: "cache_prompt", ParamType: "boolean", Value: true},
		system.IntentToolsParallel:             {Kind: bodyParam, Param: "parallel_tool_calls", ParamType: "boolean"},
	},

	// Ollama. Only the OpenAI-compatible /v1 surface is reachable through the
	// openai adapter; Ollama-specific knobs live on the native /api endpoints
	// and are reported as native_body_param for completeness.
	// https://docs.ollama.com/api/openai-compatibility
	// https://docs.ollama.com/api
	// https://docs.ollama.com/capabilities/thinking
	"ollama": {
		system.IntentReasoningEffort:           {Kind: bodyParam, Param: "reasoning_effort", ParamType: "enum", EnumValues: []string{"low", "medium", "high", "max", "none"}, Hint: "native /api/chat equivalent: think level; gpt-oss accepts only low|medium|high"},
		system.IntentResponseFormatJSON:        {Kind: bodyParam, Param: "response_format", ParamType: "object", Value: jsonObjectFormat, Hint: "native /api/chat: format: \"json\""},
		system.IntentResponseFormatSchema:      {Kind: bodyParam, Param: "response_format", ParamType: "object", Hint: "type json_schema; native /api/chat: format: <JSON Schema object>"},
		system.IntentSamplingTopK:              {Kind: nativeParam, Param: "options.top_k", ParamType: "number", Hint: "native /api/chat only"},
		system.IntentSamplingMinP:              {Kind: nativeParam, Param: "options.min_p", ParamType: "number", Hint: "native /api/chat only"},
		system.IntentSamplingRepetitionPenalty: {Kind: nativeParam, Param: "options.repeat_penalty", ParamType: "number", Hint: "native /api/chat only"},
		system.IntentContextNumCtx:             {Kind: nativeParam, Param: "options.num_ctx", ParamType: "number", Hint: "native /api/chat only; on /v1 use a Modelfile PARAMETER num_ctx"},
		system.IntentStreamIncludeUsage:        {Kind: bodyParam, Param: "stream_options.include_usage", ParamType: "boolean", Value: true},
	},

	// OpenRouter.
	// https://openrouter.ai/docs/api-reference/parameters
	// https://openrouter.ai/docs/use-cases/reasoning-tokens
	// https://openrouter.ai/docs/features/structured-outputs
	// https://openrouter.ai/docs/features/prompt-caching
	"openrouter": {
		system.IntentReasoningDisable:          {Kind: bodyParam, Param: "reasoning.effort", ParamType: "string", Value: "none", Hint: "models with reasoning.mandatory reject it; reasoning.exclude only hides output"},
		system.IntentReasoningEnable:           {Kind: bodyParam, Param: "reasoning.enabled", ParamType: "boolean", Value: true},
		system.IntentReasoningEffort:           {Kind: bodyParam, Param: "reasoning.effort", ParamType: "enum", EnumValues: []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}, Hint: "per-model allowlist in reasoning.supported_efforts"},
		system.IntentReasoningBudget:           {Kind: bodyParam, Param: "reasoning.max_tokens", ParamType: "number"},
		system.IntentReasoningFormat:           {Kind: bodyParam, Param: "reasoning.exclude", ParamType: "boolean", Hint: "true hides reasoning from the response without disabling it"},
		system.IntentResponseFormatJSON:        {Kind: bodyParam, Param: "response_format", ParamType: "object", Value: jsonObjectFormat},
		system.IntentResponseFormatSchema:      {Kind: bodyParam, Param: "response_format", ParamType: "object", Hint: `{"type":"json_schema","json_schema":{"name":...,"schema":{...},"strict":true}}; only endpoints listing structured_outputs`},
		system.IntentSamplingTopK:              {Kind: bodyParam, Param: "top_k", ParamType: "number"},
		system.IntentSamplingMinP:              {Kind: bodyParam, Param: "min_p", ParamType: "number"},
		system.IntentSamplingRepetitionPenalty: {Kind: bodyParam, Param: "repetition_penalty", ParamType: "number"},
		system.IntentSamplingTopA:              {Kind: bodyParam, Param: "top_a", ParamType: "number"},
		system.IntentStreamIncludeUsage:        {Kind: bodyParam, Param: "stream_options.include_usage", ParamType: "boolean", Value: true, Hint: "usage is always included in the last chunk"},
		system.IntentCachePrompt:               {Kind: bodyParam, Param: "cache_control", ParamType: "object", Value: map[string]any{"type": "ephemeral"}, Hint: "top-level applies to the last cacheable block; also per content block"},
		system.IntentToolsParallel:             {Kind: bodyParam, Param: "parallel_tool_calls", ParamType: "boolean"},
	},

	// Venice.
	// https://docs.venice.ai/api-reference/endpoint/chat/completions
	"venice": {
		system.IntentReasoningDisable:          {Kind: bodyParam, Param: "venice_parameters.disable_thinking", ParamType: "boolean", Value: true, Hint: "adds /no_think and strips the thinking block"},
		system.IntentReasoningEffort:           {Kind: bodyParam, Param: "reasoning_effort", ParamType: "enum", EnumValues: []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}, Hint: "top-level takes precedence over reasoning.effort"},
		system.IntentReasoningFormat:           {Kind: bodyParam, Param: "venice_parameters.strip_thinking_response", ParamType: "boolean", Hint: "true suppresses <think> blocks server-side"},
		system.IntentSamplingTopK:              {Kind: bodyParam, Param: "top_k", ParamType: "number", Hint: "integer >= 0"},
		system.IntentSamplingMinP:              {Kind: bodyParam, Param: "min_p", ParamType: "number", Hint: "[0,1]"},
		system.IntentSamplingRepetitionPenalty: {Kind: bodyParam, Param: "repetition_penalty", ParamType: "number", Hint: ">= 0, 1.0 = off"},
		system.IntentResponseFormatJSON:        {Kind: bodyParam, Param: "response_format", ParamType: "object", Value: jsonObjectFormat},
		system.IntentResponseFormatSchema:      {Kind: bodyParam, Param: "response_format", ParamType: "object", Hint: `{"type":"json_schema","json_schema":{...}}`},
		system.IntentStreamIncludeUsage:        {Kind: bodyParam, Param: "stream_options.include_usage", ParamType: "boolean", Value: true},
		system.IntentToolsParallel:             {Kind: bodyParam, Param: "parallel_tool_calls", ParamType: "boolean"},
	},

	// LiteLLM Proxy: a gateway whose unified params are translated per
	// upstream provider. Reasoning knobs are provider-neutral here, so they
	// are safe at stack level (still gated on reasoning evidence).
	// https://docs.litellm.ai/docs/completion/input
	// https://docs.litellm.ai/docs/reasoning_content
	// https://docs.litellm.ai/docs/providers/anthropic
	"litellm": {
		system.IntentReasoningDisable:     {Kind: bodyParam, Param: "reasoning_effort", ParamType: "string", Value: "none"},
		system.IntentReasoningEnable:      {Kind: bodyParam, Param: "reasoning_effort", ParamType: "string", Value: "medium", Hint: "any level other than none enables; Anthropic/Bedrock/Vertex upstreams also accept thinking {type: enabled, budget_tokens >= 1024}"},
		system.IntentReasoningEffort:      {Kind: bodyParam, Param: "reasoning_effort", ParamType: "enum", EnumValues: []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}},
		system.IntentReasoningBudget:      {Kind: bodyParam, Param: "thinking.budget_tokens", ParamType: "number", Hint: "Anthropic/Bedrock/Vertex upstreams only; minimum 1024"},
		system.IntentResponseFormatJSON:   {Kind: bodyParam, Param: "response_format", ParamType: "object", Value: jsonObjectFormat},
		system.IntentResponseFormatSchema: {Kind: bodyParam, Param: "response_format", ParamType: "object", Hint: "translated per provider; enforced client-side only with enable_json_schema_validation"},
		system.IntentStreamIncludeUsage:   {Kind: bodyParam, Param: "stream_options.include_usage", ParamType: "boolean", Value: true},
		system.IntentToolsParallel:        {Kind: bodyParam, Param: "parallel_tool_calls", ParamType: "boolean"},
	},

	// Anthropic Messages API (what the claudeai adapter targets). The
	// reasoning toggles come from the claude family defaults.
	// https://platform.claude.com/docs/en/api/messages
	// https://platform.claude.com/docs/en/build-with-claude/thinking
	// https://platform.claude.com/docs/en/build-with-claude/effort
	// https://platform.claude.com/docs/en/build-with-claude/structured-outputs
	"anthropic": {
		system.IntentReasoningFormat:      {Kind: bodyParam, Param: "thinking.display", ParamType: "enum", EnumValues: []string{"summarized", "omitted", "updates"}, Hint: "invalid with thinking.type disabled; updates needs a beta header"},
		system.IntentResponseFormatSchema: {Kind: bodyParam, Param: "output_config.format", ParamType: "object", Hint: `{"type":"json_schema","schema":{...,"additionalProperties":false}}; no schema-less JSON mode`},
		system.IntentSamplingTopK:         {Kind: bodyParam, Param: "top_k", ParamType: "number", Hint: "rejected on newest models; incompatible with thinking on older ones"},
		system.IntentCachePrompt:          {Kind: bodyParam, Param: "cache_control", ParamType: "object", Value: map[string]any{"type": "ephemeral"}, Hint: "top-level applies to the last cacheable block; ttl 5m|1h"},
	},
}

// ollamaThinkBindings maps the reasoning toggle onto what Ollama's
// OpenAI-compatible /v1 endpoint documents: reasoning_effort, where "none"
// turns thinking off. The native /api/chat boolean `think` is reported for
// completeness but is unreachable through the openai adapter. Stack
// knowledge for the "ollama" table above.
// https://docs.ollama.com/api/openai-compatibility
// https://docs.ollama.com/capabilities/thinking
func ollamaThinkBindings() map[string]*system.ParamBinding {
	return map[string]*system.ParamBinding{
		system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", ParamType: "string", Value: "none", Hint: "native /api/chat: think: false"},
		system.IntentReasoningEnable:  {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", ParamType: "string", Value: "medium", Hint: "on /v1 any level other than none enables thinking (gpt-oss maps low/medium/high); native /api/chat: think: true"},
	}
}

// rewriteBindingsForOllama rewrites template-kwarg reasoning bindings onto
// what Ollama's /v1 endpoint honors (chat_template_kwargs is not passed
// through): boolean toggles become the reasoning_effort "none" / medium
// pair, effort enums keep their levels on reasoning_effort. Intents with no
// Ollama equivalent (e.g. a numeric budget kwarg) are dropped. Stack
// knowledge for the "ollama" table above.
func rewriteBindingsForOllama(bindings map[string]*system.ParamBinding) {
	if bindings == nil {
		return
	}
	toggle := false
	for intent, b := range bindings {
		if b == nil || b.Kind != system.BindingKindTemplateKwarg {
			continue
		}
		switch intent {
		case system.IntentReasoningDisable, system.IntentReasoningEnable:
			toggle = true
		case system.IntentReasoningEffort:
			bindings[intent] = &system.ParamBinding{
				Kind: system.BindingKindBodyParam, Param: "reasoning_effort", ParamType: "enum",
				EnumValues: b.EnumValues, Hint: "native /api/chat equivalent: think: <level>",
			}
		default:
			delete(bindings, intent)
		}
	}
	if toggle {
		for intent, b := range ollamaThinkBindings() {
			bindings[intent] = b
		}
	}
}

// stackParameters lists the standard OpenAI chat-completions params each
// stack documents as supported on its OpenAI-compatible endpoint (composition
// rule 5: `parameters` is the preset's documented list). See the doc-URL
// comments on each stack's stackBindings table above for the sources.
var stackParameters = map[string][]string{
	"vllm":       {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	"sglang":     {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	"llamacpp":   {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	"ollama":     {"model", "messages", "temperature", "top_p", "stream", "stream_options", "stop", "max_tokens", "presence_penalty", "frequency_penalty", "seed", "response_format", "tools", "reasoning_effort"},
	"openrouter": {"model", "messages", "temperature", "top_p", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	// Venice. https://docs.venice.ai/api-reference/endpoint/chat/completions
	// (logit_bias is not documented by Venice — deliberately omitted)
	"venice":  {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	"litellm": {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	// Anthropic Messages API wire names (not OpenAI's): the claudeai adapter
	// targets /v1/messages. https://platform.claude.com/docs/en/api/messages
	"anthropic": {"model", "messages", "system", "max_tokens", "stream", "temperature", "top_p", "top_k", "stop_sequences", "tools", "tool_choice", "metadata"},
	// OpenAI's own API defines the standard surface.
	// https://platform.openai.com/docs/api-reference/chat/create
	"openai": openaiChatParameters,
}

// openaiChatParameters is the standard OpenAI chat-completions request
// surface, by definition supported on api.openai.com.
var openaiChatParameters = []string{"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"}

// TransportFor returns the adapter apiType required by an apiStack preset.
// The table is config.StackTransport (config cannot import this package),
// so the loader's validation and this package agree on the nine presets;
// legacy adapter names are apiType values and are not accepted here.
func TransportFor(stack string) (string, bool) {
	adapter, ok := config.StackTransport[stack]
	return adapter, ok
}

// StackFor returns the spec stack name for a configured apiStack, or ""
// when it is empty or not one of the nine presets (no chat API spec).
func StackFor(stack string) string {
	if _, ok := config.StackTransport[stack]; ok {
		return stack
	}
	return ""
}

// gatewayStacks are OpenAI-compatible gateways / hosted vendors that
// translate requests for upstream providers: model-family template kwargs
// and vendor-native params do not apply to them.
var gatewayStacks = map[string]bool{"venice": true, "openrouter": true, "litellm": true, "anthropic": true, "openai": true}

// familyNativeVendor maps a family to the stack whose own API its
// body-param defaults describe. o-series' default is the standard OpenAI
// reasoning_effort body param, valid wherever the generic openai preset
// forwards standard fields. gemini has no native vendor here (no gemini
// preset exists yet) and stays unreachable on gateway stacks.
var familyNativeVendor = map[string]string{"claude": "anthropic", "o-series": "openai"}
