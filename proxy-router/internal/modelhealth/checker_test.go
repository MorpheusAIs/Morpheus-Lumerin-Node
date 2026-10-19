package modelhealth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/attestation"
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
	modelUntagged     = common.HexToHash("0x06")
	modelChatTagged   = common.HexToHash("0x07")
	modelImage        = common.HexToHash("0x08")
	modelDecisions    = common.HexToHash("0x09")
)

type mockDeps struct {
	bids       []*structs.Bid
	bidsErr    error
	tags       map[common.Hash][]string
	names      map[common.Hash]string
	modelIDs   []common.Hash
	adapter    aiengine.AIEngineStream
	adapterErr error
	teeStatus  TeeStatusProvider
	configs    map[common.Hash]config.ModelConfig
	apiDetect  ApiDetector
}

// mockTeeStatus returns canned backend attestation snapshots per model and
// records ReattestBackend calls. When onReattest is set, it is applied to the
// snapshot map before the checker reads the status back.
type mockTeeStatus struct {
	snapshots  map[string]*attestation.BackendAttestationSnapshot
	reattested map[string]int
	onReattest func(modelID string)
}

func (m *mockTeeStatus) GetStatus(modelID string) *attestation.BackendAttestationSnapshot {
	return m.snapshots[modelID]
}

func (m *mockTeeStatus) ReattestBackend(ctx context.Context, modelID string, endpoint string) error {
	if m.reattested == nil {
		m.reattested = make(map[string]int)
	}
	m.reattested[modelID]++
	if m.onReattest != nil {
		m.onReattest(modelID)
	}
	return nil
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
	for i, id := range m.modelIDs {
		configs[i] = m.configs[id]
	}
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

func (a *mathSolvingAdapter) Decisions(ctx context.Context, req *gcs.DecisionsRequest, cb gcs.CompletionCallback) error {
	if a.promptErr != nil {
		return a.promptErr
	}
	resp := gcs.DecisionsResponse{
		Model: "health-probe",
		Answers: map[string]json.RawMessage{
			"ok": json.RawMessage(`{"noul":true,"confidence":1}`),
		},
		Usage: gcs.DecisionsUsage{InputTokens: 1, OutputTokens: 1},
	}
	return cb(ctx, gcs.NewChunkDecisions(resp), nil)
}

func (a *mathSolvingAdapter) ApiType() string { return "openai" }

func newTestChecker(deps *mockDeps) *Checker {
	return NewChecker(Deps{
		Adapters:     deps,
		Bids:         deps,
		Models:       deps,
		ModelConfigs: deps,
		TeeStatus:    deps.teeStatus,
		ApiDetect:    deps.apiDetect,
	}, time.Hour, time.Second, 0, 0, lib.NewTestLogger())
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

func TestCheckAllDecisionsProbe(t *testing.T) {
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelDecisions)},
		tags: map[common.Hash][]string{
			// Catalog still stamps modelType LLM; Decision-class tag wins.
			modelDecisions: {"LLM", "TypeSafe", "Decision"},
		},
		names:    map[common.Hash]string{modelDecisions: "jev-1.13-ts"},
		modelIDs: []common.Hash{modelDecisions},
		adapter:  &mathSolvingAdapter{},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	dec := reportByID(t, checker.GetReports(), modelDecisions)
	require.Equal(t, system.ModelHealthStatusHealthy, dec.Status)
	require.Equal(t, string(structs.ModelTypeDECISIONS), dec.ModelType)
	require.Equal(t, "jev-1.13-ts", dec.ModelName)
	require.Nil(t, dec.PromptCorrect)
	require.NotZero(t, dec.LastHealthy)
	require.Zero(t, dec.HttpStatus)
}

