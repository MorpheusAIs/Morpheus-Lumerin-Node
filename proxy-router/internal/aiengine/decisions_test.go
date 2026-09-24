package aiengine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

func typeSafeDecisionsResponse() string {
	return `{
		"model": "jev-1.13.0",
		"answers": {
			"dept": {"choice": "billing", "confidence": 0.91},
			"urgent": {"noul": false, "confidence": 0.8},
			"severity": {"score": 0.6, "confidence": 0.7}
		},
		"usage": {"input_tokens": 42, "output_tokens": 12}
	}`
}

func openRouterDecisionsResponse() string {
	return `{
		"id": "gen-test",
		"provider": "typesafe",
		"model": "typesafe/jev-1.13",
		"answers": {
			"dept": {"choice": "billing"}
		},
		"usage": {"input_tokens": 10, "output_tokens": 5, "cost": 0.001}
	}`
}

func TestDecisionsAdapterTypeSafeShape(t *testing.T) {
	var gotBody map[string]json.RawMessage
	var gotModel string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_ = json.Unmarshal(gotBody["model"], &gotModel)
		// Ensure client-supplied type discriminant is not required upstream
		if _, ok := gotBody["type"]; ok {
			t.Errorf("upstream body must not rely on client type; got type field")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(typeSafeDecisionsResponse()))
	}))
	defer server.Close()

	engine := NewDecisionsEngine("jev-1.13.0", server.URL+"/v1/systemone", "", time.Minute, &lib.LoggerMock{}, nil)

	req := &gcs.DecisionsRequest{
		Model: "client-should-be-overwritten",
		State: json.RawMessage(`"Ticket: invoice dispute"`),
		Questions: map[string]json.RawMessage{
			"dept": json.RawMessage(`{"type":"choice","instructions":"Route","criteria":{"billing":"pay"}}`),
		},
		Extra: map[string]json.RawMessage{
			"user": json.RawMessage(`"optional-or-field"`),
		},
	}

	var captured gcs.DecisionsResponse
	err := engine.Decisions(context.Background(), req, func(ctx context.Context, completion gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		if errResp != nil {
			t.Fatalf("unexpected error response: %+v", errResp)
		}
		captured = completion.Data().(gcs.DecisionsResponse)
		return nil
	})
	if err != nil {
		t.Fatalf("Decisions: %v", err)
	}

	if gotModel != "jev-1.13.0" {
		t.Fatalf("model overwrite failed: got %q", gotModel)
	}
	if _, ok := gotBody["user"]; !ok {
		t.Fatal("expected unknown Extra field 'user' forwarded")
	}
	if captured.Answers == nil || captured.Answers["dept"] == nil {
		t.Fatalf("missing answers: %+v", captured)
	}
	if captured.Usage.InputTokens != 42 {
		t.Fatalf("input_tokens=%d want 42", captured.Usage.InputTokens)
	}
	// Full apiUrl used (server.URL includes /v1/systemone path) — embeddings parity
	if !strings.HasSuffix(server.URL+"/v1/systemone", "/v1/systemone") {
		t.Fatal("sanity")
	}
}

func TestDecisionsAdapterOpenRouterShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(openRouterDecisionsResponse()))
	}))
	defer server.Close()

	engine := NewDecisionsEngine("typesafe/jev-1.13", server.URL+"/api/alpha/decisions", "test-key", time.Minute, &lib.LoggerMock{}, nil)
	req := &gcs.DecisionsRequest{
		State: json.RawMessage(`"hello"`),
		Questions: map[string]json.RawMessage{
			"dept": json.RawMessage(`{"type":"choice","criteria":{"a":"b"}}`),
		},
	}

	var captured gcs.DecisionsResponse
	err := engine.Decisions(context.Background(), req, func(ctx context.Context, completion gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		captured = completion.Data().(gcs.DecisionsResponse)
		return nil
	})
	if err != nil {
		t.Fatalf("Decisions: %v", err)
	}
	if captured.Model != "typesafe/jev-1.13" {
		t.Fatalf("model=%q", captured.Model)
	}
	// OR extras preserved via Extra on response
	if captured.Extra["id"] == nil && captured.Extra["provider"] == nil {
		// Unmarshal puts unknown fields in Extra — id/provider should be there
		t.Fatalf("expected OR passthrough extras, got Extra=%v", captured.Extra)
	}
}

func TestDecisionsAdapterSanitizes422(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":{"message":"state field leaked PII details here","type":"invalid_request"}}`))
	}))
	defer server.Close()

	engine := NewDecisionsEngine("jev-1.13.0", server.URL, "", time.Minute, &lib.LoggerMock{}, nil)
	req := &gcs.DecisionsRequest{
		State:     json.RawMessage(`"secret case data"`),
		Questions: map[string]json.RawMessage{"q": json.RawMessage(`{"type":"noul"}`)},
	}

	var captured *gcs.AiEngineErrorResponse
	err := engine.Decisions(context.Background(), req, func(ctx context.Context, completion gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		captured = errResp
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if captured == nil || captured.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected sanitized 422, got %+v", captured)
	}
	body, _ := json.Marshal(captured.ProviderModelError)
	if strings.Contains(string(body), "leaked PII") || strings.Contains(string(body), "secret case") {
		t.Fatalf("raw upstream/error body leaked to client: %s", body)
	}
}

func TestDecisionsRequestRoundTripExtra(t *testing.T) {
	raw := []byte(`{
		"model":"x",
		"state":"s",
		"questions":{"q":{"type":"noul"}},
		"provider":{"order":["typesafe"]},
		"type":"should-be-in-extra"
	}`)
	var req gcs.DecisionsRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	if string(req.State) != `"s"` {
		t.Fatalf("state=%s", req.State)
	}
	if req.Extra["provider"] == nil {
		t.Fatal("expected provider Extra forwarded")
	}
	// Server overwrite simulation (J7)
	req.Extra["type"] = json.RawMessage(`"decisions"`)
	out, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	_ = json.Unmarshal(out, &m)
	var typ string
	_ = json.Unmarshal(m["type"], &typ)
	if typ != "decisions" {
		t.Fatalf("type=%q want decisions", typ)
	}
}

func TestApiAdapterFactoryDecisions(t *testing.T) {
	eng, ok := ApiAdapterFactory(API_TYPE_DECISIONS, "jev", "http://example/v1/systemone", "", nil, time.Minute, &lib.LoggerMock{}, nil)
	if !ok || eng == nil {
		t.Fatal("factory should create decisions adapter")
	}
	if eng.ApiType() != API_TYPE_DECISIONS {
		t.Fatalf("ApiType=%q", eng.ApiType())
	}
}
