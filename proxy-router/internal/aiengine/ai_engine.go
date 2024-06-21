package aiengine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	api "github.com/sashabaranov/go-openai"
)

type AiEngine struct {
	client *api.Client
}

type ResponderFlusher interface {
	http.ResponseWriter
	http.Flusher
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

type CompletionCallback func(completion api.ChatCompletionStreamResponse) error

func requestChatCompletionStream(ctx context.Context, request *api.ChatCompletionRequest, callback CompletionCallback) (*api.ChatCompletionStreamResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", os.Getenv("OPENAI_BASE_URL")+"/chat/completions", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Handle the completion of the stream
		// if line == "data: [DONE]" {
		// 	fmt.Println("Stream completed.")

		// 	completion := &api.ChatCompletionStreamResponse{
		// 		Choices: []api.ChatCompletionStreamChoice{
		// 			{
		// 				Delta: api.ChatCompletionStreamChoiceDelta{
		// 					Content: "[DONE]",
		// 				},
		// 			},
		// 		},
		// 	}

		// 	return completion, nil
		// }

		if strings.HasPrefix(line, "data: ") {
			data := line[6:] // Skip the "data: " prefix
			// fmt.Println("data: ", data)
			var completion api.ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &completion); err != nil {
				fmt.Printf("Error decoding response: %v\n", err)
				continue
			}
			// Call the callback function with the unmarshalled completion
			callback(completion)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %v", err)
	}

	return nil, err
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

type ChunkSubmit func(*api.ChatCompletionStreamResponse) error

func (aiEngine *AiEngine) PromptStream(ctx context.Context, req interface{}, chunkSubmitCallback interface{}) (*api.ChatCompletionStreamResponse, error) {
	request := req.(*api.ChatCompletionRequest)
	chunkCallback := chunkSubmitCallback.(func(*api.ChatCompletionStreamResponse) error)

	resp, err := requestChatCompletionStream(ctx, request, func(completion api.ChatCompletionStreamResponse) error {
		return chunkCallback(&completion)
	})

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}

	return resp, err
}

func (aiEngine *AiEngine) PromptCb(ctx *gin.Context, body *openai.ChatCompletionRequest) {
	var response interface{}
	var err error

	if body.Stream {
		response, err = aiEngine.PromptStream(ctx, body, func(response *openai.ChatCompletionStreamResponse) error {

			marshalledResponse, err := json.Marshal(response)
			if err != nil {
				return err
			}

			ctx.Writer.Header().Set("Content-Type", "text/event-stream")

			_, err = ctx.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", marshalledResponse)))
			if err != nil {
				return err
			}

			ctx.Writer.Flush()
			return nil
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, response)
		return
	} else {
		response, err = aiEngine.Prompt(ctx, body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, response)
		return
	}
}
