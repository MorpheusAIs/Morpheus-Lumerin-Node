package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestModelConfigV2ParsesModelFamilyAndApiStack(t *testing.T) {
	raw := `{"models":[{"modelId":"0x01","modelName":"qwen3-235b","apiType":"openai","apiStack":"vllm","modelFamily":"qwen3","apiUrl":"http://h:8000/v1/chat/completions"}]}`
	var v2 ModelConfigsV2
	require.NoError(t, json.Unmarshal([]byte(raw), &v2))
	require.Len(t, v2.Models, 1)
	require.Equal(t, "openai", v2.Models[0].ApiType)
	require.Equal(t, "vllm", v2.Models[0].ApiStack)
	require.Equal(t, "qwen3", v2.Models[0].ModelFamily)
}

func TestModelConfigModelFamilyAndApiStackAreOptional(t *testing.T) {
	var cfg ModelConfig
	require.NoError(t, json.Unmarshal([]byte(`{"modelName":"m","apiType":"openai","apiUrl":"http://h/v1"}`), &cfg))
	require.Equal(t, "", cfg.ModelFamily)
	require.Equal(t, "", cfg.ApiStack)
}

// StackTransport is the single preset→adapter table: exactly the nine
// presets, all on the openai adapter except anthropic on claudeai, and no
// legacy adapter names (those are apiType values, not stacks).
func TestStackTransportTable(t *testing.T) {
	require.Len(t, StackTransport, 9)
	for _, stack := range []string{"openai", "vllm", "sglang", "llamacpp", "ollama", "venice", "openrouter", "litellm"} {
		require.Equal(t, "openai", StackTransport[stack], stack)
	}
	require.Equal(t, "claudeai", StackTransport["anthropic"])
	for _, legacy := range []string{"claudeai", "prodia-sd", "prodia-sdxl", "prodia-v2", "hyperbolic-sd"} {
		require.NotContains(t, StackTransport, legacy, "legacy adapter %q must not be an apiStack value", legacy)
	}
}

// IsChatTransport must accept exactly the transports StackTransport actually
// serves chat over, and reject a legacy image-only apiType and the empty
// value.
func TestIsChatTransport(t *testing.T) {
	require.True(t, IsChatTransport("openai"))
	require.True(t, IsChatTransport("claudeai"))
	require.False(t, IsChatTransport("prodia-v2"))
	require.False(t, IsChatTransport(""))
}

func TestValidateApiStackAcceptsEveryPresetOnItsTransport(t *testing.T) {
	for stack, adapter := range StackTransport {
		require.NoError(t, ValidateApiStack("0x01", ModelConfig{ApiType: adapter, ApiStack: stack}), stack)
	}
}

func TestValidateApiStackRejectsUnknown(t *testing.T) {
	err := ValidateApiStack("0x01", ModelConfig{ApiType: "openai", ApiStack: "bogus"})
	require.EqualError(t, err, `model 0x01: unknown apiStack "bogus"`)
	// a legacy adapter name is an apiType, never an apiStack
	err = ValidateApiStack("0x02", ModelConfig{ApiType: "claudeai", ApiStack: "claudeai"})
	require.EqualError(t, err, `model 0x02: unknown apiStack "claudeai"`)
}

func TestValidateApiStackRejectsTransportMismatch(t *testing.T) {
	err := ValidateApiStack("0x01", ModelConfig{ApiType: "claudeai", ApiStack: "vllm"})
	require.EqualError(t, err, `model 0x01: apiStack "vllm" requires apiType "openai", got "claudeai"`)
	err = ValidateApiStack("0x02", ModelConfig{ApiType: "openai", ApiStack: "anthropic"})
	require.EqualError(t, err, `model 0x02: apiStack "anthropic" requires apiType "claudeai", got "openai"`)
}

func TestValidateApiStackEmptyIsAcceptedForEveryLegacyApiType(t *testing.T) {
	for _, legacy := range []string{"openai", "claudeai", "prodia-sd", "prodia-sdxl", "prodia-v2", "hyperbolic-sd"} {
		require.NoError(t, ValidateApiStack("0x01", ModelConfig{ApiType: legacy}), legacy)
	}
}

// --- Init()-level coverage: the loader applies ValidateApiStack on both
// config shapes, stores the normalized value, and degrades a model with an
// invalid apiStack (field cleared, model kept) instead of rejecting the file.

type noopValidator struct{}

func (noopValidator) Struct(interface{}) error { return nil }

type fakeChain struct{}

func (fakeChain) ModelExists(context.Context, common.Hash) (bool, error) { return true, nil }

type fakeConn struct{}

func (fakeConn) TryConnect(context.Context, string) error { return nil }

