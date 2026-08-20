package genericchatstorage

import (
	"encoding/json"
	"strings"
)

type ChatMessagePartType string

const (
	ChatMessagePartTypeText     ChatMessagePartType = "text"
	ChatMessagePartTypeImageURL ChatMessagePartType = "image_url"
)

type ImageURLDetail string

const (
	ImageURLDetailHigh ImageURLDetail = "high"
	ImageURLDetailLow  ImageURLDetail = "low"
	ImageURLDetailAuto ImageURLDetail = "auto"
)

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// MultiContent []ChatMessagePart `json:"multiContent",omitempty`

	// This property isn't in the official documentation, but it's in
	// the documentation for the official library for python:
	// - https://github.com/openai/openai-python/blob/main/chatml.md
	// - https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	Name string `json:"name,omitempty"`

	// For Role=tool prompts this should be set to the ID given in the assistant's prior request to call a tool.
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// A stored multi-content (vision) prompt has "content" as an ARRAY of typed
// parts — openai.ChatCompletionMessage marshals MultiContent that way — while
// plain prompts store a string. Unmarshalling an array into `Content string`
// errors, which made asCompletionRequest reject the whole request and silently
// drop the turn (text response included) from replayed history. Accept both
// shapes: keep a string as-is, flatten an array to its "text" parts.
func (m *ChatCompletionMessage) UnmarshalJSON(data []byte) error {
	type plain ChatCompletionMessage // method-free alias to avoid recursion
	aux := struct {
		*plain
		Content json.RawMessage `json:"content"`
	}{plain: (*plain)(m)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Content) == 0 || string(aux.Content) == "null" {
		m.Content = ""
		return nil
	}

	var s string
	if err := json.Unmarshal(aux.Content, &s); err == nil {
		m.Content = s
		return nil
	}

	var parts []struct {
		Type ChatMessagePartType `json:"type"`
		Text string              `json:"text"`
	}
	if err := json.Unmarshal(aux.Content, &parts); err != nil {
		return err
	}
	texts := make([]string, 0, len(parts))
	for _, p := range parts {
		if p.Type == ChatMessagePartTypeText && p.Text != "" {
			texts = append(texts, p.Text)
		}
	}
	m.Content = strings.Join(texts, "\n")
	return nil
}

type ChatCompletionDelta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
	Role             string `json:"role"`
}

type ChatCompletionResponseFormat struct {
	Type ChatCompletionResponseFormatType `json:"type,omitempty"`
}

type ChatCompletionResponseFormatType string
