package chat

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/style"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sashabaranov/go-openai"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list        list.Model
	mode        ChatMode
	modelId     string
	sessionId   string
	isLocalChat bool

	textarea textarea.Model
	viewport viewport.Model
	err      error
	client   *client.ApiGatewayClient

	messages           []string
	completionChunkSub chan string // receive chat stream chunks
	messagesContext    []openai.ChatCompletionMessage
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) SetLocalModel(modelId string) model {
	m.mode = Chat
	m.modelId = modelId
	return m
}

func (m model) SetRemoteSession(sessionId string) model {
	m.mode = Chat
	m.sessionId = sessionId
	return m
}

func New(opt *common.Options) model {
	items := []list.Item{}

	if opt.UseLocalModel {
		for i := range opt.LocalModels {
			localModel := opt.LocalModels[i].(map[string]interface{})
			name := localModel["Name"].(string)
			desc := localModel["Id"].(string)
			items = append(items, item{title: name, desc: desc})
		}
	} else {
		for i := range opt.RemoteModels {
			remoteModel := opt.RemoteModels[i].(map[string]interface{})
			name := remoteModel["Name"].(string)
			desc := remoteModel["Id"].(string)
			items = append(items, item{title: name, desc: desc})
		}
	}

	ta := textarea.New()
	ta.Placeholder = "Send your prompt..."
	ta.Focus()

	ta.Prompt = "┃ "

	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = style.Clear
	ta.FocusedStyle.Placeholder = style.Placeholder

	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(78, 15)
	vp.SetContent(common.ChatWelcomeMessage)
	vp.Style = style.Viewport

	m := model{
		list:               list.New(items, list.NewDefaultDelegate(), 0, 0),
		mode:               SelectMode,
		textarea:           ta,
		viewport:           vp,
		err:                nil,
		messages:           []string{},
		completionChunkSub: make(chan string),
		client:             opt.Client,
		isLocalChat:        opt.UseLocalModel,
		messagesContext:    []openai.ChatCompletionMessage{},
	}

	if opt.Session != "" {
		m = m.SetRemoteSession(opt.Session)
	}

	if m.isLocalChat {
		m.list.Title = "Local Models"
	} else {
		m.list.Title = "Remote Models"
	}

	return m
}
