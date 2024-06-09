package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/sashabaranov/go-openai"
)

func SkipTestGetProviders(t *testing.T) {
	client := NewApiGatewayClient("http://localhost:8082", http.DefaultClient)
	res, err := client.GetAllProviders(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("providers: ", res)

}

func SkipTestCreateProvider(t *testing.T) {

	mnemonic, err := hdwallet.NewMnemonic(128)

	if err != nil {
		t.Fatal(err)
	}

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)

	if err != nil {
		t.Fatal(err)
	}

	account := wallet.Accounts()[0]

	address := account.Address.Hex()

	if err != nil {
		t.Fatal(err)
	}
	client := NewApiGatewayClient("http://localhost:8082", http.DefaultClient)

	res, err := client.CreateNewProvider(context.Background(), address, 1, "localhost:8082")

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("providers: ", res)
}

func SkipTestHealthCheck(t *testing.T) {
	client := NewApiGatewayClient("http://localhost:8082", &http.Client{})
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

func SkipTestFiles(t *testing.T) {
	client := NewApiGatewayClient("http://localhost:8082", http.DefaultClient)
	_, err := client.GetProxyRouterFiles(context.Background())

	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateAndStreamChatCompletionMessage(t *testing.T) {
	client := NewApiGatewayClient("http://localhost:8082", http.DefaultClient)

	prompt := "What is the meaning of life?"
	messages := []*ChatCompletionMessage{
		{
			Content: "The meaning of life is 42.",
		},
	}

	results := make([]ChatCompletionStreamResponse, 0)

	_, err := client.PromptStream(context.Background(), prompt, messages, func(msg ChatCompletionStreamResponse) error {
		results = append(results, msg)
		fmt.Println("msg: ", msg)
		return nil
	})

	if err != nil && err != io.EOF {
		t.Fatal(err)
	}

	fmt.Println("chat completion result: ", results)

	if len(results) == 0 {
		t.Fatal("invalid chat completion result: ", results)
	}

	if results[0].Choices[0].Delta.Content == "" {
		t.Fatal("invalid chat completion text result")
	}
}

func TestCreateChatCompletionMessage(t *testing.T) {

	os.Setenv("OPENAI_BASE_URL", "http://localhost:8082/v1")

	client := NewApiGatewayClient("http://localhost:8082", http.DefaultClient)

	prompt := "What is the meaning of life?"
	messages := []ChatCompletionMessage{
		{
			Role:    "user",
			Content: "The meaning of life is 42.",
		},
	}

	res, err := client.Prompt(context.Background(), prompt, messages)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("chat completion result: ", res)

	result := res.(openai.ChatCompletionResponse)

	if result.Choices == nil || len(result.Choices) == 0 {
		t.Fatal("invalid chat completion result")
	}

	if result.Choices[0].Message.Content == "" {
		t.Fatal("invalid chat completion text result")
	}
}
