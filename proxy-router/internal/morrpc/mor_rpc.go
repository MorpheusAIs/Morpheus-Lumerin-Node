package morrpc

import (
	"encoding/json"
	"fmt"
	"time"

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

type RpcResponse struct {
	ID     string                 `json:"id"`
	Result map[string]interface{} `json:"result"`
	Error  RpcError               `json:"error"`
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

type MorRpc struct{}

func NewMorRpc() *MorRpc {
	return &MorRpc{}
}

// Provider Node Communication

func (m *MorRpc) InitiateSessionResponse(providerPubKey string, userAddr string, providerPrivateKeyHex string, requestId string) (*RpcResponse, error) {
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   providerPubKey,
		"user":      userAddr,
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

func (m *MorRpc) SessionPromptResponse(message string, providerPrivateKeyHex string, requestId string) (*RpcResponse, error) {
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

func (m *MorRpc) ResponseError(message string, privateKeyHex string, requestId string) (*RpcResponse, error) {
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

func (m *MorRpc) AuthError(privateKeyHex string, requestId string) (*RpcResponse, error) {
	return m.ResponseError("Failed to authenticate signature", privateKeyHex, requestId)
}

func (m *MorRpc) OutOfCapacityError(privateKeyHex string, requestId string) (*RpcResponse, error) {
	return m.ResponseError("Provider at capacity", privateKeyHex, requestId)
}

func (m *MorRpc) SessionClosedError(privateKeyHex string, requestId string) (*RpcResponse, error) {
	return m.ResponseError("Session is closed", privateKeyHex, requestId)
}

func (m *MorRpc) SpendLimitError(privateKeyHex string, requestId string) (*RpcResponse, error) {
	return m.ResponseError("Over spend limit", privateKeyHex, requestId)
}

// Session Report

func (m *MorRpc) SessionReport(sessionID string, start uint, end uint, prompts uint, tokens uint, reqs []ReqObject, providerPrivateKeyHex string, requestId string) (*RpcResponse, error) {
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

func (m *MorRpc) generateReport(sessionID string, start uint, end uint, prompts uint, tokens uint, reqs []ReqObject) *SessionReport {
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

func (m *MorRpc) InitiateSessionRequest(user string, provider string, userPubKey string, spend float64, userPrivateKeyHex string, requestId string) (*RpcMessage, error) {
	method := "session.request"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"timestamp": timestamp,
		"user":      user,
		"provider":  provider,
		"key":       userPubKey,
		"spend":     fmt.Sprintf("%f", spend),
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

func (m *MorRpc) SessionPromptRequest(sessionID string, prompt string, providerPubKey string, userPrivateKeyHex string, requestId string) (*RpcMessage, error) {
	method := "session.prompt"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   prompt,
		"sessionid": sessionID,
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

func (m *MorRpc) SessionCloseRequest(sessionID string, userPrivateKeyHex string, requestId string) (*RpcMessage, error) {
	method := "session.close"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"sessionid": sessionID,
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

func (m *MorRpc) generateTimestamp() int64 {
	now := time.Now()
	return now.UnixMilli()
}

// https://goethereumbook.org/signature-generate/
func (m *MorRpc) generateSignature(params map[string]interface{}, privateKeyHex string) (string, error) {
	resultStr, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}
	hash := crypto.Keccak256Hash([]byte(resultStr))
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}
	return string(signature), nil
}

// https://goethereumbook.org/signature-verify/
func (m *MorRpc) VerifySignature(params []byte, signature string, publicKeyBytes []byte) bool {
	var jsonParams map[string]interface{}
	err := json.Unmarshal([]byte(params), &jsonParams)
	if err != nil {
		return false
	}
	delete(jsonParams, "signature")

	resultStr, err := json.Marshal(jsonParams)
	if err != nil {
		return false
	}

	hash := crypto.Keccak256Hash([]byte(resultStr))
	signatureNoRecoverID := signature[:len(signature)-1] // remove recovery ID
	return crypto.VerifySignature(publicKeyBytes, hash.Bytes(), []byte(signatureNoRecoverID))
}
