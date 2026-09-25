package apiprobe

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseContextLimit(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		status int
		limit  int64
		source string
		ok     bool
	}{
		{
			name:   "vllm sentence form, nested error object",
			body:   `{"error":{"message":"This model's maximum context length is 40960 tokens. However, you requested 41000 tokens (40000 in the messages, 1000 in the completion). Please reduce the length of the messages or completion.","type":"BadRequestError","param":null,"code":400}}`,
			status: 400, limit: 40960, source: ContextSourceOpenAI, ok: true,
		},
		{
			name:   "vllm sentence form, flat legacy envelope",
			body:   `{"object":"error","message":"This model's maximum context length is 8192 tokens. However, you requested 9000 tokens.","type":"BadRequestError","param":null,"code":400}`,
			status: 400, limit: 8192, source: ContextSourceOpenAI, ok: true,
		},
		{
			name:   "vllm input length form",
			body:   `{"error":{"message":"Input length (5000) exceeds model's maximum context length (4096).","type":"BadRequestError","code":400}}`,
			status: 400, limit: 4096, source: ContextSourceVLLM, ok: true,
		},
		{
			name:   "vllm input length form with 'the'",
			body:   `{"error":{"message":"Input length (5000) exceeds the model's maximum context length (2048)."}}`,
			status: 400, limit: 2048, source: ContextSourceVLLM, ok: true,
		},
		{
			name:   "openai context_length_exceeded",
			body:   `{"error":{"message":"This model's maximum context length is 128000 tokens. However, your messages resulted in 130000 tokens. Please reduce the length of the messages.","type":"invalid_request_error","param":"messages","code":"context_length_exceeded"}}`,
			status: 400, limit: 128000, source: ContextSourceOpenAI, ok: true,
		},
		{
			name:   "llama.cpp structured n_ctx beats the sentence",
			body:   `{"error":{"code":400,"message":"request (9000 tokens) exceeds the available context size (8192 tokens), try increasing it","type":"exceed_context_size_error","n_prompt_tokens":9000,"n_ctx":4096}}`,
			status: 400, limit: 4096, source: ContextSourceLlamaCpp, ok: true,
		},
		{
			name:   "llama.cpp type without n_ctx falls back to the sentence",
			body:   `{"error":{"code":400,"message":"request (9000 tokens) exceeds the available context size (8192 tokens), try increasing it","type":"exceed_context_size_error"}}`,
			status: 400, limit: 8192, source: ContextSourceLlamaCpp, ok: true,
		},
		{
			name:   "llama.cpp type with invalid n_ctx and no sentence",
			body:   `{"error":{"code":400,"message":"context overflow","type":"exceed_context_size_error","n_ctx":0}}`,
			status: 400, limit: 0, source: ContextSourceLlamaCpp, ok: false,
		},
		{
			name:   "tgi must be <=",
			body:   "{\"error\":\"Input validation error: `inputs` tokens + `max_new_tokens` must be <= 4096. Given: 4000 `inputs` tokens and 200 `max_new_tokens`\",\"error_type\":\"validation\"}",
			status: 422, limit: 4096, source: ContextSourceTGI, ok: true,
		},
		{
			name:   "tgi input length reports the input bound under its own source",
			body:   "{\"error\":\"Input validation error: `inputs` must have less than 4095 tokens. Given: 5000\",\"error_type\":\"validation\"}",
			status: 422, limit: 4095, source: ContextSourceTGIInput, ok: true,
		},
		{
			name:   "tgi max_new_tokens validation is not a window",
			body:   "{\"error\":\"Input validation error: `max_new_tokens` must be <= 1000. Given: 5000\",\"error_type\":\"validation\"}",
			status: 422, ok: false,
		},
		{
			name:   "tgi top_n_tokens validation is not a window",
			body:   "{\"error\":\"Input validation error: `top_n_tokens` must be <= 5. Given: 10\",\"error_type\":\"validation\"}",
			status: 422, ok: false,
		},
		{
			name:   "a parameter range error is not a window",
			body:   `{"error":{"message":"temperature must be <= 2","type":"invalid_request_error"}}`,
			status: 400, ok: false,
		},
		{
			name:   "anthropic prompt too long with totals",
			body:   `{"type":"error","error":{"type":"invalid_request_error","message":"prompt is too long: 213462 tokens > 200000 maximum"}}`,
			status: 400, limit: 200000, source: ContextSourceAnthropic, ok: true,
		},
		{
			name:   "anthropic prompt too long without a number is recognised but unset",
			body:   `{"type":"error","error":{"type":"invalid_request_error","message":"prompt is too long"}}`,
			status: 400, limit: 0, source: ContextSourceAnthropic, ok: false,
		},
		{
			name:   "plain text body",
			body:   "This model's maximum context length is 2048 tokens.",
			status: 400, limit: 2048, source: ContextSourceOpenAI, ok: true,
		},
		{
			name:   "unknown status is accepted",
			body:   `{"error":{"message":"This model's maximum context length is 2048 tokens."}}`,
			status: 0, limit: 2048, source: ContextSourceOpenAI, ok: true,
		},
		{
			name:   "5xx is not a request rejection",
			body:   `{"error":{"message":"This model's maximum context length is 2048 tokens."}}`,
			status: 500, ok: false,
		},
		{
			name:   "2xx is not an error",
			body:   `{"error":{"message":"This model's maximum context length is 2048 tokens."}}`,
			status: 200, ok: false,
		},
		{
			name:   "unrelated error",
			body:   `{"error":{"message":"Rate limit exceeded","code":"rate_limit_exceeded"}}`,
			status: 429, ok: false,
		},
		{
			name:   "out of range number is ignored",
			body:   `{"error":{"message":"This model's maximum context length is 99999999999 tokens."}}`,
			status: 400, ok: false,
		},
		{
			name:   "zero is ignored",
			body:   `{"error":{"message":"Input length (5000) exceeds model's maximum context length (0)."}}`,
			status: 400, ok: false,
		},
		{name: "empty", body: "", status: 400, ok: false},
		{name: "garbage", body: "\x00\xff{{{", status: 400, ok: false},
		{name: "json but not an object", body: `[1,2,3]`, status: 400, ok: false},
		{name: "error is a number", body: `{"error":42}`, status: 400, ok: false},
		{name: "detail is a number", body: `{"detail":42}`, status: 400, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			limit, source, ok := ParseContextLimit([]byte(tc.body), tc.status)
			require.Equal(t, tc.ok, ok)
			require.Equal(t, tc.limit, limit)
			require.Equal(t, tc.source, source)
		})
	}
}