func newTestLoader(t *testing.T, content string) *ModelConfigLoader {
	t.Helper()
	path := filepath.Join(t.TempDir(), "models-config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return NewModelConfigLoader(path, "", noopValidator{}, fakeChain{}, fakeConn{}, lib.NewTestLogger())
}

func TestInitV2AcceptsApiStackAndLegacyApiTypes(t *testing.T) {
	l := newTestLoader(t, `{"models":[
		{"modelId":"0x01","modelName":"qwen3-32b","apiType":"openai","apiStack":"vllm","apiUrl":"http://h/v1/chat/completions"},
		{"modelId":"0x02","modelName":"claude-sonnet-4-5","apiType":"claudeai","apiStack":"anthropic","apiUrl":"https://api.anthropic.com/v1/messages"},
		{"modelId":"0x03","modelName":"sd-xl","apiType":"prodia-v2","apiUrl":"https://inference.prodia.com/v2"}]}`)
	require.NoError(t, l.Init())
	ids, _ := l.GetAll()
	require.Len(t, ids, 3)
	require.Equal(t, "vllm", l.ModelConfigFromID("0x01").ApiStack)
	require.Equal(t, "anthropic", l.ModelConfigFromID("0x02").ApiStack)
	require.Equal(t, "", l.ModelConfigFromID("0x03").ApiStack)
}

// An invalid apiStack is a mistake in an advertisement-only field: the loader
// logs it, clears the field for that model and keeps serving every model.
// Rejecting the file would silently serve nothing, because main.go only
// warns on an Init error.
func TestInitV2IgnoresBadApiStackAndKeepsEveryModel(t *testing.T) {
	l := newTestLoader(t, `{"models":[
		{"modelId":"0x01","modelName":"ok","apiType":"openai","apiStack":"vllm","apiUrl":"http://h/v1"},
		{"modelId":"0x02","modelName":"mismatch","apiType":"openai","apiStack":"anthropic","apiUrl":"http://h/v1"},
		{"modelId":"0x03","modelName":"unknown","apiType":"openai","apiStack":"bogus","apiUrl":"http://h/v1"}]}`)
	require.NoError(t, l.Init())
	ids, _ := l.GetAll()
	require.Len(t, ids, 3, "every model is stored; only the bad apiStack is dropped")
	require.Equal(t, "vllm", l.ModelConfigFromID("0x01").ApiStack)
	require.Equal(t, "", l.ModelConfigFromID("0x02").ApiStack, "transport mismatch: apiStack dropped, model kept")
	require.Equal(t, "mismatch", l.ModelConfigFromID("0x02").ModelName)
	require.Equal(t, "", l.ModelConfigFromID("0x03").ApiStack, "unknown preset: apiStack dropped, model kept")
	require.Equal(t, "unknown", l.ModelConfigFromID("0x03").ModelName)
}

func TestInitLegacyMapIgnoresUnknownApiStack(t *testing.T) {
	l := newTestLoader(t, `{"0x01":{"modelName":"m","apiType":"openai","apiStack":"bogus","apiUrl":"http://h/v1"}}`)
	require.NoError(t, l.Init())
	ids, _ := l.GetAll()
	require.Len(t, ids, 1)
	require.Equal(t, "", l.ModelConfigFromID("0x01").ApiStack)
	require.Equal(t, "m", l.ModelConfigFromID("0x01").ModelName)
}

// apiStack is trimmed and lowercased before validation, so " vLLM " is the
// vllm preset and a whitespace-only value means unset — a stray space in an
// existing file must not change what the provider serves.
func TestValidateApiStackNormalizesWhitespaceAndCase(t *testing.T) {
	require.Equal(t, "vllm", NormalizeApiStack(" vLLM\t"))
	require.Equal(t, "", NormalizeApiStack("   "))
	require.NoError(t, ValidateApiStack("0x01", ModelConfig{ApiType: "openai", ApiStack: " VLLM "}))
	require.NoError(t, ValidateApiStack("0x01", ModelConfig{ApiType: "claudeai", ApiStack: "Anthropic"}))
	for _, legacy := range []string{"openai", "claudeai", "prodia-sd", "prodia-sdxl", "prodia-v2", "hyperbolic-sd"} {
		require.NoError(t, ValidateApiStack("0x01", ModelConfig{ApiType: legacy, ApiStack: " "}), legacy)
	}
	// errors report the normalized value
	require.EqualError(t, ValidateApiStack("0x02", ModelConfig{ApiType: "openai", ApiStack: " Bogus "}), `model 0x02: unknown apiStack "bogus"`)
	require.EqualError(t, ValidateApiStack("0x03", ModelConfig{ApiType: "openai", ApiStack: "ANTHROPIC"}), `model 0x03: apiStack "anthropic" requires apiType "claudeai", got "openai"`)
}

func TestInitV2NormalizesApiStack(t *testing.T) {
	l := newTestLoader(t, `{"models":[
		{"modelId":"0x01","modelName":"a","apiType":"openai","apiStack":" Vllm ","apiUrl":"http://h/v1"},
		{"modelId":"0x02","modelName":"b","apiType":"openai","apiStack":" ","apiUrl":"http://h/v1"}]}`)
	require.NoError(t, l.Init())
	ids, _ := l.GetAll()
	require.Len(t, ids, 2)
	require.Equal(t, "vllm", l.ModelConfigFromID("0x01").ApiStack, "stored normalized")
	require.Equal(t, "", l.ModelConfigFromID("0x02").ApiStack, "whitespace-only reads as unset")
}

func TestInitLegacyMapNormalizesApiStack(t *testing.T) {
	l := newTestLoader(t, `{"0x01":{"modelName":"m","apiType":"claudeai","apiStack":" ANTHROPIC ","apiUrl":"http://h/v1"}}`)
	require.NoError(t, l.Init())
	require.Equal(t, "anthropic", l.ModelConfigFromID("0x01").ApiStack)
}

func TestInitLegacyMapAcceptsApiStack(t *testing.T) {
	l := newTestLoader(t, `{"0x01":{"modelName":"qwen3-32b","apiType":"openai","apiStack":"sglang","apiUrl":"http://h/v1"}}`)
	require.NoError(t, l.Init())
	require.Equal(t, "sglang", l.ModelConfigFromID("0x01").ApiStack)
}
