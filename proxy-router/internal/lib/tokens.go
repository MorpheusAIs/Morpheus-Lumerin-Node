package lib

import (
	"encoding/json"

	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
	"fmt"
)

// CountTokens counts the number of tokens in a text string using tiktoken
func CountTokens(text string) int {
	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		// Fallback to rough estimation if tokenizer fails
		fmt.Println("Error getting tokenizer: ", err)
		return (len(text) + 3) / 4
	}
	ids, _, _ := enc.Encode(text)
	return len(ids)
}

// CountPromptTokens calculates the total tokens in the chat request messages
func CountPromptTokens(messages []openai.ChatCompletionMessage) int {
	var totalTokens int
	for _, msg := range messages {
		// Count role tokens (approximately 1 token for role)
		totalTokens += 1
		// Count content tokens
		if len(msg.MultiContent) > 0 {
			// Handle multi-content messages
			for _, part := range msg.MultiContent {
				if part.Type == openai.ChatMessagePartTypeText {
					totalTokens += CountTokens(part.Text)
				}
				// Note: Image tokens would need special handling based on resolution
			}
		} else {
			totalTokens += CountTokens(msg.Content)
		}
		// Count name tokens if present
		if msg.Name != "" {
			totalTokens += CountTokens(msg.Name)
		}
	}
	// Add overhead tokens for message formatting (approximately 3 tokens per message)
	totalTokens += len(messages) * 3
	return totalTokens
}

// UsageSetter is an interface for types that can set their originalJSON usage field
type UsageSetter interface {
	SetOriginalJSONUsage(usageBytes []byte)
}

// UpdateUsagePtr updates usage in a response where Usage is a pointer field
// It takes a double pointer to usage so it can create if nil
func UpdateUsagePtr(usage **openai.Usage, promptTokens, completionTokens int, setter UsageSetter) {
	if usage == nil {
		return
	}
	if *usage == nil {
		*usage = &openai.Usage{}
	}
	(*usage).PromptTokens = promptTokens
	(*usage).CompletionTokens = completionTokens
	(*usage).TotalTokens = promptTokens + completionTokens

	// Update the originalJSON to ensure the usage is reflected in marshalled output
	if setter != nil {
		usageBytes, err := json.Marshal(*usage)
		if err == nil {
			setter.SetOriginalJSONUsage(usageBytes)
		}
	}
}

// UpdateUsage updates usage in a response where Usage is a value field (not pointer)
func UpdateUsage(usage *openai.Usage, promptTokens, completionTokens int, setter UsageSetter) {
	if usage == nil {
		return
	}
	usage.PromptTokens = promptTokens
	usage.CompletionTokens = completionTokens
	usage.TotalTokens = promptTokens + completionTokens

	// Update the originalJSON to ensure the usage is reflected in marshalled output
	if setter != nil {
		usageBytes, err := json.Marshal(usage)
		if err == nil {
			setter.SetOriginalJSONUsage(usageBytes)
		}
	}
}
