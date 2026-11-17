package chat

import (
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.mode == SelectMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			// An open is in flight. Swallow keys rather than start a second
			// one, which would escrow MOR twice for one selection.
			if m.opening {
				return m, nil
			}
			// A failed open is showing in place of the list; any key
			// dismisses it and returns to the picker.
			if m.err != nil {
				m.err = nil
				return m, nil
			}
			if msg.String() == "enter" {
				// The list is empty when no model has an active bid, and a
				// filter can match nothing. Either way there is no selection
				// to act on, and the bare type assertion panicked.
				i, ok := m.list.SelectedItem().(item)
				if !ok {
					return m, nil
				}

				if m.isLocalChat {
					return m.SetLocalModel(i.desc), nil
				}

				m.opening = true
				m.openingModel = i.title
				return m, tea.Batch(
					openSession(m.client, i.desc, m.sessionDuration),
					m.spinner.Tick,
				)
			}
		case sessionOpened:
			m.opening = false
			if msg.err != nil {
				m.err = msg.err
				return m, nil
			}
			return m.SetRemoteSession(msg.sessionId), nil
		case spinner.TickMsg:
			if !m.opening {
				return m, nil
			}
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.width = msg.Width
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
					// Clear the conversation the model is shown, not just the
					// lines on screen. Leaving messagesContext populated here
					// would mean a cleared-looking chat that still resends
					// everything said before it.
					m.viewport.SetContent(common.ChatWelcomeMessage)
					m.messages = []string{}
					m.messagesContext = nil
					m.pending = ""
					m.stopReason = ""
					m.err = nil
				default:
					m.messages = append(m.messages, userMessage(val), botMessage(""))

					m.viewport.SetContent(strings.Join(m.messages, "\n\n") + "\n\n")

					m.viewport.GotoBottom()

					// Held, not appended: see messagesContext's comment. The
					// turn joins the transcript once there is a reply to pair
					// it with.
					m.pending = val
					m.stopReason = ""

					return m, tea.Batch(tiCmd, vpCmd, streamCompletions(m, val), waitForCompletion(m.completionChunkSub))
				}
			}
		case streamEvent:
			switch {
			case msg.err != nil:
				// Show the failure and stop waiting. The placeholder bubble
				// added for the reply is dropped unless some text arrived
				// before the failure, in which case the partial reply is kept.
				m.err = msg.err
				if n := len(m.messages); n > 0 && m.messages[n-1] == botMessage("") {
					m.messages = m.messages[:n-1]
				}
				// The transcript is left exactly as it was. Dropping the held
				// prompt is what stops a failed turn being resent as though
				// the model had answered it.
				m.pending = ""
				m.viewport.SetContent(strings.Join(m.messages, "\n\n") + "\n\n")
				return m, tea.Batch(tiCmd, vpCmd)

			case msg.done:
				// End of stream: record the exchange, so the next turn carries
				// it and the model can refer back to it. The router holds no
				// conversational state, so this client-side transcript is the
				// only memory the conversation has.
				if m.pending != "" {
					m.messagesContext = append(m.messagesContext,
						openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: m.pending},
						openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: msg.text},
					)
					m.pending = ""
				}
				// A provider that stopped early says so here. Anything other
				// than a normal stop is worth showing: a reply cut off at the
				// token limit otherwise reads as a complete, oddly short
				// answer. Silence is treated as normal, because whether the
				// reason is sent at all is the provider's choice.
				if r := msg.finishReason; r != "" && r != string(openai.FinishReasonStop) {
					m.stopReason = r
				}
				// Nothing more is coming, so do not re-arm the waiting command.
				return m, tea.Batch(tiCmd, vpCmd)
			}

			m.err = nil
			newMessage := botMessage(msg.text)
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
