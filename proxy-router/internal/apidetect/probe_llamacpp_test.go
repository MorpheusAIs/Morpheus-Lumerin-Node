package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

// Shapes follow tools/server/README.md (the /props and /v1/models samples)
// and server-context.cpp get_res_props / get_res_models on master.
func llamaCppProps(caps map[string]any, modalities map[string]any) map[string]any {
	props := map[string]any{
		"default_generation_settings": map[string]any{
			"id":     0,
			"n_ctx":  8192,
			"params": map[string]any{"n_predict": -1, "grammar": ""},
		},
		"total_slots":   1,
		"model_path":    "../models/Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
		"chat_template": qwen3Template,
		"build_info":    "b11177-abcdef0",
	}
	if caps != nil {
		props["chat_template_caps"] = caps
	}
	if modalities != nil {
		props["modalities"] = modalities
	}
	return props
}

func llamaCppFullCaps(toolCalls, parallel, effort bool) map[string]any {
	return map[string]any{
		"supports_string_content":      true,
		"supports_typed_content":       false,
		"supports_tools":               toolCalls,
		"supports_tool_calls":          toolCalls,
		"supports_parallel_tool_calls": parallel,
		"supports_system_role":         true,
		"supports_preserve_reasoning":  false,
		"supports_reasoning_effort":    effort,
		"supports_object_arguments":    true,
	}
}

// The README sample plus the n_ctx / ftype keys server-context.cpp adds.
func llamaCppModels(meta any, capabilities []string) map[string]any {
	return map[string]any{
		"object": "list",
		"data": []map[string]any{{
			"id":       "../models/Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
			"object":   "model",
			"created":  1735142223,
			"owned_by": "llamacpp",
			"meta":     meta,
		}},
		"models": []map[string]any{{
			"name":         "../models/Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
			"model":        "../models/Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
			"type":         "model",
			"capabilities": capabilities,
			"details":      map[string]any{"format": "gguf"},
		}},
	}
}

func llamaCppMeta(nCtx, nCtxTrain any, ftype any) map[string]any {
	meta := map[string]any{
		"vocab_type": 2,
		"n_vocab":    128256,
		"n_embd":     4096,
		"n_params":   8030261312,
		"size":       4912898304,
	}
	if nCtx != nil {
		meta["n_ctx"] = nCtx
	}
	if nCtxTrain != nil {
		meta["n_ctx_train"] = nCtxTrain
	}
	if ftype != nil {
		meta["ftype"] = ftype
	}
	return meta
}

func detectLlamaCpp(t *testing.T, props, models map[string]any) (*system.ModelApiSpec, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/props", jsonHandler(props))
	if models != nil {
		mux.HandleFunc("/v1/models", jsonHandler(models))
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{
		ModelName: "qwen3-32b", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions", ApiKey: "secret-key",
	})
	require.NotNil(t, api)
	require.Equal(t, "llamacpp", api.Stack)
	return api, strings.Join(trace, "\n")
}

