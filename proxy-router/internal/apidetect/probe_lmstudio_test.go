package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/stretchr/testify/require"
)

// Shapes: https://lmstudio.ai/docs/developer/rest/list (v1) and
// https://lmstudio.ai/docs/developer/rest/endpoints (v0).
func lmstudioV1Gemma() map[string]any {
	return map[string]any{
		"type": "llm", "publisher": "google", "key": "google/gemma-4-26b-a4b", "display_name": "Gemma 4 26B A4B",
		"architecture": "gemma4",
		"quantization": map[string]any{"name": "Q4_K_M", "bits_per_weight": 4},
		"size_bytes":   17990911801, "params_string": "26B-A4B",
		"loaded_instances": []map[string]any{{
			"id":     "google/gemma-4-26b-a4b",
			"config": map[string]any{"context_length": 4096, "eval_batch_size": 512, "parallel": 4, "flash_attention": true, "num_experts": 8, "offload_kv_cache_to_gpu": true},
		}},
		"max_context_length": 262144, "format": "gguf",
		"capabilities": map[string]any{
			"vision": true, "trained_for_tool_use": true,
			"reasoning": map[string]any{"allowed_options": []string{"off", "on"}, "default": "on"},
		},
	}
}

func lmstudioV1Embedding() map[string]any {
	return map[string]any{
		"type": "embedding", "publisher": "gaianet", "key": "text-embedding-nomic-embed-text-v1.5-embedding",
		"display_name": "Nomic Embed Text v1.5",
		"quantization": map[string]any{"name": "F16", "bits_per_weight": 16},
		"size_bytes":   274290560, "params_string": nil, "loaded_instances": []any{},
		"max_context_length": 2048, "format": "gguf",
	}
}

func probeLMStudioAt(t *testing.T, srvURL, modelName, apiKey string) (apispec.Evidence, []string) {
	t.Helper()
	ctx, tr := withTrace(context.Background())
	var ev apispec.Evidence
	require.True(t, newTestDetector().probeLMStudio(ctx, srvURL, modelName, apiKey, &ev))
	return ev, tr.lines
}

func TestProbeLMStudioV1Facts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{lmstudioV1Gemma(), lmstudioV1Embedding()}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, trace := probeLMStudioAt(t, srv.URL, "google/gemma-4-26b-a4b", "")
	require.Equal(t, "lmstudio", ev.Stack)
	require.Equal(t, "google/gemma-4-26b-a4b", ev.ServedModelID)
	require.Equal(t, "gemma4", ev.Architecture)

	c := ev.Caps
	require.Equal(t, int64(4096), *c.Context)
	require.Equal(t, int64(262144), *c.ContextTrained)
	require.Equal(t, "Q4_K_M", c.Quantization)
	require.Equal(t, []string{"image"}, c.Modalities.Input)
	require.True(t, *c.Tools)
	require.True(t, *c.Reasoning)
	require.Nil(t, c.Embeddings)
	require.Equal(t, map[string]string{
		apispec.CapFieldContext:         "lmstudio:/api/v1/models.loaded_instances[].config.context_length",
		apispec.CapFieldContextTrained:  "lmstudio:/api/v1/models.max_context_length",
		apispec.CapFieldQuantization:    "lmstudio:/api/v1/models.quantization.name",
		apispec.CapFieldModalitiesInput: "lmstudio:/api/v1/models.capabilities.vision",
		apispec.CapFieldTools:           "lmstudio:/api/v1/models.capabilities.trained_for_tool_use",
		apispec.CapFieldReasoning:       "lmstudio:/api/v1/models.capabilities.reasoning.allowed_options",
	}, c.Sources)

	require.Empty(t, ev.ReasoningEfforts, "off/on is a toggle, not a level list")
	require.Empty(t, ev.EffortsSource)
	require.Equal(t, "on", ev.ThinkingDefault)
	require.Equal(t, "lmstudio:/api/v1/models.capabilities.reasoning.default", ev.ThinkingDefaultSource)

	require.Contains(t, trace, "identified lmstudio by /api/v1/models")
	require.Contains(t, trace, "lmstudio /api/v1/models loaded_instances[].config.context_length 4096 -> context")
	require.Contains(t, trace, "lmstudio /api/v1/models max_context_length 262144 -> contextTrained")
	require.Contains(t, trace, "lmstudio /api/v1/models capabilities.trained_for_tool_use true -> tools")
}

