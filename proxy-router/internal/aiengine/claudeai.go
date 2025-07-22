package aiengine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

// ClaudeAIResponse represents the top-level structure of the ClaudeAI JSON response.
type ClaudeAIResponse struct {
	Content      []ClaudeAIContent `json:"content"`
	ID           string            `json:"id"`
	Model        string            `json:"model"`
	Role         string            `json:"role"`
	StopReason   string            `json:"stop_reason"`
	StopSequence *string           `json:"stop_sequence"`
	Type         string            `json:"type"`
	Usage        ClaudeAIUsage     `json:"usage"`
}

// ClaudeAIContent represents each item in the "content" array.
type ClaudeAIContent struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// ClaudeAIUsage represents the usage statistics of the request/response.
type ClaudeAIUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type ClaudeAIStreamResponse struct {
	Type         string                     `json:"type"`
	Delta        ClaudeAIStreamDelta        `json:"delta"`
	ContentBlock ClaudeAIStreamContentBlock `json:"content_block"`
	Message      ClaudeAIStreamMessage      `json:"message"`
}

type ClaudeAIStreamMessage struct {
	ID    string `json:"id"`
	Role  string `json:"role"`
	Model string `json:"model"`
}

type ClaudeAIStreamDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeAIStreamContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

const API_TYPE_CLAUDEAI = "claudeai"

type ClaudeAI struct {
	baseURL   string
	apiKey    string
	modelName string
	client    *http.Client
	log       lib.ILogger
}

func NewClaudeAIEngine(modelName, baseURL, apiKey string, log lib.ILogger) *ClaudeAI {
	if baseURL != "" {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}
	return &ClaudeAI{
		baseURL:   baseURL,
		modelName: modelName,
		apiKey:    apiKey,
		client:    &http.Client{},
		log:       log,
	}
}

func (a *ClaudeAI) Prompt(ctx context.Context, compl *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	compl.Model = a.modelName
	compl.MaxTokens = 1024
	requestBody, err := json.Marshal(compl)
	if err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/messages", bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	if a.apiKey != "" {
		req.Header.Set("x-api-key", a.apiKey)
	}
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)
	req.Header.Set(c.HEADER_CONNECTION, c.CONNECTION_KEEP_ALIVE)
	if compl.Stream {
		req.Header.Set(c.HEADER_ACCEPT, c.CONTENT_TYPE_EVENT_STREAM)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	a.log.Debugf("AI Model responded with status code: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return a.readError(ctx, resp.Body, cb)
	}

	if isContentTypeStream(resp.Header) {
		return a.readStream(ctx, resp.Body, cb)
	}

	return a.readResponse(ctx, resp.Body, cb)
}

func (a *ClaudeAI) readResponse(ctx context.Context, body io.Reader, cb gcs.CompletionCallback) error {
	var compl ClaudeAIResponse
	if err := json.NewDecoder(body).Decode(&compl); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	var openaiCompl gcs.ChatCompletionResponseExtra
	openaiCompl.ID = compl.ID
	openaiCompl.Model = compl.Model
	openaiCompl.Choices = make([]openai.ChatCompletionChoice, len(compl.Content))
	for i, content := range compl.Content {
		openaiCompl.Choices[i].Message.Content = content.Text
		openaiCompl.Choices[i].Message.Role = compl.Role
	}
	openaiCompl.Usage.PromptTokens = compl.Usage.InputTokens
	openaiCompl.Usage.CompletionTokens = compl.Usage.OutputTokens
	openaiCompl.Usage.TotalTokens = compl.Usage.InputTokens + compl.Usage.OutputTokens

	chunk := gcs.NewChunkText(&openaiCompl)
	err := cb(ctx, chunk, nil)
	if err != nil {
		return fmt.Errorf("callback failed: %v", err)
	}

	return nil
}

func (a *ClaudeAI) readError(ctx context.Context, body io.Reader, cb gcs.CompletionCallback) error {
	var aiEngineErrorResponse interface{}
	if err := json.NewDecoder(body).Decode(&aiEngineErrorResponse); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	err := cb(ctx, nil, gcs.NewAiEngineErrorResponse(aiEngineErrorResponse))
	if err != nil {
		return fmt.Errorf("callback failed: %v", err)
	}
	return nil
}

func (a *ClaudeAI) readStream(ctx context.Context, body io.Reader, cb gcs.CompletionCallback) error {
	var model string
	var role string
	var messageID string

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, StreamDataPrefix) {
			data := line[len(StreamDataPrefix):] // Skip the "data: " prefix

			var compl ClaudeAIStreamResponse
			if err := json.Unmarshal([]byte(data), &compl); err != nil {
				return fmt.Errorf("error decoding response: %s\n%s", err, line)
			}
			if compl.Type == "message_stop" {
				return nil
			}

			if compl.Message.ID != "" {
				messageID = compl.Message.ID
			}
			if compl.Message.Role != "" {
				role = compl.Message.Role
			}
			if compl.Message.Model != "" {
				model = compl.Message.Model
			}
			if compl.Delta.Text != "" || compl.ContentBlock.Text != "" {
				openaiCompl := gcs.ChatCompletionStreamResponseExtra{}
				openaiCompl.Choices = make([]openai.ChatCompletionStreamChoice, 1)
				openaiCompl.Choices[0].Delta.Content = compl.Delta.Text
				openaiCompl.Choices[0].Delta.Role = role
				openaiCompl.ID = messageID
				openaiCompl.Model = model
				openaiCompl.Created = time.Now().Unix()

				// Call the callback function with the unmarshalled completion
				chunk := gcs.NewChunkStreaming(&openaiCompl)
				err := cb(ctx, chunk, nil)
				if err != nil {
					return fmt.Errorf("callback failed: %v", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}

	return nil
}

func (a *ClaudeAI) AudioTranscription(ctx context.Context, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	return fmt.Errorf("audio transcription not supported")
}

func (a *ClaudeAI) AudioSpeech(ctx context.Context, prompt *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	return fmt.Errorf("audio speech not supported")
}

func (a *ClaudeAI) ApiType() string {
	return API_TYPE_CLAUDEAI
}

var _ AIEngineStream = &OpenAI{}
