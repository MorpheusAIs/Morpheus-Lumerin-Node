package apidetect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

const vllmTestKey = "sk-vllm-operator"

var vllmStructuredWant = []string{"json_object", "json_schema", "grammar", "regex", "choice"}

// tokenizeHandler answers like vLLM's /tokenize (count + max_model_len) and
// records the model the request named.
func tokenizeHandler(t *testing.T, maxModelLen int, models *[]string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var body map[string]string
		require.NoError(t, json.Unmarshal(raw, &body))
		require.Equal(t, "x", body["prompt"])
		*models = append(*models, body["model"])
		jsonHandler(map[string]any{"count": 1, "max_model_len": maxModelLen, "tokens": []int{87}})(w, r)
	}
}

func bearerGuard(key string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+key {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func detectVLLM(t *testing.T, srvURL, modelName, apiKey string) (*system.ModelApiSpec, string) {
	t.Helper()
	api, lines := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{
		ModelName: modelName, ApiType: "openai", ApiURL: srvURL + "/v1/chat/completions", ApiKey: apiKey,
	})
	require.NotNil(t, api)
	require.Equal(t, "vllm", api.Stack)
	return api, strings.Join(lines, "\n")
}

func TestProbeVLLMContextFromModelsList(t *testing.T) {
	var tokenized []string
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.30.0"}))
	mux.HandleFunc("/v1/models", jsonHandler(map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "Qwen/Qwen3-235B-A22B", "object": "model", "owned_by": "vllm", "max_model_len": 40960},
		},
	}))
	mux.HandleFunc("/tokenize", tokenizeHandler(t, 8192, &tokenized))
	srv, log := loggingServer(t, mux)

	api, trace := detectVLLM(t, srv.URL, "Qwen/Qwen3-235B-A22B", "")

	require.NotNil(t, api.Capabilities)
	require.NotNil(t, api.Capabilities.Context)
	require.EqualValues(t, 40960, *api.Capabilities.Context)
	require.Equal(t, "vllm:/v1/models.max_model_len", api.Evidence[system.EvidenceKeyCapability("context")])
	require.Equal(t, vllmStructuredWant, api.Capabilities.Structured)
	require.Equal(t, "vllm:/version", api.Evidence[system.EvidenceKeyCapability("structured")])
	require.Nil(t, api.Capabilities.Tools, "tool support is not discoverable from vllm's endpoints")
	require.Contains(t, trace, "identified vllm by /version")
	require.Contains(t, trace, "vllm: /v1/models data[].max_model_len = 40960 -> capabilities.context")
	require.Contains(t, trace, "capabilities.structured")
	require.False(t, log.sawPath("/tokenize"), "the listing answered; no fallback request")
	require.Empty(t, tokenized)
	log.requireAnonymous(t)
}

func TestProbeVLLMGuardedModelsListFallsBackToTokenize(t *testing.T) {
	var tokenized []string
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.30.0"}))
	mux.HandleFunc("/v1/models", bearerGuard(vllmTestKey, jsonHandler(map[string]any{
		"object": "list",
		"data":   []map[string]any{{"id": "Qwen/Qwen3-32B", "owned_by": "vllm", "max_model_len": 32768}},
	})))
	mux.HandleFunc("/tokenize", tokenizeHandler(t, 32768, &tokenized))
	srv, log := loggingServer(t, mux)

	t.Run("wrong key: /v1/models 401, /tokenize is open", func(t *testing.T) {
		tokenized = nil
		api, trace := detectVLLM(t, srv.URL, "Qwen/Qwen3-32B", "not-the-operator-key")

		require.NotNil(t, api.Capabilities)
		require.NotNil(t, api.Capabilities.Context)
		require.EqualValues(t, 32768, *api.Capabilities.Context)
		require.Equal(t, "vllm:/tokenize.max_model_len", api.Evidence[system.EvidenceKeyCapability("context")])
		require.Equal(t, vllmStructuredWant, api.Capabilities.Structured)
		require.Equal(t, []string{"Qwen/Qwen3-32B"}, tokenized, "the configured name is sent when the listing gave no served id")
		require.Contains(t, trace, "vllm: /tokenize max_model_len = 32768 -> capabilities.context")
		require.Contains(t, trace, "GET "+srv.URL+"/v1/models -> HTTP 401")

		log.mu.Lock()
		defer log.mu.Unlock()
		for i, p := range log.paths {
			switch p {
			case "/version", "/tokenize":
				require.Equal(t, "", log.auth[i], "%s is documented open: requested without a key", p)
			case "/v1/models":
				require.Equal(t, "Bearer not-the-operator-key", log.auth[i], "/v1/models keeps the key")
			}
		}
	})

	t.Run("right key: the listing answers and /tokenize is skipped", func(t *testing.T) {
		tokenized = nil
		api, _ := detectVLLM(t, srv.URL, "Qwen/Qwen3-32B", vllmTestKey)
		require.EqualValues(t, 32768, *api.Capabilities.Context)
		require.Equal(t, "vllm:/v1/models.max_model_len", api.Evidence[system.EvidenceKeyCapability("context")])
		require.Empty(t, tokenized)
	})
}