func TestProbeLMStudioV1Embedding(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{lmstudioV1Gemma(), lmstudioV1Embedding()}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeLMStudioAt(t, srv.URL, "text-embedding-nomic-embed-text-v1.5-embedding", "")
	c := ev.Caps
	require.True(t, *c.Embeddings)
	require.Equal(t, "lmstudio:/api/v1/models.type", c.Sources[apispec.CapFieldEmbeddings])
	require.Equal(t, int64(2048), *c.ContextTrained)
	require.Nil(t, c.Context, "no loaded instance: no served window")
	require.Equal(t, "F16", c.Quantization)
	require.Nil(t, c.Tools)
	require.Nil(t, c.Reasoning)
	require.Nil(t, c.Modalities)
	require.Empty(t, ev.Architecture)
	require.Empty(t, ev.ThinkingDefault)
	require.Empty(t, ev.ThinkingDefaultSource)
	require.Empty(t, ev.ReasoningEfforts)
	require.Empty(t, ev.EffortsSource)
}

func TestProbeLMStudioV1ReasoningLevels(t *testing.T) {
	entry := lmstudioV1Gemma()
	entry["capabilities"] = map[string]any{"reasoning": map[string]any{"allowed_options": []string{"off", "low", "medium", "high"}, "default": "medium"}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{entry}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeLMStudioAt(t, srv.URL, "google/gemma-4-26b-a4b", "")
	require.Equal(t, []string{"none", "low", "medium", "high"}, ev.ReasoningEfforts)
	require.Equal(t, "lmstudio:/api/v1/models.capabilities.reasoning.allowed_options", ev.EffortsSource)
	require.True(t, *ev.Caps.Reasoning)
	require.Equal(t, "medium", ev.ThinkingDefault)
	require.Nil(t, ev.Caps.Tools, "trained_for_tool_use absent: no tools fact")
	require.Nil(t, ev.Caps.Modalities, "vision absent: no modality fact")
}

func TestProbeLMStudioV1ReasoningOnlyOff(t *testing.T) {
	entry := lmstudioV1Gemma()
	entry["capabilities"] = map[string]any{"trained_for_tool_use": false, "reasoning": map[string]any{"allowed_options": []string{"off"}, "default": "off"}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{entry}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeLMStudioAt(t, srv.URL, "google/gemma-4-26b-a4b", "")
	require.Nil(t, ev.Caps.Reasoning, "only off is not a documented shape: no reasoning fact")
	require.Nil(t, ev.Caps.Tools, "trained_for_tool_use false is not a tools false report")
	require.NotContains(t, ev.Caps.Sources, apispec.CapFieldReasoning)
	require.NotContains(t, ev.Caps.Sources, apispec.CapFieldTools)
	require.Empty(t, ev.ReasoningEfforts)
	require.Equal(t, "off", ev.ThinkingDefault)
}

func TestProbeLMStudioV1SmallestLoadedInstanceWins(t *testing.T) {
	entry := lmstudioV1Gemma()
	entry["loaded_instances"] = []map[string]any{
		{"id": "a", "config": map[string]any{"context_length": 32768}},
		{"id": "b", "config": map[string]any{"context_length": 8192}},
		{"id": "c", "config": map[string]any{}},
		{"id": "d", "config": map[string]any{"context_length": -1}},
		{"id": "e", "config": map[string]any{"context_length": 0}},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{entry}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeLMStudioAt(t, srv.URL, "google/gemma-4-26b-a4b", "")
	require.Equal(t, int64(8192), *ev.Caps.Context)
}

func TestProbeLMStudioV1RejectsOutOfRangeAndFractionalCounts(t *testing.T) {
	entry := lmstudioV1Gemma()
	entry["max_context_length"] = 1 << 31
	entry["loaded_instances"] = []map[string]any{{"id": "a", "config": map[string]any{"context_length": 4096.5}}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{entry}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeLMStudioAt(t, srv.URL, "google/gemma-4-26b-a4b", "")
	require.Nil(t, ev.Caps.Context)
	require.Nil(t, ev.Caps.ContextTrained)
}

// Each documented field with a wrong type or an out-of-range value leaves its
// capability unset and its evidence key absent.
func TestProbeLMStudioV1RejectsWrongTypes(t *testing.T) {
	cases := map[string]struct {
		mutate func(entry map[string]any)
		check  func(t *testing.T, ev apispec.Evidence)
	}{
		"quantization.name non-string": {
			func(e map[string]any) { e["quantization"] = map[string]any{"name": 4} },
			func(t *testing.T, ev apispec.Evidence) {
				require.Empty(t, ev.Caps.Quantization)
				require.NotContains(t, ev.Caps.Sources, apispec.CapFieldQuantization)
			},
		},
		"quantization absent": {
			func(e map[string]any) { delete(e, "quantization") },
			func(t *testing.T, ev apispec.Evidence) { require.Empty(t, ev.Caps.Quantization) },
		},
		"vision non-bool": {
			func(e map[string]any) { e["capabilities"] = map[string]any{"vision": "true"} },
			func(t *testing.T, ev apispec.Evidence) {
				require.Nil(t, ev.Caps.Modalities)
				require.NotContains(t, ev.Caps.Sources, apispec.CapFieldModalitiesInput)
			},
		},
		"trained_for_tool_use non-bool": {
			func(e map[string]any) { e["capabilities"] = map[string]any{"trained_for_tool_use": 1} },
			func(t *testing.T, ev apispec.Evidence) {
				require.Nil(t, ev.Caps.Tools)
				require.NotContains(t, ev.Caps.Sources, apispec.CapFieldTools)
			},
		},
		"type non-string": {
			func(e map[string]any) { e["type"] = 7 },
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Embeddings) },
		},
		"architecture non-string": {
			func(e map[string]any) { e["architecture"] = []string{"gemma4"} },
			func(t *testing.T, ev apispec.Evidence) { require.Empty(t, ev.Architecture) },
		},
		"allowed_options non-string entries": {
			func(e map[string]any) {
				e["capabilities"] = map[string]any{"reasoning": map[string]any{"allowed_options": []any{1, true, nil}}}
			},
			func(t *testing.T, ev apispec.Evidence) {
				require.Nil(t, ev.Caps.Reasoning)
				require.Empty(t, ev.ReasoningEfforts)
				require.Empty(t, ev.EffortsSource)
			},
		},
		"allowed_options empty": {
			func(e map[string]any) {
				e["capabilities"] = map[string]any{"reasoning": map[string]any{"allowed_options": []string{}}}
			},
			func(t *testing.T, ev apispec.Evidence) {
				require.Nil(t, ev.Caps.Reasoning)
				require.Empty(t, ev.ReasoningEfforts)
			},
		},
		"allowed_options non-list": {
			func(e map[string]any) {
				e["capabilities"] = map[string]any{"reasoning": map[string]any{"allowed_options": "on"}}
			},
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Reasoning) },
		},
		"reasoning.default non-string": {
			func(e map[string]any) {
				e["capabilities"] = map[string]any{"reasoning": map[string]any{"allowed_options": []string{"off", "on"}, "default": true}}
			},
			func(t *testing.T, ev apispec.Evidence) {
				require.True(t, *ev.Caps.Reasoning)
				require.Empty(t, ev.ThinkingDefault)
				require.Empty(t, ev.ThinkingDefaultSource)
			},
		},
		"reasoning.default absent": {
			func(e map[string]any) {
				e["capabilities"] = map[string]any{"reasoning": map[string]any{"allowed_options": []string{"off", "on"}}}
			},
			func(t *testing.T, ev apispec.Evidence) { require.Empty(t, ev.ThinkingDefault) },
		},
		"reasoning non-object": {
			func(e map[string]any) { e["capabilities"] = map[string]any{"reasoning": "on"} },
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Reasoning) },
		},
		"max_context_length string": {
			func(e map[string]any) { e["max_context_length"] = "262144" },
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.ContextTrained) },
		},
		"max_context_length non-positive": {
			func(e map[string]any) { e["max_context_length"] = -262144 },
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.ContextTrained) },
		},
		"loaded_instances non-list": {
			func(e map[string]any) {
				e["loaded_instances"] = map[string]any{"config": map[string]any{"context_length": 4096}}
			},
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Context) },
		},
		"loaded_instances context_length string": {
			func(e map[string]any) {
				e["loaded_instances"] = []map[string]any{{"id": "a", "config": map[string]any{"context_length": "4096"}}}
			},
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Context) },
		},
		"loaded_instances only non-positive": {
			func(e map[string]any) {
				e["loaded_instances"] = []map[string]any{{"id": "a", "config": map[string]any{"context_length": -1}}}
			},
			func(t *testing.T, ev apispec.Evidence) {
				require.Nil(t, ev.Caps.Context)
				require.NotContains(t, ev.Caps.Sources, apispec.CapFieldContext)
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			entry := lmstudioV1Gemma()
			tc.mutate(entry)
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{entry}}))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			ev, _ := probeLMStudioAt(t, srv.URL, "google/gemma-4-26b-a4b", "")
			require.Equal(t, "lmstudio", ev.Stack)
			tc.check(t, ev)
		})
	}
}

