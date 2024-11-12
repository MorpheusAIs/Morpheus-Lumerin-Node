package aiengine

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/completion"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
)

type RemoteModel struct {
	service   ProxyService
	sessionID common.Hash
}

type ProxyService interface {
	SendPromptV2(ctx context.Context, sessionID common.Hash, prompt *openai.ChatCompletionRequest, cb completion.CompletionCallback) (interface{}, error)
}

func (p *RemoteModel) Prompt(ctx context.Context, prompt *openai.ChatCompletionRequest, cb completion.CompletionCallback) error {
	_, err := p.service.SendPromptV2(ctx, p.sessionID, prompt, cb)
	return err
}

func (p *RemoteModel) ApiType() string {
	return "remote"
}

var _ AIEngineStream = &RemoteModel{}
