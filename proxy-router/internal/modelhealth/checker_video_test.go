package modelhealth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
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

const videoTestApiKey = "sk-vendor-secret-0001"

// videoVendorStub answers every request with status, as a video vendor's
// free probe endpoints would, and counts what it received.
type videoVendorStub struct {
	*httptest.Server
	mu     sync.Mutex
	hits   []string // method, path and Authorization of each request
	status atomic.Int32
	delay  atomic.Int64 // nanoseconds to wait before answering
}

func newVideoVendorStub(t *testing.T, status int) *videoVendorStub {
	t.Helper()
	v := &videoVendorStub{}
	v.status.Store(int32(status))
	v.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v.mu.Lock()
		v.hits = append(v.hits, r.Method+" "+r.URL.Path+" "+r.Header.Get("Authorization"))
		v.mu.Unlock()
		if d := time.Duration(v.delay.Load()); d > 0 {
			select {
			case <-r.Context().Done():
			case <-time.After(d):
			}
		}
		status := int(v.status.Load())
		if r.Header.Get("Authorization") == "" {
			status = http.StatusUnauthorized // both vendors refuse a keyless request
		}
		// The answers the real vendors give: the BytePlus task list, the
		// Renderful "not found" for an unknown generation, an error otherwise.
		body := `{"error":{"code":"AuthenticationError","message":"key of account 1234567890 rejected"}}`
		switch status {
		case http.StatusOK:
			body = `{"total":0,"items":[]}`
		case http.StatusNotFound:
			body = `{"error":"Not found","message":"Generation task not found"}`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(v.Close)
	return v
}

func (v *videoVendorStub) requests() []string {
	v.mu.Lock()
	defer v.mu.Unlock()
	return slices.Clone(v.hits)
}

func bytePlusConfig(t *testing.T, stub *videoVendorStub, modelName string) config.ModelConfig {
	return videoModelConfig(t, `{"modelName":"`+modelName+`","apiType":"byteplus-video","apiUrl":"`+stub.URL+`/api/v3","apiKey":"`+videoTestApiKey+`"}`)
}

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

// A hosted video model is checked with its vendor's free probe, never with
// the LLM prompt, and publishes its job guide.
func TestCheckModelVideoApiTypesProbedWithGuide(t *testing.T) {
	bytePlusVendor := newVideoVendorStub(t, http.StatusOK)
	renderfulVendor := newVideoVendorStub(t, http.StatusNotFound)
	bytePlusCfg := bytePlusConfig(t, bytePlusVendor, "seedance-1-0-pro-250528")
	renderfulCfg := videoModelConfig(t, `{"modelName":"seedance-2.5-t2v","apiType":"renderful","apiUrl":"`+renderfulVendor.URL+`/api/v1","apiKey":"`+videoTestApiKey+`"}`)

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
		require.Equal(t, system.ModelHealthStatusHealthy, r.Status, tc.id.Hex())
		require.Empty(t, r.ErrorKind, tc.id.Hex())
		require.True(t, r.HasActiveBid)
		require.Equal(t, bidIDFor(tc.id).Hex(), r.BidID)
		require.Equal(t, string(tc.modelType), r.ModelType)
		require.Nil(t, r.PromptCorrect, "no prompt was judged")
		require.NotZero(t, r.LastHealthy)

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
		require.NotContains(t, string(raw), strings.TrimPrefix(bytePlusVendor.URL, "http://"))
		require.NotContains(t, string(raw), strings.TrimPrefix(renderfulVendor.URL, "http://"))
	}

	require.Equal(t, []string{
		"GET /api/v3/contents/generations/tasks Bearer " + videoTestApiKey,
		"GET /api/v3/contents/generations/tasks Bearer " + videoTestApiKey,
	}, bytePlusVendor.requests(), "one free probe per BytePlus model")
	require.Equal(t, []string{
		"GET /api/v1/generations/morpheus-health-probe Bearer " + videoTestApiKey,
		"GET /api/v1/generations/morpheus-health-probe ",
	}, renderfulVendor.requests(), "the key, then the keyless control request")
	require.Zero(t, adapter.calls.Load(), "a video model is never sent the LLM prompt")
	require.Zero(t, adapters.gets.Load(), "no chat adapter is built for a video model")
}

