package aiengine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

// Refs #933: default bufio.Scanner rejects a single SSE line over 64 KiB.
func TestReadStreamAcceptsSSELineOverDefaultScanLimit(t *testing.T) {
	content := strings.Repeat("x", bufio.MaxScanTokenSize+1024)
	payload, err := json.Marshal(gcs.ChatCompletionStreamResponseExtra{
		ChatCompletionStreamResponse: openai.ChatCompletionStreamResponse{
			ID: "chunk-1",
			Choices: []openai.ChatCompletionStreamChoice{{
				Delta: openai.ChatCompletionStreamChoiceDelta{Content: content},
			}},
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var body bytes.Buffer
	body.WriteString(StreamDataPrefix)
	body.Write(payload)
	body.WriteByte('\n')
	body.WriteString(StreamDataPrefix)
	body.WriteString(StreamDone)
	body.WriteByte('\n')

	engine := NewOpenAIEngine("test-model", "http://example.invalid", "", 0, &lib.LoggerMock{}, nil)

	var gotContent string
	var sawDone bool
	err = engine.readStream(context.Background(), &body, func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		if errResp != nil {
			t.Fatalf("unexpected error response: %+v", errResp)
		}
		if chunk == nil {
			return nil
		}
		switch chunk.Type() {
		case gcs.ChunkTypeText:
			gotContent += chunk.String()
		case gcs.ChunkTypeControl:
			sawDone = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("readStream: %v", err)
	}
	if gotContent != content {
		t.Fatalf("content length = %d, want %d", len(gotContent), len(content))
	}
	if !sawDone {
		t.Fatal("expected stream done control chunk")
	}
}

func TestNewSSEScannerRejectsNothingUnderRaisedCap(t *testing.T) {
	line := strings.Repeat("y", bufio.MaxScanTokenSize+1)
	scanner := newSSEScanner(strings.NewReader(line + "\n"))
	if !scanner.Scan() {
		t.Fatalf("Scan failed: %v", scanner.Err())
	}
	if got := scanner.Text(); got != line {
		t.Fatalf("got %d bytes, want %d", len(got), len(line))
	}
}
