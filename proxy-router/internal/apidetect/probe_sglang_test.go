package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

// Field set of GET /model_info in SGLang v0.5.20 http_server.py.
func sglangModelInfo(overrides map[string]any) map[string]any {
	info := map[string]any{
		"model_path":                "Qwen/Qwen3-VL-8B-Instruct",
		"tokenizer_path":            "Qwen/Qwen3-VL-8B-Instruct",
		"is_generation":             true,
		"preferred_sampling_params": nil,
		"weight_version":            "default",
		"served_model_name":         "qwen3-vl",
		"reasoning_parser":          "qwen3",
		"tool_call_parser":          "qwen25",
		"has_image_understanding":   true,
		"has_audio_understanding":   false,
		"model_type":                "qwen3_vl",
		"architectures":             []string{"Qwen3VLForConditionalGeneration"},
	}
	for k, v := range overrides {
		info[k] = v
	}
	return info
}

func sglangModelsList(id string, maxModelLen any) map[string]any {
	return map[string]any{
		"object": "list",
		"data": []map[string]any{{
			"id":            id,
			"object":        "model",
			"created":       1700000000,
			"owned_by":      "sglang",
			"root":          id,
			"max_model_len": maxModelLen,
		}},
	}
}

func detectSGLang(t *testing.T, srvURL, modelName string) (*system.ModelApiSpec, []string) {
	t.Helper()
	api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{
		ModelName: modelName, ApiType: "openai", ApiURL: srvURL + "/v1/chat/completions",
	})
	require.NotNil(t, api)
	require.Equal(t, "sglang", api.Stack)
	return api, trace
}

func TestSGLangModelInfoFacts(t *testing.T) {
	var aliasHits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(nil)))
	mux.HandleFunc("/get_model_info", func(w http.ResponseWriter, r *http.Request) {
		aliasHits.Add(1)
		jsonHandler(sglangModelInfo(nil))(w, r)
	})
	mux.HandleFunc("/v1/models", jsonHandler(sglangModelsList("Qwen/Qwen3-VL-8B-Instruct", 40960)))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api, trace := detectSGLang(t, srv.URL, "qwen3-vl-8b-instruct")
	require.Zero(t, aliasHits.Load(), "/model_info answered; the deprecated alias is not asked")
	require.Contains(t, trace, `identified sglang by /model_info; served model path "Qwen/Qwen3-VL-8B-Instruct"`)

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.NotNil(t, caps.Reasoning)
	require.True(t, *caps.Reasoning)
	require.Equal(t, "sglang:/model_info.reasoning_parser", api.Evidence["capabilities.reasoning"])
	require.NotNil(t, caps.Tools)
	require.True(t, *caps.Tools)
	require.Equal(t, "sglang:/model_info.tool_call_parser", api.Evidence["capabilities.tools"])
	require.NotNil(t, caps.Modalities)
	require.Equal(t, []string{"image"}, caps.Modalities.Input)
	require.Equal(t, "sglang:/model_info.has_image_understanding", api.Evidence["capabilities.modalities.input"])
	require.Empty(t, caps.Modalities.Output)
	require.Nil(t, caps.Embeddings, "a generation model reports no embeddings fact")
	require.Equal(t, []string{"json_object", "json_schema", "grammar", "regex"}, caps.Structured)
	require.Equal(t, "sglang:/model_info", api.Evidence["capabilities.structured"])
	require.NotNil(t, caps.Context)
	require.Equal(t, int64(40960), *caps.Context)
	require.Equal(t, "sglang:/v1/models.max_model_len", api.Evidence["capabilities.context"])
	require.Equal(t, "qwen3", api.ModelFamily)
	require.Contains(t, trace, `family "qwen3": from backend-reported architecture "Qwen3VLForConditionalGeneration"`)
	require.Contains(t, trace, `/model_info.architectures[0] "Qwen3VLForConditionalGeneration" -> architecture`)
	require.Contains(t, trace, `/model_info.reasoning_parser "qwen3" -> reasoning`)
	require.Contains(t, trace, `/model_info.tool_call_parser "qwen25" -> tools`)
	require.Contains(t, trace, "/model_info.has_image_understanding -> modalities.input image")
	require.Contains(t, trace, "/v1/models.max_model_len 40960 -> context")
}

func TestSGLangDeprecatedAliasFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/get_model_info", jsonHandler(sglangModelInfo(map[string]any{
		"has_audio_understanding": true,
	})))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api, trace := detectSGLang(t, srv.URL, "alias")
	require.Contains(t, trace, `identified sglang by /get_model_info; served model path "Qwen/Qwen3-VL-8B-Instruct"`)
	require.Equal(t, []string{"image", "audio"}, api.Capabilities.Modalities.Input)
	require.Equal(t, "sglang:/get_model_info.has_audio_understanding", api.Evidence["capabilities.modalities.input"],
		"evidence names the endpoint that answered")
	require.Nil(t, api.Capabilities.Context, "no /v1/models answer leaves context unset")
	require.Equal(t, "sglang:/get_model_info", api.Evidence["capabilities.structured"])
}

func TestSGLangEmbeddingModel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(map[string]any{
		"model_path":              "BAAI/bge-m3",
		"served_model_name":       "bge-m3",
		"is_generation":           false,
		"reasoning_parser":        nil,
		"tool_call_parser":        nil,
		"has_image_understanding": false,
		"model_type":              "xlm-roberta",
		"architectures":           []string{"XLMRobertaModel"},
	})))
	mux.HandleFunc("/v1/models", jsonHandler(sglangModelsList("BAAI/bge-m3", 8192)))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api, trace := detectSGLang(t, srv.URL, "bge-m3")
	caps := api.Capabilities
	require.NotNil(t, caps.Embeddings)
	require.True(t, *caps.Embeddings)
	require.Equal(t, "sglang:/model_info.is_generation", api.Evidence["capabilities.embeddings"])
	require.Equal(t, []string{"embeddings"}, caps.Modalities.Output)
	require.Equal(t, "sglang:/model_info.is_generation", api.Evidence["capabilities.modalities.output"])
	require.Empty(t, caps.Modalities.Input)
	require.Nil(t, caps.Tools, "a null parser is no tools fact")
	require.Nil(t, caps.Reasoning, "a null parser is no reasoning fact")
	require.Equal(t, int64(8192), *caps.Context)
	require.Contains(t, trace, "/model_info.is_generation false -> embeddings, modalities.output embeddings")
}

func TestSGLangServedModelNameOnlyWhenPathEmpty(t *testing.T) {
	t.Run("model_path wins", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(nil)))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		_, trace := detectSGLang(t, srv.URL, "alias")
		require.Contains(t, trace, `identified sglang by /model_info; served model path "Qwen/Qwen3-VL-8B-Instruct"`)
		require.NotContains(t, trace, "served_model_name stands in")
	})

	t.Run("served_model_name stands in for an empty model_path", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(map[string]any{"model_path": ""})))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		_, trace := detectSGLang(t, srv.URL, "alias")
		require.Contains(t, trace, `identified sglang by /model_info; served model path ""`)
		require.Contains(t, trace, "/model_info.served_model_name stands in for the empty model_path")
	})
}

func TestSGLangContextRejectsOutOfRangeAndUnmatched(t *testing.T) {
	t.Run("out of range", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(nil)))
		mux.HandleFunc("/v1/models", jsonHandler(sglangModelsList("Qwen/Qwen3-VL-8B-Instruct", 0)))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		api, trace := detectSGLang(t, srv.URL, "alias")
		require.Nil(t, api.Capabilities.Context)
		require.NotContains(t, api.Evidence, "capabilities.context")
		require.Contains(t, trace, "/v1/models.max_model_len 0 is out of range; ignored")
	})

	t.Run("no matching entry", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(nil)))
		mux.HandleFunc("/v1/models", jsonHandler(map[string]any{
			"object": "list",
			"data": []map[string]any{
				{"id": "other/model-a", "max_model_len": 4096},
				{"id": "other/model-b", "max_model_len": 2048},
			},
		}))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		api, _ := detectSGLang(t, srv.URL, "alias")
		require.Nil(t, api.Capabilities.Context)
	})
}

func TestSGLangArchitectureDoesNotOverrideEarlierEvidence(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(nil)))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, tr := withTrace(context.Background())
	ev := &apispec.Evidence{Architecture: "LlamaForCausalLM"}
	require.True(t, newTestDetector().probeSGLang(ctx, srv.URL, "", ev))
	require.Equal(t, "LlamaForCausalLM", ev.Architecture)
	for _, line := range tr.lines {
		require.NotContains(t, line, "architectures[0]")
	}
}

