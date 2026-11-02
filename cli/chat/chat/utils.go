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
	text string
	err  error
	done bool
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
		if err != nil {
			m.completionChunkSub <- streamEvent{err: err}
		} else {
			m.completionChunkSub <- streamEvent{done: true}
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