func TestParseContextLimitBoundsInput(t *testing.T) {
	padding := strings.Repeat(" ", MaxErrorBody)
	body := padding + `{"error":{"message":"This model's maximum context length is 2048 tokens."}}`
	_, _, ok := ParseContextLimit([]byte(body), 400)
	require.False(t, ok, "text past the bound is not read")

	body = `{"error":{"message":"This model's maximum context length is 2048 tokens.` + padding + `"}}`
	limit, _, ok := ParseContextLimit([]byte(body), 400)
	require.True(t, ok, "a body truncated at the bound is read as text")
	require.Equal(t, int64(2048), limit)
}

func TestParseRejectedParams(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		params []string
		ok     bool
	}{
		{
			name:   "openai unrecognized argument",
			body:   `{"error":{"message":"Unrecognized request argument supplied: chat_template_kwargs","type":"invalid_request_error","param":null,"code":null}}`,
			params: []string{"chat_template_kwargs"}, ok: true,
		},
		{
			name:   "openai unrecognized arguments, plural list",
			body:   `{"error":{"message":"Unrecognized request arguments supplied: top_k, min_p"}}`,
			params: []string{"top_k", "min_p"}, ok: true,
		},
		{
			name:   "fastapi pydantic extra_forbidden detail array",
			body:   `{"detail":[{"type":"extra_forbidden","loc":["body","reasoning_effort"],"msg":"Extra inputs are not permitted","input":"none"},{"type":"missing","loc":["body","messages"],"msg":"Field required"}]}`,
			params: []string{"reasoning_effort"}, ok: true,
		},
		{
			name:   "pydantic nested loc joins with dots",
			body:   `{"detail":[{"type":"extra_forbidden","loc":["body","extra_body","thinking"],"msg":"Extra inputs are not permitted"}]}`,
			params: []string{"extra_body.thinking"}, ok: true,
		},
		{
			name:   "pydantic loc with an index",
			body:   `{"detail":[{"type":"extra_forbidden","loc":["body","messages",0,"reasoning"],"msg":"Extra inputs are not permitted"}]}`,
			params: []string{"messages.reasoning"}, ok: true,
		},
		{
			name: "pydantic loc not under body is ignored",
			body: `{"detail":[{"type":"extra_forbidden","loc":["query","foo"],"msg":"Extra inputs are not permitted"}]}`,
			ok:   false,
		},
		{
			name:   "vllm stringified pydantic repr",
			body:   `{"error":{"message":"[{'type': 'extra_forbidden', 'loc': ('body', 'chat_template_kwargs'), 'msg': 'Extra inputs are not permitted', 'input': {'enable_thinking': False}}]","type":"BadRequestError","param":null,"code":400}}`,
			params: []string{"chat_template_kwargs"}, ok: true,
		},
		{
			name:   "vllm stringified repr, nested loc",
			body:   `{"error":{"message":"[{'type': 'extra_forbidden', 'loc': ('body', 'extra_body', 'top_k'), 'msg': 'Extra inputs are not permitted'}]"}}`,
			params: []string{"extra_body.top_k"}, ok: true,
		},
		{
			name:   "litellm unsupported params list",
			body:   `{"error":{"message":"litellm.UnsupportedParamsError: anthropic does not support parameters: ['top_k', 'reasoning_effort'], for model=claude-3. To drop these, set litellm.drop_params=True","type":"None","param":"None","code":"400"}}`,
			params: []string{"top_k", "reasoning_effort"}, ok: true,
		},
		{
			name:   "litellm unsupported params dict",
			body:   `{"error":{"message":"litellm.UnsupportedParamsError: openai does not support parameters: {'min_p': 0.1, 'top_k': 40}, for model=gpt-4o"}}`,
			params: []string{"min_p", "top_k"}, ok: true,
		},
		{
			name:   "vllm auto tool choice refusal",
			body:   `{"error":{"message":"\"auto\" tool choice requires --enable-auto-tool-choice and --tool-call-parser to be set","type":"BadRequestError","param":null,"code":400}}`,
			params: []string{ToolChoiceParam}, ok: true,
		},
		{
			name:   "two forms in one body are deduped",
			body:   `{"error":{"message":"Unrecognized request argument supplied: top_k"},"detail":[{"type":"extra_forbidden","loc":["body","top_k"]}]}`,
			params: []string{"top_k"}, ok: true,
		},
		{
			name:   "invalid names are dropped",
			body:   `{"error":{"message":"Unrecognized request argument supplied: bad name, ok_name, ` + strings.Repeat("x", 70) + `"}}`,
			params: []string{"ok_name"}, ok: true,
		},
		{name: "unrelated", body: `{"error":{"message":"Rate limit exceeded"}}`, ok: false},
		{name: "empty", body: "", ok: false},
		{name: "garbage", body: "\x00\xff{{{'loc'", ok: false},
		{name: "detail is a string", body: `{"detail":"Not Found"}`, ok: false},
		{name: "detail items are numbers", body: `{"detail":[1,2]}`, ok: false},
		{name: "loc is not an array", body: `{"detail":[{"type":"extra_forbidden","loc":"body"}]}`, ok: false},
		{name: "loc holds objects", body: `{"detail":[{"type":"extra_forbidden","loc":[{"a":1},"foo"]}]}`, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params, ok := ParseRejectedParams([]byte(tc.body))
			require.Equal(t, tc.ok, ok)
			require.Equal(t, tc.params, params)
		})
	}
}

