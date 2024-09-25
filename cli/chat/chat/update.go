package chat

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.mode == SelectMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if msg.String() == "enter" {
				i := m.list.SelectedItem().(item)
				fmt.Println(i.title)

				if m.isLocalChat {
					m = m.SetLocalModel(i.desc)
				} else {
					defaultSessionDuration := big.NewInt(3600)
					session, err := m.client.OpenSession(context.Background(), &client.SessionRequest{
						ModelId:         i.desc,
						SessionDuration: defaultSessionDuration,
					})
					if err != nil {
						m.err = err
						return m, nil
					}
					m = m.SetRemoteSession(session.SessionId)
				}

				return m, nil
			}
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
		}
	} else if m.mode == Chat {
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

				m.textarea.Reset()

				switch val {
				case "/c", "/clear":
					m.viewport.SetContent(common.ChatWelcomeMessage)
					m.messages = []string{}
				default:
					m.messages = append(m.messages, userMessage(val), botMessage(""))

					m.viewport.SetContent(strings.Join(m.messages, "\n\n") + "\n\n")

					m.viewport.GotoBottom()

					return m, tea.Batch(tiCmd, vpCmd, streamCompletions(m, val), waitForCompletion(m.completionChunkSub))
				}
			}
		case completionMsg:
			newMessage := botMessage(string(msg))
			trimmedMessages := m.messages[:len(m.messages)-1]

			m.messages = append(trimmedMessages, wordWrap(newMessage, 78))
			messageContent := strings.Join(m.messages, "\n\n")
			m.viewport.SetContent(messageContent + "\n\n")

			return m, tea.Batch(tiCmd, vpCmd, waitForCompletion(m.completionChunkSub))
		}

		return m, tea.Batch(tiCmd, vpCmd)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}
