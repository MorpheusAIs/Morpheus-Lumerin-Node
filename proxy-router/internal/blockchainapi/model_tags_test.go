package blockchainapi

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
)

func TestDetectModelTypeDecisionsPrecedence(t *testing.T) {
	cases := []struct {
		name string
		tags []string
		want structs.ModelType
	}{
		{"decision singular wins over LLM", []string{"LLM", "TypeSafe", "Decision"}, structs.ModelTypeDECISIONS},
		{"decisions plural", []string{"decisions"}, structs.ModelTypeDECISIONS},
		{"typesafe alone", []string{"typesafe"}, structs.ModelTypeDECISIONS},
		{"systemone", []string{"SystemOne"}, structs.ModelTypeDECISIONS},
		{"system-one", []string{"system-one"}, structs.ModelTypeDECISIONS},
		{"case insensitive Decision", []string{"llm", "DECISION"}, structs.ModelTypeDECISIONS},
		{"llm alone still LLM", []string{"LLM"}, structs.ModelTypeLLM},
		{"embedding unchanged", []string{"embeddings"}, structs.ModelTypeEMBEDDING},
		{"bare jev is not decisions", []string{"jev", "llm"}, structs.ModelTypeLLM},
		{"unknown", []string{"image"}, structs.ModelTypeUnknown},
		{"video", []string{"video"}, structs.ModelTypeVIDEO},
		{"t2v alias", []string{"T2V"}, structs.ModelTypeVIDEO},
		{"llm+video stays a chat model", []string{"LLM", "Video"}, structs.ModelTypeLLM},
		{"video+llm stays a chat model", []string{"video", "llm"}, structs.ModelTypeLLM},
		{"decisions still first", []string{"video", "decisions"}, structs.ModelTypeDECISIONS},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectModelType(tc.tags)
			if got != tc.want {
				t.Fatalf("DetectModelType(%v)=%q want %q", tc.tags, got, tc.want)
			}
		})
	}
}
