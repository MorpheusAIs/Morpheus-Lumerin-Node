package aiengine

import (
	"context"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type History struct {
	engine             AIEngineStream
	storage            gcs.ChatStorageInterface
	chatID             common.Hash
	modelID            common.Hash
	forwardChatContext bool
	log                lib.ILogger
}

func NewHistory(engine AIEngineStream, storage gcs.ChatStorageInterface, chatID, modelID common.Hash, forwardChatContext bool, log lib.ILogger) *History {
	return &History{
		engine:             engine,
		storage:            storage,
		chatID:             chatID,
		modelID:            modelID,
		forwardChatContext: forwardChatContext,
		log:                log,
	}
}

func (h *History) Prompt(ctx context.Context, prompt *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	isLocal := h.engine.ApiType() != "remote"
	completions := make([]gcs.Chunk, 0)
	startTime := time.Now()

	history, err := h.storage.LoadChatFromFile(h.chatID.Hex())
	if err != nil {
		h.log.Warnf("failed to load chat history: %v", err)
	}

	adjustedPrompt := prompt
	if h.forwardChatContext {
		adjustedPrompt = history.AppendChatHistory(prompt)
	}

	err = h.engine.Prompt(ctx, adjustedPrompt, func(ctx context.Context, completion gcs.Chunk, errorBody *gcs.AiEngineErrorResponse) error {
		if completion != nil {
			completions = append(completions, completion)
		}
		return cb(ctx, completion, errorBody)
	})
	if err != nil {
		return err
	}
	endTime := time.Now()

	err = h.storage.StorePromptResponseToFile(h.chatID.Hex(), isLocal, h.modelID.Hex(), prompt, completions, startTime, endTime)
	if err != nil {
		h.log.Errorf("failed to store prompt response: %v", err)
	}

	return err
}

func (h *History) AudioTranscription(ctx context.Context, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	isLocal := h.engine.ApiType() != "remote"
	completions := make([]gcs.Chunk, 0)
	startTime := time.Now()

	err := h.engine.AudioTranscription(ctx, prompt, func(ctx context.Context, completion gcs.Chunk, errorBody *gcs.AiEngineErrorResponse) error {
		if completion != nil {
			completions = append(completions, completion)
		}
		return cb(ctx, completion, errorBody)
	})
	if err != nil {
		return err
	}
	endTime := time.Now()

	err = h.storage.StorePromptResponseToFile(h.chatID.Hex(), isLocal, h.modelID.Hex(), prompt, completions, startTime, endTime)
	if err != nil {
		h.log.Errorf("failed to store prompt response: %v", err)
	}

	return err
}

func (h *History) AudioSpeech(ctx context.Context, prompt *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	isLocal := h.engine.ApiType() != "remote"
	completions := make([]gcs.Chunk, 0)
	startTime := time.Now()

	err := h.engine.AudioSpeech(ctx, prompt, func(ctx context.Context, completion gcs.Chunk, errorBody *gcs.AiEngineErrorResponse) error {
		if completion != nil {
			completions = append(completions, completion)
		}
		return cb(ctx, completion, errorBody)
	})
	if err != nil {
		return err
	}
	endTime := time.Now()

	err = h.storage.StorePromptResponseToFile(h.chatID.Hex(), isLocal, h.modelID.Hex(), prompt, completions, startTime, endTime)
	if err != nil {
		h.log.Errorf("failed to store prompt response: %v", err)
	}

	return err
}

func (h *History) Embeddings(ctx context.Context, prompt *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) error {
	isLocal := h.engine.ApiType() != "remote"
	completions := make([]gcs.Chunk, 0)
	startTime := time.Now()

	err := h.engine.Embeddings(ctx, prompt, func(ctx context.Context, completion gcs.Chunk, errorBody *gcs.AiEngineErrorResponse) error {
		if completion != nil {
			completions = append(completions, completion)
		}
		return cb(ctx, completion, errorBody)
	})
	if err != nil {
		return err
	}
	endTime := time.Now()

	err = h.storage.StorePromptResponseToFile(h.chatID.Hex(), isLocal, h.modelID.Hex(), prompt, completions, startTime, endTime)
	if err != nil {
		h.log.Errorf("failed to store prompt response: %v", err)
	}

	return err
}

func (h *History) ApiType() string {
	return h.engine.ApiType()
}

var _ AIEngineStream = &History{}
