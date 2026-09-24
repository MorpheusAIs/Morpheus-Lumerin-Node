package genericchatstorage

import (
	"encoding/json"
	"strings"
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

func TestDecisionsRequestStateRawMessageShapes(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantOK  bool
		wantErr string
	}{
		{"string state", `{"state":"hello","questions":{"q":{"type":"noul"}}}`, true, ""},
		{"object state", `{"state":{"ticket":"invoice"},"questions":{"q":{"type":"noul"}}}`, true, ""},
		{"array state", `{"state":["a","b"],"questions":{"q":{"type":"noul"}}}`, true, ""},
		{"missing state", `{"questions":{"q":{"type":"noul"}}}`, false, "state is required"},
		{"null state", `{"state":null,"questions":{"q":{"type":"noul"}}}`, false, "state is required"},
		{"empty string state", `{"state":"","questions":{"q":{"type":"noul"}}}`, false, "state must be non-empty"},
		{"whitespace string", `{"state":"   ","questions":{"q":{"type":"noul"}}}`, false, "state must be non-empty"},
		{"empty questions", `{"state":"ok","questions":{}}`, false, "questions"},
		{"missing questions", `{"state":"ok"}`, false, "questions"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req DecisionsRequest
			if err := json.Unmarshal([]byte(tc.raw), &req); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			err := req.Validate()
			if tc.wantOK {
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				out, mErr := json.Marshal(req)
				if mErr != nil {
					t.Fatal(mErr)
				}
				var round map[string]json.RawMessage
				_ = json.Unmarshal(out, &round)
				if round["state"] == nil {
					t.Fatal("state dropped on marshal")
				}
				// object/array must survive round-trip byte-identical to input state
				var original map[string]json.RawMessage
				_ = json.Unmarshal([]byte(tc.raw), &original)
				if string(round["state"]) != string(original["state"]) {
					t.Fatalf("state round-trip %s != %s", round["state"], original["state"])
				}
				return
			}
			if err == nil {
				t.Fatal("expected Validate error")
			}
			if tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err=%q want substring %q", err.Error(), tc.wantErr)
			}
		})
	}
}
