package aiengine

import (
	"context"

	"github.com/ollama/ollama/api"
)

type AiEngine struct {
}

func NewAiEngine() *AiEngine {
	return &AiEngine{}
}

func (aiEngine *AiEngine) Prompt(ctx context.Context, req interface {}) (*api.ChatResponse, error) {
	request := req.(*api.ChatRequest)
	client, err := api.ClientFromEnvironment()

	if err != nil {
		return nil, err
	}

	var response *api.ChatResponse

	client.Chat(ctx, request, func(res api.ChatResponse) error {
		response = &res
		return nil
	})

	return response, nil
}