// A probe the vendor refuses is reported like any failed probe: unhealthy
// with the HTTP status, degraded when throttled, connection when the vendor
// does not answer. The vendor's message is not in the report.
func TestCheckModelVideoProbeFailures(t *testing.T) {
	// Port 1 refuses at once on the CI's Linux runners, so the case cannot
	// turn into a timeout or reach a listener that took a freed port.
	refused := &videoVendorStub{Server: &httptest.Server{URL: "http://127.0.0.1:1"}}
	hanging := newVideoVendorStub(t, http.StatusOK)
	hanging.delay.Store(int64(3 * time.Second)) // the test checker's timeout is 1 s

	cases := []struct {
		name       string
		stub       *videoVendorStub
		status     string
		errorKind  string
		httpStatus int
	}{
		{"revoked key", newVideoVendorStub(t, http.StatusUnauthorized), system.ModelHealthStatusUnhealthy, system.ModelHealthErrorBadResponse, http.StatusUnauthorized},
		{"no permission", newVideoVendorStub(t, http.StatusForbidden), system.ModelHealthStatusUnhealthy, system.ModelHealthErrorBadResponse, http.StatusForbidden},
		{"throttled", newVideoVendorStub(t, http.StatusTooManyRequests), system.ModelHealthStatusDegraded, system.ModelHealthErrorRateLimited, http.StatusTooManyRequests},
		{"vendor down", newVideoVendorStub(t, http.StatusServiceUnavailable), system.ModelHealthStatusUnhealthy, system.ModelHealthErrorBadResponse, http.StatusServiceUnavailable},
		{"unreachable", refused, system.ModelHealthStatusUnhealthy, system.ModelHealthErrorConnection, 0},
		{"no answer in time", hanging, system.ModelHealthStatusUnhealthy, system.ModelHealthErrorTimeout, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := &mockDeps{
				bids:     []*structs.Bid{bidFor(modelVideoBytePlus)},
				tags:     map[common.Hash][]string{modelVideoBytePlus: {"video"}},
				modelIDs: []common.Hash{modelVideoBytePlus},
				configs:  map[common.Hash]config.ModelConfig{modelVideoBytePlus: bytePlusConfig(t, tc.stub, "seedance-1-0-pro-250528")},
				adapter:  &countingAdapter{},
			}
			checker, _ := newVideoTestChecker(deps)
			checker.checkAll(context.Background(), common.Address{})

			r := reportByID(t, checker.GetReports(), modelVideoBytePlus)
			require.Equal(t, tc.status, r.Status)
			require.Equal(t, tc.errorKind, r.ErrorKind)
			require.Equal(t, tc.httpStatus, r.HttpStatus)
			require.Zero(t, r.LastHealthy)
			require.NotNil(t, r.Video, "the guide is published whatever the probe says")
			raw, err := json.Marshal(r)
			require.NoError(t, err)
			require.NotContains(t, string(raw), "1234567890", "the vendor's message stays in the provider's log")
		})
	}
}

