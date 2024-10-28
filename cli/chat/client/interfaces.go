package client

import (
	"math/big"

	"github.com/sashabaranov/go-openai"
)

type SessionRequest struct {
	ModelId         string   `json:"modelId" validate:"required"`
	SessionDuration *big.Int `json:"sessionDuration" validate:"required"`
}

type SessionStakeRequest struct {
	Approval    string `json:"approval" validate:"required"`
	ApprovalSig string `json:"approvalSig" validate:"required"`
	Stake       uint64 `json:"stake" validate:"required,number"`
}

type Session struct {
	SessionId string `json:"sessionId"`
}

type CloseSessionRequest struct {
	SessionId string `json:"id" validate:"required"`
}

type SessionListItem struct {
	Bid             string `json:"BidId"`
	Sesssion        string `json:"Id"`
	ModelORAgent    string `json:"ModelAgentId"`
	PricePerSecond  uint64 `json:"PricePerSecond"`
	Provider        string `json:"Provider"`
	User            string `json:"User"`
	ClosedAt        uint64 `json:"ClosedAt"`
	CloseoutReceipt string `json:"CloseoutReceipt"`
	CloseoutType    uint   `json:"CloseoutType"`
	EndsAt          uint64 `json:"EndsAt"`
}

type WalletRequest struct {
	PrivateKey string `json:"privateKey" validate:"required"`
}

type ChatCompletionMessage = openai.ChatCompletionMessage
