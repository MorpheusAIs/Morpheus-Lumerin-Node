package blockchainapi

import (
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
)

// IsTeeModel returns true if the model has the "tee" tag.
// Enables both P-node attestation (consumer) and backend LLM attestation (provider).
func IsTeeModel(tags []string) bool {
	for _, raw := range tags {
		if strings.ToLower(raw) == "tee" {
			return true
		}
	}
	return false
}

func isDecisionTag(tag string) bool {
	switch tag {
	case "decision", "decisions", "typesafe", "systemone", "system-one":
		return true
	default:
		return false
	}
}

// DetectModelType classifies a model from its tags.
// Decision-class tags (decision/decisions/typesafe/systemone/system-one) take
// precedence over LLM when both are present — required for catalog rows like
// jev-1.13-ts that carry both LLM and Decision tags.
func DetectModelType(tags []string) structs.ModelType {
	normalized := make([]string, 0, len(tags))
	for _, raw := range tags {
		normalized = append(normalized, strings.ToLower(raw))
	}

	for _, tag := range normalized {
		if isDecisionTag(tag) {
			return structs.ModelTypeDECISIONS
		}
	}

	for _, tag := range normalized {
		switch tag {
		case "stt", "transcribe", "s2t", "speech", "speech-to-text", "speech2text":
			return structs.ModelTypeSTT
		case "tts", "text-to-speech", "text2speech", "t2s":
			return structs.ModelTypeTTS
		case "embedding", "embeddings":
			return structs.ModelTypeEMBEDDING
		case "llm", "textgeneration", "text2text", "text-to-text", "t2t":
			return structs.ModelTypeLLM
		}
	}

	return structs.ModelTypeUnknown
}
