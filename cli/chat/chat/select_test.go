package chat

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const selectTestModelID = "0x2222222222222222222222222222222222222222222222222222222222222222"

// newPickerModel builds a model sitting in the remote model picker with one
// selectable model, against the given stub router.
func newPickerModel(t *testing.T, handler http.HandlerFunc) model {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	m := New(&common.Options{
		UseLocalModel: false,
		RemoteModels: []interface{}{
			map[string]interface{}{"Name": "mock-model", "Id": selectTestModelID},
		},
		Client: client.NewApiGatewayClient(srv.URL, http.DefaultClient),
	})
	m.list.Select(0)
	return m
}

// flatten collapses all whitespace, so an assertion about rendered text is
// not defeated by wherever the wrapper chose to break a line.
func flatten(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// awaitSessionOpened runs a command the way bubbletea's loop does, following
// tea.BatchMsg into its children, and returns the first sessionOpened it
// yields.
func awaitSessionOpened(t *testing.T, cmd tea.Cmd) sessionOpened {
	t.Helper()
	if cmd == nil {
		t.Fatal("no command returned; the open was never started")
	}

	switch msg := cmd().(type) {
	case sessionOpened:
		return msg
	case tea.BatchMsg:
		for _, c := range msg {
			if c == nil {
				continue
			}
			if opened, ok := c().(sessionOpened); ok {
				return opened
			}
		}
	}

	t.Fatal("selecting a model produced no sessionOpened message")
	return sessionOpened{}
}

// TestSelectingAModelDoesNotBlockUpdate is the regression test for a picker
// that froze on Enter.
//
// OpenSession was called inside Update, so bubbletea's event loop sat on an
// HTTP request and an on-chain transaction -- no repaint, no keypresses, not
// even ctrl+c. Update must hand the work back as a command instead.
func TestSelectingAModelDoesNotBlockUpdate(t *testing.T) {
	var requests int32

	m := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		// Stands in for the time a real open spends settling on chain.
		time.Sleep(2 * time.Second)
		_, _ = w.Write([]byte(`{"sessionID":"0xopened"}`))
	})

	start := time.Now()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	elapsed := time.Since(start)

	if elapsed > 500*time.Millisecond {
		t.Errorf("Update blocked for %v; the open must run as a command, not inline", elapsed)
	}
	if n := atomic.LoadInt32(&requests); n != 0 {
		t.Errorf("Update itself made %d request(s); it should only return a command", n)
	}

	got := updated.(model)
	if !got.opening {
		t.Error("the model should record that an open is in flight")
	}
	if got.mode != SelectMode {
		t.Errorf("mode = %v, want SelectMode until the session is actually open", got.mode)
	}

	// The command is what performs the open.
	opened := awaitSessionOpened(t, cmd)
	if opened.err != nil {
		t.Fatalf("the open should have succeeded, got %v", opened.err)
	}
	if opened.sessionId != "0xopened" {
		t.Errorf("sessionId = %q, want the opened session", opened.sessionId)
	}
	if n := atomic.LoadInt32(&requests); n != 1 {
		t.Errorf("the command made %d request(s), want exactly 1", n)
	}
}

// TestOpeningStateIsVisible covers the other half of the freeze: with the work
// off the Update loop there has to be something on screen saying so, or a slow
// open still looks like a hang.
func TestOpeningStateIsVisible(t *testing.T) {
	m := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := updated.(model).View()

	if !strings.Contains(flatten(view), "mock-model") {
		t.Errorf("the waiting screen should name the model being opened:\n%s", view)
	}
	if !strings.Contains(strings.ToLower(flatten(view)), "opening") {
		t.Errorf("the waiting screen should say an open is under way:\n%s", view)
	}
}

// TestFailedOpenIsRenderedInThePicker is the regression test for an error that
// was set but never drawn: View only rendered m.err inside its Chat branch, so
// a refused open left the picker looking untouched with no explanation.
func TestFailedOpenIsRenderedInThePicker(t *testing.T) {
	m := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {})
	m.opening = true
	m.openingModel = "mock-model"

	updated, _ := m.Update(sessionOpened{err: errors.New("insufficient allowance")})
	got := updated.(model)

	if got.opening {
		t.Error("a finished attempt should clear the opening flag")
	}
	if got.mode != SelectMode {
		t.Errorf("mode = %v, want SelectMode after a failed open", got.mode)
	}
	if got.err == nil {
		t.Fatal("a failed open must leave an error on the model")
	}

	view := got.View()
	if !strings.Contains(flatten(view), "insufficient allowance") {
		t.Errorf("the picker must render the reason the open failed:\n%s", view)
	}
}

