package aiengine

import (
	"context"
	"net/http"
	"os"

	"fmt"

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

//TODO: IMPLEMENT RESPONSE BUFFER FLUSHING
func (aiEngine *AiEngine) PromptStream(ctx context.Context, req interface{}) (*api.ChatCompletionStreamResponse, error) {
	request := req.(*api.ChatCompletionRequest)
	fmt.Printf("Chat request: %+v\n", request)
	stream, err := aiEngine.client.CreateChatCompletionStream(
		ctx,
		*request,
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}

	var res api.ChatCompletionStreamResponse
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
