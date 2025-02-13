package client

import (
	"encoding/base64"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type ChatCompletionStreamResponse = openai.ChatCompletionStreamResponse
type ChatCompletionRequest = openai.ChatCompletionRequest
type CompletionCallback = func(completion interface{}) error
type ChatCompletionStreamChoice = openai.ChatCompletionStreamChoice
type ChatCompletionStreamChoiceDelta = openai.ChatCompletionStreamChoiceDelta

func StringTo32Byte(s string) [32]byte {
	var array [32]byte

	// Convert the string to a byte slice
	byteSlice := []byte(s)

	// Copy the byte slice into the array
	copy(array[:], byteSlice)

	return array
}

func FormatBasicAuthHeader(login, pass string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", login, pass)))
	return fmt.Sprintf("Basic %s", encoded)
}
