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

// ChatCompletionRequestExtra is the OpenAI chat completion request type used by
// the embedded gateway. It is exported as an alias so external callers (the
// NodeNeo gateway) can build requests without importing internal packages.
// All fields from openai.ChatCompletionRequest are honoured (Tools, ToolChoice,
// ParallelToolCalls, ResponseFormat, StreamOptions, Seed, LogitBias, …) plus
// any unknown JSON keys captured in Extra.
type ChatCompletionRequestExtra = gcs.OpenAICompletionRequestExtra

// RawChunkCallback receives upstream chat completion chunks verbatim, suitable
// for an OpenAI-compatible gateway that needs to relay provider fields the
// typed structs may not capture (delta.tool_calls, delta.reasoning_content,
// stream usage events, finish_reason, custom Morpheus extensions, etc.).
//
// chunkJSON is the raw JSON of either a streaming delta
// (ChatCompletionStreamResponseExtra) or a non-streaming response
// (ChatCompletionResponseExtra), preserving the original key order and any
// provider-specific fields. isLast is true on the terminal control chunk; in
// that case chunkJSON is nil — the caller should write the SSE [DONE] sentinel
// or close the response.
type RawChunkCallback func(chunkJSON json.RawMessage, isLast bool) error

// SendChatCompletion forwards an OpenAI-compatible chat completion request to
// the upstream Morpheus provider and surfaces each response chunk verbatim via
// cb. This is the entry point for OpenAI-compatible local gateways (e.g. the
// NodeNeo AI Gateway used by Cursor / Zed / Claude Desktop / LangChain).
//
// Compared to SendPromptWithMessagesAndParams, this method:
//   - Accepts the full request object (including Tools, ToolChoice,
//     ResponseFormat, StreamOptions, ParallelToolCalls, Seed, LogitBias and
//     any vendor-specific fields preserved in Extra).
//   - Relays each chunk's original JSON unchanged so tool_calls,
//     reasoning_content, finish_reason, and usage flow through untouched.
//
// req.Model is overwritten with sessionID so the proxy-router can route, which
// matches what the existing SendPrompt* methods do; the upstream provider
// rewrites it again to its own model name before responding.
func (s *SDK) SendChatCompletion(ctx context.Context, sessionID string, req *ChatCompletionRequestExtra, cb RawChunkCallback) error {
	if req == nil {
		return fmt.Errorf("nil chat completion request")
	}
	if cb == nil {
		return fmt.Errorf("nil chunk callback")
	}

	id := common.HexToHash(sessionID)
	if err := s.ensureProviderForSession(ctx, id); err != nil {
		return redactError(err)
	}

	req.Model = sessionID

	internalCB := func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		if errResp != nil {
			// errResp.ProviderModelError can include the upstream's
			// raw URL/IP — redact before bubbling to consumers.
			return fmt.Errorf("provider error: %s", redactProviderEndpointsString(fmt.Sprintf("%v", errResp.ProviderModelError)))
		}
		if chunk.Type() == gcs.ChunkTypeControl {
			return cb(nil, true)
		}

		data := chunk.Data()
		if data == nil {
			return nil
		}
		raw, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal chunk: %w", err)
		}

		return cb(raw, !chunk.IsStreaming())
	}

	_, err := s.proxySender.SendPromptV2(ctx, id, req, internalCB)
	return redactError(err)
}

// EmbeddingsRequest is the OpenAI embeddings request type used by the embedded
// gateway. Exported as an alias so external callers do not need to import
// internal proxy-router packages. All fields from openai.EmbeddingRequest are
// honoured, plus any unknown JSON keys captured in Extra.
type EmbeddingsRequest = gcs.EmbeddingsRequest

// SendEmbeddings forwards an OpenAI-compatible embeddings request to the
// upstream Morpheus provider and returns the response JSON verbatim. Embeddings
// are non-streaming: the upstream emits a single response chunk followed by a
// control chunk.
//
// The returned RawMessage is the provider's full EmbeddingsResponse (id,
// object, data:[…vectors…], model, usage, plus any vendor-specific extras).
// The caller is expected to relay it to its HTTP client unchanged.
func (s *SDK) SendEmbeddings(ctx context.Context, sessionID string, req *EmbeddingsRequest) (json.RawMessage, error) {
	if req == nil {
		return nil, fmt.Errorf("nil embeddings request")
	}

	id := common.HexToHash(sessionID)
	if err := s.ensureProviderForSession(ctx, id); err != nil {
		return nil, redactError(err)
	}

	var responseJSON json.RawMessage

	internalCB := func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		if errResp != nil {
			return fmt.Errorf("provider error: %s", redactProviderEndpointsString(fmt.Sprintf("%v", errResp.ProviderModelError)))
		}
		if chunk.Type() == gcs.ChunkTypeControl {
			return nil
		}
		data := chunk.Data()
		if data == nil {
			return nil
		}
		b, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal embeddings chunk: %w", err)
		}
		responseJSON = b
		return nil
	}

	_, err := s.proxySender.SendEmbeddings(ctx, id, req, internalCB)
	if err != nil {
		return nil, redactError(err)
	}
	return responseJSON, nil
}
