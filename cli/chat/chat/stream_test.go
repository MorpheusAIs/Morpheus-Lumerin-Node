package chat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
)

// newChatModel builds a model already in Chat mode against a stub router,
// which is the state `chat --session <id>` starts in.
func newChatModel(t *testing.T, handler http.HandlerFunc) model {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return New(&common.Options{
		Session:       "0xsession",
		Client:        client.NewApiGatewayClient(srv.URL, http.DefaultClient),
		UseLocalModel: false,
	})
}

// TestRefusedTurnIsRendered is the regression test for a chat that showed an
// empty reply bubble and then waited forever: a refused completion has to
// reach the UI as a visible error.
func TestRefusedTurnIsRendered(t *testing.T) {
	m := newChatModel(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid username or password"}`))
	})

	if m.mode != Chat {
		t.Fatalf("a session id should start the model in Chat mode, got %v", m.mode)
	}

	// The placeholder bubble the UI adds while waiting for a reply.
	m.messages = append(m.messages, userMessage("hello"), botMessage(""))

	// Run the turn as Update does, then take the event off the channel.
	go streamCompletions(m, "hello")()
	ev, ok := waitForCompletion(m.completionChunkSub)().(streamEvent)
	if !ok {
		t.Fatal("expected a streamEvent from the waiting command")
	}
	if ev.err == nil {
		t.Fatal("a refused request must produce an error event, not silence")
	}
	if !strings.Contains(ev.err.Error(), "invalid username or password") {
		t.Errorf("the error should carry the router's reason, got %v", ev.err)
	}

	// Feed it back in, as the bubbletea loop would.
	updated, _ := m.Update(ev)
	got, ok := updated.(model)
	if !ok {
		t.Fatalf("Update returned %T, want model", updated)
	}

	if got.err == nil {
		t.Fatal("Update must record the error on the model that gets rendered")
	}
	if view := got.View(); !strings.Contains(view, "invalid username or password") {
		t.Errorf("View must show the failure; got:\n%s", view)
	}
	for _, line := range got.messages {
		if line == botMessage("") {
			t.Error("the empty placeholder reply should have been dropped")
		}
	}
}

// TestSuccessfulTurnRenders covers the normal path, so the error handling
// above cannot pass by breaking ordinary replies.
func TestSuccessfulTurnRenders(t *testing.T) {
	m := newChatModel(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`data: {"id":"1","choices":[{"delta":{"content":"pong"}}]}` + "\n"))
		_, _ = w.Write([]byte("data: [DONE]\n"))
	})
	m.messages = append(m.messages, userMessage("ping"), botMessage(""))

	go streamCompletions(m, "ping")()

	first := waitForCompletion(m.completionChunkSub)().(streamEvent)
	if first.err != nil {
		t.Fatalf("unexpected error: %v", first.err)
	}
	if first.text != "pong" {
		t.Errorf("text = %q, want %q", first.text, "pong")
	}

	updated, _ := m.Update(first)
	got, ok2 := updated.(model)
	if !ok2 {
		t.Fatalf("Update returned %T, want model", updated)
	}
	if got.err != nil {
		t.Errorf("a good turn must clear any previous error, got %v", got.err)
	}
	if !strings.Contains(got.View(), "pong") {
		t.Errorf("View should show the reply; got:\n%s", got.View())
	}

	// The stream must then end, so the UI stops waiting for more.
	second := waitForCompletion(m.completionChunkSub)().(streamEvent)
	if !second.done {
		t.Errorf("the stream should finish with a done event, got %+v", second)
	}
}

// TestTurnAlwaysEnds is the leak guard: whether the turn succeeds or fails, a
// final event must arrive, otherwise the waiting command blocks forever.
func TestTurnAlwaysEnds(t *testing.T) {
	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name: "refused",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"over rate limit"}`))
			},
		},
		{
			name: "empty but successful body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("data: [DONE]\n"))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newChatModel(t, tc.handler)
			go streamCompletions(m, "hello")()

			ev := waitForCompletion(m.completionChunkSub)().(streamEvent)
			if ev.err == nil && !ev.done {
				t.Errorf("the turn must end with an error or done, got %+v", ev)
			}
		})
	}
}
