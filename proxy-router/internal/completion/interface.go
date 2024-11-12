package completion

import "context"

type CompletionCallback func(ctx context.Context, completion *ChunkImpl) error

type ChunkType string

const (
	ChunkTypeText  ChunkType = "text"
	ChunkTypeImage ChunkType = "image"
	ChunkTypeVideo ChunkType = "video"
)

type ChunkImpl struct {
	Data        interface{}
	IsStreaming bool
	Tokens      int
	Type        ChunkType
}
