package aiengine

import (
	"context"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/ethereum/go-ethereum/common"
)

type RemoteModel struct {
	service   ProxyService
	sessionID common.Hash
}

type ProxyService interface {
	SendPromptV2(ctx context.Context, sessionID common.Hash, prompt *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) (interface{}, error)
	SendAudioTranscriptionV2(ctx context.Context, sessionID common.Hash, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) (interface{}, error)
	GetModelIdSession(ctx context.Context, sessionID common.Hash) (common.Hash, error)
	GetAgentTools(ctx context.Context, sessionID common.Hash) (string, error)
	CallAgentTool(ctx context.Context, sessionID common.Hash, toolName string, input map[string]interface{}) (string, error)
}

func (p *RemoteModel) Prompt(ctx context.Context, prompt *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	_, err := p.service.SendPromptV2(ctx, p.sessionID, prompt, cb)
	return err
}

func (p *RemoteModel) AudioTranscription(ctx context.Context, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	_, err := p.service.SendAudioTranscriptionV2(ctx, p.sessionID, prompt, cb)
	return err
}

func (p *RemoteModel) ApiType() string {
	return "remote"
}

var _ AIEngineStream = &RemoteModel{}
