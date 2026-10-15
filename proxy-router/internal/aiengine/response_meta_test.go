package aiengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

type metaBackend struct {
	status  int
	headers map[string]string
	body    string
}

func promptWithMeta(t *testing.T, b metaBackend, stream bool) (*ResponseMeta, []gcs.Chunk, *gcs.AiEngineErrorResponse) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range b.headers {
			w.Header().Set(k, v)
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(b.status)
		_, _ = w.Write([]byte(b.body))
	}))
	defer server.Close()

	engine := NewOpenAIEngine("test-model", server.URL, "", time.Minute, &lib.LoggerMock{}, nil)
	meta := &ResponseMeta{}
	var chunks []gcs.Chunk
	var errResp *gcs.AiEngineErrorResponse
	req := &gcs.OpenAICompletionRequestExtra{ChatCompletionRequest: openai.ChatCompletionRequest{Stream: stream}}
	err := engine.Prompt(WithResponseMeta(context.Background(), meta), req, func(ctx context.Context, chunk gcs.Chunk, e *gcs.AiEngineErrorResponse) error {
		if e != nil {
			errResp = e
		}
		if chunk != nil {
			chunks = append(chunks, chunk)
		}
		return nil
	})
	require.NoError(t, err)
	return meta, chunks, errResp
}

func TestResponseMetaNonStreaming(t *testing.T) {
	body := `{"id":"x","object":"chat.completion","model":"qwen3-32b-served","system_fingerprint":"fp_1",
	"choices":[{"index":0,"message":{"role":"assistant","content":"4","reasoning_content":"2+2"},"finish_reason":"stop"}],
	"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`
	meta, chunks, errResp := promptWithMeta(t, metaBackend{status: 200, headers: map[string]string{
		"Server":                                 "uvicorn",
		"X-LiteLLM-Version":                      "1.70.0",
		"OpenAI-Processing-Ms":                   "12",
		"X-RateLimit-Limit-Requests":             "100",
		"anthropic-ratelimit-requests-remaining": "99",
		"X-Request-Id":                           "req-1",
		"X-Secret-Thing":                         "leak",
		"Set-Cookie":                             "session=abc",
	}, body: body}, false)
	require.Nil(t, errResp)
	require.Len(t, chunks, 1)

	require.Equal(t, 200, meta.StatusCode)
	require.Equal(t, "qwen3-32b-served", meta.Model)
	require.Equal(t, "fp_1", meta.SystemFingerprint)
	require.Equal(t, &UsageInfo{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}, meta.Usage)
	require.True(t, meta.ReasoningSeen)
	require.Nil(t, meta.ErrorBody)
	require.Equal(t, map[string]string{
		"server":                                 "uvicorn",
		"content-type":                           "application/json",
		"x-litellm-version":                      "1.70.0",
		"openai-processing-ms":                   "12",
		"x-ratelimit-limit-requests":             "100",
		"anthropic-ratelimit-requests-remaining": "99",
		"x-request-id":                           "req-1",
	}, meta.Headers)
}

func TestResponseMetaReasoningDetection(t *testing.T) {
	cases := []struct {
		name string
		msg  string
		want bool
	}{
		{"plain answer", `{"role":"assistant","content":"4"}`, false},
		{"reasoning_content", `{"role":"assistant","content":"4","reasoning_content":"think"}`, true},
		{"reasoning (ollama / gpt-oss)", `{"role":"assistant","content":"4","reasoning":"think"}`, true},
		{"empty reasoning fields", `{"role":"assistant","content":"4","reasoning_content":"","reasoning":""}`, false},
		{"think block in content", `{"role":"assistant","content":"<think>2+2</think>4"}`, true},
		{"reasoning_details (openrouter)", `{"role":"assistant","content":"4","reasoning_details":[{"type":"reasoning.text","text":"think"}]}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"id":"x","object":"chat.completion","model":"m","choices":[{"index":0,"message":` + tc.msg + `,"finish_reason":"stop"}]}`
			meta, _, _ := promptWithMeta(t, metaBackend{status: 200, body: body}, false)
			require.Equal(t, tc.want, meta.ReasoningSeen)
			require.Nil(t, meta.Usage)
		})
	}
}

