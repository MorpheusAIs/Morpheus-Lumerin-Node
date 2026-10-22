package apiprobe

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"strings"
	"testing"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

// wire marshals the request the way the OpenAI adapter does and decodes it
// into a generic tree for assertions.
func wire(t *testing.T, req *gcs.OpenAICompletionRequestExtra) map[string]any {
	t.Helper()
	raw, err := json.Marshal(req)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestNewProbeRequest(t *testing.T) {
	body := wire(t, NewProbeRequest("What is 1+1?"))
	require.Equal(t, float64(1), body["max_tokens"])
	require.Nil(t, body["stream"])
	msgs := body["messages"].([]any)
	require.Len(t, msgs, 1)
	require.Equal(t, map[string]any{"role": "user", "content": "What is 1+1?"}, msgs[0])

	require.Equal(t, DefaultPrompt, NewProbeRequest("").Messages[0].Content)
}

func TestWithDummyToolAuto(t *testing.T) {
	body := wire(t, WithDummyToolAuto(NewProbeRequest("")))
	require.Equal(t, "auto", body["tool_choice"])
	tools := body["tools"].([]any)
	require.Len(t, tools, 1)
	tool := tools[0].(map[string]any)
	require.Equal(t, "function", tool["type"])
	fn := tool["function"].(map[string]any)
	require.Equal(t, DummyToolName, fn["name"])
	require.Equal(t, map[string]any{"type": "object", "properties": map[string]any{}}, fn["parameters"])
	require.Equal(t, float64(1), body["max_tokens"])
}

func TestWithJSONSchema(t *testing.T) {
	body := wire(t, WithJSONSchema(NewProbeRequest("")))
	rf := body["response_format"].(map[string]any)
	require.Equal(t, "json_schema", rf["type"])
	schema := rf["json_schema"].(map[string]any)
	require.Equal(t, "probe", schema["name"])
	require.Equal(t, "object", schema["schema"].(map[string]any)["type"])
}

func TestWithImagePart(t *testing.T) {
	body := wire(t, WithImagePart(NewProbeRequest("Look.")))
	msgs := body["messages"].([]any)
	require.Len(t, msgs, 1)
	msg := msgs[0].(map[string]any)
	require.Equal(t, "user", msg["role"])
	parts := msg["content"].([]any)
	require.Len(t, parts, 2)
	require.Equal(t, map[string]any{"type": "text", "text": "Look."}, parts[0])
	require.Equal(t, map[string]any{"type": "image_url", "image_url": map[string]any{"url": TinyPNGDataURL}}, parts[1])

	t.Run("a request without a user message gets one", func(t *testing.T) {
		req := &gcs.OpenAICompletionRequestExtra{}
		body := wire(t, WithImagePart(req))
		msg := body["messages"].([]any)[0].(map[string]any)
		require.Equal(t, DefaultPrompt, msg["content"].([]any)[0].(map[string]any)["text"])
	})
}

func TestTinyPNGDecodes(t *testing.T) {
	const prefix = "data:image/png;base64,"
	require.True(t, strings.HasPrefix(TinyPNGDataURL, prefix))
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(TinyPNGDataURL, prefix))
	require.NoError(t, err)
	img, err := png.Decode(bytes.NewReader(raw))
	require.NoError(t, err)
	require.Equal(t, 1, img.Bounds().Dx())
	require.Equal(t, 1, img.Bounds().Dy())
}

