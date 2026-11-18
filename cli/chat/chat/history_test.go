package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"
)

// sseReply writes one content frame and the terminating marker, which is the
// shape the router sends. finishReason is attached to a final frame when set,
// the way a provider reports stopping early.
func sseReply(w http.ResponseWriter, text, finishReason string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	frame := map[string]any{
		"id":      "1",
		"object":  "chat.completion.chunk",
		"choices": []map[string]any{{"index": 0, "delta": map[string]any{"content": text}}},
	}
	body, _ := json.Marshal(frame)
	fmt.Fprintf(w, "data: %s\n", body)

	if finishReason != "" {
		final := map[string]any{
			"id":      "1",
			"object":  "chat.completion.chunk",
			"choices": []map[string]any{{"index": 0, "delta": map[string]any{"content": ""}, "finish_reason": finishReason}},
		}
		body, _ = json.Marshal(final)
		fmt.Fprintf(w, "data: %s\n", body)
	}

	fmt.Fprint(w, "data: [DONE]\n")
}

// recorder captures the decoded request body of every completion call, so a
// test can assert on the conversation actually put on the wire.
type recorder struct {
	mu   sync.Mutex
	sent [][]openai.ChatCompletionMessage
}

func (r *recorder) record(body io.Reader) {
	var req struct {
		Messages []openai.ChatCompletionMessage `json:"messages"`
	}
	_ = json.NewDecoder(body).Decode(&req)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sent = append(r.sent, req.Messages)
}

func (r *recorder) nth(i int) []openai.ChatCompletionMessage {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i >= len(r.sent) {
		return nil
	}
	return r.sent[i]
}

func (r *recorder) calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sent)
}

// chatModelWith builds a model already in Chat mode against a stub router that
// records request bodies, which is the state `chat --session <id>` starts in.
func chatModelWith(t *testing.T, reply, finishReason string) (model, *recorder) {
	t.Helper()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Body)
		sseReply(w, reply, finishReason)
	}))
	t.Cleanup(srv.Close)

	m := New(&common.Options{
		Session: "0xsession",
		Client:  client.NewApiGatewayClient(srv.URL, http.DefaultClient),
	})
	if m.mode != Chat {
		t.Fatalf("a session id should start the model in Chat mode, got %v", m.mode)
	}
	return m, rec
}

// runTurn drives one full turn the way the bubbletea loop does: type the
// prompt, press Enter, then drain the stream events back into Update.
func runTurn(t *testing.T, m model, prompt string) model {
	t.Helper()

	m.textarea.SetValue(prompt)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)

	if m.pending != prompt {
		t.Fatalf("the prompt should be held while the turn is in flight, got %q", m.pending)
	}

	go streamCompletions(m, prompt)()
	for i := 0; i < 20; i++ {
		ev, ok := waitForCompletion(m.completionChunkSub)().(streamEvent)
		if !ok {
			t.Fatal("expected a streamEvent from the waiting command")
		}
		updated, _ := m.Update(ev)
		m = updated.(model)
		if ev.done || ev.err != nil {
			return m
		}
	}
	t.Fatal("the turn never ended")
	return m
}

// TestCompletedTurnJoinsTheTranscript is the base of the feature: a finished
// exchange has to be recorded, with both halves and in order.
func TestCompletedTurnJoinsTheTranscript(t *testing.T) {
	m, _ := chatModelWith(t, "Marseille", "stop")
	m = runTurn(t, m, "name a city")

	if len(m.messagesContext) != 2 {
		t.Fatalf("a completed turn should add exactly 2 messages, got %d: %+v",
			len(m.messagesContext), m.messagesContext)
	}
	if got := m.messagesContext[0]; got.Role != openai.ChatMessageRoleUser || got.Content != "name a city" {
		t.Errorf("first message should be the user prompt, got %+v", got)
	}
	if got := m.messagesContext[1]; got.Role != openai.ChatMessageRoleAssistant || got.Content != "Marseille" {
		t.Errorf("second message should be the assistant reply, got %+v", got)
	}
	if m.pending != "" {
		t.Errorf("the held prompt should be cleared once recorded, got %q", m.pending)
	}
}

