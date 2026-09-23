package aiengine

import (
	"context"
	"strings"
	"testing"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
)

// capturingProxyService records the prompt as SendPromptV2 receives it, which
// is the first moment the request is no longer ours to edit.
type capturingProxyService struct {
	ProxyService
	got *gcs.OpenAICompletionRequestExtra
}

func (s *capturingProxyService) SendPromptV2(_ context.Context, _ common.Hash, prompt *gcs.OpenAICompletionRequestExtra, _ gcs.CompletionCallback) (interface{}, error) {
	s.got = prompt
	return nil, nil
}

func promptWith(content string) *gcs.OpenAICompletionRequestExtra {
	req := &gcs.OpenAICompletionRequestExtra{}
	req.Messages = []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: content},
	}
	return req
}

func TestRemoteModelPromptRedactsBeforeSend(t *testing.T) {
	const key = "4c0883a69102937d6231471b5dbb6204fe512961708279e2f1f0a9d2c1b5c8ea"

	tests := []struct {
		name string
		mode lib.SecretRedactMode
		in   string
		want string
	}{
		{
			name: "bare key is scrubbed under the default mode",
			mode: lib.SecretRedactHybrid,
			in:   "sweep " + key,
			want: "sweep " + lib.RedactedPrivateKey,
		},
		{
			name: "tx hash reaches the provider under hybrid",
			mode: lib.SecretRedactHybrid,
			in:   "explain tx 0x" + key,
			want: "explain tx 0x" + key,
		},
		{
			name: "tx hash is scrubbed under strict",
			mode: lib.SecretRedactStrict,
			in:   "explain tx 0x" + key,
			want: "explain tx " + lib.RedactedPrivateKey,
		},
		{
			name: "off forwards the prompt verbatim",
			mode: lib.SecretRedactOff,
			in:   "sweep " + key,
			want: "sweep " + key,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &capturingProxyService{}
			model := &RemoteModel{service: svc, redactSecrets: tt.mode, log: lib.NewTestLogger()}

			if err := model.Prompt(context.Background(), promptWith(tt.in), nil); err != nil {
				t.Fatalf("Prompt: %v", err)
			}
			if svc.got == nil {
				t.Fatal("SendPromptV2 was never called")
			}
			if got := svc.got.Messages[0].Content; got != tt.want {
				t.Errorf("provider received\n got: %q\nwant: %q", got, tt.want)
			}
			if tt.mode != lib.SecretRedactOff && strings.Contains(svc.got.Messages[0].Content, key) &&
				!strings.Contains(tt.want, key) {
				t.Error("private key leaked to the provider")
			}
		})
	}
}

// A RemoteModel built without an explicit mode must still scrub: the zero
// value of SecretRedactMode is the empty string, which ParseSecretRedactMode
// maps to hybrid, but a struct literal bypasses that. Guard the wiring so an
// unconfigured node never silently ships keys.
func TestAiEngineDefaultsToHybridRedaction(t *testing.T) {
	engine := NewAiEngine(nil, nil, nil, nil, 0, "", lib.NewTestLogger())
	if engine.redactSecrets != lib.SecretRedactHybrid {
		t.Errorf("unset PROXY_REDACT_SECRETS gave %q, want %q", engine.redactSecrets, lib.SecretRedactHybrid)
	}
}
