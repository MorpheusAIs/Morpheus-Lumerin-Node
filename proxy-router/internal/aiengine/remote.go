package aiengine

import (
	"context"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type RemoteModel struct {
	service       ProxyService
	sessionID     common.Hash
	redactSecrets lib.SecretRedactMode
	log           lib.ILogger
}

type ProxyService interface {
	SendPromptV2(ctx context.Context, sessionID common.Hash, prompt *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) (interface{}, error)
	SendAudioTranscriptionV2(ctx context.Context, sessionID common.Hash, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) (interface{}, error)
	SendAudioSpeech(ctx context.Context, sessionID common.Hash, prompt *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) (interface{}, error)
	SendEmbeddings(ctx context.Context, sessionID common.Hash, prompt *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) (interface{}, error)
	GetModelIdSession(ctx context.Context, sessionID common.Hash) (common.Hash, error)
	GetAgentTools(ctx context.Context, sessionID common.Hash) (string, error)
	CallAgentTool(ctx context.Context, sessionID common.Hash, toolName string, input map[string]interface{}) (string, error)
}

// Prompt forwards a chat completion to the provider holding the session.
//
// This is the last point the request is still ours: past SendPromptV2 it is
// signed, encrypted to the provider's public key, and gone. Wallet secrets get
// scrubbed here rather than at the HTTP handler so the bundled local model
// still sees whatever the user typed — nothing leaves the machine on that
// path.
//
// The scrub is in place. With PROXY_FORWARD_CHAT_CONTEXT off (the default) the
// History wrapper hands us the same request it later stores, so local chat
// history holds the redacted text too; with forwarding on it stores its own
// pre-scrub copy. Either way nothing unredacted reaches the provider, which is
// what this guards.
func (p *RemoteModel) Prompt(ctx context.Context, prompt *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	if n := lib.RedactMessagesSecrets(prompt.Messages, p.redactSecrets); n > 0 && p.log != nil {
		// Count only: logging the match would move the secret into the logs.
		p.log.Warnf("redacted %d wallet secret(s) from prompt before sending to provider", n)
	}
	_, err := p.service.SendPromptV2(ctx, p.sessionID, prompt, cb)
	return err
}

func (p *RemoteModel) AudioTranscription(ctx context.Context, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	_, err := p.service.SendAudioTranscriptionV2(ctx, p.sessionID, prompt, cb)
	return err
}

func (p *RemoteModel) AudioSpeech(ctx context.Context, prompt *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	_, err := p.service.SendAudioSpeech(ctx, p.sessionID, prompt, cb)
	return err
}

func (p *RemoteModel) Embeddings(ctx context.Context, prompt *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) error {
	_, err := p.service.SendEmbeddings(ctx, p.sessionID, prompt, cb)
	return err
}

func (p *RemoteModel) ApiType() string {
	return "remote"
}

var _ AIEngineStream = &RemoteModel{}
