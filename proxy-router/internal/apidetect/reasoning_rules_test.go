package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

func veniceEntryWithEfforts(id string, efforts ...string) map[string]any {
	return map[string]any{"id": id, "model_spec": map[string]any{"capabilities": map[string]any{
		"supportsReasoning": true, "supportsReasoningEffort": true, "reasoningEffortOptions": efforts,
	}}}
}

func TestDetectVeniceListingEffortsDecideAlwaysOn(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"data": []map[string]any{
		veniceEntryWithEfforts("minimax-m27", "none", "low", "medium", "high"),
		veniceEntryWithEfforts("kimi-k2-7-code", "low", "medium", "high"),
	}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := NewDetector(lib.NewTestLogger(), DefaultOptions())
	minimax, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "minimax-m27", ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	require.Equal(t, "venice", minimax.Stack)
	require.Equal(t, system.ThinkingModeControllable, minimax.Thinking.Mode, "venice offers effort none, so the M2.x always-on rule does not apply")
	require.Equal(t, "venice_parameters.disable_thinking", minimax.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, []string{"none", "low", "medium", "high"}, minimax.Bindings[system.IntentReasoningEffort].EnumValues)
	require.Contains(t, strings.Join(trace, "\n"), `venice listing offers reasoning efforts ["none" "low" "medium" "high"]`)

	kimi := d.Detect(context.Background(), config.ModelConfig{ModelName: "kimi-k2-7-code", ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	require.Equal(t, "kimi", kimi.ModelFamily)
	require.Equal(t, system.ThinkingModeAlwaysOn, kimi.Thinking.Mode, "dashed K2.7-Code is always on and venice offers no none")
	require.Nil(t, kimi.Bindings[system.IntentReasoningDisable], "no off switch is advertised for a model that cannot stop")
	require.Equal(t, "reasoning.effort", kimi.Bindings[system.IntentReasoningEffort].Param)
	require.Equal(t, []string{"low", "medium", "high"}, kimi.Bindings[system.IntentReasoningEffort].EnumValues)
}

func TestDetectLiteLLMTwoHopCarriesVeniceEfforts(t *testing.T) {
	venice, veniceLog := registryServer(t, veniceEntryWithEfforts("minimax-m27", "none", "low", "medium", "high"), 0)
	mapHostToVendor(t, venice.URL, "venice")
	group := map[string]any{"model_group": "my-chat", "supported_openai_params": []string{"temperature", "tools"}, "supports_reasoning": false}
	litellm, _ := litellmServer(t, group, []map[string]any{
		deployment("my-chat", "openai/minimax-m27:include_venice_system_prompt=false", venice.URL+"/api/v1", nil),
	})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "venice", api.Stack)
	require.Equal(t, "litellm", api.Via)
	require.Equal(t, "minimax", api.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	effort := api.Bindings[system.IntentReasoningEffort]
	require.NotNil(t, effort, "the reasoning object survives the LiteLLM filter, unlike reasoning_effort")
	require.Equal(t, "reasoning.effort", effort.Param)
	require.Equal(t, []string{"none", "low", "medium", "high"}, effort.EnumValues)
	require.Contains(t, trace, `upstream venice listing offers reasoning efforts ["none" "low" "medium" "high"]`)
	veniceLog.requireAnonymous(t)
}

func TestDetectVeniceEffortLevelsNeedSupportsReasoningEffort(t *testing.T) {
	entry := veniceEntryWithEfforts("minimax-m27", "none", "low", "high")
	entry["model_spec"].(map[string]any)["capabilities"].(map[string]any)["supportsReasoningEffort"] = false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"data": []map[string]any{entry}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := NewDetector(lib.NewTestLogger(), DefaultOptions())
	api, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "minimax-m27", ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode, "levels the listing does not accept prove no off switch")
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
	require.Contains(t, strings.Join(trace, "\n"), "levels ignored")
}

func vllmServerWithRoot(t *testing.T, id, root string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/version", jsonHandler(map[string]any{"version": "0.18.0"}))
	mux.HandleFunc("/v1/models", jsonHandler(map[string]any{"object": "list", "data": []map[string]any{{"id": id, "root": root, "owned_by": "vllm"}}}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestDetectVLLMWeightsNameTheModel(t *testing.T) {
	d := NewDetector(lib.NewTestLogger(), DefaultOptions())
	detectServed := func(id, root string) (*system.ModelApiSpec, string) {
		srv := vllmServerWithRoot(t, id, root)
		api, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: id, ApiType: "openai", ApiURL: srv.URL + "/v1/chat/completions"})
		return api, strings.Join(trace, "\n")
	}

	hybrid, trace := detectServed("qwen3-32b-thinking", "Qwen/Qwen3-32B")
	require.Equal(t, system.ThinkingModeControllable, hybrid.Thinking.Mode, "a hybrid served under a thinking label keeps its switch")
	require.Equal(t, "chat_template_kwargs.enable_thinking", hybrid.Bindings[system.IntentReasoningDisable].Param)
	require.Contains(t, trace, `vllm serves "qwen3-32b-thinking" from weights "Qwen3-32B"`)

	checkpoint, _ := detectServed("qwen3-235b", "/models/Qwen3-235B-A22B-Thinking-2507")
	require.Nil(t, checkpoint.Bindings[system.IntentReasoningDisable], "a thinking release served under a plain label gets no template switch")
	require.Equal(t, system.ThinkingModeTunable, checkpoint.Thinking.Mode)

	local, trace := detectServed("kimi-k2.5", "/models/ft")
	require.Equal(t, "chat_template_kwargs.thinking", local.Bindings[system.IntentReasoningDisable].Param, "weights named after no family leave the served name in charge")
	require.NotContains(t, trace, "/models/", "a local path stays out of the trace")
}

func TestDetectVeniceNoEffortDropsTheEffortKnob(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"data": []map[string]any{{"id": "qwen3-235b", "model_spec": map[string]any{"capabilities": map[string]any{
		"supportsReasoning": true, "supportsReasoningEffort": false,
	}}}}}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := NewDetector(lib.NewTestLogger(), DefaultOptions())
	api, trace := d.DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "qwen3-235b", ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	require.Equal(t, "venice", api.Stack)
	require.Nil(t, api.Bindings[system.IntentReasoningEffort], "venice says the model takes no effort level")
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Contains(t, strings.Join(trace, "\n"), "supportsReasoningEffort false")
}

func TestDetectLiteLLMTwoHopCarriesVeniceNoEffort(t *testing.T) {
	entry := map[string]any{"id": "qwen3-235b", "model_spec": map[string]any{"capabilities": map[string]any{"supportsReasoning": true, "supportsReasoningEffort": false}}}
	venice, _ := registryServer(t, entry, 0)
	mapHostToVendor(t, venice.URL, "venice")
	group := map[string]any{"model_group": "my-chat", "supported_openai_params": []string{"temperature", "tools"}, "supports_reasoning": false}
	litellm, _ := litellmServer(t, group, []map[string]any{deployment("my-chat", "openai/qwen3-235b", venice.URL+"/api/v1", nil)})

	api, trace := detectVia(t, litellm.URL, "my-chat", DefaultOptions())
	require.Equal(t, "venice", api.Stack)
	require.Nil(t, api.Bindings[system.IntentReasoningEffort])
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param)
	require.Contains(t, trace, "upstream venice listing says the model takes no effort level")
}
