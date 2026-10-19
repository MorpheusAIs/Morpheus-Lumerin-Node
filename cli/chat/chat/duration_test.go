package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	tea "github.com/charmbracelet/bubbletea"
)

// TestSessionDurationReachesTheRequest proves the --duration value is what the
// session is actually opened with, rather than the previously hardcoded 3600.
//
// It selects a model the way a keypress does, so it covers the whole path from
// Options through the model to the request body.
func TestSessionDurationReachesTheRequest(t *testing.T) {
	const modelID = "0x1111111111111111111111111111111111111111111111111111111111111111"

	for _, tc := range []struct {
		name string
		opt  int64
		want float64
	}{
		{name: "explicit duration is used", opt: 900, want: 900},
		{name: "unset falls back to the default", opt: 0, want: float64(common.DefaultSessionDuration)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotDuration float64
			var sawSessionCall bool

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/session") {
					sawSessionCall = true
					var body map[string]interface{}
					_ = json.NewDecoder(r.Body).Decode(&body)
					if v, ok := body["sessionDuration"].(float64); ok {
						gotDuration = v
					}
					_, _ = w.Write([]byte(`{"sessionID":"0xopened"}`))
					return
				}
				w.WriteHeader(http.StatusNotFound)
			}))
			defer srv.Close()

			m := New(&common.Options{
				UseLocalModel:   false,
				RemoteModels:    []interface{}{map[string]interface{}{"Name": "mock-model", "Id": modelID}},
				Client:          client.NewApiGatewayClient(srv.URL, http.DefaultClient),
				SessionDuration: tc.opt,
			})

			if m.sessionDuration != int64(tc.want) {
				t.Fatalf("model carries duration %d, want %d", m.sessionDuration, int64(tc.want))
			}

			// Selecting a model is what opens the session.
			m.list.Select(0)
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			got := updated.(model)

			if !sawSessionCall {
				t.Fatal("selecting a model should have opened a session")
			}
			if got.err != nil {
				t.Fatalf("opening the session failed: %v", got.err)
			}
			if got.sessionId != "0xopened" {
				t.Errorf("sessionId = %q, want the opened session", got.sessionId)
			}
			if gotDuration != tc.want {
				t.Errorf("sessionDuration on the wire = %v, want %v", gotDuration, tc.want)
			}
		})
	}
}

// TestDurationDefault guards the documented default.
func TestDurationDefault(t *testing.T) {
	if common.DefaultSessionDuration != 600 {
		t.Errorf("DefaultSessionDuration = %d, want 600", common.DefaultSessionDuration)
	}
	opt := &common.Options{}
	if got := opt.Duration(); got != 600 {
		t.Errorf("unset Duration() = %d, want 600", got)
	}
	opt.SessionDuration = 3600
	if got := opt.Duration(); got != 3600 {
		t.Errorf("explicit Duration() = %d, want 3600", got)
	}
	opt.SessionDuration = -5
	if got := opt.Duration(); got != 600 {
		t.Errorf("a negative duration should fall back to the default, got %d", got)
	}
}
