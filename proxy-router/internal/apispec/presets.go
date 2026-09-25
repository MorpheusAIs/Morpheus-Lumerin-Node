package apispec

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// Static per-stack remap tables. Keep entries literal to the cited docs: a
// wrong binding is worse than a missing one, since consumers translate
// requests with it. Reasoning toggles are model properties (family and
// template layers); stacks contribute only model-independent reasoning knobs.

const (
	bodyParam   = system.BindingKindBodyParam
	nativeParam = system.BindingKindNativeBodyParam
)

type bindingSet = map[string]*system.ParamBinding

var jsonObjectFormat = map[string]any{"type": "json_object"}

var stackBindings = map[string]bindingSet{
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

	// Native /api knobs are reported as native_body_param; only /v1 is reachable
	// through the openai adapter.
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

	// Reasoning knobs are provider-neutral on LiteLLM, so they sit at stack level.
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

	// Reasoning toggles come from the claude family defaults.
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

// https://docs.ollama.com/api/openai-compatibility
// https://docs.ollama.com/capabilities/thinking
func ollamaThinkBindings() map[string]*system.ParamBinding {
	return map[string]*system.ParamBinding{
		system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", ParamType: "string", Value: "none", Hint: "native /api/chat: think: false"},
		system.IntentReasoningEnable:  {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", ParamType: "string", Value: "medium", Hint: "on /v1 any level other than none enables thinking (gpt-oss maps low/medium/high); native /api/chat: think: true"},
	}
}

// Ollama's /v1 endpoint does not pass chat_template_kwargs through.
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

// Sources: the URLs on each stack's stackBindings table.
var stackParameters = map[string][]string{
	"vllm":       {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	"sglang":     {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	"llamacpp":   {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	"ollama":     {"model", "messages", "temperature", "top_p", "stream", "stream_options", "stop", "max_tokens", "presence_penalty", "frequency_penalty", "seed", "response_format", "tools", "reasoning_effort"},
	"openrouter": {"model", "messages", "temperature", "top_p", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	// https://docs.venice.ai/api-reference/endpoint/chat/completions (logit_bias
	// undocumented, omitted)
	"venice":  {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	"litellm": {"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"},
	// Anthropic Messages names, not OpenAI's.
	// https://platform.claude.com/docs/en/api/messages
	"anthropic": {"model", "messages", "system", "max_tokens", "stream", "temperature", "top_p", "top_k", "stop_sequences", "tools", "tool_choice", "metadata"},
	// https://platform.openai.com/docs/api-reference/chat/create
	"openai": openaiChatParameters,
}

var openaiChatParameters = []string{"model", "messages", "temperature", "top_p", "n", "stream", "stream_options", "stop", "max_tokens", "max_completion_tokens", "presence_penalty", "frequency_penalty", "logit_bias", "logprobs", "top_logprobs", "user", "seed", "response_format", "tools", "tool_choice", "parallel_tool_calls", "reasoning_effort"}

// Normalized again so a ModelConfig built outside the loader still resolves.
func StackFor(stack string) string {
	stack = config.NormalizeApiStack(stack)
	if _, ok := config.StackTransport[stack]; ok {
		return stack
	}
	return ""
}

var gatewayStacks = map[string]bool{"venice": true, "openrouter": true, "litellm": true, "anthropic": true, "openai": true}

// The stack whose own API a family's body-param defaults describe. gemini
// is absent: no gemini preset exists.
var familyNativeVendor = map[string]string{"claude": "anthropic", "o-series": "openai"}