// Session failures flip a video model like any other; a healthy vendor probe
// ends a streak it can see the cause of (here 5xx), and a failing probe does
// not end it.
func TestCheckAllVideoProbeEndsSessionFailures(t *testing.T) {
	vendor := newVideoVendorStub(t, http.StatusOK)
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelVideoBytePlus)},
		tags:     map[common.Hash][]string{modelVideoBytePlus: {"video"}},
		modelIDs: []common.Hash{modelVideoBytePlus},
		configs:  map[common.Hash]config.ModelConfig{modelVideoBytePlus: bytePlusConfig(t, vendor, "seedance-1-0-pro-250528")},
		adapter:  &countingAdapter{},
	}
	checker, _ := newVideoTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelVideoBytePlus).Status)

	for i := 0; i < DefaultMaxConsecutiveErrors; i++ {
		checker.ReportPromptFailure(modelVideoBytePlus, http.StatusInternalServerError)
	}
	r := reportByID(t, checker.GetReports(), modelVideoBytePlus)
	require.Equal(t, system.ModelHealthStatusUnhealthy, r.Status)
	require.Zero(t, r.HttpStatus, "a 5xx streak carries no status, as before")
	lastHealthy := r.LastHealthy
	require.NotZero(t, lastHealthy)

	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelVideoBytePlus).Status)

	checker.ReportPromptFailure(modelVideoBytePlus, http.StatusInternalServerError)
	require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelVideoBytePlus).Status,
		"one failure after a healthy probe is the start of a new streak")

	// A failing probe does not end the streak, and keeps the last healthy time.
	vendor.status.Store(http.StatusUnauthorized)
	checker.checkAll(context.Background(), common.Address{})
	r = reportByID(t, checker.GetReports(), modelVideoBytePlus)
	require.Equal(t, system.ModelHealthStatusUnhealthy, r.Status)
	require.Equal(t, http.StatusUnauthorized, r.HttpStatus)
	require.GreaterOrEqual(t, r.LastHealthy, lastHealthy)
	for i := 1; i < DefaultMaxConsecutiveErrors; i++ {
		checker.ReportPromptFailure(modelVideoBytePlus, http.StatusInternalServerError)
	}
	r = reportByID(t, checker.GetReports(), modelVideoBytePlus)
	require.Equal(t, system.ModelHealthStatusUnhealthy, r.Status)
	require.Equal(t, system.ModelHealthErrorSessionErrors, r.ErrorKind)
	require.Zero(t, r.HttpStatus, "the probe's 401 does not stay on a session flip")
}

// The probe cannot see an empty balance (402) or a model the account has not
// activated (403, 404). A passing probe then reports degraded rather than
// healthy: the model stays routable so a job can show it fixed, one more
// failure flips it back at once, and a success heals it.
func TestCheckAllVideoProbeBlindStreakDegrades(t *testing.T) {
	for _, blind := range []int{http.StatusPaymentRequired, http.StatusForbidden, http.StatusNotFound} {
		t.Run(http.StatusText(blind), func(t *testing.T) {
			vendor := newVideoVendorStub(t, http.StatusOK)
			deps := &mockDeps{
				bids:     []*structs.Bid{bidFor(modelVideoBytePlus)},
				tags:     map[common.Hash][]string{modelVideoBytePlus: {"video"}},
				modelIDs: []common.Hash{modelVideoBytePlus},
				configs:  map[common.Hash]config.ModelConfig{modelVideoBytePlus: bytePlusConfig(t, vendor, "seedance-1-0-pro-250528")},
				adapter:  &countingAdapter{},
			}
			checker, _ := newVideoTestChecker(deps)
			checker.checkAll(context.Background(), common.Address{})

			for i := 0; i < DefaultMaxConsecutiveErrors; i++ {
				checker.ReportPromptFailure(modelVideoBytePlus, blind)
			}
			require.Equal(t, system.ModelHealthStatusUnhealthy, reportByID(t, checker.GetReports(), modelVideoBytePlus).Status)

			checker.checkAll(context.Background(), common.Address{})
			r := reportByID(t, checker.GetReports(), modelVideoBytePlus)
			require.Equal(t, system.ModelHealthStatusDegraded, r.Status, "the probe passes, the jobs did not")
			require.Equal(t, system.ModelHealthErrorSessionErrors, r.ErrorKind)
			require.Equal(t, blind, r.HttpStatus)

			checker.ReportPromptFailure(modelVideoBytePlus, blind)
			require.Equal(t, system.ModelHealthStatusUnhealthy, reportByID(t, checker.GetReports(), modelVideoBytePlus).Status,
				"one more failure flips it back")

			// The hold outlasts quiet sweeps: nothing has shown it fixed.
			for i := 0; i < 3; i++ {
				checker.checkAll(context.Background(), common.Address{})
				r = reportByID(t, checker.GetReports(), modelVideoBytePlus)
				require.Equal(t, system.ModelHealthStatusDegraded, r.Status, "sweep %d", i)
				require.Equal(t, blind, r.HttpStatus)
			}
			checker.ReportPromptSuccess(modelVideoBytePlus)
			require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelVideoBytePlus).Status)
			checker.checkAll(context.Background(), common.Address{})
			require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), modelVideoBytePlus).Status)
		})
	}
}