func TestResponseMetaStreaming(t *testing.T) {
	chunk := func(delta string) string {
		return `data: {"id":"x","object":"chat.completion.chunk","model":"served-id","system_fingerprint":"fp_s","choices":[{"index":0,"delta":` + delta + `}]}` + "\n\n"
	}
	sse := chunk(`{"role":"assistant","content":"<thi"}`) +
		chunk(`{"content":"nk>plan</think>"}`) +
		chunk(`{"content":"4"}`) +
		`data: {"id":"x","object":"chat.completion.chunk","model":"served-id","choices":[],"usage":{"prompt_tokens":7,"completion_tokens":3,"total_tokens":10}}` + "\n\n" +
		"data: [DONE]\n\n"
	meta, chunks, errResp := promptWithMeta(t, metaBackend{status: 200, headers: map[string]string{"Content-Type": "text/event-stream"}, body: sse}, true)
	require.Nil(t, errResp)
	require.Len(t, chunks, 5)
	require.Equal(t, 200, meta.StatusCode)
	require.Equal(t, "served-id", meta.Model)
	require.Equal(t, "fp_s", meta.SystemFingerprint)
	require.True(t, meta.ReasoningSeen, "<think> split across deltas")
	require.Equal(t, &UsageInfo{PromptTokens: 7, CompletionTokens: 3, TotalTokens: 10}, meta.Usage)
	require.Equal(t, "text/event-stream", meta.Headers["content-type"])
}

func TestResponseMetaStreamingReasoningDelta(t *testing.T) {
	sse := `data: {"id":"x","object":"chat.completion.chunk","model":"m","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"hmm"}}]}` + "\n\n" +
		`data: {"id":"x","object":"chat.completion.chunk","model":"m","choices":[{"index":0,"delta":{"content":"4"}}]}` + "\n\n" +
		"data: [DONE]\n\n"
	meta, _, _ := promptWithMeta(t, metaBackend{status: 200, headers: map[string]string{"Content-Type": "text/event-stream"}, body: sse}, true)
	require.True(t, meta.ReasoningSeen)

	plain := `data: {"id":"x","object":"chat.completion.chunk","model":"m","choices":[{"index":0,"delta":{"content":"4"}}]}` + "\n\ndata: [DONE]\n\n"
	meta, _, _ = promptWithMeta(t, metaBackend{status: 200, headers: map[string]string{"Content-Type": "text/event-stream"}, body: plain}, true)
	require.False(t, meta.ReasoningSeen)
	require.Nil(t, meta.Usage)
}

func TestResponseMetaErrorBody(t *testing.T) {
	body := `{"error":{"message":"This model's maximum context length is 4096 tokens.","type":"invalid_request_error","code":"context_length_exceeded"}}`
	meta, chunks, errResp := promptWithMeta(t, metaBackend{status: 400, headers: map[string]string{"X-Request-Id": "e-1"}, body: body}, false)
	require.NotNil(t, errResp)
	require.Equal(t, 400, errResp.StatusCode)
	require.Empty(t, chunks)
	require.Equal(t, 400, meta.StatusCode)
	require.Equal(t, body, string(meta.ErrorBody))
	require.Equal(t, "e-1", meta.Headers["x-request-id"])
	require.Empty(t, meta.Model)
	require.False(t, meta.ReasoningSeen)
}

func TestResponseMetaErrorBodyCapAndNonJSON(t *testing.T) {
	big := strings.Repeat("x", MaxErrorBodyMeta+100)
	meta, _, errResp := promptWithMeta(t, metaBackend{status: 503, body: big}, false)
	require.NotNil(t, errResp)
	require.Equal(t, 503, meta.StatusCode)
	require.Len(t, meta.ErrorBody, MaxErrorBodyMeta)
}

func TestResponseMetaHeaderValueCapAndJoin(t *testing.T) {
	m := &ResponseMeta{}
	h := http.Header{}
	h.Add("X-RateLimit-Reset", "1")
	h.Add("X-RateLimit-Reset", "2")
	h.Set("Server", strings.Repeat("s", maxHeaderValueMeta*2))
	h.Set("Authorization", "Bearer nope")
	m.recordResponse(429, h)
	require.Equal(t, "1, 2", m.Headers["x-ratelimit-reset"])
	require.Len(t, m.Headers["server"], maxHeaderValueMeta)
	require.NotContains(t, m.Headers, "authorization")
	require.Equal(t, 429, m.StatusCode)
}

func TestResponseMetaAbsentLeavesAdapterUnchanged(t *testing.T) {
	require.Nil(t, responseMetaFrom(context.Background()))
	require.Nil(t, responseMetaFrom(nil))
	ctx := WithResponseMeta(context.Background(), nil)
	require.Nil(t, responseMetaFrom(ctx))
	var m *ResponseMeta
	m.recordResponse(200, nil)
	m.recordCompletion(nil)
	m.recordStreamChunk(nil)
	m.recordError(500, []byte("x"))

	resp := promptViaMockBackend(t, http.StatusBadRequest, `{"error":{"message":"bad"}}`)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
