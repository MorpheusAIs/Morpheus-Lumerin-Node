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
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		APIType: api.APITypeOpenAI,
		HTTPClient: &http.Client{},
	})
	
	fmt.Printf("client: %+v\r\n", *client)
	fmt.Printf("base url: %+v\r\n", os.Getenv("OPENAI_BASE_URL"))
	fmt.Printf("request: %+v\r\n", *request)

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

