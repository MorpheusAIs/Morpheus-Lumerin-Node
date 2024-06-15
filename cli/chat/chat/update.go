package chat

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dwisiswant0/chatgptui/common"
	"github.com/dwisiswant0/chatgptui/style"
	"github.com/sashabaranov/go-openai"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			val := m.textarea.Value()
			if val == "" {
				return m, nil
			}

			switch val {
			case "/c", "/clear":
				m.viewport.SetContent(common.ChatWelcomeMessage)
				m.messages = []string{}
			default:
				m.messages = append(m.messages, userMessage(val), botMessage(""))

				m.viewport.SetContent(strings.Join(m.messages, "\n\n"))

				m.textarea.Reset()
				return m, tea.Batch(tiCmd, vpCmd, streamCompletions(m, val), waitForCompletion(m.completionChunkSub))
			}
		}
	case completionMsg:
		// fmt.Println("completion message received: ", msg)

		newMessage := string(msg)
		// m.messageChunks = append(m.messageChunks, newMessage)

		// if m.messageChunks[0] == newMessage {
		newMessage = botMessage(string(msg))
		// }
		// fmt.Println("messages: ", m.messages)
		m.messages = append(m.messages[:len(m.messages)-1], wordWrap(newMessage, 78))

		m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
		// fmt.Println("messages: ", m.messages)
		// fmt.Println("messageChunks: ", m.messageChunks)
		m.textarea.Reset()

		return m, tea.Batch(tiCmd, vpCmd, waitForCompletion(m.completionChunkSub))
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func userMessage(val string) string {
	return fmt.Sprintf("%s %s", style.Sender.Render("👤:"), val)
}

func botMessage(val string) string {
	return fmt.Sprintf("%s %s", style.Response.Render("🤖:"), val)
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

type completionMsg string

// Simulate a process that sends events at an irregular interval in real time.
// In this case, we'll send events on the channel at a random interval between
// 100 to 1000 milliseconds. As a command, Bubble Tea will run this
// asynchronously.
func streamCompletions(m model, val string) tea.Cmd {
	return func() tea.Msg {

		newChunk := ""

		m.err = m.sendChat(val, func(completion *openai.ChatCompletionStreamResponse) error {

			newChunk += completion.Choices[0].Delta.Content

			m.completionChunkSub <- newChunk

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
