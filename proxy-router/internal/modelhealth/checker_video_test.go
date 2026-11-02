package modelhealth

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var (
	modelVideoBytePlus     = common.HexToHash("0x7a")
	modelVideoRenderful    = common.HexToHash("0x7b")
	modelVideoUntagged     = common.HexToHash("0x7c")
	modelVideoTaggedOpenAI = common.HexToHash("0x7d")
)

const (
	videoTestApiKey = "sk-vendor-secret-0001"
	videoTestApiURL = "https://ark.internal.example/api/v3"
)

// countingAdapter answers like mathSolvingAdapter but counts every request
// that would reach a backend.
type countingAdapter struct {
	mathSolvingAdapter
	calls atomic.Int32
}

func (a *countingAdapter) Prompt(ctx context.Context, compl *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	a.calls.Add(1)
	return a.mathSolvingAdapter.Prompt(ctx, compl, cb)
}

func (a *countingAdapter) Embeddings(ctx context.Context, req *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) error {
	a.calls.Add(1)
	return a.mathSolvingAdapter.Embeddings(ctx, req, cb)
}

func (a *countingAdapter) Decisions(ctx context.Context, req *gcs.DecisionsRequest, cb gcs.CompletionCallback) error {
	a.calls.Add(1)
	return a.mathSolvingAdapter.Decisions(ctx, req, cb)
}

func (a *countingAdapter) AudioTranscription(ctx context.Context, req *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	a.calls.Add(1)
	return a.mathSolvingAdapter.AudioTranscription(ctx, req, cb)
}

func (a *countingAdapter) AudioSpeech(ctx context.Context, req *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	a.calls.Add(1)
	return a.mathSolvingAdapter.AudioSpeech(ctx, req, cb)
}

// countingAdapters counts GetAdapter calls on top of mockDeps.
type countingAdapters struct {
	*mockDeps
	gets atomic.Int32
}

func (c *countingAdapters) GetAdapter(ctx context.Context, chatID, modelID, sessionID common.Hash, storeChatContext, forwardChatContext bool) (aiengine.AIEngineStream, error) {
	c.gets.Add(1)
	return c.mockDeps.GetAdapter(ctx, chatID, modelID, sessionID, storeChatContext, forwardChatContext)
}

func newVideoTestChecker(deps *mockDeps) (*Checker, *countingAdapters) {
	adapters := &countingAdapters{mockDeps: deps}
	return NewChecker(Deps{
		Adapters:     adapters,
		Bids:         deps,
		Models:       deps,
		ModelConfigs: deps,
	}, time.Hour, time.Second, 0, 0, lib.NewTestLogger()), adapters
}

func videoModelConfig(t *testing.T, raw string) config.ModelConfig {
	t.Helper()
	var cfg config.ModelConfig
	require.NoError(t, json.Unmarshal([]byte(raw), &cfg))
	return cfg
}

func guideFieldNames(guide *videojob.Guide) []string {
	names := make([]string, 0, len(guide.Fields))
	for _, f := range guide.Fields {
		names = append(names, f.Name)
	}
	return names
}

func TestCheckModelVideoApiTypesAreSkippedWithGuide(t *testing.T) {
	bytePlusCfg := videoModelConfig(t, `{"modelName":"seedance-1-0-pro-250528","apiType":"byteplus-video","apiUrl":"`+videoTestApiURL+`","apiKey":"`+videoTestApiKey+`"}`)
	renderfulCfg := videoModelConfig(t, `{"modelName":"seedance-2.5-t2v","apiType":"renderful","apiUrl":"https://api.renderful.example/api/v1","apiKey":"`+videoTestApiKey+`"}`)

	adapter := &countingAdapter{}
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelVideoBytePlus), bidFor(modelVideoRenderful), bidFor(modelVideoUntagged)},
		tags: map[common.Hash][]string{
			modelVideoBytePlus:  {"video"},
			modelVideoRenderful: {"T2V"},
			// The apiType decides even without a video tag.
			modelVideoUntagged: {"llm"},
		},
		names: map[common.Hash]string{
			modelVideoBytePlus:  "Seedance 1.0 Pro",
			modelVideoRenderful: "Seedance 2.5",
			modelVideoUntagged:  "Seedance untagged",
		},
		modelIDs: []common.Hash{modelVideoBytePlus, modelVideoRenderful, modelVideoUntagged},
		configs: map[common.Hash]config.ModelConfig{
			modelVideoBytePlus:  bytePlusCfg,
			modelVideoRenderful: renderfulCfg,
			modelVideoUntagged:  bytePlusCfg,
		},
		adapter: adapter,
	}

	checker, adapters := newVideoTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	reports := checker.GetReports()
	require.Len(t, reports, 3)

	cases := []struct {
		id        common.Hash
		modelType structs.ModelType
		modelName string
		maxDur    int
	}{
		{modelVideoBytePlus, structs.ModelTypeVIDEO, "seedance-1-0-pro-250528", 12},
		{modelVideoRenderful, structs.ModelTypeVIDEO, "seedance-2.5-t2v", 30},
		{modelVideoUntagged, structs.ModelTypeLLM, "seedance-1-0-pro-250528", 12},
	}
	for _, tc := range cases {
		r := reportByID(t, reports, tc.id)
		require.Equal(t, system.ModelHealthStatusSkipped, r.Status, tc.id.Hex())
		require.Empty(t, r.ErrorKind, tc.id.Hex())
		require.True(t, r.HasActiveBid)
		require.Equal(t, bidIDFor(tc.id).Hex(), r.BidID)
		require.Equal(t, string(tc.modelType), r.ModelType)
		require.Nil(t, r.PromptCorrect, "no prompt was judged")
		require.Zero(t, r.LastHealthy, "skipped is not healthy")

		require.NotNil(t, r.Video, "video guide for %s", tc.id.Hex())
		require.Equal(t, videojob.ObjectGuide, r.Video.Object)
		require.Equal(t, tc.modelName, r.Video.ModelName)
		names := guideFieldNames(r.Video)
		require.Contains(t, names, videojob.FieldPrompt)
		require.Contains(t, names, videojob.FieldDurationSeconds)
		require.Positive(t, r.Video.Quote.SecondsPerWorkUnit)
		for _, f := range r.Video.Fields {
			if f.Name == videojob.FieldPrompt {
				require.True(t, f.Required, "prompt is the one required field")
			}
			if f.Name == videojob.FieldDurationSeconds {
				require.NotNil(t, f.Maximum)
				require.Equal(t, tc.maxDur, *f.Maximum, "the vendor limit for the model")
			}
		}

		// The health report is public: nothing private from models-config
		// may travel with the guide.
		raw, err := json.Marshal(r)
		require.NoError(t, err)
		require.Contains(t, string(raw), `"video":{"object":"video.guide"`)
		require.NotContains(t, string(raw), videoTestApiKey)
		require.NotContains(t, string(raw), "ark.internal.example")
	}

	require.Zero(t, adapter.calls.Load(), "a video model is never sent a probe")
	require.Zero(t, adapters.gets.Load(), "no adapter is even built for a video model")
}