// An untagged model must still be probed. DetectModelType returns Unknown when
// a model carries no recognised tag, and skipping publishes no health for it at
// all, which drops the bid out of consumer-facing routable listings even though
// the backend serves normally. Only the model owner can add tags, so a provider
// cannot fix this at source.
func TestCheckAllUntaggedModelProbesAsLLM(t *testing.T) {
	deps := &mockDeps{
		bids: []*structs.Bid{
			bidFor(modelUntagged), bidFor(modelChatTagged),
			bidFor(modelImage), bidFor(modelTTS),
		},
		tags: map[common.Hash][]string{
			modelUntagged:   {},
			modelChatTagged: {"chat"},
			modelImage:      {"image"},
			modelTTS:        {"tts"},
		},
		modelIDs: []common.Hash{modelUntagged, modelChatTagged, modelImage, modelTTS},
		adapter:  &mathSolvingAdapter{},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	reports := checker.GetReports()

	untagged := reportByID(t, reports, modelUntagged)
	require.Equal(t, system.ModelHealthStatusHealthy, untagged.Status)
	require.Equal(t, string(structs.ModelTypeUnknown), untagged.ModelType)
	require.NotNil(t, untagged.PromptCorrect)
	require.True(t, *untagged.PromptCorrect)

	// "chat" is not in DetectModelType, so it resolves to Unknown and must take
	// the same LLM probe. The create-model API tells contributors to tag "chat"
	// while DetectModelType looks for llm/textgeneration/t2t, so this is a
	// common way to end up untyped while plainly being a text model.
	chat := reportByID(t, reports, modelChatTagged)
	require.Equal(t, system.ModelHealthStatusHealthy, chat.Status)
	require.Equal(t, string(structs.ModelTypeUnknown), chat.ModelType)
	require.NotNil(t, chat.PromptCorrect)
	require.True(t, *chat.PromptCorrect)

	// "image" is also Unknown to DetectModelType, but it must NOT inherit the
	// LLM fallback: a text prompt fired at an image backend costs an upstream
	// call and reports a misleading unhealthy rather than information.
	image := reportByID(t, reports, modelImage)
	require.Equal(t, system.ModelHealthStatusSkipped, image.Status)
	require.Nil(t, image.PromptCorrect)

	// A known non-LLM type still skips: no text probe is sent to a speech model.
	tts := reportByID(t, reports, modelTTS)
	require.Equal(t, system.ModelHealthStatusSkipped, tts.Status)
}

// A guessed type that turns out wrong must NOT report unhealthy. Most of the
// registry carries no tags at all, so mediaOnly cannot catch an untagged image
// or video model and it reaches the LLM probe. unhealthy is excluded by every
// session health policy, including the default permissive one, while skipped is
// not — so reporting unhealthy on a failed guess would hard-block a bid that was
// only unroutable before, which is worse than the silence this patch fixes.
// A tagged model keeps reporting unhealthy: there the type is known, so a failed
// probe really is evidence about the backend.
func TestCheckAllGuessedTypeProbeFailureSkipsNotUnhealthy(t *testing.T) {
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelUntagged), bidFor(modelLLM)},
		tags: map[common.Hash][]string{
			modelUntagged: {},
			modelLLM:      {"llm"},
		},
		modelIDs: []common.Hash{modelUntagged, modelLLM},
		adapter:  &mathSolvingAdapter{promptErr: errors.New("not a text model")},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	reports := checker.GetReports()

	// Guessed type, probe failed: stays out of the way rather than blocking.
	untagged := reportByID(t, reports, modelUntagged)
	require.Equal(t, system.ModelHealthStatusSkipped, untagged.Status)
	require.Equal(t, string(structs.ModelTypeUnknown), untagged.ModelType)

	// Known type, same failure: unhealthy, because the probe was appropriate.
	llm := reportByID(t, reports, modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
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

func (a *errorRespondingAdapter) Decisions(ctx context.Context, req *gcs.DecisionsRequest, cb gcs.CompletionCallback) error {
	return cb(ctx, nil, gcs.NewAiEngineErrorResponse(a.statusCode, map[string]interface{}{"error": "upstream error"}))
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
	require.Equal(t, system.ModelHealthStatusDegraded, llm.Status)
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

func TestCheckAllTeeModelAttestationGate(t *testing.T) {
	teeDeps := func(status TeeStatusProvider) *mockDeps {
		return &mockDeps{
			bids:      []*structs.Bid{bidFor(modelLLM)},
			tags:      map[common.Hash][]string{modelLLM: {"llm", "tee"}},
			modelIDs:  []common.Hash{modelLLM},
			adapter:   &mathSolvingAdapter{},
			teeStatus: status,
		}
	}

	t.Run("failed attestation reports tee_unverified and skips the probe", func(t *testing.T) {
		deps := teeDeps(&mockTeeStatus{snapshots: map[string]*attestation.BackendAttestationSnapshot{
			modelLLM.Hex(): {ModelID: modelLLM.Hex(), Status: attestation.StatusFailed},
		}})
		checker := newTestChecker(deps)
		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.Equal(t, system.ModelHealthStatusTeeUnverified, llm.Status)
		require.Equal(t, system.ModelHealthErrorTeeAttestation, llm.ErrorKind)
		require.Nil(t, llm.PromptCorrect, "backend must not be probed when TEE attestation failed")
	})

	t.Run("missing snapshot reports tee_unverified", func(t *testing.T) {
		deps := teeDeps(&mockTeeStatus{snapshots: map[string]*attestation.BackendAttestationSnapshot{}})
		checker := newTestChecker(deps)
		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.Equal(t, system.ModelHealthStatusTeeUnverified, llm.Status)
	})

	t.Run("passed attestation probes normally", func(t *testing.T) {
		deps := teeDeps(&mockTeeStatus{snapshots: map[string]*attestation.BackendAttestationSnapshot{
			modelLLM.Hex(): {ModelID: modelLLM.Hex(), Status: attestation.StatusPassed},
		}})
		checker := newTestChecker(deps)
		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
		require.NotNil(t, llm.PromptCorrect)
	})

	t.Run("first sweep skips re-attestation, later sweeps re-attest regardless of status", func(t *testing.T) {
		tee := &mockTeeStatus{snapshots: map[string]*attestation.BackendAttestationSnapshot{
			modelLLM.Hex(): {ModelID: modelLLM.Hex(), Status: attestation.StatusPassed},
		}}
		checker := newTestChecker(teeDeps(tee))

		checker.checkAll(context.Background(), common.Address{})
		require.Zero(t, tee.reattested[modelLLM.Hex()], "first sweep must rely on the startup attestation")

		checker.checkAll(context.Background(), common.Address{})
		checker.checkAll(context.Background(), common.Address{})
		require.Equal(t, 2, tee.reattested[modelLLM.Hex()], "re-attestation must run on every later sweep, even when passed")
	})

	t.Run("failed attestation self-heals when re-attestation passes", func(t *testing.T) {
		tee := &mockTeeStatus{snapshots: map[string]*attestation.BackendAttestationSnapshot{
			modelLLM.Hex(): {ModelID: modelLLM.Hex(), Status: attestation.StatusFailed},
		}}
		tee.onReattest = func(modelID string) {
			tee.snapshots[modelID] = &attestation.BackendAttestationSnapshot{ModelID: modelID, Status: attestation.StatusPassed}
		}
		checker := newTestChecker(teeDeps(tee))
		checker.firstTeeSweepDone = true

		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.Equal(t, system.ModelHealthStatusHealthy, llm.Status, "a stale startup failure must heal once re-attestation passes")
		require.Equal(t, 1, tee.reattested[modelLLM.Hex()])
	})

	t.Run("no TEE status provider probes normally", func(t *testing.T) {
		deps := teeDeps(nil)
		checker := newTestChecker(deps)
		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	})
}

func TestReportPromptFailureFlipsAfterThreshold(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}
	checker := newTestChecker(deps) // threshold defaults to 3
	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelLLM).Status)

	checker.ReportPromptFailure(modelLLM, 0)
	checker.ReportPromptFailure(modelLLM, 0)
	require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelLLM).Status,
		"below the threshold the status must not change")

	checker.ReportPromptFailure(modelLLM, 0)
	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
	require.Equal(t, system.ModelHealthErrorSessionErrors, llm.ErrorKind)
}