func TestApplyBinding(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(req *gcs.OpenAICompletionRequestExtra)
		binding *system.ParamBinding
		want    map[string]any // wire keys to check
		err     error
	}{
		{
			name:    "template kwarg lands in the typed chat_template_kwargs",
			binding: &system.ParamBinding{Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", Value: false},
			want:    map[string]any{"chat_template_kwargs": map[string]any{"enable_thinking": false}},
		},
		{
			name: "a second kwarg keeps the first",
			prepare: func(req *gcs.OpenAICompletionRequestExtra) {
				req.ChatTemplateKwargs = map[string]any{"thinking": true}
			},
			binding: &system.ParamBinding{Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", Value: false},
			want:    map[string]any{"chat_template_kwargs": map[string]any{"thinking": true, "enable_thinking": false}},
		},
		{
			name:    "body param on a known field",
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "reasoning_effort", Value: "low"},
			want:    map[string]any{"reasoning_effort": "low"},
		},
		{
			name:    "body param on an unknown field goes to Extra",
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "top_k", Value: 40},
			want:    map[string]any{"top_k": float64(40)},
		},
		{
			name:    "nested path in Extra",
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "extra_body.thinking.type", Value: "disabled"},
			want:    map[string]any{"extra_body": map[string]any{"thinking": map[string]any{"type": "disabled"}}},
		},
		{
			name: "nested path merges into an existing Extra object",
			prepare: func(req *gcs.OpenAICompletionRequestExtra) {
				req.Extra = map[string]json.RawMessage{"extra_body": json.RawMessage(`{"top_k":5,"thinking":{"budget":100}}`)}
			},
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "extra_body.thinking.type", Value: "disabled"},
			want:    map[string]any{"extra_body": map[string]any{"top_k": float64(5), "thinking": map[string]any{"budget": float64(100), "type": "disabled"}}},
		},
		{
			name: "an array value extends a known array field",
			prepare: func(req *gcs.OpenAICompletionRequestExtra) {
				req.Stop = []string{"a"}
			},
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "stop", Value: []string{"b"}},
			want:    map[string]any{"stop": []any{"a", "b"}},
		},
		{
			name: "an array value extends an Extra array",
			prepare: func(req *gcs.OpenAICompletionRequestExtra) {
				req.Extra = map[string]json.RawMessage{"plugins": json.RawMessage(`[{"id":"x"}]`)}
			},
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "plugins", Value: []any{map[string]any{"id": "web"}}},
			want:    map[string]any{"plugins": []any{map[string]any{"id": "x"}, map[string]any{"id": "web"}}},
		},
		{
			name:    "an object value on a known object field",
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "response_format", Value: map[string]any{"type": "json_object"}},
			want:    map[string]any{"response_format": map[string]any{"type": "json_object"}},
		},
		{
			name:    "tool_entry appends to an empty tools list",
			binding: &system.ParamBinding{Kind: system.BindingKindToolEntry, Param: "tools", Value: map[string]any{"googleSearch": map[string]any{}}},
			want:    map[string]any{"tools": []any{map[string]any{"googleSearch": map[string]any{}}}},
		},
		{
			name: "tool_entry keeps the typed tools",
			prepare: func(req *gcs.OpenAICompletionRequestExtra) {
				WithDummyToolAuto(req)
			},
			binding: &system.ParamBinding{Kind: system.BindingKindToolEntry, Param: "tools", Value: map[string]any{"type": "web_search_20250305", "name": "web_search"}},
			want: map[string]any{"tools": []any{
				map[string]any{"type": "function", "function": map[string]any{"name": DummyToolName, "description": "Does nothing.", "parameters": map[string]any{"type": "object", "properties": map[string]any{}}}},
				map[string]any{"type": "web_search_20250305", "name": "web_search"},
			}},
		},
		{
			name:    "system_prompt is not a request parameter",
			binding: &system.ParamBinding{Kind: system.BindingKindSystemPrompt, Hint: "/no_think"},
			err:     ErrUnsupportedBinding,
		},
		{
			name:    "native_body_param is unreachable on this endpoint",
			binding: &system.ParamBinding{Kind: system.BindingKindNativeBodyParam, Param: "think", Value: false},
			err:     ErrUnsupportedBinding,
		},
		{
			name:    "a binding without a value cannot be applied",
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "reasoning_effort"},
			err:     ErrBindingWithoutValue,
		},
		{
			name:    "a binding without a parameter cannot be applied",
			binding: &system.ParamBinding{Kind: system.BindingKindBodyParam, Value: 1},
			err:     ErrBindingWithoutParam,
		},
		{
			name:    "a tool_entry that is not an object is refused",
			binding: &system.ParamBinding{Kind: system.BindingKindToolEntry, Param: "tools", Value: "web"},
			err:     ErrToolEntryNotObject,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := NewProbeRequest("q")
			if tc.prepare != nil {
				tc.prepare(req)
			}
			err := ApplyBinding(req, tc.binding)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			body := wire(t, req)
			for key, want := range tc.want {
				require.Equal(t, want, body[key], key)
			}
			require.Equal(t, float64(1), body["max_tokens"], "the probe bound survives")
		})
	}
}

