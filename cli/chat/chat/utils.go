package chat

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"
)

type ChatMode int

const (
	SelectMode ChatMode = iota
	Chat
)

// openSessionTimeout bounds a session open started from the model picker.
//
// The API client is built with a zero-value http.Client, which carries no
// timeout of its own, so without this the request can wait indefinitely on a
// router that has stopped answering. The value is deliberately generous:
// opening a session settles an on-chain transaction, which on a busy network
// takes tens of seconds.
const openSessionTimeout = 3 * time.Minute

// sessionOpened carries the outcome of a session open back to Update.
type sessionOpened struct {
	sessionId string
	err       error
}

// openSession opens a session for the given model off the Update loop.
//
// Opening a session is an HTTP round trip plus an on-chain transaction, so it
// routinely takes tens of seconds. Doing that inside Update blocks bubbletea's
// event loop for the whole duration: no repaint, no keypresses, not even
// ctrl+c, and no way for the user to tell a slow open from a dead one.
func openSession(c *client.ApiGatewayClient, modelId string, duration int64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), openSessionTimeout)
		defer cancel()

		session, err := c.OpenSession(ctx, &client.SessionRequest{
			ModelId:         modelId,
			SessionDuration: big.NewInt(duration),
		})
		if err != nil {
			return sessionOpened{err: err}
		}
		return sessionOpened{sessionId: session.SessionId}
	}
}

// streamEvent is one update from an in-flight completion: more text, a
// failure, or the end of the stream.
//
// Failures travel the same channel as chunks on purpose. The command waiting
// on that channel is the only thing that turns an update into something the UI
// renders, so a failure that never reaches it leaves the chat waiting on a
// reply that is not coming.
type streamEvent struct {
	// text is the reply accumulated so far, not the latest delta. The
	// terminal (done) event repeats it, because by then it is the only clean
	// copy of the reply: what the UI keeps has been word-wrapped and given a
	// speaker prefix, which is not something to send back to a model.
	text string
	err  error
	done bool
	// finishReason is the provider's stated reason for stopping, carried on
	// the done event: "length" when it ran out of room, "content_filter" when
	// it was blocked, and so on. Empty means the provider said nothing, which
	// is treated as a normal completion -- see the done branch in Update.
	finishReason string
}

// WordWrap wraps the given string to fit within a specified width.
func wordWrap(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var wrappedText strings.Builder
	currentLineLength := 0

	for _, word := range words {
		wordLength := len(word)
		if currentLineLength+wordLength+5 > width { // +1 for the space
			if currentLineLength > 0 {
				wrappedText.WriteString("\n")
				currentLineLength = 0
			}
		}

		if currentLineLength > 0 {
			wrappedText.WriteString(" ")
			currentLineLength++
		}
		wrappedText.WriteString(word)
		currentLineLength += wordLength
	}

	return wrappedText.String()
}

func streamCompletions(m model, val string) tea.Cmd {
	return func() tea.Msg {
		newChunk := ""
		finishReason := ""

		// Note this runs on a value copy of the model, so assigning to m.err
		// here would be discarded. The result has to be sent, not stored.
		err := m.sendChat(val, func(completion interface{}) error {
			switch c := (completion).(type) {
			case *openai.ChatCompletionStreamResponse:
				// A frame can legitimately carry no choices; indexing [0]
				// unconditionally panics when one does.
				if len(c.Choices) == 0 {
					return nil
				}
				// The reason arrives on the last frame carrying choices, and
				// that frame's delta is usually empty, so it has to be kept
				// here rather than read off the final chunk later.
				if r := string(c.Choices[0].FinishReason); r != "" {
					finishReason = r
				}
				if content := c.Choices[0].Delta.Content; content != "[DONE]" {
					newChunk += content
					m.completionChunkSub <- streamEvent{text: newChunk}
				}
			case map[string]interface{}:
				url, ok := c["imageUrl"].(string)
				if !ok {
					return nil
				}
				newChunk += url
				m.completionChunkSub <- streamEvent{text: newChunk}
			}

			return nil
		})

		// Always close out the turn, so the waiting command stops waiting:
		// with the error if there was one, otherwise with a plain end-of-stream.
		// The done event carries the whole reply and the stop reason, which is
		// what lets Update record the turn and say why it ended.
		if err != nil {
			m.completionChunkSub <- streamEvent{err: err}
		} else {
			m.completionChunkSub <- streamEvent{done: true, text: newChunk, finishReason: finishReason}
		}

		return nil
	}
}

// A command that waits for the next event on a channel.
func waitForCompletion(sub chan streamEvent) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func userMessage(val string) string {
	return fmt.Sprintf("%s %s", style.Sender.Render("👤:"), val)
}

func botMessage(val string) string {
	return fmt.Sprintf("%s %s", style.Response.Render("🤖:"), val)
}

func (m model) sendChat(prompt string, streamResponse client.CompletionCallback) error {
	ctx := context.Background()

	_, err := m.client.PromptStream(ctx, prompt, m.messagesContext, m.modelId, m.sessionId, streamResponse)

	if err != nil {
		return err
	}

	return nil
}