func TestProbeLMStudioV0RejectsWrongTypes(t *testing.T) {
	base := func() map[string]any {
		return map[string]any{
			"id": "qwen2-vl-7b-instruct", "object": "model", "type": "vlm", "publisher": "mlx-community", "arch": "qwen2_vl",
			"compatibility_type": "mlx", "quantization": "4bit", "state": "not-loaded", "max_context_length": 32768, "capabilities": []string{"tool_use"},
		}
	}
	cases := map[string]struct {
		mutate func(entry map[string]any)
		check  func(t *testing.T, ev apispec.Evidence)
	}{
		"quantization non-string": {
			func(e map[string]any) { e["quantization"] = map[string]any{"name": "4bit"} },
			func(t *testing.T, ev apispec.Evidence) {
				require.Empty(t, ev.Caps.Quantization)
				require.NotContains(t, ev.Caps.Sources, apispec.CapFieldQuantization)
			},
		},
		"quantization absent": {
			func(e map[string]any) { delete(e, "quantization") },
			func(t *testing.T, ev apispec.Evidence) { require.Empty(t, ev.Caps.Quantization) },
		},
		"max_context_length out of range": {
			func(e map[string]any) { e["max_context_length"] = 1 << 31 },
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Context) },
		},
		"max_context_length fractional": {
			func(e map[string]any) { e["max_context_length"] = 32768.5 },
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Context) },
		},
		"max_context_length string": {
			func(e map[string]any) { e["max_context_length"] = "32768" },
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Context) },
		},
		"capabilities non-strings": {
			func(e map[string]any) { e["capabilities"] = []any{1, true, map[string]any{"tool_use": true}} },
			func(t *testing.T, ev apispec.Evidence) {
				require.Nil(t, ev.Caps.Tools)
				require.NotContains(t, ev.Caps.Sources, apispec.CapFieldTools)
			},
		},
		"capabilities non-list": {
			func(e map[string]any) { e["capabilities"] = "tool_use" },
			func(t *testing.T, ev apispec.Evidence) { require.Nil(t, ev.Caps.Tools) },
		},
		"type non-string": {
			func(e map[string]any) { e["type"] = 1 },
			func(t *testing.T, ev apispec.Evidence) {
				require.Nil(t, ev.Caps.Modalities)
				require.Nil(t, ev.Caps.Embeddings)
			},
		},
		"arch non-string": {
			func(e map[string]any) { e["arch"] = 1 },
			func(t *testing.T, ev apispec.Evidence) { require.Empty(t, ev.Architecture) },
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			entry := base()
			tc.mutate(entry)
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v0/models", jsonHandler(map[string]any{"object": "list", "data": []map[string]any{entry}}))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			ev, _ := probeLMStudioAt(t, srv.URL, "qwen2-vl-7b-instruct", "")
			require.Equal(t, "lmstudio", ev.Stack)
			tc.check(t, ev)
		})
	}
}

