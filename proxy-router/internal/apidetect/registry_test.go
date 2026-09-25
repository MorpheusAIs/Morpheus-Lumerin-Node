package apidetect

import (
	"context"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec/registry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

// withoutRegistry keeps a vendor-probe fixture about the vendor's own facts:
// the real model ids those fixtures use are in the snapshot, and the registry
// would otherwise fill what the probe deliberately left unset.
func withoutRegistry(t *testing.T) {
	t.Helper()
	prev := registryMatch
	registryMatch = func(...string) (*apispec.RegistryFacts, int, bool) { return nil, -1, false }
	t.Cleanup(func() { registryMatch = prev })
}

func registryEvidence(api *system.ModelApiSpec) map[string]string {
	out := map[string]string{}
	for k, v := range api.Evidence {
		if strings.HasPrefix(v, "registry:") {
			out[k] = v
		}
	}
	return out
}

func registryLines(trace []string) []string {
	var out []string
	for _, l := range trace {
		if strings.HasPrefix(l, "registry:") {
			out = append(out, l)
		}
	}
	return out
}

// A hostname-recognised vendor without a probe (openai) still gets the
// snapshot's view of the model.
func TestRegistryFillsHostnameRecognisedOpenAIModel(t *testing.T) {
	origin, hits := stubVendorHost(t, "openai")
	api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "gpt-4o-mini", ApiType: "openai", ApiURL: origin + "/v1/chat/completions"})
	require.NotNil(t, api)
	require.Zero(t, hits.Load(), "the registry is offline")
	require.Equal(t, "openai", api.Stack)

	source := apispec.RegistrySource("models.dev+litellm", registry.Date())
	require.Equal(t, "registry:models.dev+litellm@"+registry.Date(), source)
	caps := api.Capabilities
	require.NotNil(t, caps)
	require.EqualValues(t, 128000, *caps.Context)
	require.EqualValues(t, 16384, *caps.MaxOutput)
	require.Equal(t, true, *caps.Tools)
	require.Equal(t, []string{"json_schema"}, caps.Structured)
	require.Equal(t, []string{"text", "image", "pdf"}, caps.Modalities.Input)
	for _, field := range []string{"context", "maxOutput", "tools", "structured", "webSearch", "promptCache", "modalities.input", "modalities.output"} {
		require.Equal(t, source, api.Evidence[system.EvidenceKeyCapability(field)], field)
	}
	require.Equal(t, []string{`registry: models.dev+litellm@` + registry.Date() + ` matched "bare model name"`}, registryLines(trace), "the registry line names the candidate kind, not the id")
}

func TestRegistryMissIsTraced(t *testing.T) {
	origin, _ := stubVendorHost(t, "openai")
	api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "no-such-model-xyz", ApiType: "openai", ApiURL: origin + "/v1/chat/completions"})
	require.NotNil(t, api)
	require.Nil(t, api.Capabilities)
	require.Equal(t, []string{"registry: models.dev+litellm@" + registry.Date() + " has no entry for the served id, bare name or configured name"}, registryLines(trace))
}

// A self-hosted engine serves whatever it was started with; the snapshot's
// row for the same model id says nothing about that deployment.
func TestRegistryNotConsultedForSelfHostedEngines(t *testing.T) {
	vllm, _ := vllmServer(t, "deepseek-ai/DeepSeek-R1")
	api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "deepseek-r1", ApiType: "openai", ApiURL: vllm.URL + "/v1/chat/completions"})
	require.NotNil(t, api)
	require.Equal(t, "vllm", api.Stack)
	require.Empty(t, registryEvidence(api))
	require.Empty(t, registryLines(trace))
	if api.Capabilities != nil {
		require.Nil(t, api.Capabilities.Context)
		require.Nil(t, api.Capabilities.MaxOutput)
	}
}

// Rule 10: the registry fills only what nothing else set. A declared value
// and a vendor-reported value each keep their own evidence; the registry
// supplies the rest and the far-below note fires on the served window.
func TestRegistryFillsOnlyUnsetFields(t *testing.T) {
	t.Run("declared beats registry", func(t *testing.T) {
		origin, _ := stubVendorHost(t, "openai")
		declared := &config.DeclaredCapabilities{Context: int64p(1000), Modalities: &config.DeclaredModalities{Input: []string{"text"}}}
		api, _ := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "gpt-4o-mini", ApiType: "openai", ApiURL: origin + "/v1/chat/completions", Capabilities: declared})
		require.NotNil(t, api)
		require.EqualValues(t, 1000, *api.Capabilities.Context)
		require.Equal(t, []string{"text"}, api.Capabilities.Modalities.Input)
		require.Equal(t, apispec.SourceDeclared, api.Evidence["capabilities.context"])
		require.Equal(t, apispec.SourceDeclared, api.Evidence["capabilities.modalities.input"])
		require.EqualValues(t, 16384, *api.Capabilities.MaxOutput)
		require.True(t, strings.HasPrefix(api.Evidence["capabilities.maxOutput"], "registry:"))
	})

	t.Run("vendor facts beat registry", func(t *testing.T) {
		srv, fx := serveVendor(t, "groq", "/openai/v1/models/llama-3.3-70b-versatile", map[string]any{"id": "llama-3.3-70b-versatile", "context_window": 4096})
		api, trace := detectVendor(t, srv, config.ModelConfig{ModelName: "llama-3.3-70b-versatile", ApiKey: vendorTestKey})
		require.Equal(t, int32(1), fx.hits.Load())
		require.NotNil(t, api)
		require.EqualValues(t, 4096, *api.Capabilities.Context)
		require.Equal(t, "groq:/openai/v1/models/{id}.context_window", api.Evidence["capabilities.context"])
		require.EqualValues(t, 32768, *api.Capabilities.MaxOutput)
		require.Equal(t, apispec.RegistrySource("models.dev", registry.Date()), api.Evidence["capabilities.maxOutput"])
		require.Contains(t, trace, "context: served window 4096 is far below the registry's 128000")
	})
}