// Only the streak's account statuses matter, not their order: a 402 early in
// a streak of 5xx still holds the model.
func TestCheckAllVideoProbeBlindStreakAnyOrder(t *testing.T) {
	for name, streak := range map[string][]int{
		"402 first": {http.StatusPaymentRequired, http.StatusBadGateway, http.StatusBadGateway},
		"402 last":  {http.StatusBadGateway, http.StatusBadGateway, http.StatusPaymentRequired},
		"then none": {http.StatusPaymentRequired, http.StatusPaymentRequired, 0},
	} {
		t.Run(name, func(t *testing.T) {
			vendor := newVideoVendorStub(t, http.StatusOK)
			deps := &mockDeps{
				bids:     []*structs.Bid{bidFor(modelVideoBytePlus)},
				tags:     map[common.Hash][]string{modelVideoBytePlus: {"video"}},
				modelIDs: []common.Hash{modelVideoBytePlus},
				configs:  map[common.Hash]config.ModelConfig{modelVideoBytePlus: bytePlusConfig(t, vendor, "seedance-1-0-pro-250528")},
				adapter:  &countingAdapter{},
			}
			checker, _ := newVideoTestChecker(deps)
			checker.checkAll(context.Background(), common.Address{})
			require.Len(t, streak, DefaultMaxConsecutiveErrors)
			for _, status := range streak {
				checker.ReportPromptFailure(modelVideoBytePlus, status)
			}
			r := reportByID(t, checker.GetReports(), modelVideoBytePlus)
			require.Equal(t, system.ModelHealthStatusUnhealthy, r.Status)
			require.Equal(t, http.StatusPaymentRequired, r.HttpStatus, "the flip shows the account status")

			checker.checkAll(context.Background(), common.Address{})
			r = reportByID(t, checker.GetReports(), modelVideoBytePlus)
			require.Equal(t, system.ModelHealthStatusDegraded, r.Status)
			require.Equal(t, http.StatusPaymentRequired, r.HttpStatus)
		})
	}
}

// A model tagged video but served through a chat adapter has no video probe:
// it stays skipped, and each sweep starts its session-failure streak over.
func TestCheckAllResetsTaggedVideoChatModelFailures(t *testing.T) {
	deps := &mockDeps{
		bids:     []*structs.Bid{bidFor(modelVideoTaggedOpenAI)},
		tags:     map[common.Hash][]string{modelVideoTaggedOpenAI: {"video"}},
		modelIDs: []common.Hash{modelVideoTaggedOpenAI},
		configs: map[common.Hash]config.ModelConfig{
			modelVideoTaggedOpenAI: {ModelName: "some-model", ApiType: "openai", ApiURL: "http://h/v1/chat/completions"},
		},
		adapter: &countingAdapter{},
	}
	checker, _ := newVideoTestChecker(deps)
	checker.checkAll(context.Background(), common.Address{})
	for i := 0; i < DefaultMaxConsecutiveErrors; i++ {
		checker.ReportPromptFailure(modelVideoTaggedOpenAI, http.StatusInternalServerError)
	}
	require.Equal(t, system.ModelHealthStatusUnhealthy, reportByID(t, checker.GetReports(), modelVideoTaggedOpenAI).Status)

	checker.checkAll(context.Background(), common.Address{})
	require.Equal(t, system.ModelHealthStatusSkipped, reportByID(t, checker.GetReports(), modelVideoTaggedOpenAI).Status)
	checker.ReportPromptFailure(modelVideoTaggedOpenAI, http.StatusInternalServerError)
	require.Equal(t, system.ModelHealthStatusSkipped, reportByID(t, checker.GetReports(), modelVideoTaggedOpenAI).Status)
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

// checkAll paces requests that reach a backend; a video model's free vendor
// probe is one of them.
func TestCheckAllPacesVideoProbes(t *testing.T) {
	vendor := newVideoVendorStub(t, http.StatusNotFound)
	cfg := videoModelConfig(t, `{"modelName":"seedance-2.5-t2v","apiType":"renderful","apiUrl":"`+vendor.URL+`/api/v1","apiKey":"k"}`)
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
		require.Equal(t, system.ModelHealthStatusHealthy, reportByID(t, checker.GetReports(), id).Status)
	}
	require.EqualValues(t, 1, sleeps.Load(), "the second vendor probe waits for the pace, as any probe does")
	require.Len(t, vendor.requests(), 4, "two Renderful models, each the keyed probe and the keyless control")
}
