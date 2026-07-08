package modelhealth

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

func TestGenerateArithmeticPrompt(t *testing.T) {
	for i := 0; i < 100; i++ {
		a, b, prompt := generateArithmeticPrompt()
		require.GreaterOrEqual(t, a, 1)
		require.LessOrEqual(t, a, 50)
		require.GreaterOrEqual(t, b, 1)
		require.LessOrEqual(t, b, 50)
		require.Contains(t, prompt, fmt.Sprintf("%d+%d", a, b))
	}
}

func TestVerifyArithmeticAnswer(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected int
		want     bool
	}{
		{"bare number", "4", 4, true},
		{"number in sentence", "The answer is 42.", 42, true},
		{"echoed equation", "2+2=4", 4, true},
		{"wrong answer", "5", 4, false},
		{"substring of larger number", "14", 4, false},
		{"expected contains other digits", "142", 42, false},
		{"empty", "", 4, false},
		{"newline wrapped", "\n12\n", 12, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, verifyArithmeticAnswer(tt.response, tt.expected))
		})
	}
}

var (
	modelLLM          = common.HexToHash("0x01")
	modelEmbedding    = common.HexToHash("0x02")
	modelNoBid        = common.HexToHash("0x03")
	modelTTS          = common.HexToHash("0x04")
	modelUnconfigured = common.HexToHash("0x05")
)

type mockDeps struct {
	bids       []*structs.Bid
	bidsErr    error
	tags       map[common.Hash][]string
	names      map[common.Hash]string
	modelIDs   []common.Hash
	adapter    aiengine.AIEngineStream
	adapterErr error
}

func (m *mockDeps) GetActiveBidsByProvider(ctx context.Context, provider common.Address, offset *big.Int, limit uint8, order registries.Order) ([]*structs.Bid, error) {
	if m.bidsErr != nil {
		return nil, m.bidsErr
	}
	from := int(offset.Int64())
	if from >= len(m.bids) {
		return nil, nil
	}
	to := from + int(limit)
	if to > len(m.bids) {
		to = len(m.bids)
	}
	return m.bids[from:to], nil
}

func (m *mockDeps) GetModelNameAndTags(ctx context.Context, modelID common.Hash) (string, []string, error) {
	tags, ok := m.tags[modelID]
	if !ok {
		return "", nil, errors.New("model not found")
	}
	return m.names[modelID], tags, nil
}

func (m *mockDeps) GetAll() ([]common.Hash, []config.ModelConfig) {
	configs := make([]config.ModelConfig, len(m.modelIDs))
	return m.modelIDs, configs
}

func (m *mockDeps) GetAdapter(ctx context.Context, chatID, modelID, sessionID common.Hash, storeChatContext, forwardChatContext bool) (aiengine.AIEngineStream, error) {
	return m.adapter, m.adapterErr
}

// mathSolvingAdapter parses the arithmetic prompt and answers it correctly,
// and returns a non-empty vector for embeddings probes.
type mathSolvingAdapter struct {
	promptErr error
}

func (a *mathSolvingAdapter) Prompt(ctx context.Context, compl *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	if a.promptErr != nil {
		return a.promptErr
	}
	var x, y int
	_, err := fmt.Sscanf(compl.Messages[0].Content, "What is %d+%d?", &x, &y)
	if err != nil {
		return err
	}
	resp := &gcs.ChatCompletionResponseExtra{
		ChatCompletionResponse: openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: fmt.Sprintf("%d", x+y)}},
			},
		},
	}
	return cb(ctx, gcs.NewChunkText(resp), nil)
}

func (a *mathSolvingAdapter) Embeddings(ctx context.Context, req *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) error {
	resp := gcs.EmbeddingsResponse{
		EmbeddingResponse: openai.EmbeddingResponse{
			Data: []openai.Embedding{{Embedding: []float32{0.1, 0.2}}},
		},
	}
	return cb(ctx, gcs.NewChunkEmbedding(resp), nil)
}

func (a *mathSolvingAdapter) AudioTranscription(ctx context.Context, req *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	return errors.New("not implemented")
}

func (a *mathSolvingAdapter) AudioSpeech(ctx context.Context, req *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	return errors.New("not implemented")
}

func (a *mathSolvingAdapter) ApiType() string { return "openai" }

func newTestChecker(deps *mockDeps) *Checker {
	return NewChecker(Deps{
		Adapters:     deps,
		Bids:         deps,
		Models:       deps,
		ModelConfigs: deps,
	}, time.Hour, time.Second, 0, lib.NewTestLogger())
}