func TestProbeLMStudioV1MatchesBareIDAndSingleEntry(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{lmstudioV1Gemma()}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeLMStudioAt(t, srv.URL, "openai/google/gemma-4-26b-a4b", "")
	require.Equal(t, "google/gemma-4-26b-a4b", ev.ServedModelID)

	ev, _ = probeLMStudioAt(t, srv.URL, "unrelated-alias", "")
	require.Equal(t, "google/gemma-4-26b-a4b", ev.ServedModelID, "a single-entry listing is the model")
}

func TestProbeLMStudioV1NoMatchLeavesNoFacts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{lmstudioV1Gemma(), lmstudioV1Embedding()}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, trace := probeLMStudioAt(t, srv.URL, "unrelated-alias", "")
	require.Equal(t, "lmstudio", ev.Stack)
	require.True(t, ev.Caps.Empty())
	require.Empty(t, ev.ServedModelID)
	require.Contains(t, trace, "lmstudio /api/v1/models lists no entry for the configured model; no capability facts")
}

func TestProbeLMStudioFallsBackToV0(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/models", jsonHandler(map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "qwen2-vl-7b-instruct", "object": "model", "type": "vlm", "publisher": "mlx-community", "arch": "qwen2_vl",
				"compatibility_type": "mlx", "quantization": "4bit", "state": "not-loaded", "max_context_length": 32768, "capabilities": []string{"tool_use"}},
			{"id": "text-embedding-nomic-embed-text-v1.5", "object": "model", "type": "embeddings", "publisher": "nomic-ai", "arch": "nomic-bert",
				"compatibility_type": "gguf", "quantization": "Q4_0", "state": "not-loaded", "max_context_length": 2048},
		},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, trace := probeLMStudioAt(t, srv.URL, "qwen2-vl-7b-instruct", "")
	require.Equal(t, "lmstudio", ev.Stack)
	require.Equal(t, "qwen2_vl", ev.Architecture)
	require.Equal(t, "qwen2-vl-7b-instruct", ev.ServedModelID)
	c := ev.Caps
	require.Equal(t, int64(32768), *c.Context)
	require.Nil(t, c.ContextTrained)
	require.Equal(t, []string{"image"}, c.Modalities.Input)
	require.Equal(t, "4bit", c.Quantization)
	require.True(t, *c.Tools)
	require.Nil(t, c.Embeddings)
	require.Nil(t, c.Reasoning)
	require.Equal(t, map[string]string{
		apispec.CapFieldContext:         "lmstudio:/api/v0/models.max_context_length",
		apispec.CapFieldModalitiesInput: "lmstudio:/api/v0/models.type",
		apispec.CapFieldQuantization:    "lmstudio:/api/v0/models.quantization",
		apispec.CapFieldTools:           "lmstudio:/api/v0/models.capabilities",
	}, c.Sources)
	require.Empty(t, ev.ThinkingDefault)
	require.Contains(t, trace, "identified lmstudio by /api/v0/models")
	require.Contains(t, trace, "lmstudio /api/v0/models max_context_length 32768 -> context")

	ev, _ = probeLMStudioAt(t, srv.URL, "text-embedding-nomic-embed-text-v1.5", "")
	require.True(t, *ev.Caps.Embeddings)
	require.Equal(t, "lmstudio:/api/v0/models.type", ev.Caps.Sources[apispec.CapFieldEmbeddings])
	require.Nil(t, ev.Caps.Modalities)
	require.Nil(t, ev.Caps.Tools)
}