func TestSGLangClassifierIsNotEmbeddings(t *testing.T) {
	for _, task := range []string{"classify", "none"} {
		t.Run(task, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(map[string]any{
				"model_path":       "Qwen/Qwen3-Reranker-0.6B",
				"is_generation":    false,
				"reasoning_parser": nil,
				"tool_call_parser": nil,
				"architectures":    []string{"Qwen3ForSequenceClassification"},
				"embedding":        map[string]any{"family": "qwen3", "task": task},
			})))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			api, trace := detectSGLang(t, srv.URL, "alias")
			require.Nil(t, api.Capabilities.Embeddings)
			require.NotContains(t, api.Evidence, "capabilities.embeddings")
			require.NotContains(t, api.Evidence, "capabilities.modalities.output")
			require.Contains(t, trace, `/model_info.is_generation false but embedding.task "`+task+`" -> no embeddings fact`)
		})
	}

	t.Run("embed task keeps the fact", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(map[string]any{
			"is_generation": false,
			"embedding":     map[string]any{"task": "embed"},
		})))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		api, _ := detectSGLang(t, srv.URL, "alias")
		require.NotNil(t, api.Capabilities.Embeddings)
		require.True(t, *api.Capabilities.Embeddings)
	})

	t.Run("malformed embedding block does not suppress", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(map[string]any{
			"is_generation": false,
			"embedding":     map[string]any{"task": 3},
		})))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		api, _ := detectSGLang(t, srv.URL, "alias")
		require.NotNil(t, api.Capabilities.Embeddings)
	})
}

func TestSGLangContextFromServedModelNameWithLoRA(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(nil)))
	mux.HandleFunc("/v1/models", jsonHandler(map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "qwen3-vl", "object": "model", "root": "qwen3-vl", "max_model_len": 40960},
			{"id": "lora-a", "object": "model", "root": "/lora/a", "parent": "qwen3-vl", "max_model_len": nil},
		},
	}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api, trace := detectSGLang(t, srv.URL, "alias")
	require.NotNil(t, api.Capabilities.Context)
	require.Equal(t, int64(40960), *api.Capabilities.Context, "served_model_name selects the base card")
	require.Contains(t, trace, "/v1/models.max_model_len 40960 -> context")
}

func TestSGLangIgnoresWrongTypedFacts(t *testing.T) {
	t.Run("wrong JSON types", func(t *testing.T) {
		mux := http.NewServeMux()
		// A model path with no family keeps the family-default bindings (and the
		// reasoning fact derived from them) out of the picture.
		mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(map[string]any{
			"model_path":              "org/plain-model",
			"served_model_name":       "plain",
			"reasoning_parser":        7,
			"tool_call_parser":        "",
			"has_image_understanding": "yes",
			"has_audio_understanding": 1,
			"is_generation":           "false",
			"architectures":           []string{},
		})))
		mux.HandleFunc("/v1/models", jsonHandler(sglangModelsList("plain", "40960")))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		api, trace := detectSGLang(t, srv.URL, "alias")
		caps := api.Capabilities
		require.NotNil(t, caps)
		require.Nil(t, caps.Reasoning)
		require.Nil(t, caps.Tools)
		require.Nil(t, caps.Embeddings)
		require.Nil(t, caps.Modalities)
		require.Nil(t, caps.Context)
		require.Empty(t, api.ModelFamily)
		for _, line := range trace {
			require.NotContains(t, line, "architectures[0]")
			require.NotContains(t, line, "max_model_len")
		}
		require.Equal(t, []string{"json_object", "json_schema", "grammar", "regex"}, caps.Structured)
	})

	t.Run("max_model_len at the upper bound", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/model_info", jsonHandler(sglangModelInfo(nil)))
		mux.HandleFunc("/v1/models", jsonHandler(sglangModelsList("Qwen/Qwen3-VL-8B-Instruct", int64(1)<<31)))
		srv := httptest.NewServer(mux)
		defer srv.Close()

		api, trace := detectSGLang(t, srv.URL, "alias")
		require.Nil(t, api.Capabilities.Context)
		require.Contains(t, trace, "/v1/models.max_model_len 2.147483648e+09 is out of range; ignored")
	})
}
