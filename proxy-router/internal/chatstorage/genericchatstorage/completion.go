package genericchatstorage

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type CompletionCallback func(ctx context.Context, completion Chunk) error

type ChunkType string

const (
	ChunkTypeText    ChunkType = "text"
	ChunkTypeImage   ChunkType = "image"
	ChunkTypeVideo   ChunkType = "video"
	ChunkTypeControl ChunkType = "control-message"
)

type ChunkText struct {
	data        *openai.ChatCompletionResponse
	isStreaming bool
	tokenCount  int
}

func NewChunkText(data *openai.ChatCompletionResponse) *ChunkText {
	return &ChunkText{
		data: data,
	}
}

func (c *ChunkText) IsStreaming() bool {
	return false
}

func (c *ChunkText) Tokens() int {
	return c.data.Usage.CompletionTokens
}

func (c *ChunkText) Type() ChunkType {
	return ChunkTypeText
}

func (c *ChunkText) String() string {
	return c.data.Choices[0].Message.Content
}

func (c *ChunkText) Data() interface{} {
	return c.data
}

type ChunkStreaming struct {
	data *openai.ChatCompletionStreamResponse
}

func NewChunkStreaming(data *openai.ChatCompletionStreamResponse) *ChunkStreaming {
	return &ChunkStreaming{
		data: data,
	}
}

func (c *ChunkStreaming) IsStreaming() bool {
	return true
}

func (c *ChunkStreaming) Tokens() int {
	return len(c.data.Choices)
}

func (c *ChunkStreaming) Type() ChunkType {
	return ChunkTypeText
}

func (c *ChunkStreaming) String() string {
	return c.data.Choices[0].Delta.Content
}

func (c *ChunkStreaming) Data() interface{} {
	return c.data
}

type ChunkControl struct {
	message string
}

func NewChunkControl(message string) *ChunkControl {
	return &ChunkControl{
		message: message,
	}
}

func (c *ChunkControl) IsStreaming() bool {
	return true
}

func (c *ChunkControl) Tokens() int {
	return 0
}

func (c *ChunkControl) Type() ChunkType {
	return ChunkTypeControl
}

func (c *ChunkControl) String() string {
	return ""
}

func (c *ChunkControl) Data() interface{} {
	return c.message
}

type ChunkImage struct {
	data *ImageGenerationResult
}

func NewChunkImage(data *ImageGenerationResult) *ChunkImage {
	return &ChunkImage{
		data: data,
	}
}

func (c *ChunkImage) IsStreaming() bool {
	return false
}

func (c *ChunkImage) Tokens() int {
	return 1
}

func (c *ChunkImage) Type() ChunkType {
	return ChunkTypeImage
}

func (c *ChunkImage) String() string {
	return c.data.ImageUrl
}

func (c *ChunkImage) Data() interface{} {
	return c.data
}

type ChunkVideo struct {
	data *VideoGenerationResult
}

func NewChunkVideo(data *VideoGenerationResult) *ChunkVideo {
	return &ChunkVideo{
		data: data,
	}
}

func (c *ChunkVideo) IsStreaming() bool {
	return false
}

func (c *ChunkVideo) Tokens() int {
	return 1
}

func (c *ChunkVideo) Type() ChunkType {
	return ChunkTypeVideo
}

func (c *ChunkVideo) String() string {
	return c.data.VideoRawContent
}

func (c *ChunkVideo) Data() interface{} {
	return c.data
}

type ChunkBase64Image struct {
	data *ImageBase64Result
}

func NewChunkBase64Image(data *ImageBase64Result) *ChunkBase64Image {
	return &ChunkBase64Image{
		data: data,
	}
}

func (c *ChunkBase64Image) IsStreaming() bool {
	return false
}

func (c *ChunkBase64Image) Tokens() int {
	return 1
}

func (c *ChunkBase64Image) Type() ChunkType {
	return ChunkTypeImage
}

func (c *ChunkBase64Image) String() string {
	return c.data.Base64Image
}

func (c *ChunkBase64Image) Data() interface{} {
	return c.data
}

type Chunk interface {
	IsStreaming() bool
	Tokens() int
	Type() ChunkType
	String() string
	Data() interface{}
}

var _ Chunk = &ChunkText{}
var _ Chunk = &ChunkImage{}
var _ Chunk = &ChunkControl{}
var _ Chunk = &ChunkStreaming{}
var _ Chunk = &ChunkVideo{}
var _ Chunk = &ChunkBase64Image{}
