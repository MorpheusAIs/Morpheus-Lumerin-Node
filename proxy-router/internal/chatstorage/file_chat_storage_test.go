package chatstorage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
)

func TestStorePromptResponseToFile_DecisionsRequest(t *testing.T) {
	dir := t.TempDir()
	cs := NewChatStorage(dir)

	prompt := &gcs.DecisionsRequest{
		Model: "test-model",
		State: "secret-case-state-must-not-be-title",
		Questions: map[string]json.RawMessage{
			"q1": json.RawMessage(`{"type":"boolean"}`),
		},
	}
	resp := gcs.NewChunkDecisions(gcs.DecisionsResponse{
		Model: "test-model",
		Answers: map[string]json.RawMessage{
			"q1": json.RawMessage(`true`),
		},
		Usage: gcs.DecisionsUsage{InputTokens: 1, OutputTokens: 1},
	})

	now := time.Now()
	if err := cs.StorePromptResponseToFile("chat1", true, "modelhex", prompt, []gcs.Chunk{resp}, now, now); err != nil {
		t.Fatalf("StorePromptResponseToFile: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "chat1.json"))
	if err != nil {
		t.Fatalf("read stored chat: %v", err)
	}
	var hist gcs.ChatHistory
	if err := json.Unmarshal(raw, &hist); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if hist.Title != "Decisions" {
		t.Errorf("title = %q, want %q (short title only; J2)", hist.Title, "Decisions")
	}
	if hist.Title == prompt.State {
		t.Error("title must not be full state (J2)")
	}
	if len(hist.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(hist.Messages))
	}
	if hist.Messages[0].Response == "" {
		t.Error("expected non-empty response")
	}
}

func TestStorePromptResponseToFile_EmbeddingsRequest(t *testing.T) {
	dir := t.TempDir()
	cs := NewChatStorage(dir)

	prompt := &gcs.EmbeddingsRequest{}
	now := time.Now()
	if err := cs.StorePromptResponseToFile("emb1", true, "modelhex", prompt, nil, now, now); err != nil {
		t.Fatalf("StorePromptResponseToFile embeddings: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "emb1.json")); err != nil {
		t.Fatalf("expected emb1.json: %v", err)
	}
}
