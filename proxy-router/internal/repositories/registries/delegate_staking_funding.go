package registries

import (
	"context"
	"errors"
	"math/big"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/delegatestakingfunding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DelegateStakingFunding reads a hot wallet's delegate-staking pool and runs
// its upkeep: funder approvals, the withdrawal queue and lock release.
//
// The views that take a block read at that block, or at the latest one when
// it is nil, so a caller combining several of them can read them all at once.
type DelegateStakingFunding struct {
	// deps
	delegateStakingFunding *delegatestakingfunding.DelegateStakingFunding
	log                    lib.ILogger
}

func NewDelegateStakingFunding(diamondAddr common.Address, client i.ContractBackend, log lib.ILogger) *DelegateStakingFunding {
	dsf, err := delegatestakingfunding.NewDelegateStakingFunding(diamondAddr, client)
	if err != nil {
		panic("invalid delegate staking funding ABI")
	}
	return &DelegateStakingFunding{
		delegateStakingFunding: dsf,
		log:                    log,
	}
}

func (g *DelegateStakingFunding) GetPoolSession(ctx context.Context, sessionID common.Hash) (*delegatestakingfunding.IDelegateStakingStoragePoolSession, error) {
	session, err := g.delegateStakingFunding.GetPoolSession(&bind.CallOpts{Context: ctx}, sessionID)
	if err != nil {
		return nil, stage1Error(err)
	}
	return &session, nil
}

// IsPoolSession reports whether the session was opened from a pool. A diamond
// without the Stage 1 cut has no pool sessions.
func (g *DelegateStakingFunding) IsPoolSession(ctx context.Context, sessionID common.Hash) (bool, error) {
	session, err := g.GetPoolSession(ctx, sessionID)
	if errors.Is(err, ErrStage1NotDeployed) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return session.Exists, nil
}

// GetAvailablePoolStake is the stake the next pool session of hot can draw,
// after pending withdrawals and the funders' caps.
func (g *DelegateStakingFunding) GetAvailablePoolStake(ctx context.Context, hot common.Address, block *big.Int) (*big.Int, error) {
	stake, err := g.delegateStakingFunding.GetAvailablePoolStake(&bind.CallOpts{Context: ctx, BlockNumber: block}, hot)
	if err != nil {
		return nil, stage1Error(err)
	}
	return stake, nil
}

func (g *DelegateStakingFunding) GetPoolBalances(ctx context.Context, hot common.Address, block *big.Int) (*delegatestakingfunding.IDelegateStakingStoragePoolBalances, error) {
	balances, err := g.delegateStakingFunding.GetPoolBalances(&bind.CallOpts{Context: ctx, BlockNumber: block}, hot)
	if err != nil {
		return nil, stage1Error(err)
	}
	return &balances, nil
}

func (g *DelegateStakingFunding) GetPoolFunders(ctx context.Context, hot common.Address, offset *big.Int, limit uint8) ([]delegatestakingfunding.IDelegateStakingStoragePoolFunder, *big.Int, error) {
	page, err := g.delegateStakingFunding.GetPoolFunders(&bind.CallOpts{Context: ctx}, hot, offset, big.NewInt(int64(limit)))
	if err != nil {
		return nil, nil, stage1Error(err)
	}
	return page.Funders, page.Total, nil
}

// GetPoolWithdrawalQueue pages the withdrawal queue in FIFO order starting at
// cursor; the zero address is the head, and nextCursor is zero at the tail.
func (g *DelegateStakingFunding) GetPoolWithdrawalQueue(ctx context.Context, hot common.Address, cursor common.Address, limit uint8) (withdrawals []delegatestakingfunding.IDelegateStakingStoragePoolWithdrawal, nextCursor common.Address, total *big.Int, err error) {
	page, err := g.delegateStakingFunding.GetPoolWithdrawalQueue(&bind.CallOpts{Context: ctx}, hot, cursor, big.NewInt(int64(limit)))
	if err != nil {
		return nil, common.Address{}, nil, stage1Error(err)
	}
	return page.Withdrawals, page.NextCursor, page.Total, nil
}

// GetPoolLockBuckets pages the unreleased buckets of stake that pool sessions
// returned with a day lock.
func (g *DelegateStakingFunding) GetPoolLockBuckets(ctx context.Context, hot common.Address, offset *big.Int, limit uint8) ([]delegatestakingfunding.IDelegateStakingStoragePoolLockBucket, *big.Int, error) {
	page, err := g.delegateStakingFunding.GetPoolLockBuckets(&bind.CallOpts{Context: ctx}, hot, offset, big.NewInt(int64(limit)))
	if err != nil {
		return nil, nil, stage1Error(err)
	}
	return page.Buckets, page.Total, nil
}

func (g *DelegateStakingFunding) GetDelegateStakingParameters(ctx context.Context, block *big.Int) (*delegatestakingfunding.IDelegateStakingStorageDelegateStakingParameters, error) {
	params, err := g.delegateStakingFunding.GetDelegateStakingParameters(&bind.CallOpts{Context: ctx, BlockNumber: block})
	if err != nil {
		return nil, stage1Error(err)
	}
	return &params, nil
}

func (g *DelegateStakingFunding) GetDelegateStakingLimits(ctx context.Context, block *big.Int) (*delegatestakingfunding.IDelegateStakingStorageDelegateStakingLimits, error) {
	limits, err := g.delegateStakingFunding.GetDelegateStakingLimits(&bind.CallOpts{Context: ctx, BlockNumber: block})
	if err != nil {
		return nil, stage1Error(err)
	}
	return &limits, nil
}

// ApprovePoolFunderTx builds an approvePoolFunder transaction without sending
// it. Use with opts.NoSend = true for escalation support. It lets funder fund
// the pool of hot; the sender must be hot or its delegate-staking delegate.
func (g *DelegateStakingFunding) ApprovePoolFunderTx(opts *bind.TransactOpts, hot common.Address, funder common.Address) (*types.Transaction, error) {
	tx, err := g.delegateStakingFunding.ApprovePoolFunder(opts, hot, funder)
	if err != nil {
		return nil, stage1Error(err)
	}
	return tx, nil
}

// RevokePoolFunderApprovalTx builds a revokePoolFunderApproval transaction
// without sending it. Use with opts.NoSend = true for escalation support. It
// stops draws from and top-ups by funder at once and never blocks the
// funder's withdrawal.
func (g *DelegateStakingFunding) RevokePoolFunderApprovalTx(opts *bind.TransactOpts, hot common.Address, funder common.Address) (*types.Transaction, error) {
	tx, err := g.delegateStakingFunding.RevokePoolFunderApproval(opts, hot, funder)
	if err != nil {
		return nil, stage1Error(err)
	}
	return tx, nil
}

// ServePoolWithdrawalsTx builds a servePoolWithdrawals transaction without
// sending it. Use with opts.NoSend = true for escalation support. It pays up
// to maxEntries queued withdrawals of hot's pool from its free stake. Anyone
// may send it; the money goes to the funders.
func (g *DelegateStakingFunding) ServePoolWithdrawalsTx(opts *bind.TransactOpts, hot common.Address, maxEntries uint8) (*types.Transaction, error) {
	tx, err := g.delegateStakingFunding.ServePoolWithdrawals(opts, hot, big.NewInt(int64(maxEntries)))
	if err != nil {
		return nil, stage1Error(err)
	}
	return tx, nil
}

// ReleasePoolLocksTx builds a releasePoolLocks transaction without sending it.
// Use with opts.NoSend = true for escalation support. It frees up to
// maxBuckets matured lock buckets of hot's pool, then serves queued
// withdrawals. Anyone may send it.
func (g *DelegateStakingFunding) ReleasePoolLocksTx(opts *bind.TransactOpts, hot common.Address, maxBuckets uint8) (*types.Transaction, error) {
	tx, err := g.delegateStakingFunding.ReleasePoolLocks(opts, hot, big.NewInt(int64(maxBuckets)))
	if err != nil {
		return nil, stage1Error(err)
	}
	return tx, nil
}