func TestReportPromptSuccessResetsStreakAndHeals(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	// success in the middle of a streak resets the counter
	checker.ReportPromptFailure(modelLLM, 0)
	checker.ReportPromptFailure(modelLLM, 0)
	checker.ReportPromptSuccess(modelLLM)
	checker.ReportPromptFailure(modelLLM, 0)
	checker.ReportPromptFailure(modelLLM, 0)
	require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelLLM).Status)

	// third consecutive failure flips it...
	checker.ReportPromptFailure(modelLLM, 0)
	require.Equal(t, system.ModelHealthStatusUnhealthy, reportByID(t, checker.GetReports(), modelLLM).Status)

	// ...and a successful real prompt heals it right away
	checker.ReportPromptSuccess(modelLLM)
	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.Empty(t, llm.ErrorKind)
	require.NotZero(t, llm.LastHealthy)
}

func TestHealthyProbeResetsFailureStreak(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}
	checker := newTestChecker(deps)

	checker.ReportPromptFailure(modelLLM, 0)
	checker.ReportPromptFailure(modelLLM, 0)

	// a healthy scheduled sweep clears the streak: two more failures stay
	// below the threshold again
	checker.checkAll(context.Background(), common.Address{})
	checker.ReportPromptFailure(modelLLM, 0)
	checker.ReportPromptFailure(modelLLM, 0)
	require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelLLM).Status)
}

