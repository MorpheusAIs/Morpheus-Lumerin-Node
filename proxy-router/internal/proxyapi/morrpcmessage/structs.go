package morrpcmesssage

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type SessionReq struct {
	Signature lib.HexString  `json:"signature" validate:"required,hexadecimal"`
	User      common.Address `json:"user"      validate:"required,eth_address"`
	Key       lib.HexString  `json:"key"       validate:"required,hexadecimal"`
	Spend     lib.BigInt     `json:"spend"     validate:"required,number,gt=0"`
	Timestamp string         `json:"timestamp" validate:"required,timestamp"`
	BidID     common.Hash    `json:"bidid"     validate:"required,hex32"`
}

type SessionPromptReq struct {
	Signature lib.HexString `json:"signature" validate:"required,hexadecimal"`
	SessionID common.Hash   `json:"sessionid" validate:"required,hexadecimal"`
	Message   string        `json:"message"   validate:"required"`
	Timestamp string        `json:"timestamp" validate:"required,timestamp"`
}
