package storages

import (
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type Session struct {
	Id            string
	UserAddr      string
	ProviderAddr  string
	AgentUsername string
	EndsAt        *big.Int
	ModelID       string

	TPSScaled1000Arr []int
	TTFTMsArr        []int
	InputTokens      int
	OutputTokens     int
	FailoverEnabled  bool
	DirectPayment    bool
}

type User struct {
	Addr   string
	PubKey string
	Url    string
}

type PromptActivity struct {
	SessionID string
	StartTime int64
	EndTime   int64
}

type AgentUser struct {
	Username    string
	Password    string
	Perms       []string
	Allowances  map[string]lib.BigInt
	IsConfirmed bool
}

type AllowanceRequest struct {
	Username  string
	Token     string
	Allowance lib.BigInt
}
