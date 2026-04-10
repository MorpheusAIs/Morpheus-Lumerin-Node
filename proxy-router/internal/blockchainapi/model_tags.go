package blockchainapi

import (
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
)

// IsTeeModel returns true if the model has any TEE tag ("tee" or "tee-gpu").
// Used for Phase 1 P-node attestation on the consumer side.
func IsTeeModel(tags []string) bool {
	for _, raw := range tags {
		tag := strings.ToLower(raw)
		if tag == "tee" || tag == "tee-gpu" {
			return true
		}
	}
	return false
}

// IsTeeGPUModel returns true if the model has the "tee-gpu" tag.
// Used for Phase 2 backend LLM attestation on the provider side.
func IsTeeGPUModel(tags []string) bool {
	for _, raw := range tags {
		if strings.ToLower(raw) == "tee-gpu" {
			return true
		}
	}
	return false
}

func DetectModelType(tags []string) structs.ModelType {
	for _, raw := range tags {
		tag := strings.ToLower(raw)
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
