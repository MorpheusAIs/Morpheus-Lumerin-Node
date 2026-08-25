package genericchatstorage

import (
	"encoding/json"
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

// ChatMessage.Prompt is an `interface{}`, so a history read back from disk holds
// a map[string]interface{}, NOT an OpenAiCompletionRequest. A plain type
// assertion therefore ALWAYS fails after a JSON round-trip — and because the
// `ok` was discarded, it failed silently: every stored turn was dropped and the
// model was handed a conversation with no history. That is why chat memory never
// worked. Accept both shapes.
func asCompletionRequest(prompt interface{}) (OpenAiCompletionRequest, bool) {
	switch p := prompt.(type) {
	case OpenAiCompletionRequest:
		return p, true
	case *OpenAiCompletionRequest:
		if p == nil {
			return OpenAiCompletionRequest{}, false
		}
		return *p, true
	case nil:
		return OpenAiCompletionRequest{}, false
	default:
		// Loaded from JSON: re-marshal the generic map back through the concrete type.
		raw, err := json.Marshal(prompt)
		if err != nil {
			return OpenAiCompletionRequest{}, false
		}
		var req OpenAiCompletionRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return OpenAiCompletionRequest{}, false
		}
		return req, true
	}
}

func (h *ChatHistory) AppendChatHistory(req *OpenAICompletionRequestExtra) *OpenAICompletionRequestExtra {
	if h == nil {
		return req
	}

	messagesWithHistory := make([]openai.ChatCompletionMessage, 0)
	for _, chat := range h.Messages {
		// Only append chat completion messages to history, skip audio transcriptions.
		chatReq, ok := asCompletionRequest(chat.Prompt)
		if !ok || len(chatReq.Messages) == 0 {
			continue
		}

		// The turn the user actually took is the LAST message of the stored
		// prompt. Taking Messages[0] only holds if the client sends exactly one
		// message per request; a client that sends the running transcript
		// (as most OpenAI-compatible clients do — though the bundled desktop
		// UI and CLI send only the latest turn) would otherwise have its
		// FIRST message replayed on every turn.
		last := chatReq.Messages[len(chatReq.Messages)-1]

		// An image-only multi-content turn flattens to empty text. Forwarding
		// it as an empty-content user message can be rejected by strict
		// providers, so skip the pair — the same treatment such turns got
		// before multi-content prompts survived the round-trip at all.
		if last.Content == "" {
			continue
		}

		messagesWithHistory = append(messagesWithHistory, openai.ChatCompletionMessage{
			Role:    last.Role,
			Content: last.Content,
		})
		messagesWithHistory = append(messagesWithHistory, openai.ChatCompletionMessage{
			Role:    "assistant",
			Content: chat.Response,
		})
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
