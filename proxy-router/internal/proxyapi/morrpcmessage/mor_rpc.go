package morrpcmesssage

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type MORRPCMessage struct{}

func NewMorRpc() *MORRPCMessage {
	return &MORRPCMessage{}
}

// RESPONSES

func (m *MORRPCMessage) PongResponce(requestId string, providerPrKey lib.HexString, nonce lib.HexString) (*RpcResponse, error) {
	params := PongRes{
		Nonce: nonce,
	}
	signature, err := m.generateSignature(params, providerPrKey)
	if err != nil {
		return &RpcResponse{}, err
	}

	params.Signature = signature

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return &RpcResponse{}, err
	}

	paramsJSON := json.RawMessage(paramsBytes)

	return &RpcResponse{
		ID:     requestId,
		Result: &paramsJSON,
	}, nil
}

func (m *MORRPCMessage) InitiateSessionResponse(providerPubKey lib.HexString, userAddr common.Address, bidID common.Hash, providerPrivateKeyHex lib.HexString, requestID string, chainID *big.Int) (*RpcResponse, error) {
	timestamp := m.generateTimestamp()

	approval, err := lib.EncodeAbiParameters(approvalAbi, []interface{}{bidID, chainID, userAddr, big.NewInt(int64(timestamp))})
	if err != nil {
		return &RpcResponse{}, err
	}
	approvalSig, err := lib.SignEthMessageV2(approval, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}

	params := SessionRes{
		PubKey:      providerPubKey,
		Approval:    approval,
		ApprovalSig: approvalSig,
		User:        userAddr,
		Timestamp:   timestamp,
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}

	params.Signature = signature

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return &RpcResponse{}, err
	}

	paramsJSON := json.RawMessage(paramsBytes)

	return &RpcResponse{
		ID:     requestID,
		Result: &paramsJSON,
	}, nil
}

func (m *MORRPCMessage) SessionReportResponse(tps uint32, ttfp uint32, sessionID common.Hash, providerPrivateKeyHex lib.HexString, requestID string, chainID *big.Int) (*RpcResponse, error) {
	timestamp := m.generateTimestamp()

	report, err := lib.EncodeAbiParameters(sessionReportAbi, []interface{}{sessionID, chainID, big.NewInt(int64(timestamp)), tps, ttfp})
	if err != nil {
		return &RpcResponse{}, err
	}

	signedReport, err := lib.SignEthMessageV2(report, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}

	params := SessionReportRes{
		Timestamp:    timestamp,
		Message:      report,
		SignedReport: signedReport,
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}

	params.Signature = signature

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return &RpcResponse{}, err
	}

	paramsJSON := json.RawMessage(paramsBytes)

	return &RpcResponse{
		ID:     requestID,
		Result: &paramsJSON,
	}, nil
}

func (m *MORRPCMessage) SessionPromptResponse(message string, providerPrivateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	params := SessionPromptRes{
		Message:   message,
		Timestamp: m.generateTimestamp(),
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}
	params.Signature = signature

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return &RpcResponse{}, err
	}

	paramsJSON := json.RawMessage(paramsBytes)

	return &RpcResponse{
		ID:     requestId,
		Result: &paramsJSON,
	}, nil
}

func (m *MORRPCMessage) SessionReport(sessionID string, start uint, end uint, prompts uint, tokens uint, reqs []ReqObject, providerPrivateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	params := ReportRes{
		Message: &SessionReport{
			SessionID: sessionID,
			Start:     start,
			End:       end,
			Prompts:   prompts,
			Tokens:    tokens,
			Reqs:      reqs,
		},
		Timestamp: m.generateTimestamp(),
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}
	params.Signature = signature

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return &RpcResponse{}, err
	}

	paramsJSON := json.RawMessage(paramsBytes)

	return &RpcResponse{
		ID:     requestId,
		Result: &paramsJSON,
	}, nil
}

func (m *MORRPCMessage) CallAgentToolResponse(message string, providerPrivateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	params := CallAgentToolRes{
		Message:   message,
		Timestamp: m.generateTimestamp(),
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}
	params.Signature = signature

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return &RpcResponse{}, err
	}

	paramsJSON := json.RawMessage(paramsBytes)

	return &RpcResponse{
		ID:     requestId,
		Result: &paramsJSON,
	}, nil
}

func (m *MORRPCMessage) GetAgentToolsResponse(message string, providerPrivateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	params := GetAgentToolsRes{
		Message:   message,
		Timestamp: m.generateTimestamp(),
	}

	signature, err := m.generateSignature(params, providerPrivateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}
	params.Signature = signature

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return &RpcResponse{}, err
	}

	paramsJSON := json.RawMessage(paramsBytes)

	return &RpcResponse{
		ID:     requestId,
		Result: &paramsJSON,
	}, nil
}

// ERRORS