func TestReportPromptFailureRateLimitedStreakFlipsDegraded(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}
	checker := newTestChecker(deps) // threshold defaults to 3
	checker.checkAll(context.Background(), common.Address{})

	// a streak that is pure upstream throttling flips to degraded, not unhealthy
	checker.ReportPromptFailure(modelLLM, 429)
	checker.ReportPromptFailure(modelLLM, 429)
	checker.ReportPromptFailure(modelLLM, 429)

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusDegraded, llm.Status)
	require.Equal(t, system.ModelHealthErrorSessionErrors, llm.ErrorKind)
	require.Equal(t, 429, llm.HttpStatus)

	// a successful real prompt heals session-flipped degraded right away
	checker.ReportPromptSuccess(modelLLM)
	llm = reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusHealthy, llm.Status)
	require.Empty(t, llm.ErrorKind)
	require.Zero(t, llm.HttpStatus)
}

func TestReportPromptFailureMixedStreakFlipsUnhealthy(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	// any non-429 failure in the streak means real errors, not just throttling
	checker.ReportPromptFailure(modelLLM, 429)
	checker.ReportPromptFailure(modelLLM, 500)
	checker.ReportPromptFailure(modelLLM, 429)

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
	require.Equal(t, system.ModelHealthErrorSessionErrors, llm.ErrorKind)
}

func TestReportPromptFailureDegradedEscalatesToUnhealthy(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	checker.ReportPromptFailure(modelLLM, 429)
	checker.ReportPromptFailure(modelLLM, 429)
	checker.ReportPromptFailure(modelLLM, 429)
	require.Equal(t, system.ModelHealthStatusDegraded, reportByID(t, checker.GetReports(), modelLLM).Status)

	// a non-429 failure joining the still-open streak escalates to unhealthy
	checker.ReportPromptFailure(modelLLM, 500)
	require.Equal(t, system.ModelHealthStatusUnhealthy, reportByID(t, checker.GetReports(), modelLLM).Status)
}

func TestReportPromptFailureWithoutPriorReport(t *testing.T) {
	checker := newTestChecker(&mockDeps{})

	for i := 0; i < DefaultMaxConsecutiveErrors; i++ {
		checker.ReportPromptFailure(modelLLM, 0)
	}

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusUnhealthy, llm.Status)
	require.Equal(t, system.ModelHealthErrorSessionErrors, llm.ErrorKind)
	require.True(t, llm.HasActiveBid)
}

func TestReportTeeFailureImmediate(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelLLM).Status)

	checker.ReportTeeFailure(modelLLM)

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, system.ModelHealthStatusTeeUnverified, llm.Status)
	require.Equal(t, system.ModelHealthErrorTeeAttestation, llm.ErrorKind)
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

func TestCheckModelAttachesDeclaredApiSpec(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM, modelNoBid},
		adapter:  &mathSolvingAdapter{},
		configs: map[common.Hash]config.ModelConfig{
			modelLLM:   {ModelName: "qwen3-32b", ApiType: "openai", ApiStack: "vllm", ApiURL: "http://llm:8000/v1/chat/completions"},
			modelNoBid: {ModelName: "llama-3.1-8b", ApiType: "openai", ApiStack: "ollama", ApiURL: "http://ollama:11434/v1/chat/completions"},
		},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	reports := checker.GetReports()

	llm := reportByID(t, reports, modelLLM)
	require.NotNil(t, llm.Api)
	require.Equal(t, "vllm", llm.Api.Stack)
	require.Equal(t, "qwen3", llm.Api.ModelFamily)
	require.Equal(t, "chat_template_kwargs.enable_thinking", llm.Api.Bindings[system.IntentReasoningDisable].Param)
	require.NotZero(t, llm.Api.DeclaredAt)

	noBid := reportByID(t, reports, modelNoBid)
	require.Equal(t, system.ModelHealthStatusNoBid, noBid.Status)
	require.NotNil(t, noBid.Api)
	require.Equal(t, "ollama", noBid.Api.Stack)
}

