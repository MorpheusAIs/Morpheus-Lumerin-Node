package blockchainapi

import (
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
)

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
