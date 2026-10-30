package morrpcmesssage

import (
	"encoding/json"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
)

// Content encoding markers and capability names for consumer -> provider
// content encryption. Provider -> consumer content has always been ECIES to
// the consumer's key; these mirror that scheme in the other direction.
const (
	// EncECIES marks a content field as hex(ECIES-secp256k1(plaintext)) under
	// the receiving node's key: the same primitive and encoding that
	// SessionPromptRes.Message has always used toward the consumer.
	EncECIES = "ecies"

	// CapPromptEncryptionECIES is advertised by a provider that accepts
	// EncECIES content on session.prompt, session.prompt.stream.chunk/end and
	// agent.call_tool.
	CapPromptEncryptionECIES = "prompt-enc:ecies"
)

type SessionReq struct {
	Signature lib.HexString  `json:"signature,omitempty" validate:"required,hexadecimal"`
	User      common.Address `json:"user"      validate:"required,eth_addr"`
	Key       lib.HexString  `json:"key"       validate:"required,hexadecimal"`
	Spend     lib.BigInt     `json:"spend"     validate:"required"`
	Timestamp uint64         `json:"timestamp" validate:"required,number"`
	BidID     common.Hash    `json:"bidid"     validate:"required,hex32"`
	// Enc declares the content encoding this consumer will use for the
	// session's prompts (EncECIES). Omitted by consumers that send unencrypted.
	Enc string `json:"enc,omitempty" validate:"omitempty,oneof=ecies"`
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
	Enc       string        `json:"enc,omitempty" validate:"omitempty,oneof=ecies"`
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
	// Enc is EncECIES when Message is ciphertext under the provider's key.
	// Empty means legacy plaintext JSON. The field is inside the signed
	// payload, so it cannot be stripped or added in transit.
	Enc       string `json:"enc,omitempty" validate:"omitempty,oneof=ecies"`
	Timestamp uint64 `json:"timestamp" validate:"required,number"`
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
	Timestamp uint64        `json:"timestamp" validate:"required,number"`
	Signature lib.HexString `json:"signature" validate:"required,hexadecimal"`
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

// CapabilitiesReq asks a provider which optional MORRPC features it supports.
// Same shape as PingReq. Legacy providers answer with an "unknown method"
// error, which consumers treat as "no optional capabilities".
type CapabilitiesReq struct {
	Nonce     lib.HexString `json:"nonce"     validate:"required,hexadecimal"`
	Signature lib.HexString `json:"signature" validate:"required,hexadecimal"`
}

// CapabilitiesRes is signed by the provider key over nonce + capabilities so a
// path attacker cannot forge or alter the advertised feature set.
type CapabilitiesRes struct {
	Nonce        lib.HexString `json:"nonce"        validate:"required,hexadecimal"`
	Capabilities []string      `json:"capabilities" validate:"required"`
	Timestamp    uint64        `json:"timestamp"    validate:"required,number"`
	Signature    lib.HexString `json:"signature"    validate:"required,hexadecimal"`
}

type PongRes struct {
	Nonce   lib.HexString `json:"nonce"     validate:"required,hexadecimal"`
	Version string        `json:"version,omitempty"   validate:"omitempty"`
	// Models is set after signing (like Signature) and excluded from the
	// signed payload: consumers with an older PongRes struct drop the unknown
	// field on unmarshal, so signature verification keeps working for them.
	Models    []system.ModelHealthReport `json:"models,omitempty" validate:"omitempty"`
	Signature lib.HexString              `json:"signature" validate:"required,hexadecimal"`
}

// Audio streaming message types
type SessionPromptStreamStartReq struct {
	SessionID   common.Hash   `json:"sessionid" validate:"required,hex32"`
	StreamID    string        `json:"streamid" validate:"required"`
	TotalChunks uint32        `json:"totalchunks" validate:"required,number"`
	FileSize    uint64        `json:"filesize" validate:"required,number"`
	ContentType string        `json:"contenttype" validate:"required"`
	Signature   lib.HexString `json:"signature" validate:"required,hexadecimal"`
	Timestamp   uint64        `json:"timestamp" validate:"required,number"`
}

type SessionPromptStreamStartRes struct {
	StreamID  string        `json:"streamid" validate:"required"`
	Status    string        `json:"status" validate:"required"`
	Signature lib.HexString `json:"signature" validate:"required,hexadecimal"`
	Timestamp uint64        `json:"timestamp" validate:"required,number"`
}

type SessionPromptStreamChunkReq struct {
	SessionID  common.Hash `json:"sessionid" validate:"required,hex32"`
	StreamID   string      `json:"streamid" validate:"required"`
	ChunkIndex uint32      `json:"chunkindex" validate:"number"`
	// ChunkData is base64(raw bytes) for legacy plaintext, or
	// hex(ECIES(raw bytes)) when Enc is EncECIES.
	ChunkData   string        `json:"chunkdata" validate:"required"`
	Enc         string        `json:"enc,omitempty" validate:"omitempty,oneof=ecies"`
	IsLastChunk bool          `json:"islastchunk"`
	Signature   lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp   uint64        `json:"timestamp" validate:"required,number"`
}

type SessionPromptStreamChunkRes struct {
	StreamID   string        `json:"streamid" validate:"required"`
	ChunkIndex uint32        `json:"chunkindex" validate:"required,number"`
	Status     string        `json:"status" validate:"required"`
	Signature  lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp  uint64        `json:"timestamp" validate:"required,number"`
}

type SessionPromptStreamEndReq struct {
	SessionID         common.Hash   `json:"sessionid" validate:"required,hex32"`
	StreamID          string        `json:"streamid" validate:"required"`
	AudioRequestParam string        `json:"audiorequestparam" validate:"required"`
	Enc               string        `json:"enc,omitempty" validate:"omitempty,oneof=ecies"`
	Signature         lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp         uint64        `json:"timestamp" validate:"required,number"`
}

type SessionPromptStreamEndRes struct {
	StreamID  string        `json:"streamid" validate:"required"`
	Message   string        `json:"message" validate:"required"`
	Signature lib.HexString `json:"signature,omitempty" validate:"required,hexadecimal"`
	Timestamp uint64        `json:"timestamp" validate:"required,number"`
}
