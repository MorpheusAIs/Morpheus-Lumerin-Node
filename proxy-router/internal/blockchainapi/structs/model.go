package structs

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Model struct {
	Id        common.Hash
	IpfsCID   common.Hash
	Fee       *big.Int `swaggertype:"integer"`
	Stake     *big.Int `swaggertype:"integer"`
	Owner     common.Address
	Name      string
	Tags      []string
	CreatedAt *big.Int `swaggertype:"integer"`
	IsDeleted bool
	ModelType ModelType // Type of the model (LLM, STT, TTS, EMBEDDING)
}

// ModelType is "LLM" or "STT" or "TTS" or "EMBEDDING"
// ModelType represents the type of model, such as LLM (Large Language Model), STT (Speech-to-Text), TTS (Text-to-Speech), or EMBEDDING.
// It is used to categorize models based on their functionality.
type ModelType string

const (
	ModelTypeLLM       ModelType = "LLM"
	ModelTypeSTT       ModelType = "STT"
	ModelTypeTTS       ModelType = "TTS"
	ModelTypeEMBEDDING ModelType = "EMBEDDING"
	ModelTypeUnknown   ModelType = "UNKNOWN" // Default type for unknown models
)