func TestProbeLlamaCppPropsFacts(t *testing.T) {
	props := llamaCppProps(llamaCppFullCaps(true, true, true), map[string]any{"vision": true, "video": false, "audio": true})
	models := llamaCppModels(llamaCppMeta(8192, 131072, "Q4_K - Medium"), []string{"completion", "multimodal"})
	api, trace := detectLlamaCpp(t, props, models)

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.NotNil(t, caps.Context)
	require.Equal(t, int64(8192), *caps.Context)
	require.Equal(t, "llamacpp:/props.default_generation_settings.n_ctx", api.Evidence["capabilities.context"])

	require.NotNil(t, caps.ContextTrained)
	require.Equal(t, int64(131072), *caps.ContextTrained)
	require.Equal(t, "llamacpp:/v1/models.meta.n_ctx_train", api.Evidence["capabilities.contextTrained"])

	require.NotNil(t, caps.Modalities)
	require.Equal(t, []string{"image", "audio"}, caps.Modalities.Input, "video false is not listed; multimodal adds nothing when /props gave modalities")
	require.Equal(t, "llamacpp:/props.modalities.audio", api.Evidence["capabilities.modalities.input"])

	require.NotNil(t, caps.Tools)
	require.True(t, *caps.Tools)
	require.Equal(t, "llamacpp:/props.chat_template_caps.supports_tool_calls", api.Evidence["capabilities.tools"])
	require.NotNil(t, caps.ToolsParallel)
	require.True(t, *caps.ToolsParallel)
	require.Equal(t, "llamacpp:/props.chat_template_caps.supports_parallel_tool_calls", api.Evidence["capabilities.toolsParallel"])

	require.Equal(t, []string{"json_object", "json_schema", "grammar"}, caps.Structured)
	require.Equal(t, "llamacpp:/props", api.Evidence["capabilities.structured"])

	require.Equal(t, "Q4_K_M", caps.Quantization)
	require.Equal(t, "llamacpp:/v1/models.meta.ftype", api.Evidence["capabilities.quantization"])

	require.Contains(t, trace, "default_generation_settings.n_ctx = 8192")
	require.Contains(t, trace, "modalities.vision = true")
	require.Contains(t, trace, "modalities.audio = true")
	require.NotContains(t, trace, "modalities.video")
	require.Contains(t, trace, "supports_tool_calls = true")
	require.Contains(t, trace, "supports_parallel_tool_calls = true")
	require.Contains(t, trace, "supports_reasoning_effort = true")
	require.Contains(t, trace, "meta.n_ctx_train = 131072")
	require.Contains(t, trace, "quantization Q4_K_M")
	require.NotContains(t, trace, "secret-key")
}

func TestProbeLlamaCppTemplateWithoutToolCalls(t *testing.T) {
	props := llamaCppProps(llamaCppFullCaps(false, false, false), map[string]any{"vision": false, "video": false, "audio": false})
	api, trace := detectLlamaCpp(t, props, nil)

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.NotNil(t, caps.Tools)
	require.False(t, *caps.Tools, "the engine computed that the template cannot render tool calls")
	require.NotNil(t, caps.ToolsParallel)
	require.False(t, *caps.ToolsParallel)
	require.Nil(t, caps.Modalities, "all-false modalities list nothing")
	require.Nil(t, caps.ContextTrained, "no /v1/models")
	require.NotContains(t, trace, "supports_reasoning_effort")
	require.Equal(t, []string{"json_object", "json_schema", "grammar"}, caps.Structured, "grammar sampling does not depend on the template")
}

func TestProbeLlamaCppEmptyCapsAndNoModalities(t *testing.T) {
	// The README /props sample: chat_template_caps is {} and modalities absent.
	props := llamaCppProps(map[string]any{}, nil)
	models := llamaCppModels(llamaCppMeta(nil, 131072, nil), []string{"completion", "multimodal"})
	api, trace := detectLlamaCpp(t, props, models)

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.Nil(t, caps.Tools)
	require.Nil(t, caps.ToolsParallel)
	require.NotNil(t, caps.Modalities)
	require.Equal(t, []string{"image"}, caps.Modalities.Input, "multimodal fills in only when /props listed no modalities")
	require.Equal(t, "llamacpp:/v1/models.models.capabilities", api.Evidence["capabilities.modalities.input"])
	require.Contains(t, trace, "multimodal")
	require.Equal(t, "", caps.Quantization)
}

