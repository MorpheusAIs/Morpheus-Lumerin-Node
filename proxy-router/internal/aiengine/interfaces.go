package aiengine

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
)

type AIEngineStream interface {
	Prompt(ctx context.Context, prompt *genericchatstorage.OpenAICompletionRequestExtra, cb genericchatstorage.CompletionCallback) error
	AudioTranscription(ctx context.Context, prompt *genericchatstorage.AudioTranscriptionRequest, cb genericchatstorage.CompletionCallback) error
	ApiType() string
}

type ModelParameters map[string]string
