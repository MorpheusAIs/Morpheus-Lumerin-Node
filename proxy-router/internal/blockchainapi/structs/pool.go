package structs

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

// DefaultPoolBatchSize is how many queued withdrawals or lock buckets one
// keeper call handles when the request does not say: the contract maximum.
const DefaultPoolBatchSize uint8 = 32

// QueryPoolHot names the hot wallet whose pool is read: the proxy-router
// wallet when omitted.
type QueryPoolHot struct {
	Hot lib.Address `form:"hot" binding:"omitempty" validate:"eth_addr"`
}

// QueryPoolFunders pages the funders of a pool. The contract accepts up to
// 128 a page.
type QueryPoolFunders struct {
	QueryPoolHot
	Offset lib.BigInt `form:"offset,default=0" binding:"omitempty" validate:"number,gte=0"`
	Limit  uint8      `form:"limit,default=10" binding:"gte=1,lte=128"`
}

// QueryPoolWithdrawals pages the withdrawal queue from Cursor, the head of the
// queue when omitted. The contract accepts up to 128 a page.
type QueryPoolWithdrawals struct {
	QueryPoolHot
	Cursor lib.Address `form:"cursor" binding:"omitempty" validate:"eth_addr"`
	Limit  uint8       `form:"limit,default=10" binding:"gte=1,lte=128"`
}

// QueryPoolLocks pages the lock buckets of a pool. The contract accepts up to
// 32 a page.
type QueryPoolLocks struct {
	QueryPoolHot
	Offset lib.BigInt `form:"offset,default=0" binding:"omitempty" validate:"number,gte=0"`
	Limit  uint8      `form:"limit,default=10" binding:"gte=1,lte=32"`
}

type PoolFunderRequest struct {
	Funder common.Address `json:"funder" binding:"required" validate:"eth_addr" example:"0x1234567890abcdef1234567890abcdef12345678"`
}

// ServePoolWithdrawalsRequest and ReleasePoolLocksRequest act on the pool of
// Hot, the proxy-router wallet when omitted. The batch size is a pointer so an
// omitted one (the default) is told apart from an explicit zero (rejected).
type ServePoolWithdrawalsRequest struct {
	Hot        *common.Address `json:"hot" binding:"omitempty" swaggertype:"string" example:"0x1234567890abcdef1234567890abcdef12345678"`
	MaxEntries *uint8          `json:"maxEntries" binding:"omitempty,gte=1,lte=32" example:"32"`
}

type ReleasePoolLocksRequest struct {
	Hot        *common.Address `json:"hot" binding:"omitempty" swaggertype:"string" example:"0x1234567890abcdef1234567890abcdef12345678"`
	MaxBuckets *uint8          `json:"maxBuckets" binding:"omitempty,gte=1,lte=32" example:"32"`
}

func PoolBatchSizeOrDefault(size *uint8) uint8 {
	if size == nil {
		return DefaultPoolBatchSize
	}

	return *size
}

type PoolBalances struct {
	Free         *lib.BigInt `json:"free" swaggertype:"string"`
	InSession    *lib.BigInt `json:"inSession" swaggertype:"string"`
	Locked       *lib.BigInt `json:"locked" swaggertype:"string"`
	PendingTotal *lib.BigInt `json:"pendingTotal" swaggertype:"string"`
}

// PoolParameters are the delegate-staking parameters in effect, with the
// contract defaults filled in for unset ones.
type PoolParameters struct {
	MaxActiveFunders   *lib.BigInt `json:"maxActiveFunders" swaggertype:"string"`
	MinPrincipal       *lib.BigInt `json:"minPrincipal" swaggertype:"string"`
	SettlementGrace    *lib.BigInt `json:"settlementGrace" swaggertype:"string"`
	MaxAutoService     *lib.BigInt `json:"maxAutoService" swaggertype:"string"`
	MaxAutoReleaseDays *lib.BigInt `json:"maxAutoReleaseDays" swaggertype:"string"`
}

// PoolLimits are the hard limits on the parameters and on the page and batch
// sizes of the pool calls.
type PoolLimits struct {
	MaxActiveFunders   *lib.BigInt `json:"maxActiveFunders" swaggertype:"string"`
	MaxAutoService     *lib.BigInt `json:"maxAutoService" swaggertype:"string"`
	MaxAutoReleaseDays *lib.BigInt `json:"maxAutoReleaseDays" swaggertype:"string"`
	MaxSettlementGrace *lib.BigInt `json:"maxSettlementGrace" swaggertype:"string"`
	MaxMinPrincipal    *lib.BigInt `json:"maxMinPrincipal" swaggertype:"string"`
}

// PoolRes is a hot wallet's delegate-staking pool. AvailableStake is what the
// next pool session can draw.
type PoolRes struct {
	Hot            common.Address `json:"hot"`
	Balances       PoolBalances   `json:"balances"`
	AvailableStake *lib.BigInt    `json:"availableStake" swaggertype:"string"`
	Parameters     PoolParameters `json:"parameters"`
	Limits         PoolLimits     `json:"limits"`
}

type PoolFunderAllowance struct {
	CumulativeFundingCap *lib.BigInt `json:"cumulativeFundingCap" swaggertype:"string"`
	Expiry               *lib.BigInt `json:"expiry" swaggertype:"string"`
	Revoked              bool        `json:"revoked"`
	LifetimeFunded       *lib.BigInt `json:"lifetimeFunded" swaggertype:"string"`
	CurrentPrincipal     *lib.BigInt `json:"currentPrincipal" swaggertype:"string"`
	Exposure             *lib.BigInt `json:"exposure" swaggertype:"string"`
	LockedExposure       *lib.BigInt `json:"lockedExposure" swaggertype:"string"`
	Pending              *lib.BigInt `json:"pending" swaggertype:"string"`
	Active               bool        `json:"active"`
	Approved             bool        `json:"approved"`
}

type PoolFunder struct {
	Funder       common.Address      `json:"funder"`
	Allowance    PoolFunderAllowance `json:"allowance"`
	InActiveList bool                `json:"inActiveList"`
}

type PoolFundersRes struct {
	Funders []PoolFunder `json:"funders"`
	Total   *lib.BigInt  `json:"total" swaggertype:"string"`
}

type PoolWithdrawal struct {
	Funder common.Address `json:"funder"`
	Amount *lib.BigInt    `json:"amount" swaggertype:"string"`
}

// PoolWithdrawalsRes is a page of the withdrawal queue in FIFO order.
// NextCursor is the zero address once the tail is reached.
type PoolWithdrawalsRes struct {
	Withdrawals []PoolWithdrawal `json:"withdrawals"`
	NextCursor  common.Address   `json:"nextCursor"`
	Total       *lib.BigInt      `json:"total" swaggertype:"string"`
}

type PoolLockBucket struct {
	ReleaseAt *lib.BigInt `json:"releaseAt" swaggertype:"string"`
	Amount    *lib.BigInt `json:"amount" swaggertype:"string"`
}

type PoolLocksRes struct {
	Locks []PoolLockBucket `json:"locks"`
	Total *lib.BigInt      `json:"total" swaggertype:"string"`
}