func TestProbeLlamaCppModelsNCtxFallsBackWhenPropsHasNone(t *testing.T) {
	props := llamaCppProps(nil, nil)
	delete(props, "default_generation_settings")
	models := llamaCppModels(llamaCppMeta(4096, 32768, 15), []string{"completion"})
	api, trace := detectLlamaCpp(t, props, models)

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.NotNil(t, caps.Context)
	require.Equal(t, int64(4096), *caps.Context)
	require.Equal(t, "llamacpp:/v1/models.meta.n_ctx", api.Evidence["capabilities.context"])
	require.NotNil(t, caps.ContextTrained)
	require.Equal(t, int64(32768), *caps.ContextTrained)
	require.Equal(t, "Q4_K_M", caps.Quantization, "integer ftype 15 is LLAMA_FTYPE_MOSTLY_Q4_K_M")
	require.Nil(t, caps.Modalities, "completion alone says nothing about images")
	require.Contains(t, trace, "meta.n_ctx = 4096: served context (/props had none)")
}

func TestProbeLlamaCppPropsNCtxBeatsModelsNCtx(t *testing.T) {
	props := llamaCppProps(nil, nil)
	models := llamaCppModels(llamaCppMeta(2048, 131072, nil), []string{"completion"})
	api, trace := detectLlamaCpp(t, props, models)

	require.Equal(t, int64(8192), *api.Capabilities.Context)
	require.Equal(t, "llamacpp:/props.default_generation_settings.n_ctx", api.Evidence["capabilities.context"])
	require.Contains(t, trace, "meta.n_ctx = 2048: /props already gave the served context")
}

func TestProbeLlamaCppMetaNullWhileLoading(t *testing.T) {
	props := llamaCppProps(nil, nil)
	models := llamaCppModels(nil, []string{"completion"})
	api, trace := detectLlamaCpp(t, props, models)

	require.Nil(t, api.Capabilities.ContextTrained)
	require.Equal(t, "", api.Capabilities.Quantization)
	require.Contains(t, trace, "meta absent or null")
}

func TestProbeLlamaCppRejectsOutOfRangeAndUnknownFtype(t *testing.T) {
	props := llamaCppProps(nil, nil)
	props["default_generation_settings"] = map[string]any{"n_ctx": 0}
	models := llamaCppModels(llamaCppMeta(1<<31, -1, "unknown, may not work"), []string{"completion"})
	api, trace := detectLlamaCpp(t, props, models)

	require.Nil(t, api.Capabilities.Context)
	require.Nil(t, api.Capabilities.ContextTrained)
	require.Equal(t, "", api.Capabilities.Quantization)
	require.Contains(t, trace, "n_ctx = 0 out of range; ignored")
	require.Contains(t, trace, "n_ctx = 2147483648 out of range; ignored")
	require.Contains(t, trace, "n_ctx_train = -1 out of range; ignored")
	require.Contains(t, trace, "has no known GGUF type name; ignored")
}

func TestProbeLlamaCppSendsKeyToBothEndpoints(t *testing.T) {
	var propsAuth, modelsAuth string
	mux := http.NewServeMux()
	mux.HandleFunc("/props", func(w http.ResponseWriter, r *http.Request) {
		propsAuth = r.Header.Get("Authorization")
		jsonHandler(llamaCppProps(nil, nil))(w, r)
	})
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		modelsAuth = r.Header.Get("Authorization")
		jsonHandler(llamaCppModels(llamaCppMeta(8192, 131072, nil), []string{"completion"}))(w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := newTestDetector().Detect(context.Background(), config.ModelConfig{
		ModelName: "qwen3-32b", ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions", ApiKey: "secret-key",
	})
	require.NotNil(t, api)
	require.Equal(t, "llamacpp", api.Stack)
	require.Equal(t, "Bearer secret-key", propsAuth)
	require.Equal(t, "Bearer secret-key", modelsAuth)
}

func TestLlamaCppFtypeLabel(t *testing.T) {
	cases := []struct {
		name string
		raw  any
		want string
		ok   bool
	}{
		{"integer Q4_K_M", float64(15), "Q4_K_M", true},
		{"integer with guessed bit", float64(15 | 1024), "Q4_K_M", true},
		{"integer F16", float64(1), "F16", true},
		{"integer MXFP4 MoE", float64(38), "MXFP4_MOE", true},
		{"removed integer", float64(5), "", false},
		{"fractional", 15.5, "", false},
		{"string plain", "Q8_0", "Q8_0", true},
		{"string with size suffix", "Q4_K - Medium", "Q4_K_M", true},
		{"string guessed prefix", "(guessed) Q5_K - Small", "Q5_K_S", true},
		{"string Q2_K medium keeps enum name", "Q2_K - Medium", "Q2_K", true},
		{"string IQ3 mix is IQ3_M", "IQ3_S mix - 3.66 bpw", "IQ3_M", true},
		{"string all F32", "all F32", "F32", true},
		{"string unknown", "unknown, may not work", "", false},
		{"other type", true, "", false},
		{"nil", nil, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := llamaCppFtypeLabel(c.raw)
			require.Equal(t, c.ok, ok)
			require.Equal(t, c.want, got)
		})
	}
	for _, ft := range llamaCppFtypes {
		var facts apispec.CapabilityFacts
		require.True(t, facts.SetQuantization(ft.name, llamaCppSrcMetaFtype), "every GGUF type name must pass the quantization label check: %s", ft.name)
	}
}

