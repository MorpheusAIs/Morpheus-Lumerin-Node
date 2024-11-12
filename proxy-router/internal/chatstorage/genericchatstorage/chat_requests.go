package genericchatstorage

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

type ChatMessageImageURL struct {
	URL    string         `json:"url,omitempty"`
	Detail ImageURLDetail `json:"detail,omitempty"`
}

type ChatMessagePart struct {
	Type     ChatMessagePartType  `json:"type,omitempty"`
	Text     string               `json:"text,omitempty"`
	ImageURL *ChatMessageImageURL `json:"image_url,omitempty"`
}

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

type ChatCompletionDelta struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type ChatCompletionResponseFormat struct {
	Type ChatCompletionResponseFormatType `json:"type,omitempty"`
}

type ChatCompletionResponseFormatType string