// TestSecondTurnResendsTheFirst is the regression test for the issue itself:
// every request used to carry exactly one message, so the model could never
// refer back to anything.
func TestSecondTurnResendsTheFirst(t *testing.T) {
	m, rec := chatModelWith(t, "Marseille", "stop")

	m = runTurn(t, m, "name a city")
	m = runTurn(t, m, "what did you just say")

	if rec.calls() != 2 {
		t.Fatalf("expected 2 completion calls, got %d", rec.calls())
	}

	first := rec.nth(0)
	if len(first) != 1 {
		t.Errorf("the opening turn should send just the prompt, got %d messages", len(first))
	}

	second := rec.nth(1)
	if len(second) != 3 {
		t.Fatalf("the second turn should send the earlier exchange plus the new prompt "+
			"(3 messages), got %d: %+v", len(second), second)
	}
	want := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "name a city"},
		{Role: openai.ChatMessageRoleAssistant, Content: "Marseille"},
		{Role: openai.ChatMessageRoleUser, Content: "what did you just say"},
	}
	for i, w := range want {
		if second[i].Role != w.Role || second[i].Content != w.Content {
			t.Errorf("message %d = %+v, want %+v", i, second[i], w)
		}
	}
}

// TestFailedTurnLeavesNoTrace: a refused turn must not leave a dangling prompt
// that the next request resends as though it had been answered.
func TestFailedTurnLeavesNoTrace(t *testing.T) {
	rec := &recorder{}
	fail := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Body)
		if fail {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid username or password"}`))
			return
		}
		sseReply(w, "Marseille", "stop")
	}))
	t.Cleanup(srv.Close)

	m := New(&common.Options{
		Session: "0xsession",
		Client:  client.NewApiGatewayClient(srv.URL, http.DefaultClient),
	})

	m = runTurn(t, m, "this one fails")
	if m.err == nil {
		t.Fatal("a refused turn should surface an error")
	}
	if len(m.messagesContext) != 0 {
		t.Errorf("a failed turn must not be recorded, got %+v", m.messagesContext)
	}
	if m.pending != "" {
		t.Errorf("the held prompt should be dropped on failure, got %q", m.pending)
	}

	// The next turn must carry only itself.
	fail = false
	m = runTurn(t, m, "this one works")

	last := rec.nth(rec.calls() - 1)
	if len(last) != 1 {
		t.Fatalf("the turn after a failure should send only its own prompt, got %d: %+v",
			len(last), last)
	}
	if last[0].Content != "this one works" {
		t.Errorf("sent %q, want the new prompt only", last[0].Content)
	}
}

// TestStopReasonIsSurfaced: a reply cut off at the provider's limit otherwise
// reads as a complete, oddly short answer.
func TestStopReasonIsSurfaced(t *testing.T) {
	m, _ := chatModelWith(t, "Marseille is a port in", "length")
	m = runTurn(t, m, "tell me about Marseille")

	if m.stopReason != string(openai.FinishReasonLength) {
		t.Errorf("stopReason = %q, want %q", m.stopReason, openai.FinishReasonLength)
	}

	view := flatten(m.View())
	if !strings.Contains(view, "output limit") {
		t.Errorf("the view should explain that the reply hit the output limit:\n%s", m.View())
	}

	// The partial reply is still worth keeping, and is still recorded.
	if len(m.messagesContext) != 2 {
		t.Fatalf("a truncated turn should still be recorded, got %+v", m.messagesContext)
	}
	if m.messagesContext[1].Content != "Marseille is a port in" {
		t.Errorf("the partial reply should be kept, got %q", m.messagesContext[1].Content)
	}
}

