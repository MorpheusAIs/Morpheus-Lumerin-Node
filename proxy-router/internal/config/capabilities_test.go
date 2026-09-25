package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func i64(v int64) *int64 { return &v }
func bp(v bool) *bool    { return &v }

func TestValidateDeclaredCapabilities(t *testing.T) {
	cases := []struct {
		name string
		caps *DeclaredCapabilities
		err  string
	}{
		{"nil block", nil, ""},
		{"empty block", &DeclaredCapabilities{}, ""},
		{"full valid block", &DeclaredCapabilities{Context: i64(32768), MaxOutput: i64(4096), Modalities: &DeclaredModalities{Input: []string{"text", "image"}, Output: []string{"text"}}, Tools: bp(true), Structured: []string{"json_schema", "grammar"}, WebSearch: bp(false), Embeddings: bp(false)}, ""},
		{"context zero", &DeclaredCapabilities{Context: i64(0)}, "model 0x01: invalid capabilities: context must be in (0, 2147483648), got 0"},
		{"context negative", &DeclaredCapabilities{Context: i64(-5)}, "model 0x01: invalid capabilities: context must be in (0, 2147483648), got -5"},
		{"context at the bound", &DeclaredCapabilities{Context: i64(1 << 31)}, "model 0x01: invalid capabilities: context must be in (0, 2147483648), got 2147483648"},
		{"context just below the bound", &DeclaredCapabilities{Context: i64(1<<31 - 1)}, ""},
		{"maxOutput zero", &DeclaredCapabilities{MaxOutput: i64(0)}, "model 0x01: invalid capabilities: maxOutput must be in (0, 2147483648), got 0"},
		{"unknown input modality", &DeclaredCapabilities{Modalities: &DeclaredModalities{Input: []string{"text", "vision"}}}, `model 0x01: invalid capabilities: modalities.input: unknown value "vision" (allowed: text, image, audio, video, pdf)`},
		{"unknown output modality", &DeclaredCapabilities{Modalities: &DeclaredModalities{Output: []string{"pdf"}}}, `model 0x01: invalid capabilities: modalities.output: unknown value "pdf" (allowed: text, image, audio, embeddings)`},
		{"duplicate modality", &DeclaredCapabilities{Modalities: &DeclaredModalities{Input: []string{"text", "text"}}}, `model 0x01: invalid capabilities: modalities.input: duplicate value "text"`},
		{"unknown structured kind", &DeclaredCapabilities{Structured: []string{"json_schema", "xml"}}, `model 0x01: invalid capabilities: structured: unknown value "xml" (allowed: json_object, json_schema, grammar, regex, choice)`},
		{"duplicate structured kind", &DeclaredCapabilities{Structured: []string{"regex", "regex"}}, `model 0x01: invalid capabilities: structured: duplicate value "regex"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDeclaredCapabilities("0x01", tc.caps)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestDeclaredCapabilitiesStrictDecodeRecordsErrorsInsteadOfFailing(t *testing.T) {
	t.Run("unknown key", func(t *testing.T) {
		var cfg ModelConfig
		require.NoError(t, json.Unmarshal([]byte(`{"modelName":"m","apiType":"openai","apiUrl":"http://h/v1","capabilities":{"context":1024,"contextTrained":4096}}`), &cfg), "the model's own decode never fails")
		require.NotNil(t, cfg.Capabilities)
		require.EqualError(t, ValidateDeclaredCapabilities("0x01", cfg.Capabilities), `model 0x01: invalid capabilities: json: unknown field "contextTrained"`)
	})
	t.Run("wrong type", func(t *testing.T) {
		var cfg ModelConfig
		require.NoError(t, json.Unmarshal([]byte(`{"modelName":"m","apiType":"openai","apiUrl":"http://h/v1","capabilities":{"context":"big"}}`), &cfg))
		err := ValidateDeclaredCapabilities("0x01", cfg.Capabilities)
		require.Error(t, err)
		require.Contains(t, err.Error(), "model 0x01: invalid capabilities: json: cannot unmarshal string")
	})
	t.Run("unknown nested key", func(t *testing.T) {
		var caps DeclaredCapabilities
		require.NoError(t, json.Unmarshal([]byte(`{"modalities":{"input":["text"],"video":true}}`), &caps))
		require.EqualError(t, ValidateDeclaredCapabilities("0x01", &caps), `model 0x01: invalid capabilities: json: unknown field "video"`)
	})
	t.Run("explicit false and every field survive", func(t *testing.T) {
		var caps DeclaredCapabilities
		require.NoError(t, json.Unmarshal([]byte(`{"context":32768,"maxOutput":4096,"modalities":{"input":["text","image"]},"tools":true,"structured":["json_schema"],"webSearch":false,"embeddings":false}`), &caps))
		require.NoError(t, ValidateDeclaredCapabilities("0x01", &caps))
		require.Equal(t, int64(32768), *caps.Context)
		require.Equal(t, int64(4096), *caps.MaxOutput)
		require.Equal(t, []string{"text", "image"}, caps.Modalities.Input)
		require.Nil(t, caps.Modalities.Output)
		require.True(t, *caps.Tools)
		require.Equal(t, []string{"json_schema"}, caps.Structured)
		require.False(t, *caps.WebSearch)
		require.False(t, *caps.Embeddings)
		raw, err := json.Marshal(caps)
		require.NoError(t, err)
		require.JSONEq(t, `{"context":32768,"maxOutput":4096,"modalities":{"input":["text","image"]},"tools":true,"structured":["json_schema"],"webSearch":false,"embeddings":false}`, string(raw))
	})
	t.Run("absent block is nil", func(t *testing.T) {
		var cfg ModelConfig
		require.NoError(t, json.Unmarshal([]byte(`{"modelName":"m","apiType":"openai","apiUrl":"http://h/v1"}`), &cfg))
		require.Nil(t, cfg.Capabilities)
	})
}

func TestDeclaredCapabilitiesEmpty(t *testing.T) {
	var nilCaps *DeclaredCapabilities
	require.True(t, nilCaps.Empty())
	require.True(t, (&DeclaredCapabilities{}).Empty())
	require.True(t, (&DeclaredCapabilities{Modalities: nil}).Empty())
	require.False(t, (&DeclaredCapabilities{Tools: bp(false)}).Empty(), "an explicit false is a declared fact")
	require.False(t, (&DeclaredCapabilities{Modalities: &DeclaredModalities{}}).Empty(), "an empty modalities object is still a set field")
	require.False(t, (&DeclaredCapabilities{Structured: []string{}}).Empty())
}

func TestCapabilityVocabularies(t *testing.T) {
	require.True(t, IsModalityInput("pdf"))
	require.False(t, IsModalityInput("embeddings"))
	require.True(t, IsModalityOutput("embeddings"))
	require.False(t, IsModalityOutput("pdf"))
	require.True(t, IsStructuredKind("choice"))
	require.False(t, IsStructuredKind("xml"))
	require.True(t, ValidCapabilityInt(1))
	require.True(t, ValidCapabilityInt(1<<31-1))
	require.False(t, ValidCapabilityInt(0))
	require.False(t, ValidCapabilityInt(1<<31))
	require.False(t, ValidCapabilityInt(-1))
}

func TestExampleModelsConfigDeclaresCapabilities(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "models-config.json.example"))
	require.NoError(t, err)
	var v2 ModelConfigsV2
	require.NoError(t, json.Unmarshal(raw, &v2))
	var found *DeclaredCapabilities
	for _, m := range v2.Models {
		require.NoError(t, ValidateDeclaredCapabilities(m.ID, m.Capabilities), m.ModelName)
		if m.Capabilities != nil {
			found = m.Capabilities
		}
	}
	require.NotNil(t, found, "the example carries one declared capabilities block")
	require.Equal(t, int64(32768), *found.Context)
	require.Equal(t, []string{"text", "image"}, found.Modalities.Input)
	require.False(t, *found.WebSearch)
}
