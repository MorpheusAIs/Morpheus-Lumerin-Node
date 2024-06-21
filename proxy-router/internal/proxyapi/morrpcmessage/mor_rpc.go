package morrpcmesssage

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type RpcError struct {
	Message string `json:"message"`
	Data    string `json:"data"`
	Code    int    `json:"code"`
}

type RpcMessage struct {
	ID     string                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

type RPCMessageV2 struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type RpcResponse struct {
	ID     string   `json:"id"`
	Result any      `json:"result"`
	Error  RpcError `json:"error"`
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

var approvalAbi = []lib.AbiParameter{
	{Type: "bytes32"},
	{Type: "uint128"},
}

type MORRPCMessage struct{}

func NewMorRpc() *MORRPCMessage {
	return &MORRPCMessage{}
}

// Provider Node Communication

func (m *MORRPCMessage) InitiateSessionResponse(providerPubKey lib.HexString, userAddr common.Address, bidID common.Hash, providerPrivateKeyHex lib.HexString, requestID string) (*RpcResponse, error) {
	timestamp := m.generateTimestamp()

	approval, err := lib.EncodeAbiParameters(approvalAbi, []interface{}{bidID, big.NewInt(timestamp)})
	if err != nil {
		return &RpcResponse{}, err
	}
	approvalSig, err := lib.SignEthMessageV2(approval, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}

	params := constants.InitiateSessionResponse{
		Message:     providerPubKey,
		Approval:    approval,
		ApprovalSig: approvalSig,
		User:        userAddr,
		Timestamp:   fmt.Sprintf("%d", timestamp),
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}

	params.Signature = signature

	return &RpcResponse{
		ID:     requestID,
		Result: params,
	}, nil
}

func (m *MORRPCMessage) SessionPromptResponse(message string, providerPrivateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   message,
		"timestamp": timestamp,
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}
	params["signature"] = signature
	return &RpcResponse{
		ID:     requestId,
		Result: params,
	}, nil
}

func (m *MORRPCMessage) ResponseError(message string, privateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   message,
		"timestamp": timestamp,
	}

	signature, err := m.generateSignature(params, privateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}
	params["signature"] = signature
	return &RpcResponse{
		ID:    requestId,
		Error: RpcError{Message: message, Data: "", Code: 400},
	}, nil
}

func (m *MORRPCMessage) AuthError(privateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	return m.ResponseError("Failed to authenticate signature", privateKeyHex, requestId)
}

func (m *MORRPCMessage) OutOfCapacityError(privateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	return m.ResponseError("Provider at capacity", privateKeyHex, requestId)
}

func (m *MORRPCMessage) SessionClosedError(privateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	return m.ResponseError("Session is closed", privateKeyHex, requestId)
}

func (m *MORRPCMessage) SpendLimitError(privateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	return m.ResponseError("Over spend limit", privateKeyHex, requestId)
}

// Session Report

func (m *MORRPCMessage) SessionReport(sessionID string, start uint, end uint, prompts uint, tokens uint, reqs []ReqObject, providerPrivateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	report := m.generateReport(sessionID, start, end, prompts, tokens, reqs)
	reportJson, err := json.Marshal(report)
	if err != nil {
		return m.ResponseError("Failed to generate report", providerPrivateKeyHex, requestId)
	}
	reportStr := string(reportJson)

	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   reportStr,
		"timestamp": timestamp,
	}
	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}
	params["signature"] = signature
	return &RpcResponse{
		ID:     requestId,
		Result: params,
	}, nil
}

func (m *MORRPCMessage) generateReport(sessionID string, start uint, end uint, prompts uint, tokens uint, reqs []ReqObject) *SessionReport {
	return &SessionReport{
		SessionID: sessionID,
		Start:     start,
		End:       end,
		Prompts:   prompts,
		Tokens:    tokens,
		Reqs:      reqs,
	}
}

// User Node Communication

func (m *MORRPCMessage) InitiateSessionRequest(user common.Address, provider common.Address, spend *big.Int, bidID common.Hash, userPrivateKeyHex lib.HexString, requestId string) (*RpcMessage, error) {
	method := "session.request"
	timestamp := m.generateTimestamp()
	pbKey, err := lib.PubKeyFromPrivate(userPrivateKeyHex)
	if err != nil {
		return &RpcMessage{}, err
	}
	params := map[string]interface{}{
		"timestamp": timestamp,
		"user":      user,
		"provider":  provider,
		"key":       pbKey,
		"spend":     spend.String(),
		"bidid":     bidID,
	}

	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return &RpcMessage{}, err
	}
	params["signature"] = signature
	return &RpcMessage{
		ID:     requestId,
		Method: method,
		Params: params,
	}, nil
}

func (m *MORRPCMessage) SessionPromptRequest(sessionID common.Hash, prompt interface{}, providerPubKey lib.HexString, userPrivateKeyHex lib.HexString, requestId string) (*RpcMessage, error) {
	method := "session.prompt"
	timestamp := m.generateTimestamp()

	promptStr, err := json.Marshal(prompt)
	if err != nil {
		return &RpcMessage{}, err
	}
	params := map[string]interface{}{
		"message":   string(promptStr),
		"sessionid": sessionID.Hex(),
		"timestamp": timestamp,
	}
	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return &RpcMessage{}, err
	}
	params["signature"] = signature
	return &RpcMessage{
		ID:     requestId,
		Method: method,
		Params: params,
	}, nil
}

func (m *MORRPCMessage) SessionCloseRequest(sessionID common.Hash, userPrivateKeyHex lib.HexString, requestId string) (*RpcMessage, error) {
	method := "session.close"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"sessionid": sessionID.Hex(),
		"timestamp": timestamp,
	}
	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return &RpcMessage{}, err
	}
	params["signature"] = signature
	return &RpcMessage{
		ID:     requestId,
		Method: method,
		Params: params,
	}, nil
}

func (m *MORRPCMessage) generateTimestamp() int64 {
	now := time.Now()
	return now.UnixMilli()
}

// https://goethereumbook.org/signature-generate/
func (m *MORRPCMessage) generateSignature(params any, privateKeyHex lib.HexString) (lib.HexString, error) {
	result, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.ToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}
	hash := crypto.Keccak256Hash(result)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (m *MORRPCMessage) VerifySignature(params any, signature lib.HexString, publicKey lib.HexString, sourceLog lib.ILogger) bool {
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		sourceLog.Error("Error marshalling params", err)
		return false
	}

	return lib.VerifySignature(paramsBytes, signature, publicKey)
}
