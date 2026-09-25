package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

// Field set of one GET /api/v1/models entry per
// https://openrouter.ai/docs/api/api-reference/models/get-models (research plan §6.1).
func openrouterEntry(id string, overrides map[string]any) map[string]any {
	entry := map[string]any{
		"id":   id,
		"name": "Vendor: Model",
		"architecture": map[string]any{
			"modality":          "text+image+file->text",
			"input_modalities":  []string{"text", "image", "file"},
			"output_modalities": []string{"text"},
			"tokenizer":         "Other",
		},
		"context_length": 163840,
		"top_provider": map[string]any{
			"context_length":        131072,
			"max_completion_tokens": 32768,
			"is_moderated":          false,
		},
		"pricing": map[string]any{"prompt": "0.0000002", "completion": "0.0000008", "web_search": "0"},
		"supported_parameters": []string{
			"include_reasoning", "max_tokens", "parallel_tool_calls", "reasoning", "reasoning_effort",
			"response_format", "seed", "structured_outputs", "temperature", "tool_choice", "tools", "top_p",
		},
		"reasoning": map[string]any{
			"mandatory":           false,
			"default_enabled":     true,
			"supported_efforts":   []string{"none", "low", "medium", "high"},
			"default_effort":      "medium",
			"supports_max_tokens": true,
		},
	}
	for k, v := range overrides {
		if v == nil {
			delete(entry, k)
		} else {
			entry[k] = v
		}
	}
	return entry
}

// Field set of one GET /api/v1/models?type=text entry per
// https://docs.venice.ai/api-reference/endpoint/models/list (research plan §6.1).
func veniceSpecEntry(id string, caps map[string]any) map[string]any {
	capabilities := map[string]any{
		"optimizedForCode":        false,
		"quantization":            "fp8",
		"supportsFunctionCalling": true,
		"supportsLogProbs":        true,
		"supportsReasoning":       true,
		"supportsReasoningEffort": true,
		"reasoningEffortOptions":  []string{"low", "medium", "high"},
		"defaultReasoningEffort":  "medium",
		"supportsResponseSchema":  true,
		"supportsVision":          true,
		"supportsMultipleImages":  true,
		"maxImages":               4,
		"supportsVideoInput":      false,
		"supportsAudioInput":      false,
		"supportsWebSearch":       true,
		"supportsXSearch":         true,
		"supportsTeeAttestation":  false,
		"supportsE2EE":            false,
	}
	for k, v := range caps {
		if v == nil {
			delete(capabilities, k)
		} else {
			capabilities[k] = v
		}
	}
	return map[string]any{
		"id":             id,
		"object":         "model",
		"type":           "text",
		"context_length": 131072,
		"model_spec": map[string]any{
			"availableContextTokens": 131072,
			"maxCompletionTokens":    16384,
			"traits":                 []string{"default"},
			"modelSource":            "https://huggingface.co/Qwen/Qwen3-235B-A22B",
			"capabilities":           capabilities,
		},
	}
}

func detectListing(t *testing.T, vendor string, entries ...map[string]any) (*system.ModelApiSpec, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"data": entries}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	mapHostToVendor(t, srv.URL, vendor)

	id, _ := entries[0]["id"].(string)
	api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{ModelName: id, ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	require.NotNil(t, api)
	require.Equal(t, vendor, api.Stack)
	return api, strings.Join(trace, "\n")
}

// detectListingByShape serves the listing on a host no vendor table knows, so
// the detector reaches probeRegistryShape through the fingerprint fallback.
func detectListingByShape(t *testing.T, entries ...map[string]any) (*system.ModelApiSpec, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", jsonHandler(map[string]any{"data": entries}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	id, _ := entries[0]["id"].(string)
	api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{ModelName: id, ApiType: "openai", ApiURL: srv.URL + "/api/v1/chat/completions"})
	require.NotNil(t, api)
	return api, strings.Join(trace, "\n")
}