func TestCheckModelWithoutApiStackHasNoApiSpec(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
		configs: map[common.Hash]config.ModelConfig{
			modelLLM: {ModelName: "qwen3-32b", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"},
		},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.NotEqual(t, system.ModelHealthStatusNoBid, llm.Status, "the model is still checked")
	require.Nil(t, llm.Api, "report carries api only for models with apiStack")
}

func TestCheckModelDeclaredAtStableAcrossSweeps(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
		configs: map[common.Hash]config.ModelConfig{
			modelLLM: {ModelName: "qwen3-32b", ApiType: "openai", ApiStack: "vllm", ApiURL: "http://llm:8000/v1/chat/completions", ModelFamily: "qwen3"},
		},
	}

	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	first := reportByID(t, checker.GetReports(), modelLLM)
	require.NotNil(t, first.Api)
	require.NotZero(t, first.Api.DeclaredAt)

	// Seeded directly: wall-clock granularity between sweeps is not reliable.
	seeded := first
	seededAPI := *first.Api
	seededAPI.DeclaredAt = 12345
	seeded.Api = &seededAPI
	checker.setReport(seeded)

	checker.checkAll(context.Background(), common.Address{})
	second := reportByID(t, checker.GetReports(), modelLLM)
	require.NotNil(t, second.Api)
	require.Equal(t, int64(12345), second.Api.DeclaredAt, "unchanged spec must keep the previous DeclaredAt")

	cfg := deps.configs[modelLLM]
	cfg.ModelFamily = "deepseek-r1"
	deps.configs[modelLLM] = cfg
	checker.checkAll(context.Background(), common.Address{})
	third := reportByID(t, checker.GetReports(), modelLLM)
	require.NotNil(t, third.Api)
	require.NotZero(t, third.Api.DeclaredAt)
	require.NotEqual(t, int64(12345), third.Api.DeclaredAt, "changed spec must refresh DeclaredAt")
}

func TestCheckModelDeclaredSpecCarriesSource(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelLLM)},
		tags:     map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM},
		adapter:  &mathSolvingAdapter{},
		configs:  map[common.Hash]config.ModelConfig{modelLLM: {ModelName: "qwen3-32b", ApiType: "openai", ApiStack: "vllm", ApiURL: "http://llm:8000/v1/chat/completions"}},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.NotNil(t, llm.Api)
	require.Equal(t, system.ApiSpecSourceDeclared, llm.Api.Source)
}

type fakeDetector struct {
	mu        sync.Mutex
	calls     int
	deadlines []time.Time
	spec      *system.ModelApiSpec
}

func (f *fakeDetector) Detect(ctx context.Context, cfg config.ModelConfig) *system.ModelApiSpec {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if dl, ok := ctx.Deadline(); ok {
		f.deadlines = append(f.deadlines, dl)
	}
	return f.spec.Clone()
}

