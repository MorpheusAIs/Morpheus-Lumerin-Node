package aiengine

import "context"

type AiEngine struct {
}

func NewAiEngine() *AiEngine {
	return &AiEngine{}
}

func (aiEngine *AiEngine) Prompt(ctx context.Context) (string, error) {
	return "Hello!", nil
}
