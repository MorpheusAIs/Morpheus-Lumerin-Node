package sessionrepo

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type sessionModel struct {
	id           common.Hash
	userAddr     common.Address
	providerAddr common.Address
	endsAt       *big.Int
	modelID      common.Hash

	tpsScaled1000Arr []int
	ttftMsArr        []int
	failoverEnabled  bool
	directPayment    bool
}

func (s *sessionModel) ID() common.Hash {
	return s.id
}

func (s *sessionModel) UserAddr() common.Address {
	return s.userAddr
}

func (s *sessionModel) ProviderAddr() common.Address {
	return s.providerAddr
}

func (s *sessionModel) EndsAt() *big.Int {
	// copy big.Int so that the original value is not modified
	return new(big.Int).Set(s.endsAt)
}

func (s *sessionModel) GetStats() (tpsScaled1000Arr []int, ttftMsArr []int) {
	return s.tpsScaled1000Arr, s.ttftMsArr
}

func (s *sessionModel) ModelID() common.Hash {
	return s.modelID
}

func (s *sessionModel) FailoverEnabled() bool {
	return s.failoverEnabled
}

func (s *sessionModel) DirectPayment() bool {
	return s.directPayment
}

func (s *sessionModel) AddStats(tpsScaled1000 int, ttftMs int) {
	s.tpsScaled1000Arr = append(s.tpsScaled1000Arr, tpsScaled1000)
	s.ttftMsArr = append(s.ttftMsArr, ttftMs)
}

func (s *sessionModel) SetFailoverEnabled(enabled bool) {
	s.failoverEnabled = enabled
}