func TestProbeVLLMLegacyListingWithoutMaxModelLen(t *testing.T) {
	var tokenized []string
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.4.3"}))
	mux.HandleFunc("/v1/models", jsonHandler(map[string]any{
		"object": "list",
		"data":   []map[string]any{{"id": "deepseek-ai/DeepSeek-R1", "object": "model", "owned_by": "vllm"}},
	}))
	mux.HandleFunc("/tokenize", tokenizeHandler(t, 16384, &tokenized))
	srv, log := loggingServer(t, mux)

	api, trace := detectVLLM(t, srv.URL, "DeepSeek-R1", "")

	require.NotNil(t, api.Capabilities)
	require.NotNil(t, api.Capabilities.Context)
	require.EqualValues(t, 16384, *api.Capabilities.Context)
	require.Equal(t, "vllm:/tokenize.max_model_len", api.Evidence[system.EvidenceKeyCapability("context")])
	require.Equal(t, []string{"deepseek-ai/DeepSeek-R1"}, tokenized, "the served id from the listing names the model")
	require.Contains(t, trace, "vllm: /tokenize max_model_len = 16384 -> capabilities.context")
	require.True(t, log.sawPath("/v1/models"))
	log.requireAnonymous(t)
}

func TestProbeVLLMTokenizeGuardedRetriesWithKey(t *testing.T) {
	var tokenized []string
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.30.0"}))
	mux.HandleFunc("/v1/models", bearerGuard(vllmTestKey, jsonHandler(map[string]any{
		"object": "list", "data": []map[string]any{{"id": "m", "owned_by": "vllm"}},
	})))
	mux.HandleFunc("/tokenize", bearerGuard(vllmTestKey, tokenizeHandler(t, 4096, &tokenized)))
	srv, log := loggingServer(t, mux)

	api, trace := detectVLLM(t, srv.URL, "m", vllmTestKey)

	require.EqualValues(t, 4096, *api.Capabilities.Context)
	require.Equal(t, "vllm:/tokenize.max_model_len", api.Evidence[system.EvidenceKeyCapability("context")])
	require.Contains(t, trace, "POST "+srv.URL+"/tokenize -> HTTP 401 without a key; retrying with the configured key")

	var tokenizeAuth []string
	log.mu.Lock()
	for i, p := range log.paths {
		if p == "/tokenize" {
			tokenizeAuth = append(tokenizeAuth, log.auth[i])
		}
	}
	log.mu.Unlock()
	require.Equal(t, []string{"", "Bearer " + vllmTestKey}, tokenizeAuth, "key-free first, the key only after a 401")
}

func TestProbeVLLMNothingKnownLeavesContextUnset(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.30.0"}))
	mux.HandleFunc("/v1/models", jsonHandler(map[string]any{"object": "list", "data": []map[string]any{{"id": "m", "owned_by": "vllm"}}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api, _ := detectVLLM(t, srv.URL, "m", "")
	require.NotNil(t, api.Capabilities)
	require.Nil(t, api.Capabilities.Context, "a 404 from /tokenize is no evidence")
	require.Equal(t, vllmStructuredWant, api.Capabilities.Structured)
}

func TestProbeVLLMRejectsNonPositiveOrHugeMaxModelLen(t *testing.T) {
	cases := []struct {
		name string
		v    any
	}{
		{"zero", 0}, {"negative", -1}, {"over the int32 bound", int64(1) << 31}, {"fraction", 1.5}, {"string", "40960"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.30.0"}))
			mux.HandleFunc("/v1/models", jsonHandler(map[string]any{
				"object": "list", "data": []map[string]any{{"id": "m", "owned_by": "vllm", "max_model_len": c.v}},
			}))
			mux.HandleFunc("/tokenize", jsonHandler(map[string]any{"count": 1, "max_model_len": c.v}))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			api, _ := detectVLLM(t, srv.URL, "m", "")
			require.Nil(t, api.Capabilities.Context)
		})
	}
}

func TestJSONPositiveInt(t *testing.T) {
	cases := []struct {
		v    any
		want int64
		ok   bool
	}{
		{float64(40960), 40960, true},
		{float64(1), 1, true},
		{float64(0), 0, false},
		{float64(-5), 0, false},
		{1.5, 0, false},
		{1e30, 0, false},
		{"7", 0, false},
		{nil, 0, false},
	}
	for _, c := range cases {
		got, ok := jsonPositiveInt(c.v)
		require.Equal(t, c.ok, ok, "%v", c.v)
		require.Equal(t, c.want, got, "%v", c.v)
	}
}
