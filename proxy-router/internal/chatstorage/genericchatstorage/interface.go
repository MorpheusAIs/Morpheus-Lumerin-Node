package genericchatstorage

import (
	"time"

	"github.com/sashabaranov/go-openai"
)

type ChatStorageInterface interface {
	LoadChatFromFile(chatID string) (*ChatHistory, error)
	StorePromptResponseToFile(chatID string, isLocal bool, modelID string, prompt interface{}, responses []Chunk, promptAt time.Time, responseAt time.Time) error
	GetChats() []Chat
	DeleteChat(chatID string) error
	UpdateChatTitle(chatID string, title string) error
}

type ChatHistory struct {
	Title    string        `json:"title"`
	ModelId  string        `json:"modelId"`
	IsLocal  bool          `json:"isLocal"`
	Messages []ChatMessage `json:"messages"`
}

func (h *ChatHistory) AppendChatHistory(req *openai.ChatCompletionRequest) *openai.ChatCompletionRequest {
	if h == nil {
		return req
	}

	messagesWithHistory := make([]openai.ChatCompletionMessage, 0)
	for _, chat := range h.Messages {
		// Only append chat completion messages to history, skip audio transcriptions
		if chatReq, ok := chat.Prompt.(OpenAiCompletionRequest); ok {
			messagesWithHistory = append(messagesWithHistory, openai.ChatCompletionMessage{
				Role:    chatReq.Messages[0].Role,
				Content: chatReq.Messages[0].Content,
			})
			messagesWithHistory = append(messagesWithHistory, openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: chat.Response,
			})
		}
	}

	messagesWithHistory = append(messagesWithHistory, req.Messages...)

	// superficial copy to avoid modifying the original request
	newReq := *req
	newReq.Messages = messagesWithHistory
	return &newReq
}

// Helper method to convert openai.ChatCompletionRequest to OpenAiCompletionRequest
func ConvertChatCompletionRequest(prompt *openai.ChatCompletionRequest) OpenAiCompletionRequest {
	messages := make([]ChatCompletionMessage, 0)
	for _, r := range prompt.Messages {
		messages = append(messages, ChatCompletionMessage{
			Content: r.Content,
			Role:    r.Role,
		})
	}

	return OpenAiCompletionRequest{
		Messages:         messages,
		Model:            prompt.Model,
		MaxTokens:        prompt.MaxTokens,
		Temperature:      prompt.Temperature,
		TopP:             prompt.TopP,
		FrequencyPenalty: prompt.FrequencyPenalty,
		PresencePenalty:  prompt.PresencePenalty,
		Stop:             prompt.Stop,
	}
}

type ChatMessage struct {
	Prompt            interface{} `json:"prompt"`
	Response          string      `json:"response"`
	PromptAt          int64       `json:"promptAt"`
	ResponseAt        int64       `json:"responseAt"`
	IsImageContent    bool        `json:"isImageContent"`
	IsVideoRawContent bool        `json:"isVideoRawContent"`
	IsAudioContent    bool        `json:"isAudioContent"`
}

type Chat struct {
	ChatID    string `json:"chatId"`
	ModelID   string `json:"modelId"`
	Title     string `json:"title"`
	IsLocal   bool   `json:"isLocal"`
	CreatedAt int64  `json:"createdAt"`
}

type OpenAiCompletionRequest struct {
	Model            string                        `json:"model"`
	Messages         []ChatCompletionMessage       `json:"messages"`
	MaxTokens        int                           `json:"max_tokens,omitempty"`
	Temperature      float32                       `json:"temperature,omitempty"`
	TopP             float32                       `json:"top_p,omitempty"`
	N                int                           `json:"n,omitempty"`
	Stream           bool                          `json:"stream,omitempty"`
	Stop             []string                      `json:"stop,omitempty"`
	PresencePenalty  float32                       `json:"presence_penalty,omitempty"`
	ResponseFormat   *ChatCompletionResponseFormat `json:"response_format,omitempty"`
	Seed             *int                          `json:"seed,omitempty"`
	FrequencyPenalty float32                       `json:"frequency_penalty,omitempty"`
	// LogitBias is must be a token id string (specified by their token ID in the tokenizer), not a word string.
	// incorrect: `"logit_bias":{"You": 6}`, correct: `"logit_bias":{"1639": 6}`
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat/create-logit_bias
	LogitBias map[string]int `json:"logit_bias,omitempty"`
	// LogProbs indicates whether to return log probabilities of the output tokens or not.
	// If true, returns the log probabilities of each output token returned in the content of message.
	// This option is currently not available on the gpt-4-vision-preview model.
	LogProbs bool `json:"logprobs,omitempty"`
	// TopLogProbs is an integer between 0 and 5 specifying the number of most likely tokens to return at each
	// token position, each with an associated log probability.
	// logprobs must be set to true if this parameter is used.
	TopLogProbs int    `json:"top_logprobs,omitempty"`
	User        string `json:"user,omitempty"`

	// Deprecated: use ToolChoice instead.
	FunctionCall any `json:"function_call,omitempty"`
	// This can be either a string or an ToolChoice object.
	ToolChoice any `json:"tool_choice,omitempty"`
}

type AudioTranscriptionRequest struct {
	Model string

	// FilePath is either an existing file in your filesystem
	FilePath string

	Prompt                 string
	Temperature            float32
	Language               string // Only for transcription.
	Format                 openai.AudioResponseFormat
	TimestampGranularities []openai.TranscriptionTimestampGranularity // Only for transcription.
	TimestampGranularity   openai.TranscriptionTimestampGranularity
}
