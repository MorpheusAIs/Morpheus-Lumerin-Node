package genericchatstorage

import (
	"encoding/json"
	"testing"
)

func TestDecisionsResponseORExtras(t *testing.T) {
	raw := []byte(`{
		"model":"typesafe/jev-1.13",
		"answers":{"q":{"noul":true}},
		"usage":{"input_tokens":1,"output_tokens":2,"cost":0.01},
		"id":"gen-1",
		"provider":"typesafe"
	}`)
	var resp DecisionsResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Usage.InputTokens != 1 || resp.Usage.Cost != 0.01 {
		t.Fatalf("usage=%+v", resp.Usage)
	}
	if string(resp.Extra["id"]) != `"gen-1"` {
		t.Fatalf("id Extra=%s", resp.Extra["id"])
	}
	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var round map[string]json.RawMessage
	_ = json.Unmarshal(out, &round)
	if round["provider"] == nil {
		t.Fatal("provider dropped on marshal")
	}
}

func TestChunkDecisionsTokens(t *testing.T) {
	chunk := NewChunkDecisions(DecisionsResponse{
		Usage: DecisionsUsage{InputTokens: 3, OutputTokens: 4},
	})
	if chunk.Tokens() != 7 {
		t.Fatalf("tokens=%d", chunk.Tokens())
	}
	if chunk.Type() != ChunkTypeDecisions {
		t.Fatalf("type=%s", chunk.Type())
	}
	if chunk.IsStreaming() {
		t.Fatal("decisions must be non-streaming")
	}
}