// TestNormalStopShowsNoNotice: "stop", and a provider that reports nothing at
// all, are both ordinary completions. Treating silence as a fault would make
// every reply from such a provider look broken.
func TestNormalStopShowsNoNotice(t *testing.T) {
	for _, reason := range []string{"stop", ""} {
		name := reason
		if name == "" {
			name = "(absent)"
		}
		t.Run(name, func(t *testing.T) {
			m, _ := chatModelWith(t, "Marseille", reason)
			m = runTurn(t, m, "name a city")

			if m.stopReason != "" {
				t.Errorf("finish_reason %q should raise no notice, got %q", reason, m.stopReason)
			}
			if v := flatten(m.View()); strings.Contains(v, "ended early") || strings.Contains(v, "output limit") {
				t.Errorf("no stop notice should be rendered:\n%s", m.View())
			}
		})
	}
}

// TestStopReasonNoticeNamesUnknownReasons guards against swallowing a value
// the library does not know about yet.
func TestStopReasonNoticeNamesUnknownReasons(t *testing.T) {
	notice := stopReasonNotice("some_new_reason")
	if !strings.Contains(notice, "some_new_reason") {
		t.Errorf("an unrecognised reason should still be named, got %q", notice)
	}
}

// TestClearResetsTheModelFacingContext covers the `/c` requirement: clearing
// the screen but not the context would leave a chat that looks empty and still
// resends everything said before it.
func TestClearResetsTheModelFacingContext(t *testing.T) {
	for _, cmd := range []string{"/c", "/clear"} {
		t.Run(cmd, func(t *testing.T) {
			m, rec := chatModelWith(t, "Marseille", "stop")
			m = runTurn(t, m, "name a city")
			if len(m.messagesContext) == 0 {
				t.Fatal("precondition: the first turn should have been recorded")
			}
			m.err = errors.New("stale")
			m.stopReason = "length"

			m.textarea.SetValue(cmd)
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			m = updated.(model)

			if len(m.messagesContext) != 0 {
				t.Errorf("%s should clear the conversation the model sees, got %+v", cmd, m.messagesContext)
			}
			if len(m.messages) != 0 {
				t.Errorf("%s should clear the lines on screen, got %+v", cmd, m.messages)
			}
			if m.err != nil || m.stopReason != "" {
				t.Errorf("%s should clear stale notices, got err=%v stopReason=%q", cmd, m.err, m.stopReason)
			}

			// And the turn after a clear carries only itself.
			m = runTurn(t, m, "fresh start")
			last := rec.nth(rec.calls() - 1)
			if len(last) != 1 {
				t.Errorf("after %s the next turn should send only its own prompt, got %d: %+v",
					cmd, len(last), last)
			}
		})
	}
}

// TestResumedSessionSaysItStartsFresh: a session outlives the process that
// opened it, so reattaching looks like picking a conversation back up. The
// transcript is in memory only, so it is not.
func TestResumedSessionSaysItStartsFresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	m := New(&common.Options{
		Session: "0xsession",
		Client:  client.NewApiGatewayClient(srv.URL, http.DefaultClient),
	})

	view := flatten(m.View())
	for _, want := range []string{"0xsession", "does not remember"} {
		if !strings.Contains(view, want) {
			t.Errorf("a resumed session should say it starts fresh and mention %q:\n%s", want, m.View())
		}
	}
}

// TestFreshChatSaysNothingAboutResuming: the notice belongs only to --session.
func TestFreshChatSaysNothingAboutResuming(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	m := New(&common.Options{
		RemoteModels: []interface{}{map[string]interface{}{"Name": "m", "Id": "0x1"}},
		Client:       client.NewApiGatewayClient(srv.URL, http.DefaultClient),
	})
	m = m.SetRemoteSession("0xopened")

	if v := flatten(m.View()); strings.Contains(v, "does not remember") {
		t.Errorf("a session opened in this run should not claim to be resumed:\n%s", m.View())
	}
}