func TestApplyBindingTypedFieldStaysTyped(t *testing.T) {
	req := NewProbeRequest("q")
	require.NoError(t, ApplyBinding(req, &system.ParamBinding{Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", Value: false}))
	require.Equal(t, map[string]any{"enable_thinking": false}, req.ChatTemplateKwargs)
	require.Empty(t, req.Extra, "a known field never lands in Extra")

	require.NoError(t, ApplyBinding(req, &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "max_tokens", Value: 7}))
	require.Equal(t, 7, req.MaxTokens)
}

func TestSetBodyPathMalformed(t *testing.T) {
	req := NewProbeRequest("q")
	require.Error(t, SetBodyPath(req, "a..b", 1))
	require.Error(t, SetBodyPath(req, "", 1))
}

func TestRenderURL(t *testing.T) {
	cases := map[string]string{
		"http://llm:8000/v1/chat/completions":         "http://llm:8000/v1/chat/completions/render",
		"http://llm:8000/v1/chat/completions/":        "http://llm:8000/v1/chat/completions/render",
		"http://llm:8000/v1":                          "http://llm:8000/v1/chat/completions/render",
		"http://llm:8000/v1/completions":              "http://llm:8000/v1/chat/completions/render",
		"http://llm:8000":                             "http://llm:8000/v1/chat/completions/render",
		"https://llm.example/api/v1/chat/completions": "https://llm.example/api/v1/chat/completions/render",
		"not a url": "",
		"/v1/chat":  "",
	}
	for in, want := range cases {
		require.Equal(t, want, RenderURL(in), in)
	}
}

func TestRenderRequestBody(t *testing.T) {
	raw, err := RenderRequestBody("qwen3", "q", nil)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body))
	require.Equal(t, "qwen3", body["model"])
	require.Equal(t, float64(1), body["max_tokens"])
	require.Nil(t, body["chat_template_kwargs"])

	raw, err = RenderRequestBody("qwen3", "q", map[string]any{"enable_thinking": false})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &body))
	require.Equal(t, map[string]any{"enable_thinking": false}, body["chat_template_kwargs"])
}

func TestRenderedTokens(t *testing.T) {
	ids, ok := RenderedTokens([]byte(`{"request_id":"r1","token_ids":[1,2,3],"sampling_params":{}}`))
	require.True(t, ok)
	require.Equal(t, []int64{1, 2, 3}, ids)
	_, ok = RenderedTokens([]byte(`{"request_id":"r1","token_ids":[]}`))
	require.False(t, ok)
	_, ok = RenderedTokens([]byte(`not json`))
	require.False(t, ok)
	_, ok = RenderedTokens(nil)
	require.False(t, ok)
}

func TestProbeRequestRoundTripsThroughAdapterMarshal(t *testing.T) {
	// The adapter sets Model and marshals; Extra must not shadow a typed
	// field the probe set.
	req := WithDummyToolAuto(NewProbeRequest("q"))
	req.Model = "m"
	require.NoError(t, ApplyBinding(req, &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "top_k", Value: 3}))
	body := wire(t, req)
	require.Equal(t, "m", body["model"])
	require.Equal(t, "auto", body["tool_choice"])
	require.Equal(t, float64(3), body["top_k"])
	require.Len(t, body["tools"].([]any), 1)
	require.Equal(t, openai.ChatMessageRoleUser, req.Messages[0].Role)
}
