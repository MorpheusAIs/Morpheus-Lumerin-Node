package chat

import (
	"context"
	"fmt"
	"strings"

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

type completionMsg string

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

		m.err = m.sendChat(val, func(completion interface{}) error {
			openAiCompletion, ok := (completion).(*openai.ChatCompletionStreamResponse)
			if ok {
				if openAiCompletion.Choices[0].Delta.Content != "[DONE]" {
					newChunk += openAiCompletion.Choices[0].Delta.Content
					m.completionChunkSub <- newChunk
				}
			} else {
				prodiaCompletion := (completion).(map[string]interface{})
				newChunk += prodiaCompletion["imageUrl"].(string)
				m.completionChunkSub <- newChunk
			}

			return nil
		})

		return nil
	}
}

// A command that waits for the next chunk on a channel.
func waitForCompletion(sub chan string) tea.Cmd {
	return func() tea.Msg {
		chunk := <-sub

		return completionMsg(chunk)
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
