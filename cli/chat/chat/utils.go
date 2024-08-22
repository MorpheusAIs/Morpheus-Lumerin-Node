package chat

import (
	"context"
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/api-gateway/client"
	"github.com/sashabaranov/go-openai"
)

func (m model) sendChat(prompt string, streamResponse client.CompletionCallback) error {
	ctx := context.Background()

	sessionId, err := m.getSessionId()

	if err != nil {
		return err
	}

	m.config.SessionId = sessionId

	m.openaiRequest.Messages = append(m.openaiRequest.Messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: prompt,
	})

	_, err = client.RequestChatCompletionStream(ctx, &m.openaiRequest, streamResponse, m.config.SessionId)

	if err != nil {
		return err
	}

	return nil
}

func (m model) getSessionId() (string, error) {

	ctx := context.Background()
fmt.Printf("model id: %s", m.config.ModelId)
	if m.config.SessionId == "" {
		session, err := m.openaiClient.OpenSession(ctx, &client.SessionRequest{
			ModelId: m.config.ModelId,
		})

		if err != nil {
			fmt.Printf("error opening session: %v\n", err)
			return "", err
		}

		if session != nil {
			m.config.SessionId = session.SessionId
		} else {
			return "", fmt.Errorf("failed to open session.  Session id is empty")
		}
		return session.SessionId, err
	}

	return m.config.SessionId, nil
}
