package aiengine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	api "github.com/sashabaranov/go-openai"
)

func SkipAiEngine_Prompt(t *testing.T) {
	os.Setenv("OPENAI_BASE_URL", "http://localhost:11434/v1")

	aiEngine := NewAiEngine()
	ctx := context.Background()
	req := &api.ChatCompletionRequest{
		Model:     "llama2",
		MaxTokens: 100,
		Messages: []api.ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello, I am a test user"},
		}, // This is a test
	}
	resp, err := aiEngine.Prompt(ctx, req)
	if err != nil {
		t.Errorf("Prompt error: %v", err)
	}
	fmt.Printf("Prompt response: %+v\n", resp)
}

func TestAiEngine_PromptStreamAPI(t *testing.T) {
	os.Setenv("OPENAI_BASE_URL", "http://localhost:11434/v1")

	aiEngine := NewAiEngine()
	ctx := context.Background()
	req := &api.ChatCompletionRequest{
		Model: "llama2",
		// MaxTokens: 100,
		Stream: true,
		Messages: []api.ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Hello, I am a test user",
			},
		}, // This is a test
	}

	choicesChannel := make(chan api.ChatCompletionStreamChoice)
	choices := []api.ChatCompletionStreamChoice{}

	go func() {
		aiEngine.PromptStream(ctx, req, func(response *api.ChatCompletionStreamResponse) error {
			fmt.Printf("chunk: %+v", response)
			choicesChannel <- response.Choices[0]

			if response.Choices[0].Delta.Content == "" {
				return errors.New("empty response")
			}
			// fmt.Printf("chunk - no error")
			return nil
		})
	}()

	timeout := time.After(60 * time.Second)
outerLoop:
	for {
		select {
		case choice := <-choicesChannel:
			choices = append(choices, choice)

			if len(choices) >= 1 {
				break outerLoop
			}
		case <-timeout:
			break outerLoop
		}
	}

	// if err != nil {
	// 	t.Errorf("error: %v", err)
	// 	// fmt.Println("error: ", err)
	// }

	if len(choices) == 0 {
		t.Errorf("invalid response: %v", choices)
	}

	// if choices[len(choices)-1].Delta.Content != "[DONE]" {
	// 	t.Errorf("invalid end of stream response: %s", choices[0].Delta.Content)
	// }

	// content := concatenateDeltaContent(choices)

	// if content == "" {
	// 	t.Errorf("content is empty")
	// }

	// if content == "" {
	// 	t.Errorf("content is empty")
	// }

	// if strings.Contains(content, "Hello ") {
	// 	t.Errorf("content is invalid: %v", content)
	// }
}

func concatenateDeltaContent(choices []api.ChatCompletionStreamChoice) string {
	var concatenatedContent string
	for _, choice := range choices {
		concatenatedContent += choice.Delta.Content
	}
	return concatenatedContent
}