func detectedQwen3() *system.ModelApiSpec {
	return &system.ModelApiSpec{Stack: "vllm", ModelFamily: "qwen3", Source: system.ApiSpecSourceDetected, Thinking: &system.ThinkingSpec{Mode: system.ThinkingModeControllable},
		Bindings: map[string]*system.ParamBinding{system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", ParamType: "boolean", Value: false}}}
}

func TestCheckModelDetectsWhenNoApiStack(t *testing.T) {
	det := &fakeDetector{spec: detectedQwen3()}
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM, modelNoBid}, adapter: &mathSolvingAdapter{}, apiDetect: det,
		configs: map[common.Hash]config.ModelConfig{
			modelLLM:   {ModelName: "qwen3-32b", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"},
			modelNoBid: {ModelName: "qwen3-8b", ApiType: "openai", ApiURL: "http://llm:8001/v1/chat/completions"},
		},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	reports := checker.GetReports()
	for _, id := range []common.Hash{modelLLM, modelNoBid} {
		r := reportByID(t, reports, id)
		require.NotNil(t, r.Api, id.Hex())
		require.Equal(t, system.ApiSpecSourceDetected, r.Api.Source)
		require.Equal(t, "qwen3", r.Api.ModelFamily)
		require.NotZero(t, r.Api.DeclaredAt)
	}
	require.Equal(t, system.ModelHealthStatusNoBid, reportByID(t, reports, modelNoBid).Status)
	require.Equal(t, 2, det.calls)
}

func TestCheckModelSkipsDetectorForNonChatTransport(t *testing.T) {
	det := &fakeDetector{spec: detectedQwen3()}
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM, modelImage}, adapter: &mathSolvingAdapter{}, apiDetect: det,
		configs: map[common.Hash]config.ModelConfig{
			modelLLM:   {ModelName: "qwen3-32b", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"},
			modelImage: {ModelName: "sd-xl", ApiType: "prodia-v2", ApiURL: "https://inference.prodia.com/v2"},
		},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	reports := checker.GetReports()

	llm := reportByID(t, reports, modelLLM)
	require.NotNil(t, llm.Api, "chat transport (openai) is still detected")
	require.Equal(t, system.ApiSpecSourceDetected, llm.Api.Source)

	image := reportByID(t, reports, modelImage)
	require.Nil(t, image.Api, "non-chat transport (prodia-v2) must not be probed")

	require.Equal(t, 1, det.calls, "the detector must be called only for the chat-transport model")
}

func TestCheckModelDeclaredStackSkipsDetector(t *testing.T) {
	det := &fakeDetector{spec: detectedQwen3()}
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM}, adapter: &mathSolvingAdapter{}, apiDetect: det,
		configs: map[common.Hash]config.ModelConfig{modelLLM: {ModelName: "deepseek-r1", ApiType: "openai", ApiStack: "sglang", ApiURL: "http://llm:8000/v1/chat/completions"}},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	llm := reportByID(t, checker.GetReports(), modelLLM)
	require.Equal(t, "sglang", llm.Api.Stack)
	require.Equal(t, system.ApiSpecSourceDeclared, llm.Api.Source)
	require.Equal(t, 0, det.calls)
}

func TestCheckModelWithoutDetectorHasNoDetectedApi(t *testing.T) {
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM, modelNoBid}, adapter: &mathSolvingAdapter{},
		configs: map[common.Hash]config.ModelConfig{
			modelLLM:   {ModelName: "qwen3-32b", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"},
			modelNoBid: {ModelName: "qwen3-8b", ApiType: "openai", ApiStack: "vllm", ApiURL: "http://llm:8001/v1/chat/completions"},
		},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	require.Nil(t, reportByID(t, checker.GetReports(), modelLLM).Api)
	require.Equal(t, system.ApiSpecSourceDeclared, reportByID(t, checker.GetReports(), modelNoBid).Api.Source)
}

func TestCheckModelDetectorNilResultHasNoApi(t *testing.T) {
	det := &fakeDetector{}
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM}, adapter: &mathSolvingAdapter{}, apiDetect: det,
		configs: map[common.Hash]config.ModelConfig{modelLLM: {ModelName: "unknown-model-x", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"}},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	require.Nil(t, reportByID(t, checker.GetReports(), modelLLM).Api)
	require.Equal(t, 1, det.calls)
}

func TestCheckModelDetectBudgetIsCappedByTimeout(t *testing.T) {
	newDeps := func(det *fakeDetector) *mockDeps {
		return &mockDeps{
			bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
			modelIDs: []common.Hash{modelLLM}, adapter: &mathSolvingAdapter{}, apiDetect: det,
			configs: map[common.Hash]config.ModelConfig{modelLLM: {ModelName: "qwen3-32b", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"}},
		}
	}
	det := &fakeDetector{spec: detectedQwen3()}
	start := time.Now()
	newTestChecker(newDeps(det)).checkAll(context.Background(), common.Address{})
	require.Len(t, det.deadlines, 1)
	require.WithinDuration(t, start.Add(time.Second), det.deadlines[0], 300*time.Millisecond)

	det = &fakeDetector{spec: detectedQwen3()}
	deps := newDeps(det)
	checker := NewChecker(Deps{Adapters: deps, Bids: deps, Models: deps, ModelConfigs: deps, ApiDetect: det}, time.Hour, time.Hour, 0, 0, lib.NewTestLogger())
	start = time.Now()
	checker.checkAll(context.Background(), common.Address{})
	require.Len(t, det.deadlines, 1)
	require.WithinDuration(t, start.Add(detectBudget), det.deadlines[0], 300*time.Millisecond)
}

func TestCheckModelDetectBudgetFallsBackWhenTimeoutIsZero(t *testing.T) {
	det := &fakeDetector{spec: detectedQwen3()}
	deps := &mockDeps{
		modelIDs:  []common.Hash{modelNoBid},
		apiDetect: det,
		configs:   map[common.Hash]config.ModelConfig{modelNoBid: {ModelName: "qwen3-32b", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"}},
	}
	checker := NewChecker(Deps{Adapters: deps, Bids: deps, Models: deps, ModelConfigs: deps, ApiDetect: det}, time.Hour, 0, 0, 0, lib.NewTestLogger())
	start := time.Now()
	checker.checkAll(context.Background(), common.Address{})
	require.Len(t, det.deadlines, 1)
	require.WithinDuration(t, start.Add(detectBudget), det.deadlines[0], 300*time.Millisecond)
}

func TestCheckModelDetectedDeclaredAtStableAcrossSweeps(t *testing.T) {
	det := &fakeDetector{spec: detectedQwen3()}
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
		modelIDs: []common.Hash{modelLLM}, adapter: &mathSolvingAdapter{}, apiDetect: det,
		configs: map[common.Hash]config.ModelConfig{modelLLM: {ModelName: "qwen3-32b", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"}},
	}
	checker := newTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	first := reportByID(t, checker.GetReports(), modelLLM)
	require.NotZero(t, first.Api.DeclaredAt)

	seeded := first
	seededAPI := *first.Api
	seededAPI.DeclaredAt = 12345
	seeded.Api = &seededAPI
	checker.setReport(seeded)

	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, int64(12345), reportByID(t, checker.GetReports(), modelLLM).Api.DeclaredAt, "unchanged detected spec keeps DeclaredAt")

	det.mu.Lock()
	det.spec.ModelFamily = "deepseek-r1"
	det.mu.Unlock()
	checker.checkAll(context.Background(), common.Address{})
	third := reportByID(t, checker.GetReports(), modelLLM)
	require.NotEqual(t, int64(12345), third.Api.DeclaredAt, "changed detected spec refreshes DeclaredAt")
}

func declaredContext(v int64) *config.DeclaredCapabilities {
	return &config.DeclaredCapabilities{Context: &v}
}

func TestCheckModelPublishesDeclaredCapabilitiesWithoutDetection(t *testing.T) {
	t.Run("detection off", func(t *testing.T) {
		deps := &mockDeps{
			bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
			modelIDs: []common.Hash{modelLLM, modelNoBid}, adapter: &mathSolvingAdapter{},
			configs: map[common.Hash]config.ModelConfig{
				modelLLM:   {ModelName: "my-private-alias", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions", Capabilities: declaredContext(8192)},
				modelNoBid: {ModelName: "qwen3-8b", ApiType: "openai", ApiURL: "http://llm:8001/v1/chat/completions"},
			},
		}
		checker := newTestChecker(deps)
		checker.checkAll(context.Background(), common.Address{})

		llm := reportByID(t, checker.GetReports(), modelLLM)
		require.NotNil(t, llm.Api, "the declared block is published without a detector")
		require.Equal(t, "", llm.Api.Stack)
		require.Equal(t, system.ApiSpecSourceDeclared, llm.Api.Source)
		require.Equal(t, int64(8192), *llm.Api.Capabilities.Context)
		require.Equal(t, "declared", llm.Api.Evidence["capabilities.context"])
		require.NotZero(t, llm.Api.DeclaredAt)
		require.Nil(t, reportByID(t, checker.GetReports(), modelNoBid).Api, "nothing declared, nothing detected: no api block")
	})

	t.Run("transport the detector does not cover", func(t *testing.T) {
		det := &fakeDetector{spec: detectedQwen3()}
		deps := &mockDeps{
			bids: []*structs.Bid{bidFor(modelLLM)}, tags: map[common.Hash][]string{modelLLM: {"llm"}},
			modelIDs: []common.Hash{modelLLM, modelImage}, adapter: &mathSolvingAdapter{}, apiDetect: det,
			configs: map[common.Hash]config.ModelConfig{
				modelLLM:   {ModelName: "qwen3-32b", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions", Capabilities: declaredContext(4096)},
				modelImage: {ModelName: "sd-xl", ApiType: "prodia-v2", ApiURL: "https://inference.prodia.com/v2", Capabilities: declaredContext(77)},
			},
		}
		checker := newTestChecker(deps)
		checker.checkAll(context.Background(), common.Address{})
		reports := checker.GetReports()

		image := reportByID(t, reports, modelImage)
		require.NotNil(t, image.Api)
		require.Equal(t, system.ApiSpecSourceDeclared, image.Api.Source)
		require.Equal(t, int64(77), *image.Api.Capabilities.Context)

		llm := reportByID(t, reports, modelLLM)
		require.Equal(t, system.ApiSpecSourceDetected, llm.Api.Source, "a chat-transport model with a block is still detected")
		require.Equal(t, 1, det.calls)
	})
}

// probeAwareDetector answers WouldProbe from a fixed set of model names and
// records when each Detect call arrived.
type probeAwareDetector struct {
	fakeDetector
	probes map[string]bool
	at     []time.Time
}

func (p *probeAwareDetector) WouldProbe(cfg config.ModelConfig) bool { return p.probes[cfg.ModelName] }

func (p *probeAwareDetector) Detect(ctx context.Context, cfg config.ModelConfig) *system.ModelApiSpec {
	p.mu.Lock()
	p.at = append(p.at, time.Now())
	p.mu.Unlock()
	return p.fakeDetector.Detect(ctx, cfg)
}

func TestCheckAllPacesDetectionsThatProbe(t *testing.T) {
	modelA, modelB, modelC := common.HexToHash("0xa1"), common.HexToHash("0xb2"), common.HexToHash("0xc3")
	newDeps := func(det ApiDetector) *mockDeps {
		return &mockDeps{
			modelIDs: []common.Hash{modelA, modelB, modelC}, adapter: &mathSolvingAdapter{}, apiDetect: det,
			configs: map[common.Hash]config.ModelConfig{
				modelA: {ModelName: "a", ApiType: "openai", ApiURL: "http://llm:8000/v1/chat/completions"},
				modelB: {ModelName: "b", ApiType: "openai", ApiURL: "http://llm:8001/v1/chat/completions"},
				modelC: {ModelName: "c", ApiType: "openai", ApiURL: "http://llm:8002/v1/chat/completions"},
			},
		}
	}
	const delay = 150 * time.Millisecond
	// pausesDuring runs one sweep with the pause recorded instead of slept
	// and returns, per pause, which detection calls had already happened.
	pausesDuring := func(t *testing.T, deps *mockDeps, calls func() int) []int {
		t.Helper()
		checker := NewChecker(Deps{Adapters: deps, Bids: deps, Models: deps, ModelConfigs: deps, ApiDetect: deps.apiDetect}, time.Hour, time.Second, delay, 0, lib.NewTestLogger())
		var pauses []int
		checker.sleepFn = func(_ context.Context, d time.Duration) bool {
			require.Equal(t, delay, d)
			pauses = append(pauses, calls())
			return true
		}
		checker.checkAll(context.Background(), common.Address{})
		return pauses
	}

	t.Run("a cache miss after a cache miss waits probeDelay", func(t *testing.T) {
		det := &probeAwareDetector{fakeDetector: fakeDetector{spec: detectedQwen3()}, probes: map[string]bool{"a": true, "b": false, "c": true}}
		pauses := pausesDuring(t, newDeps(det), func() int { return len(det.at) })
		require.Len(t, det.at, 3, "no model has a bid, so every call is a detection")
		require.Equal(t, []int{2}, pauses, "one pause, before c's probe: b is a cache hit")
	})

	t.Run("cache hits are not paced", func(t *testing.T) {
		det := &probeAwareDetector{fakeDetector: fakeDetector{spec: detectedQwen3()}, probes: map[string]bool{}}
		pauses := pausesDuring(t, newDeps(det), func() int { return len(det.at) })
		require.Len(t, det.at, 3)
		require.Empty(t, pauses)
	})

	t.Run("a detector without WouldProbe is paced on every call", func(t *testing.T) {
		det := &fakeDetector{spec: detectedQwen3()}
		pauses := pausesDuring(t, newDeps(det), func() int { return det.calls })
		require.Equal(t, 3, det.calls)
		require.Equal(t, []int{1, 2}, pauses)
	})

	// A health probe and a detection cache miss reach the same upstream
	// account, so they pace each other in either order.
	t.Run("a cache miss after a health probe waits probeDelay", func(t *testing.T) {
		det := &probeAwareDetector{fakeDetector: fakeDetector{spec: detectedQwen3()}, probes: map[string]bool{"a": false, "b": true, "c": false}}
		deps := newDeps(det)
		deps.bids = []*structs.Bid{bidFor(modelA)}
		pauses := pausesDuring(t, deps, func() int { return len(det.at) })
		require.Len(t, det.at, 3)
		require.Equal(t, []int{1}, pauses, "b's cache miss follows a's health probe: paced; c is a cache hit without a bid")
	})

	t.Run("a health probe after a cache miss waits probeDelay", func(t *testing.T) {
		det := &probeAwareDetector{fakeDetector: fakeDetector{spec: detectedQwen3()}, probes: map[string]bool{"a": true, "b": false, "c": false}}
		deps := newDeps(det)
		deps.bids = []*structs.Bid{bidFor(modelB)}
		pauses := pausesDuring(t, deps, func() int { return len(det.at) })
		require.Len(t, det.at, 3)
		require.Equal(t, []int{1}, pauses, "b's health probe follows a's cache miss: paced; c is a cache hit without a bid")
	})

	t.Run("a cancelled pause ends the sweep", func(t *testing.T) {
		det := &probeAwareDetector{fakeDetector: fakeDetector{spec: detectedQwen3()}, probes: map[string]bool{"a": true, "b": true, "c": true}}
		deps := newDeps(det)
		checker := NewChecker(Deps{Adapters: deps, Bids: deps, Models: deps, ModelConfigs: deps, ApiDetect: deps.apiDetect}, time.Hour, time.Second, time.Hour, 0, lib.NewTestLogger())
		ctx, cancel := context.WithCancel(context.Background())
		go func() { time.Sleep(50 * time.Millisecond); cancel() }()
		checker.checkAll(ctx, common.Address{})
		require.Equal(t, 1, det.calls)
	})
}
