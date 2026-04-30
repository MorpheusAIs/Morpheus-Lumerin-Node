package mobile

import (
	"context"
	"encoding/json"
	"fmt"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/ethereum/go-ethereum/common"
	openai "github.com/sashabaranov/go-openai"
)

// StreamCallback receives streaming chunks from a chat completion.
// text is the content delta (or reasoning delta when isThinking is true),
// isLast is true on the final chunk.
type StreamCallback func(text string, isThinking bool, isLast bool) error

// ChatParams holds optional generation parameters forwarded to the provider.
// Nil fields are left at their zero values (provider uses its own defaults).
type ChatParams struct {
	Temperature      *float32
	TopP             *float32
	MaxTokens        *int
	FrequencyPenalty *float32
	PresencePenalty  *float32
}

// SendPrompt sends a chat completion request over an active session.
// If stream is true, the provider may return SSE chunks; otherwise a single JSON completion.
// In both cases, deltas are delivered through cb until the response is complete.
func (s *SDK) SendPrompt(ctx context.Context, sessionID string, prompt string, stream bool, cb StreamCallback) error {
	return s.SendPromptWithMessages(ctx, sessionID, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	}, stream, cb)
}

// SendPromptWithMessages sends a full chat transcript (OpenAI roles) over an active session.
// Use this when the app persists history locally and must re-supply prior turns after restart.
func (s *SDK) SendPromptWithMessages(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessage, stream bool, cb StreamCallback) error {
	if err := s.checkClosed(); err != nil {
		return err
	}
	id := common.HexToHash(sessionID)
	if err := s.ensureProviderForSession(ctx, id); err != nil {
		return err
	}

	req := &gcs.OpenAICompletionRequestExtra{}
	req.Model = sessionID
	req.Messages = messages
	req.Stream = stream

	internalCB := s.buildChatCallback(cb)

	_, err := s.proxySender.SendPromptV2(ctx, id, req, internalCB)
	return err
}

// SendPromptWithMessagesAndParams is like SendPromptWithMessages but allows
// setting generation parameters (temperature, top_p, max_tokens, etc.).
// Pass nil for params to get the default behavior.
// Returns the last non-control chunk's Data() serialised as JSON so the
// caller gets the full, unfiltered provider response metadata.
func (s *SDK) SendPromptWithMessagesAndParams(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessage, stream bool, params *ChatParams, cb StreamCallback) (json.RawMessage, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}
	id := common.HexToHash(sessionID)
	if err := s.ensureProviderForSession(ctx, id); err != nil {
		return nil, err
	}

	req := &gcs.OpenAICompletionRequestExtra{}
	req.Model = sessionID
	req.Messages = messages
	req.Stream = stream

	if params != nil {
		if params.Temperature != nil {
			req.Temperature = *params.Temperature
		}
		if params.TopP != nil {
			req.TopP = *params.TopP
		}
		if params.MaxTokens != nil {
			req.MaxTokens = *params.MaxTokens
		}
		if params.FrequencyPenalty != nil {
			req.FrequencyPenalty = *params.FrequencyPenalty
		}
		if params.PresencePenalty != nil {
			req.PresencePenalty = *params.PresencePenalty
		}
	}

	var lastChunkData interface{}
	baseCB := s.buildChatCallback(cb)
	internalCB := func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		if errResp == nil && chunk != nil && chunk.Type() != gcs.ChunkTypeControl {
			lastChunkData = chunk.Data()
		}
		return baseCB(ctx, chunk, errResp)
	}

	_, err := s.proxySender.SendPromptV2(ctx, id, req, internalCB)
	if err != nil {
		return nil, err
	}

	var raw json.RawMessage
	if lastChunkData != nil {
		b, merr := json.Marshal(lastChunkData)
		if merr != nil {
			return nil, fmt.Errorf("marshal last chunk: %w", merr)
		}
		raw = b
	}
	return raw, nil
}

func (s *SDK) buildChatCallback(cb StreamCallback) func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
	return func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		if errResp != nil {
			return fmt.Errorf("provider error: %v", errResp.ProviderModelError)
		}
		if chunk.Type() == gcs.ChunkTypeControl {
			return cb("", false, true)
		}
		isLast := !chunk.IsStreaming()
		text := chunk.String()
		reasoning := extractReasoningContent(chunk)
		if reasoning != "" {
			if err := cb(reasoning, true, false); err != nil {
				return err
			}
		}
		if text != "" || (reasoning == "" && isLast) {
			return cb(text, false, isLast)
		}
		return nil
	}
}

// ensureProviderForSession repopulates in-memory provider URL + pubkey after SDK restart (mobile).
func (s *SDK) ensureProviderForSession(ctx context.Context, sessionID common.Hash) error {
	sess, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}
	provAddr := sess.ProviderAddr()
	u, err := s.sessionStorage.GetUser(provAddr.Hex())
	if err != nil {
		return fmt.Errorf("provider cache: %w", err)
	}
	if u != nil {
		return nil
	}
	p, err := s.blockchain.GetProvider(ctx, provAddr)
	if err != nil {
		return fmt.Errorf("provider registry: %w", err)
	}
	if p == nil || p.Endpoint == "" {
		return fmt.Errorf("provider endpoint not found for %s", provAddr.Hex())
	}
	return s.proxySender.EnsureProviderRegistered(ctx, provAddr, p.Endpoint)
}

// extractReasoningContent returns reasoning_content from a streaming chunk.
// Uses the typed accessor on ChunkStreaming which reads from the preserved
// original JSON — no re-marshaling, no side effects on the chunk state.
func extractReasoningContent(chunk gcs.Chunk) string {
	if cs, ok := chunk.(*gcs.ChunkStreaming); ok {
		return cs.ReasoningContent()
	}
	return ""
}