func (m *MORRPCMessage) ResponseError(message string, privateKeyHex lib.HexString, requestId string) (*RpcResponse, error) {
	params2 := RpcError{
		Message: message,
		Code:    400,
		Data: RPCErrorData{
			Timestamp: m.generateTimestamp(),
		},
	}

	signature, err := m.generateSignature(params2, privateKeyHex)
	if err != nil {
		return &RpcResponse{}, err
	}
	params2.Data.Signature = &signature

	return &RpcResponse{
		ID:    requestId,
		Error: &params2,
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

// REQUESTS

func (m *MORRPCMessage) PingRequest(requestId string, userPrivateKeyHex lib.HexString, nonce lib.HexString) (*RPCMessage, error) {
	params := PingReq{
		Nonce: nonce,
	}
	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return &RPCMessage{}, err
	}
	params.Signature = signature

	serializedParams, err := json.Marshal(params)
	if err != nil {
		return &RPCMessage{}, err
	}

	return &RPCMessage{
		ID:     requestId,
		Method: "network.ping",
		Params: serializedParams,
	}, nil
}

func (m *MORRPCMessage) CallAgentToolRequest(sessionID common.Hash, toolName string, input map[string]interface{}, userPrivateKeyHex lib.HexString, requestId string) (*RPCMessage, error) {
	method := "agent.call_tool"

	message, err := json.Marshal(input)
	if err != nil {
		return &RPCMessage{}, err
	}

	params := CallAgentToolReq{
		SessionID: sessionID,
		Message:   string(message),
		ToolName:  toolName,
		Timestamp: m.generateTimestamp(),
	}

	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return &RPCMessage{}, err
	}
	params.Signature = signature

	serializedParams, err := json.Marshal(params)
	if err != nil {
		return &RPCMessage{}, err
	}

	return &RPCMessage{
		ID:     requestId,
		Method: method,
		Params: serializedParams,
	}, nil
}

func (m *MORRPCMessage) CallGetAgentToolsRequest(sessionID common.Hash, userPrivateKeyHex lib.HexString, requestId string) (*RPCMessage, error) {
	method := "agent.get_tools"

	params := GetAgentToolsReq{
		SessionID: sessionID,
		Timestamp: m.generateTimestamp(),
	}

	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return &RPCMessage{}, err
	}
	params.Signature = signature

	serializedParams, err := json.Marshal(params)
	if err != nil {
		return &RPCMessage{}, err
	}

	return &RPCMessage{
		ID:     requestId,
		Method: method,
		Params: serializedParams,
	}, nil
}

func (m *MORRPCMessage) InitiateSessionRequest(user common.Address, provider common.Address, spend *big.Int, bidID common.Hash, userPrivateKeyHex lib.HexString, requestId string) (*RPCMessage, error) {
	method := "session.request"
	pbKey, err := lib.PubKeyFromPrivate(userPrivateKeyHex)
	if err != nil {
		return &RPCMessage{}, err
	}

	params2 := SessionReq{
		Timestamp: uint64(time.Now().UnixMilli()),
		User:      user,
		Key:       pbKey,
		Spend:     lib.BigInt{Int: *spend},
		BidID:     bidID,
	}

	signature, err := m.generateSignature(params2, userPrivateKeyHex)
	if err != nil {
		return &RPCMessage{}, err
	}
	params2.Signature = signature
	serializedParams, err := json.Marshal(params2)
	if err != nil {
		return &RPCMessage{}, err
	}
	return &RPCMessage{
		ID:     requestId,
		Method: method,
		Params: serializedParams,
	}, nil
}

func (m *MORRPCMessage) SessionPromptRequest(sessionID common.Hash, prompt interface{}, providerPubKey lib.HexString, userPrivateKeyHex lib.HexString, requestId string) (*RPCMessage, error) {
	method := "session.prompt"
	promptStr, err := json.Marshal(prompt)
	if err != nil {
		return &RPCMessage{}, err
	}
	params2 := SessionPromptReq{
		Message:   string(promptStr),
		SessionID: sessionID,
		Timestamp: uint64(time.Now().UnixMilli()),
	}
	signature, err := m.generateSignature(params2, userPrivateKeyHex)
	if err != nil {
		return &RPCMessage{}, err
	}
	params2.Signature = signature

	serializedParams, err := json.Marshal(params2)
	if err != nil {
		return &RPCMessage{}, err
	}
	return &RPCMessage{
		ID:     requestId,
		Method: method,
		Params: serializedParams,
	}, nil
}

func (m *MORRPCMessage) SessionReportRequest(sessionID common.Hash, userPrivateKeyHex lib.HexString, requestId string) (*RPCMessage, error) {
	method := "session.report"

	params := SessionReportReq{
		Timestamp: m.generateTimestamp(),
		Message:   sessionID.Hex(),
	}

	signature, err := m.generateSignature(params, userPrivateKeyHex)
	if err != nil {
		return &RPCMessage{}, err
	}
	params.Signature = signature
	serializedParams, err := json.Marshal(params)
	if err != nil {
		return &RPCMessage{}, err
	}
	return &RPCMessage{
		ID:     requestId,
		Method: method,
		Params: serializedParams,
	}, nil
}

// UTILS

func (m *MORRPCMessage) VerifySignature(params any, signature lib.HexString, publicKey lib.HexString, sourceLog lib.ILogger) bool {
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		sourceLog.Error("Error marshalling params", err)
		return false
	}

	return lib.VerifySignature(paramsBytes, signature, publicKey)
}

func (m *MORRPCMessage) VerifySignatureAddr(params any, signature lib.HexString, addr common.Address, sourceLog lib.ILogger) bool {
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		sourceLog.Error("Error marshalling params", err)
		return false
	}

	return lib.VerifySignatureAddr(paramsBytes, signature, addr)
}

func (m *MORRPCMessage) generateTimestamp() uint64 {
	now := time.Now()
	return uint64(now.UnixMilli())
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