func TestParseRejectedParamsCapsCount(t *testing.T) {
	var names []string
	for i := 0; i < maxRejectedParams+10; i++ {
		names = append(names, "p"+strings.Repeat("x", i%5)+string(rune('a'+i%26))+strings.Repeat("y", i/26))
	}
	body := `{"error":{"message":"Unrecognized request arguments supplied: ` + strings.Join(names, ", ") + `"}}`
	params, ok := ParseRejectedParams([]byte(body))
	require.True(t, ok)
	require.LessOrEqual(t, len(params), maxRejectedParams)
}

func TestIsToolChoiceAutoUnsupported(t *testing.T) {
	require.True(t, IsToolChoiceAutoUnsupported([]byte(`{"error":{"message":"\"auto\" tool choice requires --enable-auto-tool-choice and --tool-call-parser to be set"}}`)))
	require.True(t, IsToolChoiceAutoUnsupported([]byte(`"auto" tool choice requires --enable-auto-tool-choice`)))
	require.False(t, IsToolChoiceAutoUnsupported([]byte(`{"error":{"message":"tool choice is fine"}}`)))
	require.False(t, IsToolChoiceAutoUnsupported(nil))
	long := strings.Repeat(" ", MaxErrorBody) + `"auto" tool choice requires --enable-auto-tool-choice`
	require.False(t, IsToolChoiceAutoUnsupported([]byte(long)))
}

func FuzzParsersNeverPanic(f *testing.F) {
	for _, seed := range []string{
		`{"error":{"message":"This model's maximum context length is 4096 tokens."}}`,
		`{"detail":[{"type":"extra_forbidden","loc":["body","x"]}]}`,
		`{"error":"must be <= 10"}`, `{"error":{"n_ctx":"x","type":"exceed_context_size_error"}}`,
		"", "\x00", "[", "{", `{"error":null,"detail":null,"message":null}`,
	} {
		f.Add([]byte(seed), 400)
	}
	f.Fuzz(func(t *testing.T, body []byte, status int) {
		ParseContextLimit(body, status)
		ParseRejectedParams(body)
		IsToolChoiceAutoUnsupported(body)
	})
}