func TestProbeLlamaCppIgnoresWrongTypedFields(t *testing.T) {
	cases := []struct {
		name  string
		nCtx  any
		train any
	}{
		{"string numbers", "8192", "131072"},
		{"fractional numbers", 8192.5, 131072.5},
		{"object numbers", map[string]any{"value": 8192}, []any{131072}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			props := llamaCppProps(map[string]any{
				"supports_tool_calls":          1,
				"supports_parallel_tool_calls": "true",
			}, map[string]any{"vision": "yes", "audio": 1, "video": map[string]any{}})
			props["default_generation_settings"] = map[string]any{"n_ctx": c.nCtx}
			models := llamaCppModels(llamaCppMeta(c.nCtx, c.train, nil), []string{"completion"})
			api, trace := detectLlamaCpp(t, props, models)

			caps := api.Capabilities
			require.NotNil(t, caps)
			require.Nil(t, caps.Context)
			require.Nil(t, caps.ContextTrained)
			require.Nil(t, caps.Modalities)
			require.Nil(t, caps.Tools)
			require.Nil(t, caps.ToolsParallel)
			for _, field := range []string{"capabilities.context", "capabilities.contextTrained", "capabilities.modalities.input", "capabilities.tools", "capabilities.toolsParallel"} {
				require.NotContains(t, api.Evidence[field], "llamacpp:", "%s must not be sourced from a wrong-typed value", field)
			}
			require.NotContains(t, trace, "n_ctx = ")
			require.NotContains(t, trace, "n_ctx_train = ")
			require.NotContains(t, trace, "modalities.")
			require.NotContains(t, trace, "supports_tool_calls = ")
			require.NotContains(t, trace, "supports_parallel_tool_calls = ")
		})
	}
}

func TestProbeLlamaCppQuotesRawFtypeInTrace(t *testing.T) {
	props := llamaCppProps(nil, nil)
	models := llamaCppModels(llamaCppMeta(nil, nil, "Q4_K - Medium\nforged trace line"), []string{"completion"})
	api, trace := detectLlamaCpp(t, props, models)

	require.Equal(t, "", api.Capabilities.Quantization)
	require.NotContains(t, trace, "\nforged trace line")
	require.Contains(t, trace, `meta.ftype "Q4_K - Medium\nforged trace line" has no known GGUF type name; ignored`)

	models = llamaCppModels(llamaCppMeta(nil, nil, 15), []string{"completion"})
	api, trace = detectLlamaCpp(t, props, models)
	require.Equal(t, "Q4_K_M", api.Capabilities.Quantization)
	require.Contains(t, trace, `meta.ftype "15": quantization Q4_K_M`)
}
