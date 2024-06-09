package client

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type ChatCompletionStreamResponse = openai.ChatCompletionStreamResponse
type ChatCompletionRequest = openai.ChatCompletionRequest
type CompletionCallback = func(completion ChatCompletionStreamResponse) error
type ChatCompletionStreamChoice = openai.ChatCompletionStreamChoice
type ChatCompletionStreamChoiceDelta = openai.ChatCompletionStreamChoiceDelta

// Stream the chats and pass each stream chunck to the callback function
func RequestChatCompletionStream(ctx context.Context, client openai.Client, request *ChatCompletionRequest, callback CompletionCallback) (*ChatCompletionStreamResponse, error) {

	stream, err := client.CreateChatCompletionStream(ctx, *request)

	if err != nil {
		return nil, err
	}

	streamDone := make(chan struct{})

	for {
		completion, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("Stream completed.")
			streamDone <- struct{}{}
		}

		if err != nil {

			fmt.Println("Stream failed. ", err)
			return nil, err
		}

		if completion.Choices[0].Delta.Content == "" {
			fmt.Println("Stream empty.")
			break
		}

		fmt.Println("Stream chunk. ", completion.Choices[0].Delta.Content)
		if err := callback(completion); err != nil {
			return nil, err
		}
	}

	<-streamDone

	return nil, nil
	// requestBody, err := json.Marshal(request)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to encode request: %v", err)
	// }

	// req, err := http.NewRequestWithContext(ctx, "POST", os.Getenv("OPENAI_BASE_URL")+"/chat/completions", bytes.NewReader(requestBody))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create request: %v", err)
	// }

	// apiKey := os.Getenv("OPENAI_API_KEY")
	// if apiKey != "" {
	// 	req.Header.Set("Authorization", "Bearer "+apiKey)
	// }
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Accept", "text/event-stream")
	// req.Header.Set("Connection", "keep-alive")

	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to send request: %v", err)
	// }
	// defer resp.Body.Close()

	// scanner := bufio.NewScanner(resp.Body)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	// Handle the completion of the stream
	// 	if line == "data: [DONE]" {
	// 		fmt.Println("Stream completed.")

	// 		completion := &ChatCompletionStreamResponse{
	// 			Choices: []ChatCompletionStreamChoice{
	// 				{
	// 					Delta: ChatCompletionStreamChoiceDelta{
	// 						Content: "[DONE]",
	// 					},
	// 				},
	// 			},
	// 		}

	// 		return completion, nil
	// 	}

	// 	if strings.HasPrefix(line, "data: ") {
	// 		data := line[6:] // Skip the "data: " prefix
	// 		// fmt.Println("data: ", data)
	// 		var completion ChatCompletionStreamResponse
	// 		if err := json.Unmarshal([]byte(data), &completion); err != nil {
	// 			fmt.Printf("Error decoding response: %v\n", err)
	// 			continue
	// 		}
	// 		// Call the callback function with the unmarshalled completion
	// 		callback(completion)
	// 	}
	// }

	// if err := scanner.Err(); err != nil {
	// 	return nil, fmt.Errorf("error reading stream: %v", err)
	// }

	// return nil, err
}
