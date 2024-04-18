package aiengine

import (
	"context"
	"net/http"
	"os"

	"fmt"

	api "github.com/sashabaranov/go-openai"
)

type AiEngine struct {
}

func NewAiEngine() *AiEngine {
	return &AiEngine{}
}

func (aiEngine *AiEngine) Prompt(ctx context.Context, req interface{}) (*api.ChatCompletionResponse, error) {
	request := req.(*api.ChatCompletionRequest)
	client := api.NewClientWithConfig(api.ClientConfig{
		BaseURL:    os.Getenv("OPENAI_BASE_URL"),
		APIType:    api.APITypeOpenAI,
		HTTPClient: &http.Client{},
	})

	response, err := client.CreateChatCompletion(
		ctx,
		*request,
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}
	return &response, nil
}

func (aiEngine *AiEngine) PromptStream(ctx context.Context, req interface{}) (*api.ChatCompletionStreamResponse, error) {
	request := req.(*api.ChatCompletionRequest)
	client := api.NewClientWithConfig(api.ClientConfig{
		BaseURL:    os.Getenv("OPENAI_BASE_URL"),
		APIType:    api.APITypeOpenAI,
		HTTPClient: &http.Client{},
	})

	stream, err := client.CreateChatCompletionStream(
		ctx,
		*request,
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}

	res, err := stream.Recv()

	if err != nil {
		fmt.Printf("ChatCompletion stream receive error: %v\n", err)
		return nil, err
	}

	response := &api.ChatCompletionStreamResponse{}

	for {
		response.Choices = append(response.Choices, res.Choices...)
		res, err = stream.Recv()

		if err != nil {
			fmt.Printf("ChatCompletion stream receive error: %v\n", err)
			break
		}
	}

	return response, err
}
