package lib

import (
	"encoding/json"

	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

// CountTokens counts the number of tokens in a text string using tiktoken
func CountTokens(text string) int {
	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		// Fallback to rough estimation if tokenizer fails
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

// CustomUsageSetter is an interface for types that can set custom usage fields in Extra map
type CustomUsageSetter interface {
	SetCustomUsage(fieldName string, promptTokens, completionTokens int)
}

// SetUsageFromProvider sets the usage_from_provider field with calculated tokens
// Does NOT modify the original "usage" field from the LLM
func SetUsageFromProvider(setter CustomUsageSetter, promptTokens, completionTokens int) {
	if setter == nil {
		return
	}
	setter.SetCustomUsage("usage_from_provider", promptTokens, completionTokens)
}

// SetUsageFromConsumer sets the usage_from_consumer field with calculated tokens
// Does NOT modify the original "usage" field from the LLM
func SetUsageFromConsumer(setter CustomUsageSetter, promptTokens, completionTokens int) {
	if setter == nil {
		return
	}
	setter.SetCustomUsage("usage_from_consumer", promptTokens, completionTokens)
}
