package proxyapi

import (
	"encoding/json"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

func TestProcessDecisionsRequiresTypeDiscriminant(t *testing.T) {
	payload := []byte(`{"state":"s","questions":{"q":{"type":"noul"}}}`)
	req, err := processDecisions(payload, &lib.LoggerMock{})
	if err != nil {
		t.Fatal(err)
	}
	if req != nil {
		t.Fatal("without type=decisions must return nil (not chat-steal)")
	}
}

func TestProcessDecisionsServerType(t *testing.T) {
	payload := []byte(`{
		"type":"decisions",
		"state":"Ticket dispute",
		"questions":{"urgent":{"type":"noul"}},
		"provider":{"order":["typesafe"]}
	}`)
	req, err := processDecisions(payload, &lib.LoggerMock{})
	if err != nil {
		t.Fatal(err)
	}
	if req == nil {
		t.Fatal("expected decisions request")
	}
	if string(req.State) != `"Ticket dispute"` {
		t.Fatalf("state=%s", req.State)
	}
	if _, ok := req.Extra["type"]; ok {
		t.Fatal("type should be stripped before adapter dispatch")
	}
	if req.Extra["provider"] == nil {
		t.Fatal("unknown fields should remain in Extra")
	}
	// Ensure marshal does not reintroduce a trusted-looking wrong type from client
	out, _ := json.Marshal(req)
	var m map[string]interface{}
	_ = json.Unmarshal(out, &m)
	if _, ok := m["type"]; ok {
		t.Fatalf("type should be absent after strip: %v", m["type"])
	}
}