func reportByID(t *testing.T, reports []system.ModelHealthReport, modelID common.Hash) system.ModelHealthReport {
	t.Helper()
	for _, r := range reports {
		if r.ModelID == modelID.Hex() {
			return r
		}
	}
	t.Fatalf("report for model %s not found", modelID.Hex())
	return system.ModelHealthReport{}
}

func bidFor(modelID common.Hash) *structs.Bid {
	return &structs.Bid{Id: bidIDFor(modelID), ModelAgentId: modelID}
}

// bidIDFor derives a deterministic fake bid ID from a model ID for assertions.
func bidIDFor(modelID common.Hash) common.Hash {
	return common.BytesToHash(append([]byte{0xbb}, modelID.Bytes()[1:]...))
}

func TestCheckAllStatuses(t *testing.T) {
	olderLLMBid := &structs.Bid{Id: common.HexToHash("0xdead"), ModelAgentId: modelLLM}
	deps := &mockDeps{
		// bids come back newest-first; the older duplicate for modelLLM must be ignored
		bids: []*structs.Bid{bidFor(modelLLM), bidFor(modelEmbedding), bidFor(modelTTS), olderLLMBid, bidFor(modelUnconfigured)},
		tags: map[common.Hash][]string{
			modelLLM:          {"llm"},
			modelEmbedding:    {"embedding"},
			modelTTS:          {"tts"},
			modelUnconfigured: {"llm"},
		},
		names:    map[common.Hash]string{modelLLM: "test-llm"},
		modelIDs: []common.Hash{modelLLM, modelEmbedding, modelNoBid, modelTTS},
		adapter:  &mathSolvingAdapter{},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	reports := checker.GetReports()
	require.Len(t, reports, 5)

	llm := reportByID(t, reports, modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.True(t, llm.HasActiveBid)
	require.Equal(t, bidIDFor(modelLLM).Hex(), llm.BidID)
	require.Equal(t, string(structs.ModelTypeLLM), llm.ModelType)
	require.Equal(t, "test-llm", llm.ModelName)
	require.Zero(t, llm.HttpStatus)
	require.NotNil(t, llm.PromptCorrect)
	require.True(t, *llm.PromptCorrect)
	require.NotZero(t, llm.LastHealthy)
	require.NotZero(t, llm.LastChecked)

	embedding := reportByID(t, reports, modelEmbedding)
	require.Equal(t, system.ModelHealthStatusHealthy, embedding.Status)
	require.Nil(t, embedding.PromptCorrect)

	noBid := reportByID(t, reports, modelNoBid)
	require.Equal(t, system.ModelHealthStatusNoBid, noBid.Status)
	require.False(t, noBid.HasActiveBid)
	require.Empty(t, noBid.BidID)
	require.Zero(t, noBid.LastHealthy)

	tts := reportByID(t, reports, modelTTS)
	require.Equal(t, system.ModelHealthStatusSkipped, tts.Status)

	// bid without a matching models-config entry: reported, not probed
	unconfigured := reportByID(t, reports, modelUnconfigured)
	require.Equal(t, system.ModelHealthStatusNoModel, unconfigured.Status)
	require.True(t, unconfigured.HasActiveBid)
	require.Equal(t, bidIDFor(modelUnconfigured).Hex(), unconfigured.BidID)
	require.Equal(t, string(structs.ModelTypeLLM), unconfigured.ModelType)
	require.Zero(t, unconfigured.LatencyMs)
	require.Nil(t, unconfigured.PromptCorrect)
}

func TestCheckAllPrunesRemovedModels(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM), bidFor(modelUnconfigured)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}, modelUnconfigured: {"llm"}},
		modelIDs: []common.Hash{modelLLM, modelNoBid},
		adapter:  &mathSolvingAdapter{},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	require.Len(t, checker.GetReports(), 3)

	// model removed from config and bid cancelled: both must disappear
	deps.modelIDs = []common.Hash{modelLLM}
	deps.bids = []*structs.Bid{bidFor(modelLLM)}
	checker.checkAll(context.Background(), common.Address{})

	reports := checker.GetReports()
	require.Len(t, reports, 1)
	require.Equal(t, modelLLM.Hex(), reports[0].ModelID)
}

func TestCheckAllBidsOnlyNoConfig(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelUnconfigured)},
		tags:     map[common.Hash][]string{modelUnconfigured: {"llm"}},
		modelIDs: []common.Hash{},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	reports := checker.GetReports()
	require.Len(t, reports, 1)
	require.Equal(t, system.ModelHealthStatusNoModel, reports[0].Status)
}