func TestProbeLMStudioKeyFreeFirst(t *testing.T) {
	t.Run("no key sent when auth is off", func(t *testing.T) {
		var auth []string
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
			auth = append(auth, r.Header.Get("Authorization"))
			jsonHandler(map[string]any{"models": []map[string]any{lmstudioV1Gemma()}})(w, r)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		probeLMStudioAt(t, srv.URL, "google/gemma-4-26b-a4b", "secret")
		require.Equal(t, []string{""}, auth)
	})
	t.Run("key sent only after 401", func(t *testing.T) {
		var auth []string
		guarded := func(body any) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				auth = append(auth, r.Header.Get("Authorization"))
				if r.Header.Get("Authorization") != "Bearer secret" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				jsonHandler(body)(w, r)
			}
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/models", guarded(map[string]any{"models": []map[string]any{lmstudioV1Gemma()}}))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		ev, _ := probeLMStudioAt(t, srv.URL, "google/gemma-4-26b-a4b", "secret")
		require.Equal(t, []string{"", "Bearer secret"}, auth)
		require.Equal(t, int64(4096), *ev.Caps.Context)
	})
}

func TestDetectLMStudioV1PublishesCapabilities(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"models": []map[string]any{lmstudioV1Gemma(), lmstudioV1Embedding()}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "google/gemma-4-26b-a4b")
	require.NotNil(t, api)
	require.Equal(t, "lmstudio", api.Stack)
	require.NotNil(t, api.Capabilities)
	require.Equal(t, int64(4096), *api.Capabilities.Context)
	require.Equal(t, int64(262144), *api.Capabilities.ContextTrained)
	require.True(t, *api.Capabilities.Tools)
	require.True(t, *api.Capabilities.Reasoning)
	require.Equal(t, "Q4_K_M", api.Capabilities.Quantization)
	require.Contains(t, api.Capabilities.Modalities.Input, "image")
	require.Equal(t, "lmstudio:/api/v1/models.loaded_instances[].config.context_length", api.Evidence["capabilities.context"])
	require.Equal(t, "lmstudio:/api/v1/models.max_context_length", api.Evidence["capabilities.contextTrained"])
	require.Equal(t, "lmstudio:/api/v1/models.capabilities.trained_for_tool_use", api.Evidence["capabilities.tools"])
}
