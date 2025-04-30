package morrpcmesssage

import (
	"encoding/json"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type SessionReq struct {
	Signature lib.HexString  `json:"signature,omitempty" validate:"required,hexadecimal"`
	User      common.Address `json:"user"      validate:"required,eth_addr"`
	Key       lib.HexString  `json:"key"       validate:"required,hexadecimal"`
	Spend     lib.BigInt     `json:"spend"     validate:"required"`
	Timestamp uint64         `json:"timestamp" validate:"required,number"`
	BidID     common.Hash    `json:"bidid"     validate:"required,hex32"`
}

type SessionRes struct {
	PubKey      lib.HexString  `json:"message" validate:"required,hexadecimal" swaggertype:"string"`
	Approval    lib.HexString  `json:"approval" validate:"required,hexadecimal" swaggertype:"string"`
	ApprovalSig lib.HexString  `json:"approvalSig" validate:"required,hexadecimal" swaggertype:"string"`
	User        common.Address `json:"user" validate:"required,eth_addr"`
	Timestamp   uint64         `json:"timestamp" validate:"required,timestamp"`
	Signature   lib.HexString  `json:"signature,omitempty" validate:"required,hexadecimal" swaggertype:"string"`
}

type CallAgentToolReq struct {
	SessionID common.Hash   `json:"sessionid" validate:"required,hex32"`
	ToolName  string        `json:"toolname" validate:"required"`
	Message   string        `json:"message" validate:"required"`
	Signature lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp uint64        `json:"timestamp" validate:"required,number"`
}

type CallAgentToolRes struct {
	Message   string        `json:"message" validate:"required"`
	Signature lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp uint64        `json:"timestamp" validate:"required,number"`
}

type GetAgentToolsReq struct {
	SessionID common.Hash   `json:"sessionid" validate:"required,hex32"`
	Signature lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp uint64        `json:"timestamp" validate:"required,number"`
}

type GetAgentToolsRes struct {
	Message   string        `json:"message" validate:"required"`
	Signature lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp uint64        `json:"timestamp" validate:"required,number"`
}

type SessionPromptReq struct {
	Signature lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	SessionID common.Hash   `json:"sessionid" validate:"required,hex32"`
	Message   string        `json:"message"   validate:"required"`
	Timestamp uint64        `json:"timestamp" validate:"required,number"`
}

type SessionReportReq struct {
	Signature lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Message   string        `json:"message"           validate:"required,hexadecimal"`
	Timestamp uint64        `json:"timestamp"           validate:"required,number"`
}

type SessionPromptRes struct {
	Message   string        `json:"message"             validate:"required"`
	Signature lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp uint64        `json:"timestamp"           validate:"required,number"`
}

type SessionReportRes struct {
	Message      lib.HexString `json:"message"             validate:"required"`
	SignedReport lib.HexString `json:"signedReport"        validate:"required"`
	Signature    lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp    uint64        `json:"timestamp"           validate:"required,number"`
}

type ReportRes struct {
	Signature lib.HexString  `json:"signature,omitempty" validate:"required,hexadecimal"`
	Message   *SessionReport `json:"message"             validate:"required"`
	Timestamp uint64         `json:"timestamp"           validate:"required,number"`
}

type RpcError struct {
	Message string       `json:"message"`
	Code    int          `json:"code"`
	Data    RPCErrorData `json:"data"`
}

type RPCErrorData struct {
	Timestamp uint64         `json:"timestamp" validate:"required,number"`
	Signature *lib.HexString `json:"signature" validate:"required,hexadecimal"`
}

type RPCMessage struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type RpcResponse struct {
	ID     string           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  *RpcError        `json:"error,omitempty"`
}

// SessionReport represents the detailed session report
type SessionReport struct {
	SessionID string      `json:"sessionid"`
	Start     uint        `json:"start"`
	End       uint        `json:"end"`
	Prompts   uint        `json:"prompts"`
	Tokens    uint        `json:"tokens"`
	Reqs      []ReqObject `json:"reqs"`
}

// ReqObject represents a request object within a session report
type ReqObject struct {
	Req  uint `json:"req"`
	Res  uint `json:"res"`
	Toks uint `json:"toks"`
}

type PingReq struct {
	Nonce     lib.HexString `json:"nonce"     validate:"required,hexadecimal"`
	Signature lib.HexString `json:"signature" validate:"required,hexadecimal"`
}

type PongRes struct {
	Nonce     lib.HexString `json:"nonce"     validate:"required,hexadecimal"`
	Signature lib.HexString `json:"signature" validate:"required,hexadecimal"`
}