// errorRespondingAdapter simulates an upstream backend that answers every
// request with an HTTP error status (e.g. 402 payment required, 429 rate
// limited) instead of a completion.
type errorRespondingAdapter struct {
	statusCode int
}

func (a *errorRespondingAdapter) Prompt(ctx context.Context, compl *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	return cb(ctx, nil, gcs.NewAiEngineErrorResponse(a.statusCode, map[string]interface{}{"error": "upstream error"}))
}

func (a *errorRespondingAdapter) Embeddings(ctx context.Context, req *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) error {
	return cb(ctx, nil, gcs.NewAiEngineErrorResponse(a.statusCode, map[string]interface{}{"error": "upstream error"}))
}

func (a *errorRespondingAdapter) AudioTranscription(ctx context.Context, req *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	return errors.New("not implemented")
}

func (a *errorRespondingAdapter) AudioSpeech(ctx context.Context, req *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	return errors.New("not implemented")
}

func (a *errorRespondingAdapter) ApiType() string { return "openai" }

func TestCheckAllUpstreamHTTPError(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &errorRespondingAdapter{statusCode: 402},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
	require.Equal(t, system.ModelHealthErrorBadResponse, llm.ErrorKind)
	require.Equal(t, 402, llm.HttpStatus)
}

func TestCheckAllRateLimited(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &errorRespondingAdapter{statusCode: 429},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
	require.Equal(t, system.ModelHealthErrorRateLimited, llm.ErrorKind)
	require.Equal(t, 429, llm.HttpStatus)
}

func TestCheckAllProbeFailure(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{promptErr: errors.New("failed to send request: connection refused")},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
	require.Equal(t, system.ModelHealthErrorConnection, llm.ErrorKind)
	require.Zero(t, llm.LastHealthy)
}

func TestCheckAllPreservesLastHealthy(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	healthy := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, healthy.Status)
	lastHealthy := healthy.LastHealthy
	require.NotZero(t, lastHealthy)

	deps.adapter = &mathSolvingAdapter{promptErr: context.DeadlineExceeded}
	checker.checkAll(context.Background(), common.Address{})

	unhealthy := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, unhealthy.Status)
	require.Equal(t, system.ModelHealthErrorTimeout, unhealthy.ErrorKind)
	require.Equal(t, lastHealthy, unhealthy.LastHealthy)
}

func TestActiveBidModelsPaginates(t *testing.T) {
	// 3 full pages plus a partial one; the configured model's bid sits on the
	// last page and must still be found
	var bids []*structs.Bid
	for i := 0; i < 3*bidsPageLimit+5; i++ {
		id := common.BigToHash(big.NewInt(int64(i + 1000)))
		bids = append(bids, &structs.Bid{Id: common.BigToHash(big.NewInt(int64(i))), ModelAgentId: id})
	}
	bids = append(bids, bidFor(modelLLM))

	deps := &mockDeps{
		bids:     bids,
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}

	checker := newTestChecker(deps)
	byModel, err := checker.activeBidModels(context.Background(), common.Address{})
	require.NoError(t, err)
	require.Len(t, byModel, len(bids)) // every bid targets a distinct model
	require.Equal(t, bidIDFor(modelLLM), byModel[modelLLM])
}

func TestTriggerNowQueuesAtMostOne(t *testing.T) {
	checker := newTestChecker(&mockDeps{})
	require.True(t, checker.TriggerNow())
	require.False(t, checker.TriggerNow(), "second trigger must be rejected while one is queued")
}

func TestRunConsumesTrigger(t *testing.T) {
	checker := newTestChecker(&mockDeps{})
	require.True(t, checker.TriggerNow())
	require.False(t, checker.TriggerNow())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		_ = checker.Run(ctx, common.Address{})
		close(done)
	}()

	// the Run loop must drain the queued trigger, freeing the slot
	require.Eventually(t, func() bool { return checker.TriggerNow() }, 5*time.Second, 10*time.Millisecond)

	cancel()
	<-done
}

func TestCheckAllBidsError(t *testing.T) {
	deps := &mockDeps{
		bidsErr:  errors.New("rpc unavailable"),
		modelIDs: []common.Hash{modelLLM},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	// no reports should be written when the bid lookup fails, so stale
	// results from a previous successful run are preserved
	require.Empty(t, checker.GetReports())
}