// A model the chain tags video but that is served through a chat adapter
// is skipped like any video model, and advertises no guide: there is no
// video adapter to describe it.
func TestCheckModelVideoTaggedNonVideoApiTypeSkippedWithoutGuide(t *testing.T) {
	adapter := &countingAdapter{}
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelVideoTaggedOpenAI)},
		tags:     map[common.Hash][]string{modelVideoTaggedOpenAI: {"video"}},
		modelIDs: []common.Hash{modelVideoTaggedOpenAI},
		configs: map[common.Hash]config.ModelConfig{
			modelVideoTaggedOpenAI: {ModelName: "some-model", ApiType: "openai", ApiURL: "http://h/v1/chat/completions"},
		},
		adapter: adapter,
	}

	checker, adapters := newVideoTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})

	r := reportByID(t, checker.GetReports(), modelVideoTaggedOpenAI)
	require.Equal(t, system.ModelHealthStatusSkipped, r.Status)
	require.Equal(t, string(structs.ModelTypeVIDEO), r.ModelType)
	require.Nil(t, r.Video)
	require.Nil(t, r.PromptCorrect)
	require.Zero(t, adapter.calls.Load())
	require.Zero(t, adapters.gets.Load())
}

// The guide is rendered from models-config alone, so a video model whose
// on-chain lookup fails or that has no bid yet still publishes it; only the
// status says why it is not sold.
func TestCheckModelVideoGuideWithoutBidOrMetadata(t *testing.T) {
	cfg := videoModelConfig(t, `{"modelName":"seedance-2.5-t2v","apiType":"renderful","apiUrl":"https://api.renderful.example/api/v1","apiKey":"k"}`)

	adapter := &countingAdapter{}
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelVideoRenderful)},
		// no tags entry for modelVideoRenderful: GetModelNameAndTags fails
		tags:     map[common.Hash][]string{},
		modelIDs: []common.Hash{modelVideoRenderful, modelVideoBytePlus},
		configs: map[common.Hash]config.ModelConfig{
			modelVideoRenderful: cfg,
			modelVideoBytePlus:  cfg,
		},
		adapter: adapter,
	}

	checker, _ := newVideoTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	reports := checker.GetReports()

	lookupFailed := reportByID(t, reports, modelVideoRenderful)
	require.Equal(t, system.ModelHealthStatusSkipped, lookupFailed.Status)
	require.Equal(t, system.ModelHealthErrorTypeLookup, lookupFailed.ErrorKind)
	require.NotNil(t, lookupFailed.Video)

	noBid := reportByID(t, reports, modelVideoBytePlus)
	require.Equal(t, system.ModelHealthStatusNoBid, noBid.Status)
	require.NotNil(t, noBid.Video)

	require.Zero(t, adapter.calls.Load())
}

// checkAll paces requests that reach a backend. A video model with a bid
// sends none (no probe, no detection), so it should not cost the sweep a
// probeDelay.
func TestCheckAllDoesNotPaceVideoModels(t *testing.T) {
	cfg := videoModelConfig(t, `{"modelName":"seedance-2.5-t2v","apiType":"renderful","apiUrl":"https://api.renderful.example/api/v1","apiKey":"k"}`)
	deps := &mockDeps{
		bids: []*structs.Bid{bidFor(modelVideoRenderful), bidFor(modelVideoBytePlus)},
		tags: map[common.Hash][]string{
			modelVideoRenderful: {"video"},
			modelVideoBytePlus:  {"video"},
		},
		modelIDs: []common.Hash{modelVideoRenderful, modelVideoBytePlus},
		configs: map[common.Hash]config.ModelConfig{
			modelVideoRenderful: cfg,
			modelVideoBytePlus:  cfg,
		},
		adapter: &countingAdapter{},
	}

	checker, _ := newVideoTestChecker(deps)
	checker.probeDelay = time.Minute
	var sleeps atomic.Int32
	checker.sleepFn = func(ctx context.Context, d time.Duration) bool {
		sleeps.Add(1)
		return true
	}
	checker.checkAll(context.Background(), common.Address{})

	for _, id := range []common.Hash{modelVideoRenderful, modelVideoBytePlus} {
		require.Equal(t, system.ModelHealthStatusSkipped, reportByID(t, checker.GetReports(), id).Status)
	}
	require.Zero(t, sleeps.Load(), "video models reach no backend during a sweep: no pacing")
}
