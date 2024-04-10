package morrpc

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

type RpcMessage struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
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

func (m *MorRpc) InitiateSessionResponse(providerPubKey string, userAddr string, providerPrivateKeyHex string) (RpcMessage, error) {
	method := "response.success"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   providerPubKey,
		"timestamp": timestamp,
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return RpcMessage{}, err
	}
	params["signature"] = signature
	return RpcMessage{
		Method: method,
		Params: params,
	}, nil
}

func (m *MorRpc) SessionPromptResponse(message string, providerPrivateKeyHex string) (RpcMessage, error) {
	method := "response.inference"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   message,
		"timestamp": timestamp,
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return RpcMessage{}, err
	}
	params["signature"] = signature
	return RpcMessage{
		Method: method,
		Params: params,
	}, nil
}

func (m *MorRpc) ResponseError(message string, privateKeyHex string) (RpcMessage, error) {
	method := "response.error"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   message,
		"timestamp": timestamp,
	}

	signature, err := m.generateSignature(params, privateKeyHex)
	if err != nil {
		return RpcMessage{}, err
	}
	params["signature"] = signature
	return RpcMessage{
		Method: method,
		Params: params,
	}, nil
}

func (m *MorRpc) AuthError(privateKeyHex string) (RpcMessage, error) {
	return m.ResponseError("Failed to authenticate signature.", privateKeyHex)
}

func (m *MorRpc) OutOfCapacityError(privateKeyHex string) (RpcMessage, error) {
	return m.ResponseError("Provider at capacity", privateKeyHex)
}

func (m *MorRpc) SessionClosedError(privateKeyHex string) (RpcMessage, error) {
	return m.ResponseError("Session is closed.", privateKeyHex)
}

func (m *MorRpc) SpendLimitError(privateKeyHex string) (RpcMessage, error) {
	return m.ResponseError("Over spend limit.", privateKeyHex)
}

// Session Report

func (m *MorRpc) SessionReport(sessionID string, start uint, end uint, prompts uint, tokens uint, reqs []ReqObject, providerPrivateKeyHex string) (RpcMessage, error) {
	method := "session.report"

	report := m.generateReport(sessionID, start, end, prompts, tokens, reqs)
	reportJson, err := json.Marshal(report)
	if err != nil {
		return m.ResponseError("Failed to generate report.", providerPrivateKeyHex)
	}
	reportStr := string(reportJson)

	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   reportStr,
		"timestamp": timestamp,
	}
	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return RpcMessage{}, err
	}
	params["signature"] = signature
	return RpcMessage{
		Method: method,
		Params: params,
	}, nil
}

func (m *MorRpc) generateReport(sessionID string, start uint, end uint, prompts uint, tokens uint, reqs []ReqObject) SessionReport {
	return SessionReport{
		SessionID: sessionID,
		Start:     start,
		End:       end,
		Prompts:   prompts,
		Tokens:    tokens,
		Reqs:      reqs,
	}
}

// User Node Communication

func (m *MorRpc) InitiateSessionRequest(user string, provider string, userPubKey string, spend float64, userPrivateKeyHex string) (RpcMessage, error) {
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
		return RpcMessage{}, err
	}
	params["signature"] = signature
	return RpcMessage{
		Method: method,
		Params: params,
	}, nil
}

func (m *MorRpc) SessionPromptRequest(sessionID string, prompt string, providerPubKey string, userPrivateKeyHex string) (RpcMessage, error) {
	method := "session.prompt"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"message":   prompt,
		"sessionid": sessionID,
		"timestamp": timestamp,
	}
	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return RpcMessage{}, err
	}
	params["signature"] = signature
	return RpcMessage{
		Method: method,
		Params: params,
	}, nil
}

func (m *MorRpc) SessionCloseRequest(sessionID string, userPrivateKeyHex string) (RpcMessage, error) {
	method := "session.close"
	timestamp := m.generateTimestamp()
	params := map[string]interface{}{
		"sessionid": sessionID,
		"timestamp": timestamp,
	}
	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return RpcMessage{}, err
	}
	params["signature"] = signature
	return RpcMessage{
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
	resultStr := ""

	keys := make([]string, 0)
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Concatenate the parameters in the order of the sorted keys
	key_values := make([]string, 0)
	for _, k := range keys {
		key_values = append(key_values, fmt.Sprintf("%s=%v", k, params[k]))
	}
	resultStr = strings.Join(key_values, "&")

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
func (m *MorRpc) VerifySignature(params map[string]interface{}, signature string, publicKeyBytes []byte) bool {
	result := ""

	keys := make([]string, 0)
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Concatenate the parameters in the order of the sorted keys
	for _, k := range keys {
		result += fmt.Sprintf("%s:%v", k, params[k])
	}

	hash := crypto.Keccak256Hash([]byte(result))
	signatureNoRecoverID := signature[:len(signature)-1] // remove recovery ID
	return crypto.VerifySignature(publicKeyBytes, hash.Bytes(), []byte(signatureNoRecoverID))
}
