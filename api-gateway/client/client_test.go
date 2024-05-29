package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	client := NewApiGatewayClient("http://localhost:8080", &http.Client{})
	res, err := client.HealthCheck(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("healthcheck result: ", res)

	mapResult := res.(map[string]interface{})

	if mapResult["status"] != "healthy" {
		t.Fatal("healthcheck failed")
	}

	if mapResult["version"] == "" {
		t.Fatal("invalid version result")
	}

	if mapResult["uptime"] == "" {
		t.Fatal("invalid uptime result")
	}
}

func TestFiles(t *testing.T) {
	client := NewApiGatewayClient("http://localhost:8080", http.DefaultClient)
	_, err := client.GetProxyRouterFiles(context.Background())

	if err != nil {
		t.Fatal(err)
	}
}

func SkipTestCreateAndStreamChatCompletionMessage(t *testing.T) {
	client := NewApiGatewayClient("http://localhost:8080", http.DefaultClient)

	prompt := "What is the meaning of life?"
	messages := []*ChatCompletionMessage{
		{
			Content: "The meaning of life is 42.",
		},
	}

	res, err := client.PromptStream(context.Background(), prompt, messages, func(msg ChatCompletionStreamResponse) error {
		fmt.Println("chat completion message: ", msg)
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("chat completion result: ", res)

	mapResult := res.(map[string]interface{})

	if mapResult["choices"] == nil {
		t.Fatal("invalid chat completion result")
	}

	if mapResult["choices"].([]interface{})[0].(map[string]interface{})["text"] == "" {
		t.Fatal("invalid chat completion text result")
	}
}

func SkipTestCreateChatCompletionMessage(t *testing.T) {

	os.Setenv("OPENAI_BASE_URL", "http://localhost:8080/v1")

	client := NewApiGatewayClient("http://localhost:8080", http.DefaultClient)

	prompt := "What is the meaning of life?"
	messages := []ChatCompletionMessage{
		{
			Content: "The meaning of life is 42.",
		},
	}

	res, err := client.Prompt(context.Background(), prompt, messages)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("chat completion result: ", res)

	mapResult := res.(map[string]interface{})

	if mapResult["choices"] == nil {
		t.Fatal("invalid chat completion result")
	}

	if mapResult["choices"].([]interface{})[0].(map[string]interface{})["text"] == "" {
		t.Fatal("invalid chat completion text result")
	}
}
