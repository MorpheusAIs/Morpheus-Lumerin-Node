package genericchatstorage

import (
	"testing"

	"github.com/sashabaranov/go-openai"
)

// Regression: usage-only stream chunks have no choices; String() must not panic.
func TestChunkStreaming_String_UsageOnlyNoChoices(t *testing.T) {
	c := NewChunkStreaming(&ChatCompletionStreamResponseExtra{
		ChatCompletionStreamResponse: openai.ChatCompletionStreamResponse{
			Object:  "chat.completion.chunk",
			Choices: []openai.ChatCompletionStreamChoice{},
			Usage:   &openai.Usage{TotalTokens: 42},
		},
	})
	if got := c.String(); got != "" {
		t.Fatalf("String() = %q, want empty", got)
	}
}

func TestChunkStreaming_String_NilData(t *testing.T) {
	c := &ChunkStreaming{data: nil}
	if got := c.String(); got != "" {
		t.Fatalf("String() = %q, want empty", got)
	}
}
