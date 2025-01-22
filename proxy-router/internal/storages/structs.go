package storages

import "math/big"

type Session struct {
	Id            string
	UserAddr      string
	ProviderAddr  string
	AgentUsername string
	EndsAt        *big.Int
	ModelID       string

	TPSScaled1000Arr []int
	TTFTMsArr        []int
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
