package contractmanager

import (
	"time"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate"
)

type TermsCommon interface {
	ID() string
	BlockchainState() hashrate.BlockchainState
	Seller() string
	Buyer() string
	Validator() string
	StartTime() time.Time
	Duration() time.Duration
	HashrateGHS() float64
}
