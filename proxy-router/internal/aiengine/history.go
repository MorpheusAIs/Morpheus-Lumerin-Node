package aiengine

import (
	"context"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/completion"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
)

type History struct {
	engine  AIEngineStream
	storage gcs.ChatStorageInterface
	chatID  common.Hash
	modelID common.Hash
	log     lib.ILogger
}

func NewHistory(engine AIEngineStream, storage gcs.ChatStorageInterface, chatID, modelID common.Hash, log lib.ILogger) *History {
	return &History{
		engine:  engine,
		storage: storage,
		chatID:  chatID,
		modelID: modelID,
		log:     log,
	}
}

func (h *History) Prompt(ctx context.Context, prompt *openai.ChatCompletionRequest, cb completion.CompletionCallback) error {
	isLocal := h.engine.ApiType() != "remote"
	completions := make([]*completion.ChunkImpl, 0)
	startTime := time.Now()

	history, err := h.storage.LoadChatFromFile(h.chatID.Hex())
	if err != nil {
		h.log.Warnf("failed to load chat history: %v", err)
	}

	promptWithHistory := history.AppendChatHistory(prompt)

	err = h.engine.Prompt(ctx, promptWithHistory, func(ctx context.Context, completion *completion.ChunkImpl) error {
		completions = append(completions, completion)
		return cb(ctx, completion)
	})
	if err != nil {
		return err
	}
	endTime := time.Now()

	err = h.storage.StorePromptResponseToFile(h.chatID.Hex(), isLocal, h.modelID.Hex(), promptWithHistory, completions, startTime, endTime)
	if err != nil {
		h.log.Errorf("failed to store prompt response: %v", err)
	}

	return err
}

func (h *History) ApiType() string {
	return h.engine.ApiType()
}

var _ AIEngineStream = &History{}
