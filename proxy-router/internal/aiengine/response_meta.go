package aiengine

import (
	"context"
	"net/http"
	"strings"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
)

// ResponseMeta is what one chat-completion round trip reveals about the
// backend beyond the completion itself: the status, an allow-listed header
// subset, the body's identity fields, and whether reasoning content was
// present. The OpenAI adapter fills it when the request context carries one
// (WithResponseMeta); the health checker reads it after Prompt returns. One
// value serves one request: fields are overwritten, never merged, and the
// value is not safe for concurrent use.
type ResponseMeta struct {
	StatusCode int
	// Headers holds the harvested headers under lower-case names; a repeated
	// header is joined with ", ".
	Headers           map[string]string
	Model             string
	SystemFingerprint string
	Usage             *UsageInfo
	// ReasoningSeen is true when a message or delta carried a non-empty
	// reasoning_content / reasoning / reasoning_details field, or the answer
	// content contained a <think> block (llama.cpp, vLLM without a reasoning
	// parser, Ollama raw mode).
	ReasoningSeen bool
	// ErrorBody is the raw upstream body of a non-2xx response, capped at
	// MaxErrorBodyMeta bytes.
	ErrorBody []byte

	// contentTail keeps the last len(thinkOpen)-1 bytes of streamed content so
	// a <think> tag split across deltas is still seen.
	contentTail string
}

type UsageInfo struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// MaxErrorBodyMeta bounds ErrorBody; every error format apiprobe parses fits
// in the first kilobyte.
const MaxErrorBodyMeta = 8 << 10

// maxHeaderValueMeta bounds one harvested header value; the headers of
// interest are short identifiers and counters.
const maxHeaderValueMeta = 256

const thinkOpen = "<think>"

type responseMetaKey struct{}

// WithResponseMeta asks the OpenAI adapter to fill m for the request made
// with the returned context. A nil m returns ctx unchanged.
func WithResponseMeta(ctx context.Context, m *ResponseMeta) context.Context {
	if m == nil {
		return ctx
	}
	return context.WithValue(ctx, responseMetaKey{}, m)
}

func responseMetaFrom(ctx context.Context) *ResponseMeta {
	if ctx == nil {
		return nil
	}
	m, _ := ctx.Value(responseMetaKey{}).(*ResponseMeta)
	return m
}

// harvestedHeader is the allow-list of §5: identity and rate-limit headers
// that name the serving stack (LiteLLM, OpenAI, Anthropic) without carrying
// request content.
func harvestedHeader(name string) bool {
	switch name {
	case "server", "content-type", "x-request-id":
		return true
	}
	for _, prefix := range []string{"x-litellm-", "openai-", "x-ratelimit-", "anthropic-"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func (m *ResponseMeta) recordResponse(status int, h http.Header) {
	if m == nil {
		return
	}
	m.StatusCode = status
	m.Headers = nil
	for name, values := range h {
		lower := strings.ToLower(name)
		if !harvestedHeader(lower) || len(values) == 0 {
			continue
		}
		v := strings.Join(values, ", ")
		if len(v) > maxHeaderValueMeta {
			v = v[:maxHeaderValueMeta]
		}
		if m.Headers == nil {
			m.Headers = map[string]string{}
		}
		m.Headers[lower] = v
	}
}

func (m *ResponseMeta) recordCompletion(c *gcs.ChatCompletionResponseExtra) {
	if m == nil || c == nil {
		return
	}
	m.Model = c.Model
	m.SystemFingerprint = c.SystemFingerprint
	if u := c.Usage; u.PromptTokens > 0 || u.CompletionTokens > 0 || u.TotalTokens > 0 {
		m.Usage = &UsageInfo{PromptTokens: u.PromptTokens, CompletionTokens: u.CompletionTokens, TotalTokens: u.TotalTokens}
	}
	if c.ReasoningContent() != "" {
		m.ReasoningSeen = true
	}
	for _, choice := range c.Choices {
		if choice.Message.ReasoningContent != "" || strings.Contains(choice.Message.Content, thinkOpen) {
			m.ReasoningSeen = true
		}
	}
}

func (m *ResponseMeta) recordStreamChunk(c *gcs.ChatCompletionStreamResponseExtra) {
	if m == nil || c == nil {
		return
	}
	if c.Model != "" {
		m.Model = c.Model
	}
	if c.SystemFingerprint != "" {
		m.SystemFingerprint = c.SystemFingerprint
	}
	if u := c.Usage; u != nil {
		m.Usage = &UsageInfo{PromptTokens: u.PromptTokens, CompletionTokens: u.CompletionTokens, TotalTokens: u.TotalTokens}
	}
	if c.ReasoningContent() != "" {
		m.ReasoningSeen = true
	}
	for _, choice := range c.Choices {
		if choice.Delta.ReasoningContent != "" {
			m.ReasoningSeen = true
		}
		if choice.Delta.Content != "" {
			m.noteContent(choice.Delta.Content)
		}
	}
}

func (m *ResponseMeta) noteContent(delta string) {
	window := m.contentTail + delta
	if strings.Contains(window, thinkOpen) {
		m.ReasoningSeen = true
	}
	if keep := len(thinkOpen) - 1; len(window) > keep {
		window = window[len(window)-keep:]
	}
	m.contentTail = window
}

func (m *ResponseMeta) recordError(status int, raw []byte) {
	if m == nil {
		return
	}
	if m.StatusCode == 0 {
		m.StatusCode = status
	}
	if len(raw) > MaxErrorBodyMeta {
		raw = raw[:MaxErrorBodyMeta]
	}
	m.ErrorBody = append([]byte(nil), raw...)
}