func TestOpenRouterEntryCapabilities(t *testing.T) {
	api, trace := detectListing(t, "openrouter", openrouterEntry("vendor/model", nil))

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.EqualValues(t, 163840, *caps.Context)
	require.Nil(t, caps.ContextTrained)
	require.EqualValues(t, 32768, *caps.MaxOutput)
	require.Equal(t, []string{"text", "image", "pdf"}, caps.Modalities.Input)
	require.Equal(t, []string{"text"}, caps.Modalities.Output)
	require.Equal(t, true, *caps.Tools)
	require.Equal(t, true, *caps.ToolsParallel)
	require.Equal(t, []string{"json_object", "json_schema"}, caps.Structured)
	require.Equal(t, true, *caps.WebSearch)
	require.Equal(t, true, *caps.Reasoning)
	require.Empty(t, caps.Quantization)

	require.Equal(t, map[string]string{
		"capabilities.context":           "openrouter:/models.context_length",
		"capabilities.maxOutput":         "openrouter:/models.top_provider.max_completion_tokens",
		"capabilities.modalities.input":  "openrouter:/models.architecture.input_modalities",
		"capabilities.modalities.output": "openrouter:/models.architecture.output_modalities",
		"capabilities.tools":             "openrouter:/models.supported_parameters",
		"capabilities.toolsParallel":     "openrouter:/models.supported_parameters",
		"capabilities.structured":        "openrouter:/models.supported_parameters",
		"capabilities.webSearch":         "openrouter:/models.plugins",
		"capabilities.reasoning":         "derived:bindings.reasoning",
		"thinking.default":               "openrouter:/models.reasoning.default_effort",
		"bindings.reasoning.effort":      "openrouter:/models.reasoning.supported_efforts",
	}, api.Evidence)

	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "medium", api.Thinking.Default)
	require.Equal(t, []string{"none", "low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
	require.NotNil(t, api.Bindings[system.IntentReasoningDisable])

	for _, name := range []string{"tools", "parallel_tool_calls", "response_format", "structured_outputs", "reasoning", "temperature"} {
		require.Contains(t, api.Parameters, name)
	}
	require.NotContains(t, api.Parameters, "web_search_options")

	for _, line := range []string{
		"/models.context_length 163840 -> context",
		"/models.top_provider.max_completion_tokens 32768 -> maxOutput",
		"/models.architecture.input_modalities [text image pdf] -> modalities.input [text image pdf]",
		"/models.supported_parameters lists tools -> tools",
		"/models.supported_parameters lists structured_outputs -> structured json_schema",
		"/models: the web plugin searches for any openrouter model -> webSearch",
		"/models.reasoning.supported_efforts [none low medium high] -> reasoning.effort levels",
		`/models.reasoning.default_effort "medium" -> thinking.default`,
	} {
		require.Contains(t, trace, line)
	}
}

func TestOpenRouterEntryVariants(t *testing.T) {
	t.Run("web_search_options names the model's own evidence", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/search", map[string]any{
			"supported_parameters": []string{"tools", "web_search_options"},
			"reasoning":            nil,
		}))
		require.Equal(t, true, *api.Capabilities.WebSearch)
		require.Equal(t, "openrouter:/models.supported_parameters", api.Evidence["capabilities.webSearch"])
		require.Contains(t, api.Parameters, "web_search_options")
		require.Nil(t, api.Capabilities.Reasoning)
		require.Nil(t, api.Thinking)
	})

	t.Run("top_provider.context_length only when context_length is absent", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/fallback", map[string]any{"context_length": nil}))
		require.EqualValues(t, 131072, *api.Capabilities.Context)
		require.Equal(t, "openrouter:/models.top_provider.context_length", api.Evidence["capabilities.context"])
	})

	t.Run("null and out-of-range numbers are ignored", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/nulls", map[string]any{
			"context_length": 1 << 40,
			"top_provider":   map[string]any{"context_length": nil, "max_completion_tokens": nil},
		}))
		require.Nil(t, api.Capabilities.Context)
		require.Nil(t, api.Capabilities.MaxOutput)
	})

	t.Run("mandatory reasoning is always_on and drops the toggles", func(t *testing.T) {
		api, trace := detectListing(t, "openrouter", openrouterEntry("vendor/thinker", map[string]any{
			"reasoning": map[string]any{
				"mandatory": true, "default_enabled": true, "supported_efforts": []string{"low", "medium", "high"},
				"supports_max_tokens": false,
			},
		}))
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.Equal(t, "on", api.Thinking.Default)
		require.Equal(t, "openrouter:/models.reasoning.default_enabled", api.Evidence["thinking.default"])
		for _, intent := range []string{system.IntentReasoningDisable, system.IntentReasoningEnable, system.IntentReasoningEffort, system.IntentReasoningBudget} {
			require.Nil(t, api.Bindings[intent], "%s has no knob on an always-on model", intent)
		}
		require.Equal(t, "reasoning.exclude", api.Bindings[system.IntentReasoningFormat].Param)
		require.NotContains(t, api.Evidence, "bindings.reasoning.effort")
		require.Equal(t, true, *api.Capabilities.Reasoning)
		require.Equal(t, "derived:thinking.mode", api.Evidence["capabilities.reasoning"])
		require.Contains(t, trace, "/models.reasoning.mandatory -> reasoning cannot be disabled")
		require.Contains(t, trace, "/models.reasoning.supports_max_tokens false; the reasoning.budget binding stays as documented for the stack")
	})

	t.Run("default_enabled false without a default effort is off", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/quiet", map[string]any{
			"reasoning": map[string]any{"mandatory": false, "default_enabled": false, "supported_efforts": []string{"none", "high"}},
		}))
		require.Equal(t, "off", api.Thinking.Default)
		require.Equal(t, "openrouter:/models.reasoning.default_enabled", api.Evidence["thinking.default"])
	})

	// Live listing: anthropic/claude-opus-4.8 reports default_enabled: false
	// with default_effort: "high"; the level is what applies once reasoning
	// is requested, not evidence that it runs unasked.
	t.Run("default_enabled false wins over a default effort", func(t *testing.T) {
		api, trace := detectListing(t, "openrouter", openrouterEntry("vendor/opt-in", map[string]any{
			"reasoning": map[string]any{
				"mandatory": false, "default_enabled": false, "supported_efforts": []string{"none", "low", "medium", "high"}, "default_effort": "high",
			},
		}))
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
		require.Equal(t, "off", api.Thinking.Default)
		require.Equal(t, "openrouter:/models.reasoning.default_enabled", api.Evidence["thinking.default"])
		require.Contains(t, trace, `/models.reasoning.default_enabled false -> thinking.default off (default_effort "high" applies only when reasoning is requested)`)
		require.NotContains(t, trace, `/models.reasoning.default_effort "high" -> thinking.default`)
	})

	t.Run("default_enabled absent and not mandatory publishes no default", func(t *testing.T) {
		api, trace := detectListing(t, "openrouter", openrouterEntry("vendor/unstated", map[string]any{
			"reasoning": map[string]any{"mandatory": false, "supported_efforts": []string{"low", "medium", "high"}, "default_effort": "medium"},
		}))
		require.NotNil(t, api.Thinking)
		require.Empty(t, api.Thinking.Default)
		require.NotContains(t, api.Evidence, "thinking.default")
		require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Contains(t, trace, "/models.reasoning.default_enabled absent and reasoning not mandatory; thinking.default left unset")
	})

	t.Run("mandatory reasoning takes the default effort as its level", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/forced", map[string]any{
			"reasoning": map[string]any{"mandatory": true, "supported_efforts": []string{"low", "medium", "high"}, "default_effort": "high"},
		}))
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.Equal(t, "high", api.Thinking.Default)
		require.Equal(t, "openrouter:/models.reasoning.default_effort", api.Evidence["thinking.default"])
	})

	t.Run("mandatory reasoning without a toggle or level is on", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/forced-plain", map[string]any{
			"reasoning": map[string]any{"mandatory": true},
		}))
		require.Equal(t, system.ThinkingModeAlwaysOn, api.Thinking.Mode)
		require.Equal(t, "on", api.Thinking.Default)
		require.Equal(t, "openrouter:/models.reasoning.mandatory", api.Evidence["thinking.default"])
	})

	// Live listing: openai/gpt-5.1 reports default_enabled: true with
	// default_effort: "none" — the rule-5 off switch is itself a level.
	t.Run("default_enabled true keeps a none level as reported", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/none-level", map[string]any{
			"reasoning": map[string]any{
				"mandatory": false, "default_enabled": true, "supported_efforts": []string{"none", "low", "medium", "high"}, "default_effort": "none",
			},
		}))
		require.Equal(t, "none", api.Thinking.Default)
		require.Equal(t, "openrouter:/models.reasoning.default_effort", api.Evidence["thinking.default"])
	})

	t.Run("the web plugin is asserted only for a hostname-recognized openrouter", func(t *testing.T) {
		api, trace := detectListingByShape(t, openrouterEntry("vendor/model", nil))
		require.Equal(t, "openrouter", api.Stack)
		require.NotNil(t, api.Capabilities)
		require.Nil(t, api.Capabilities.WebSearch)
		require.NotContains(t, api.Evidence, "capabilities.webSearch")
		require.Equal(t, true, *api.Capabilities.Tools, "facts read from the listing itself still apply")
		require.EqualValues(t, 163840, *api.Capabilities.Context)
		require.Contains(t, trace, "/models: openrouter-shaped listing on an unrecognized host; the web plugin is not assumed, webSearch left unset")
		require.NotContains(t, trace, "the web plugin searches for any openrouter model")
	})

	t.Run("web_search_options on an unrecognized host is still the model's own evidence", func(t *testing.T) {
		api, _ := detectListingByShape(t, openrouterEntry("vendor/search", map[string]any{
			"supported_parameters": []string{"tools", "web_search_options"},
		}))
		require.Equal(t, true, *api.Capabilities.WebSearch)
		require.Equal(t, "openrouter:/models.supported_parameters", api.Evidence["capabilities.webSearch"])
	})

	t.Run("a string-typed context_length is not a token count", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/stringy", map[string]any{
			"context_length": "163840",
			"top_provider":   map[string]any{"context_length": "131072", "max_completion_tokens": "32768"},
		}))
		require.Nil(t, api.Capabilities.Context)
		require.Nil(t, api.Capabilities.MaxOutput)
		require.NotContains(t, api.Evidence, "capabilities.context")
		require.NotContains(t, api.Evidence, "capabilities.maxOutput")
	})

	t.Run("max_completion_tokens out of range is ignored", func(t *testing.T) {
		api, trace := detectListing(t, "openrouter", openrouterEntry("vendor/huge-output", map[string]any{
			"top_provider": map[string]any{"context_length": 131072, "max_completion_tokens": 1 << 40},
		}))
		require.Nil(t, api.Capabilities.MaxOutput)
		require.EqualValues(t, 163840, *api.Capabilities.Context)
		require.Contains(t, trace, "/models.top_provider.max_completion_tokens 1099511627776 is out of range; ignored")
	})

	t.Run("non-string entries in supported_efforts and supported_parameters are dropped", func(t *testing.T) {
		api, _ := detectListing(t, "openrouter", openrouterEntry("vendor/mixed", map[string]any{
			"supported_parameters": []any{"tools", 7, "reasoning", map[string]any{"name": "temperature"}},
			"reasoning":            map[string]any{"mandatory": false, "default_enabled": true, "supported_efforts": []any{"low", 3, "high", true}},
		}))
		require.Equal(t, []string{"low", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
		require.Equal(t, true, *api.Capabilities.Tools)
		require.NotContains(t, api.Parameters, "temperature")
	})

	t.Run("a non-object reasoning field reads as no reasoning block", func(t *testing.T) {
		api, trace := detectListing(t, "openrouter", openrouterEntry("vendor/yes", map[string]any{"reasoning": "yes"}))
		require.NotNil(t, api.Thinking, "supported_parameters still lists reasoning")
		require.Empty(t, api.Thinking.Default)
		require.NotContains(t, api.Evidence, "thinking.default")
		require.NotContains(t, api.Evidence, "bindings.reasoning.effort")
		require.NotContains(t, trace, "/models.reasoning.")
	})

	t.Run("unknown modalities are filtered, none known leaves the field unset", func(t *testing.T) {
		api, trace := detectListing(t, "openrouter", openrouterEntry("vendor/odd", map[string]any{
			"architecture": map[string]any{"input_modalities": []string{"text", "smell"}, "output_modalities": []string{"hologram"}},
		}))
		require.Equal(t, []string{"text"}, api.Capabilities.Modalities.Input)
		require.Nil(t, api.Capabilities.Modalities.Output)
		require.Contains(t, trace, "/models.architecture.output_modalities [hologram] has no known entry; ignored")
	})
}

func TestVeniceSpecCapabilities(t *testing.T) {
	api, trace := detectListing(t, "venice", veniceSpecEntry("qwen3-235b", nil))

	caps := api.Capabilities
	require.NotNil(t, caps)
	require.EqualValues(t, 131072, *caps.Context)
	require.EqualValues(t, 16384, *caps.MaxOutput)
	require.Equal(t, []string{"image"}, caps.Modalities.Input)
	require.Nil(t, caps.Modalities.Output)
	require.Equal(t, true, *caps.Tools)
	require.Nil(t, caps.ToolsParallel)
	require.Equal(t, []string{"json_schema"}, caps.Structured)
	require.Equal(t, true, *caps.WebSearch)
	require.Equal(t, true, *caps.Reasoning)
	require.Equal(t, "fp8", caps.Quantization)

	require.Equal(t, map[string]string{
		"capabilities.context":          "venice:/models.model_spec.availableContextTokens",
		"capabilities.maxOutput":        "venice:/models.model_spec.maxCompletionTokens",
		"capabilities.modalities.input": "venice:/models.model_spec.capabilities.supportsVision",
		"capabilities.tools":            "venice:/models.model_spec.capabilities.supportsFunctionCalling",
		"capabilities.structured":       "venice:/models.model_spec.capabilities.supportsResponseSchema",
		"capabilities.webSearch":        "venice:/models.model_spec.capabilities.supportsWebSearch",
		"capabilities.reasoning":        "derived:bindings.reasoning",
		"capabilities.quantization":     "venice:/models.model_spec.capabilities.quantization",
		"thinking.default":              "venice:/models.model_spec.capabilities.defaultReasoningEffort",
		"bindings.reasoning.effort":     "venice:/models.model_spec.capabilities.reasoningEffortOptions",
	}, api.Evidence)

	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "medium", api.Thinking.Default)
	require.Equal(t, "reasoning_effort", api.Bindings[system.IntentReasoningEffort].Param)
	require.Equal(t, []string{"low", "medium", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
	require.Equal(t, "venice_parameters.disable_thinking", api.Bindings[system.IntentReasoningDisable].Param, "the toggle is on another param; the effort list does not drop it")

	// tools, response_format and logprobs are request parameter names the
	// venice table supplies; the capability-only names must not appear.
	for _, name := range []string{"reasoning", "vision", "web_search"} {
		require.NotContains(t, api.Parameters, name)
	}
	require.Contains(t, api.Parameters, "tools")
	for _, line := range []string{
		"/models.model_spec.availableContextTokens 131072 -> context",
		"/models.model_spec.maxCompletionTokens 16384 -> maxOutput",
		"/models.model_spec.capabilities.supportsFunctionCalling true -> tools",
		"/models.model_spec.capabilities.supportsVision -> modalities.input image",
		"/models.model_spec.capabilities.supportsWebSearch true -> webSearch",
		"/models.model_spec.capabilities.supportsResponseSchema -> structured json_schema",
		"venice listing reports supportsReasoning",
		"/models.model_spec.capabilities.reasoningEffortOptions [low medium high] -> reasoning.effort levels",
		`/models.model_spec.capabilities.defaultReasoningEffort "medium" -> thinking.default`,
		`/models.model_spec.capabilities.quantization "fp8" -> quantization`,
	} {
		require.Contains(t, trace, line)
	}
}

func TestVeniceSpecVariants(t *testing.T) {
	t.Run("supportsReasoning false is a backend false", func(t *testing.T) {
		api, trace := detectListing(t, "venice", veniceSpecEntry("llama-3.3-70b", map[string]any{
			"supportsReasoning": false, "supportsReasoningEffort": false, "reasoningEffortOptions": nil, "defaultReasoningEffort": nil,
		}))
		require.Equal(t, false, *api.Capabilities.Reasoning)
		require.Equal(t, "venice:/models.model_spec.capabilities.supportsReasoning", api.Evidence["capabilities.reasoning"])
		require.Nil(t, api.Thinking)
		require.Nil(t, api.Bindings[system.IntentReasoningEffort])
		require.Contains(t, trace, "/models.model_spec.capabilities.supportsReasoning false -> reasoning false")
	})

	t.Run("effort options need supportsReasoningEffort", func(t *testing.T) {
		api, _ := detectListing(t, "venice", veniceSpecEntry("qwen3-235b", map[string]any{"supportsReasoningEffort": false}))
		require.NotContains(t, api.Evidence, "bindings.reasoning.effort")
		require.Contains(t, api.Bindings[system.IntentReasoningEffort].EnumValues, "xhigh", "the stack table's enum stands")
	})

	t.Run("false flags set no modality and the not-available placeholder no quantization", func(t *testing.T) {
		api, _ := detectListing(t, "venice", veniceSpecEntry("text-only", map[string]any{
			"supportsVision": false, "supportsFunctionCalling": false, "supportsResponseSchema": false, "quantization": "not-available",
		}))
		require.Nil(t, api.Capabilities.Modalities)
		require.Equal(t, false, *api.Capabilities.Tools)
		require.Nil(t, api.Capabilities.Structured)
		require.Empty(t, api.Capabilities.Quantization)
		require.NotContains(t, api.Evidence, "capabilities.quantization")
	})

	t.Run("wrong-type and fractional numerics are not token counts", func(t *testing.T) {
		entry := veniceSpecEntry("odd-numbers", nil)
		spec := entry["model_spec"].(map[string]any)
		spec["availableContextTokens"] = "131072"
		spec["maxCompletionTokens"] = 1.5
		api, _ := detectListing(t, "venice", entry)
		require.Nil(t, api.Capabilities.Context)
		require.Nil(t, api.Capabilities.MaxOutput)
		require.NotContains(t, api.Evidence, "capabilities.context")
		require.NotContains(t, api.Evidence, "capabilities.maxOutput")
	})

	t.Run("absent and out-of-range numerics are ignored", func(t *testing.T) {
		entry := veniceSpecEntry("no-numbers", nil)
		spec := entry["model_spec"].(map[string]any)
		delete(spec, "availableContextTokens")
		spec["maxCompletionTokens"] = 1 << 40
		api, trace := detectListing(t, "venice", entry)
		require.Nil(t, api.Capabilities.Context)
		require.Nil(t, api.Capabilities.MaxOutput)
		require.Contains(t, trace, "/models.model_spec.maxCompletionTokens 1099511627776 is out of range; ignored")
	})

	t.Run("non-string effort options are dropped", func(t *testing.T) {
		api, _ := detectListing(t, "venice", veniceSpecEntry("qwen3-235b", map[string]any{"reasoningEffortOptions": []any{"low", 3, "high", false}}))
		require.Equal(t, []string{"low", "high"}, api.Bindings[system.IntentReasoningEffort].EnumValues)
	})

	t.Run("a quantization label that is not a plain name is ignored", func(t *testing.T) {
		api, trace := detectListing(t, "venice", veniceSpecEntry("qwen3-235b", map[string]any{"quantization": "fp8 (dynamic)"}))
		require.Empty(t, api.Capabilities.Quantization)
		require.NotContains(t, api.Evidence, "capabilities.quantization")
		require.Contains(t, trace, "/models.model_spec.capabilities.quantization is not a plain label; ignored")
	})

	t.Run("video and audio inputs", func(t *testing.T) {
		api, _ := detectListing(t, "venice", veniceSpecEntry("omni", map[string]any{"supportsVideoInput": true, "supportsAudioInput": true}))
		require.Equal(t, []string{"image", "video", "audio"}, api.Capabilities.Modalities.Input)
	})

	t.Run("supportsReasoning alone identifies the stack without a capabilities block", func(t *testing.T) {
		api, _ := detectListing(t, "venice", veniceEntry("qwen3-235b", true))
		require.Nil(t, api.Capabilities, "no fact was read, so the pre-capabilities wire shape stands")
		require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	})
}

func TestRegistryEvidenceSourcesNameTheirStack(t *testing.T) {
	for _, vendor := range []string{"openrouter", "venice"} {
		t.Run(vendor, func(t *testing.T) {
			var api *system.ModelApiSpec
			if vendor == "openrouter" {
				api, _ = detectListing(t, vendor, openrouterEntry("vendor/model", nil))
			} else {
				api, _ = detectListing(t, vendor, veniceSpecEntry("qwen3-235b", nil))
			}
			for key, source := range api.Evidence {
				if strings.HasPrefix(source, "derived:") {
					continue
				}
				require.True(t, strings.HasPrefix(source, vendor+":/models"), "%s = %s", key, source)
			}
		})
	}
}
