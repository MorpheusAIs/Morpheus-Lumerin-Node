package chat

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"

	"github.com/dwisiswant0/chatgptui/common"
	"github.com/dwisiswant0/chatgptui/style"
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
				m.messages = append(m.messages, fmt.Sprintf("%s %s", style.Sender.Render("👤:"), val))

				err := m.sendChat(val, func(completion openai.ChatCompletionStreamResponse) error {

					m.messages = append(m.messages, fmt.Sprintf(
						"%s %s",
						style.Response.Render("🤖:"),
						lipgloss.NewStyle().Width(78-5).Render(completion.Choices[0].Delta.Content)),
					)
					m.viewport.SetContent(strings.Join(m.messages, "\n"))
					m.viewport.GotoBottom()

					return nil
				})

				if err != nil {
					m.err = err
					return m, nil
				}

			}

			m.textarea.Reset()
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}
