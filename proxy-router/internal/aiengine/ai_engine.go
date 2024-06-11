package aiengine

import (
	"context"
	"net/http"
	"os"

	"fmt"

	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/api-gateway/client"
	api "github.com/sashabaranov/go-openai"
)

type AiEngine struct {
	client *api.Client
}

func NewAiEngine() *AiEngine {
	return &AiEngine{
		client: api.NewClientWithConfig(api.ClientConfig{
			BaseURL:    os.Getenv("OPENAI_BASE_URL"),
			APIType:    api.APITypeOpenAI,
			HTTPClient: &http.Client{},
		}),
	}
}

type CompletionCallback func(completion *api.ChatCompletionStreamResponse) error

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

type ChunkSubmit func(*api.ChatCompletionStreamResponse) error

func (aiEngine *AiEngine) PromptStream(ctx context.Context, req interface{}, chunkSubmitCallback interface{}) (*api.ChatCompletionStreamResponse, error) {
	request := req.(*api.ChatCompletionRequest)
	chunkCallback := chunkSubmitCallback.(func(*api.ChatCompletionStreamResponse) error)
	fmt.Println("requesting chat completion stream - request: ", request)
	resp, err := client.RequestChatCompletionStream(ctx, request, func(completion *api.ChatCompletionStreamResponse) error {

		fmt.Println("chunk - response: ", completion)
		return chunkCallback(completion)
	})

	fmt.Println("requested chat completion stream - response: ", resp, "; error: ", err)

	if err != nil {
		fmt.Printf("Stream ChatCompletion error: %v\n", err)
		return nil, err
	}

	return resp, err
}
