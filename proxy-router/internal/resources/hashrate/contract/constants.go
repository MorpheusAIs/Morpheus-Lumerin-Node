package contract

import "errors"

var (
	ErrContractClosed = errors.New("contract closed")
)

const (
	ResourceTypeHashrate        = "hashrate"
	ResourceEstimateHashrateGHS = "hashrate_ghs"
)
