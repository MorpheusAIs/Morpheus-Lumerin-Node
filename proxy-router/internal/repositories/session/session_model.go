package sessionrepo

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type SessionModel struct {
	id            common.Hash
	userAddr      common.Address
	providerAddr  common.Address
	agentUsername string
	endsAt        *big.Int
	modelID       common.Hash

	tpsScaled1000Arr []int
	ttftMsArr        []int
	failoverEnabled  bool
	directPayment    bool
}

func (s *SessionModel) ID() common.Hash {
	return s.id
}

func (s *SessionModel) UserAddr() common.Address {
	return s.userAddr
}

func (s *SessionModel) ProviderAddr() common.Address {
	return s.providerAddr
}

func (s *SessionModel) AgentUsername() string {
	return s.agentUsername
}

func (s *SessionModel) EndsAt() *big.Int {
	// copy big.Int so that the original value is not modified
	return new(big.Int).Set(s.endsAt)
}

func (s *SessionModel) GetStats() (tpsScaled1000Arr []int, ttftMsArr []int) {
	return s.tpsScaled1000Arr, s.ttftMsArr
}

func (s *SessionModel) ModelID() common.Hash {
	return s.modelID
}

func (s *SessionModel) FailoverEnabled() bool {
	return s.failoverEnabled
}

func (s *SessionModel) DirectPayment() bool {
	return s.directPayment
}

func (s *SessionModel) AddStats(tpsScaled1000 int, ttftMs int) {
	s.tpsScaled1000Arr = append(s.tpsScaled1000Arr, tpsScaled1000)
	s.ttftMsArr = append(s.ttftMsArr, ttftMs)
}

func (s *SessionModel) SetFailoverEnabled(enabled bool) {
	s.failoverEnabled = enabled
}

func (s *SessionModel) SetAgentUsername(username string) {
	s.agentUsername = username
}
