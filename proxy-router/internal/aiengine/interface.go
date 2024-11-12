package aiengine

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/completion"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

type AIEngineStream interface {
	Prompt(ctx context.Context, prompt *openai.ChatCompletionRequest, cb completion.CompletionCallback) error
	ApiType() string
}

func ApiAdapterFactory(apiType string, modelName string, url string, apikey string, log lib.ILogger) (AIEngineStream, bool) {
	switch apiType {
	case "openai":
		return NewOpenAIEngine(modelName, url, apikey, log), true
	case "prodia":
		return NewProdiaEngine(modelName, url, apikey, log), true
	}
	return nil, false
}
