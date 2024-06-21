package constants

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type InitiateSessionResponse struct {
	Message     lib.HexString  `json:"message" validate:"required,hexadecimal"`
	Approval    lib.HexString  `json:"approval" validate:"required,hexadecimal"`
	ApprovalSig lib.HexString  `json:"approvalSig" validate:"required,hexadecimal"`
	User        common.Address `json:"user" validate:"required,eth_addr"`
	Timestamp   string         `json:"timestamp" validate:"required,timestamp"`
	Signature   lib.HexString  `json:"signature,omitempty" validate:"required,hexadecimal"`
}
