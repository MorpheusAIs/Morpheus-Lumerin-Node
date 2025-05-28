package genericchatstorage

import (
	"context"
	"encoding/json"

	"github.com/sashabaranov/go-openai"
)

type CompletionCallback func(ctx context.Context, completion Chunk, aiEngineErrorResponse *AiEngineErrorResponse) error

type ChunkType string

const (
	ChunkTypeText                   ChunkType = "text"
	ChunkTypeImage                  ChunkType = "image"
	ChunkTypeVideo                  ChunkType = "video"
	ChunkTypeControl                ChunkType = "control-message"
	ChunkTypeAudioTranscriptionText ChunkType = "audio-transcription-text"
	ChunkTypeAudioTranscriptionJson ChunkType = "audio-transcription-json"
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

type ChunkImageRawContent struct {
	data *ImageRawContentResult
}

func NewChunkImageRawContent(data *ImageRawContentResult) *ChunkImageRawContent {
	return &ChunkImageRawContent{
		data: data,
	}
}

func (c *ChunkImageRawContent) IsStreaming() bool {
	return false
}

func (c *ChunkImageRawContent) Tokens() int {
	return 1
}

func (c *ChunkImageRawContent) Type() ChunkType {
	return ChunkTypeImage
}

func (c *ChunkImageRawContent) String() string {
	return c.data.ImageRawContent
}

func (c *ChunkImageRawContent) Data() interface{} {
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
var _ Chunk = &ChunkImageRawContent{}

type ChunkAudioTranscriptionText struct {
	data string
}

func NewChunkAudioTranscriptionText(data string) *ChunkAudioTranscriptionText {
	return &ChunkAudioTranscriptionText{
		data: data,
	}
}

func (c *ChunkAudioTranscriptionText) IsStreaming() bool {
	return false
}

func (c *ChunkAudioTranscriptionText) Tokens() int {
	return len(c.data)
}

func (c *ChunkAudioTranscriptionText) Type() ChunkType {
	return ChunkTypeAudioTranscriptionText
}

func (c *ChunkAudioTranscriptionText) String() string {
	return c.data
}

func (c *ChunkAudioTranscriptionText) Data() interface{} {
	return c.data
}

var _ Chunk = &ChunkAudioTranscriptionText{}

type ChunkAudioTranscriptionJson struct {
	data interface{}
}

func NewChunkAudioTranscriptionJson(data interface{}) *ChunkAudioTranscriptionJson {
	return &ChunkAudioTranscriptionJson{
		data: data,
	}
}

func (c *ChunkAudioTranscriptionJson) IsStreaming() bool {
	return false
}

func (c *ChunkAudioTranscriptionJson) Tokens() int {
	jsonData, err := json.Marshal(c.data)
	if err != nil {
		return 0
	}

	return len(jsonData)
}

func (c *ChunkAudioTranscriptionJson) Type() ChunkType {
	return ChunkTypeAudioTranscriptionJson
}

func (c *ChunkAudioTranscriptionJson) String() string {
	jsonData, err := json.Marshal(c.data)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

func (c *ChunkAudioTranscriptionJson) Data() interface{} {
	return c.data
}

var _ Chunk = &ChunkAudioTranscriptionJson{}

type AiEngineErrorResponse struct {
	ProviderModelError interface{} `json:"providerModelError"`
}

func NewAiEngineErrorResponse(ProviderModelError interface{}) *AiEngineErrorResponse {
	return &AiEngineErrorResponse{
		ProviderModelError: ProviderModelError,
	}
}
