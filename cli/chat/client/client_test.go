package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

// TestPromptStreamReportsHTTPFailures covers the case that previously left the
// chat UI waiting forever: the router refuses the request, so the body is not
// an event stream, no "data:" frame is ever found, and the function used to
// return (nil, nil) -- no reply and no error.
func TestPromptStreamReportsHTTPFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "unauthorized",
			status: http.StatusUnauthorized,
			body:   `{"error":"invalid username or password"}`,
			want:   "invalid username or password",
		},
		{
			name:   "rate limited",
			status: http.StatusTooManyRequests,
			body:   `{"error":"over rate limit"}`,
			want:   "over rate limit",
		},
		{
			name:   "server error with an empty body",
			status: http.StatusInternalServerError,
			body:   "",
			want:   "500",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			c := NewApiGatewayClient(srv.URL, http.DefaultClient)

			called := false
			_, err := c.PromptStream(context.Background(), "hello", nil, "0xmodel", "0xsession",
				func(interface{}) error { called = true; return nil })

			if err == nil {
				t.Fatal("want an error for a refused request, got nil (the caller would wait forever)")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q should mention %q", err.Error(), tc.want)
			}
			if called {
				t.Error("the completion callback must not fire for a refused request")
			}
		})
	}
}

// TestPromptStreamSkipsUndecodableFrames covers the nil dereference: a frame
// that does not unmarshal used to be logged and then dereferenced anyway.
func TestPromptStreamSkipsUndecodableFrames(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {not json\n"))
		_, _ = w.Write([]byte(`data: {"id":"1","choices":[{"delta":{"content":"ok"}}]}` + "\n"))
		_, _ = w.Write([]byte("data: [DONE]\n"))
	}))
	defer srv.Close()

	c := NewApiGatewayClient(srv.URL, http.DefaultClient)

	var got string
	_, err := c.PromptStream(context.Background(), "hello", nil, "0xmodel", "0xsession",
		func(completion interface{}) error {
			if r, ok := completion.(*openai.ChatCompletionStreamResponse); ok && len(r.Choices) > 0 {
				got += r.Choices[0].Delta.Content
			}
			return nil
		})
	if err != nil {
		t.Fatalf("an unreadable frame should be skipped, not fatal: %v", err)
	}
	if got != "ok" {
		t.Errorf("got %q, want the readable frame to still be delivered", got)
	}
}

// TestPromptStreamHandlesLargeFrames covers the CLI's instance of the 64 KB
// scanner limit reported for the proxy-router in #933: a single SSE frame
// larger than the default cap used to fail the whole stream with
// "token too long", leaving the caller with no reply at all.
func TestPromptStreamHandlesLargeFrames(t *testing.T) {
	big := strings.Repeat("x", 200*1024) // 200 KB in one frame
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`data: {"id":"1","choices":[{"delta":{"content":"` + big + `"}}]}` + "\n"))
		_, _ = w.Write([]byte("data: [DONE]\n"))
	}))
	defer srv.Close()

	c := NewApiGatewayClient(srv.URL, http.DefaultClient)

	got := 0
	_, err := c.PromptStream(context.Background(), "hello", nil, "0xmodel", "0xsession",
		func(completion interface{}) error {
			if r, ok := completion.(*openai.ChatCompletionStreamResponse); ok && len(r.Choices) > 0 {
				got += len(r.Choices[0].Delta.Content)
			}
			return nil
		})
	if err != nil {
		t.Fatalf("a large frame must not fail the stream: %v", err)
	}
	if got != len(big) {
		t.Errorf("delivered %d bytes, want %d", got, len(big))
	}
}