// TestLongOpenErrorIsWrappedNotTruncated guards the tail of a router error.
//
// lipgloss does not wrap without an explicit width, so a long message was cut
// off at the right edge of the terminal. Router errors put the revert reason
// last, so the cut removed the only part worth reading:
//
//	Could not open a session: internal error: unexpected status code: 500,
//	response: failed to send transaction: open session failed
//
// is already 128 characters before the reason even begins.
func TestLongOpenErrorIsWrappedNotTruncated(t *testing.T) {
	sized, _ := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {}).
		Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m := sized.(model)

	const reason = "SessionStakeTooLow"
	m.err = errors.New("internal error: unexpected status code: 500, response: " +
		"failed to send transaction: open session failed: execution reverted: " + reason)

	view := m.View()
	if !strings.Contains(flatten(view), reason) {
		t.Errorf("the end of a long error must survive wrapping, %q is missing:\n%s", reason, view)
	}

	// Wrapping, not just a very long line: no rendered line may exceed the
	// terminal width, or the terminal itself does the truncating.
	for _, line := range strings.Split(view, "\n") {
		if w := lipgloss.Width(line); w > 100 {
			t.Errorf("rendered line is %d cols wide, past the 100-col terminal: %q", w, line)
		}
	}
}

// TestAnyKeyDismissesTheOpenError lets the user get back to the list after a
// failure instead of being stuck on the error screen.
func TestAnyKeyDismissesTheOpenError(t *testing.T) {
	m := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {})
	m.err = errors.New("insufficient allowance")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	got := updated.(model)

	if got.err != nil {
		t.Errorf("a keypress should dismiss the error, still have %v", got.err)
	}
	if strings.Contains(flatten(got.View()), "insufficient allowance") {
		t.Error("the error should be gone from the view once dismissed")
	}
}

// TestSuccessfulOpenEntersChat is the happy path: the session id from the
// command is what the chat is then held against.
func TestSuccessfulOpenEntersChat(t *testing.T) {
	m := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {})
	m.opening = true

	updated, _ := m.Update(sessionOpened{sessionId: "0xopened"})
	got := updated.(model)

	if got.mode != Chat {
		t.Errorf("mode = %v, want Chat once the session is open", got.mode)
	}
	if got.sessionId != "0xopened" {
		t.Errorf("sessionId = %q, want the opened session", got.sessionId)
	}
	if got.opening {
		t.Error("the opening flag should be cleared once the session is open")
	}
}

// TestSecondEnterDoesNotStartASecondOpen matters more than most guards here:
// each open escrows MOR, so a double keypress must not buy two sessions.
func TestSecondEnterDoesNotStartASecondOpen(t *testing.T) {
	m := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {})

	updated, firstCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if firstCmd == nil {
		t.Fatal("the first Enter should start an open")
	}

	_, secondCmd := updated.(model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if secondCmd != nil {
		t.Error("a second Enter while an open is in flight must not start another")
	}
}

// TestCtrlCStillQuitsWhileOpening: being able to give up is the whole point of
// getting the call off the Update loop.
func TestCtrlCStillQuitsWhileOpening(t *testing.T) {
	m := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {})
	m.opening = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should return a command while an open is in flight")
	}
	// The quit message is unexported in this bubbletea version, so compare
	// against what tea.Quit itself produces rather than naming the type.
	if cmd() != tea.Quit() {
		t.Error("ctrl+c should quit while an open is in flight")
	}
}

// TestEnterWithNothingSelectedDoesNotPanic covers an empty picker -- no model
// has an active bid, or a filter matched nothing. The selection was asserted
// to item without checking, which panics on the nil the list returns.
func TestEnterWithNothingSelectedDoesNotPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	m := New(&common.Options{
		UseLocalModel: false,
		RemoteModels:  []interface{}{},
		Client:        client.NewApiGatewayClient(srv.URL, http.DefaultClient),
	})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("there is nothing to open with an empty list")
	}
	if updated.(model).opening {
		t.Error("an empty list should not put the model into the opening state")
	}
}

// TestSelectingAModelWritesNothingToStdout guards the third defect here: a
// bare fmt.Println in Update wrote the model name straight to stdout,
// underneath the full-screen program, which left stray lines on screen.
func TestSelectingAModelWritesNothingToStdout(t *testing.T) {
	m := newPickerModel(t, func(w http.ResponseWriter, r *http.Request) {})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	saved := os.Stdout
	os.Stdout = w

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	os.Stdout = saved
	_ = w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("Update wrote %q to stdout; a full-screen program must render through View", out)
	}
}

// TestLocalModelSelectionIsStillImmediate: a local model opens no session, so
// it should go straight to the chat without an in-flight state.
func TestLocalModelSelectionIsStillImmediate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	m := New(&common.Options{
		UseLocalModel: true,
		LocalModels: []interface{}{
			map[string]interface{}{"Name": "llama-local", "Id": "llama"},
		},
		Client: client.NewApiGatewayClient(srv.URL, http.DefaultClient),
	})
	m.list.Select(0)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(model)

	if got.mode != Chat {
		t.Errorf("mode = %v, want Chat straight away for a local model", got.mode)
	}
	if got.modelId != "llama" {
		t.Errorf("modelId = %q, want the selected local model", got.modelId)
	}
	if got.opening {
		t.Error("a local model opens no session, so nothing should be in flight")
	}
}
