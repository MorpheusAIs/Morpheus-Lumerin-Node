package blockchainapi

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	ErrNotSettleable = errors.New("session cannot be settled")

	// ErrPoolSelfFunder refuses the node wallet as a funder of its own pool.
	// The contract treats a hot wallet as always approved for its own pool and
	// draws its escrow without looking at the approval, so approving or
	// revoking it would send a transaction that changes nothing.
	ErrPoolSelfFunder = errors.New("the node wallet funds its own pool without an approval, so approving or revoking it changes nothing; " +
		"to stop drawing its own escrow, the node wallet sends revokeStakingAllowance for its pool, which this API does not send")
)

// SettlePreconditionError is returned by SettleExpiredSession when
// settleExpiredSession would revert for the session as it is now. It is typed
// so the HTTP controller can answer 409 rather than 500.
type SettlePreconditionError struct {
	Reason string
}

func (e *SettlePreconditionError) Error() string {
	return fmt.Sprintf("%s: %s", ErrNotSettleable, e.Reason)
}

func (e *SettlePreconditionError) Unwrap() error {
	return ErrNotSettleable
}

// poolCapacity checks the stakes that one pool open tries against the pool of
// hot.
type poolCapacity struct {
	hot common.Address
	// waited is set once a shortfall has been read again with the open loop's
	// backoff, so the later bids of the same open do not wait again.
	waited bool
}

// newPoolCapacity starts the capacity checks of a pool open by hot. It reads
// the pool once, so a diamond without the Stage 1 cut fails the open before
// any bid is rated or any provider is pinged.
func (s *BlockchainService) newPoolCapacity(ctx context.Context, hot common.Address) (*poolCapacity, error) {
	if _, err := s.delegateStakingFunding.GetAvailablePoolStake(ctx, hot, nil); err != nil {
		return nil, fmt.Errorf("failed to get available pool stake: %w", err)
	}
	return &poolCapacity{hot: hot}, nil
}

// checkPoolCapacity fails a stake the pool cannot draw now, which
// openSessionFromPool would revert on. A pool open often follows a close or a
// top-up that the RPC node answering the read has not seen yet, so the first
// shortfall of an open is read again with the open loop's backoff before it
// counts. Later bids of the same open read the pool once, as the wait has
// already passed.
func (s *BlockchainService) checkPoolCapacity(ctx context.Context, pool *poolCapacity, stake *big.Int) error {
	reads := sessionTxMaxRetries
	if pool.waited {
		reads = 1
	}

	var available *big.Int
	for attempt := 0; attempt < reads; attempt++ {
		if attempt > 0 {
			pool.waited = true
			backoff := time.Duration(attempt*3) * time.Second
			s.requestLog(ctx).Warnf("the pool of %s can stake %s, less than the %s the session needs (read %d/%d), reading it again in %s", pool.hot.Hex(), available, stake, attempt, reads, backoff)
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during pool capacity check: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		var err error
		available, err = s.delegateStakingFunding.GetAvailablePoolStake(ctx, pool.hot, nil)
		if err != nil {
			return fmt.Errorf("failed to get available pool stake: %w", err)
		}
		if available.Cmp(stake) >= 0 {
			return nil
		}
	}
	return fmt.Errorf("%w: the session needs %s, the pool of %s can stake %s", ErrPoolCapacity, stake, pool.hot.Hex(), available)
}

// poolHot is the hot wallet a pool call is about: hot, or the node wallet
// when hot is the zero address.
func (s *BlockchainService) poolHot(ctx context.Context, hot common.Address) (common.Address, error) {
	if hot != (common.Address{}) {
		return hot, nil
	}
	addr, err := s.GetMyAddress(ctx)
	if err != nil {
		return common.Address{}, lib.WrapError(ErrMyAddress, err)
	}
	return addr, nil
}

// headBlock is the latest block number, for reading several views at one
// block.
func (s *BlockchainService) headBlock(ctx context.Context) (*big.Int, error) {
	head, err := s.ethClient.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get the latest block: %w", err)
	}
	return new(big.Int).SetUint64(head), nil
}

// GetPool reports the balances of the pool of hot, the stake its next session
// can draw, and the delegate-staking parameters and limits. Every view is read
// at one block, so a session opened or a withdrawal served between two reads
// cannot show in one figure and not in another.
func (s *BlockchainService) GetPool(ctx context.Context, hot common.Address) (*structs.PoolRes, error) {
	hot, err := s.poolHot(ctx, hot)
	if err != nil {
		return nil, err
	}

	block, err := s.headBlock(ctx)
	if err != nil {
		return nil, err
	}
	balances, err := s.delegateStakingFunding.GetPoolBalances(ctx, hot, block)
	if err != nil {
		return nil, err
	}
	available, err := s.delegateStakingFunding.GetAvailablePoolStake(ctx, hot, block)
	if err != nil {
		return nil, err
	}
	params, err := s.delegateStakingFunding.GetDelegateStakingParameters(ctx, block)
	if err != nil {
		return nil, err
	}
	limits, err := s.delegateStakingFunding.GetDelegateStakingLimits(ctx, block)
	if err != nil {
		return nil, err
	}

	return &structs.PoolRes{
		Hot: hot,
		Balances: structs.PoolBalances{
			Free:         &lib.BigInt{Int: *balances.Free},
			InSession:    &lib.BigInt{Int: *balances.InSession},
			Locked:       &lib.BigInt{Int: *balances.Locked},
			PendingTotal: &lib.BigInt{Int: *balances.PendingTotal},
		},
		AvailableStake: &lib.BigInt{Int: *available},
		Parameters: structs.PoolParameters{
			MaxActiveFunders:   &lib.BigInt{Int: *params.MaxActiveFunders},
			MinPrincipal:       &lib.BigInt{Int: *params.MinPrincipal},
			SettlementGrace:    &lib.BigInt{Int: *params.SettlementGrace},
			MaxAutoService:     &lib.BigInt{Int: *params.MaxAutoService},
			MaxAutoReleaseDays: &lib.BigInt{Int: *params.MaxAutoReleaseDays},
		},
		Limits: structs.PoolLimits{
			MaxActiveFunders:   &lib.BigInt{Int: *limits.MaxActiveFunders},
			MaxAutoService:     &lib.BigInt{Int: *limits.MaxAutoService},
			MaxAutoReleaseDays: &lib.BigInt{Int: *limits.MaxAutoReleaseDays},
			MaxSettlementGrace: &lib.BigInt{Int: *limits.MaxSettlementGrace},
			MaxMinPrincipal:    &lib.BigInt{Int: *limits.MaxMinPrincipal},
		},
	}, nil
}

// GetPoolFunders pages every funder that has ever funded the pool, active or
// not. The contract accepts a limit of up to 128.
func (s *BlockchainService) GetPoolFunders(ctx context.Context, hot common.Address, offset *big.Int, limit uint8) (*structs.PoolFundersRes, error) {
	hot, err := s.poolHot(ctx, hot)
	if err != nil {
		return nil, err
	}

	funders, total, err := s.delegateStakingFunding.GetPoolFunders(ctx, hot, offset, limit)
	if err != nil {
		return nil, err
	}

	res := &structs.PoolFundersRes{
		Funders: make([]structs.PoolFunder, len(funders)),
		Total:   &lib.BigInt{Int: *total},
	}
	for i, funder := range funders {
		allowance := funder.Allowance
		res.Funders[i] = structs.PoolFunder{
			Funder: funder.Funder,
			Allowance: structs.PoolFunderAllowance{
				CumulativeFundingCap: &lib.BigInt{Int: *allowance.CumulativeFundingCap},
				Expiry:               &lib.BigInt{Int: *allowance.Expiry},
				Revoked:              allowance.Revoked,
				LifetimeFunded:       &lib.BigInt{Int: *allowance.LifetimeFunded},
				CurrentPrincipal:     &lib.BigInt{Int: *allowance.CurrentPrincipal},
				Exposure:             &lib.BigInt{Int: *allowance.Exposure},
				LockedExposure:       &lib.BigInt{Int: *allowance.LockedExposure},
				Pending:              &lib.BigInt{Int: *allowance.Pending},
				Active:               allowance.Active,
				Approved:             allowance.Approved,
			},
			InActiveList: funder.InActiveList,
		}
	}
	return res, nil
}

// GetPoolWithdrawals pages the withdrawal queue from cursor, the zero address
// being the head. The contract accepts a limit of up to 128.
func (s *BlockchainService) GetPoolWithdrawals(ctx context.Context, hot common.Address, cursor common.Address, limit uint8) (*structs.PoolWithdrawalsRes, error) {
	hot, err := s.poolHot(ctx, hot)
	if err != nil {
		return nil, err
	}

	withdrawals, nextCursor, total, err := s.delegateStakingFunding.GetPoolWithdrawalQueue(ctx, hot, cursor, limit)
	if err != nil {
		return nil, err
	}

	res := &structs.PoolWithdrawalsRes{
		Withdrawals: make([]structs.PoolWithdrawal, len(withdrawals)),
		NextCursor:  nextCursor,
		Total:       &lib.BigInt{Int: *total},
	}
	for i, withdrawal := range withdrawals {
		res.Withdrawals[i] = structs.PoolWithdrawal{
			Funder: withdrawal.Funder,
			Amount: &lib.BigInt{Int: *withdrawal.Amount},
		}
	}
	return res, nil
}

// GetPoolLocks pages the unreleased day-locked stake of the pool. The
// contract accepts a limit of up to 32.
func (s *BlockchainService) GetPoolLocks(ctx context.Context, hot common.Address, offset *big.Int, limit uint8) (*structs.PoolLocksRes, error) {
	hot, err := s.poolHot(ctx, hot)
	if err != nil {
		return nil, err
	}

	buckets, total, err := s.delegateStakingFunding.GetPoolLockBuckets(ctx, hot, offset, limit)
	if err != nil {
		return nil, err
	}

	res := &structs.PoolLocksRes{
		Locks: make([]structs.PoolLockBucket, len(buckets)),
		Total: &lib.BigInt{Int: *total},
	}
	for i, bucket := range buckets {
		res.Locks[i] = structs.PoolLockBucket{
			ReleaseAt: &lib.BigInt{Int: *bucket.ReleaseAt},
			Amount:    &lib.BigInt{Int: *bucket.Amount},
		}
	}
	return res, nil
}

// ApprovePoolFunder lets funder fund the pool of the node wallet.
func (s *BlockchainService) ApprovePoolFunder(ctx context.Context, funder common.Address) (common.Hash, error) {
	if err := s.checkPoolFunder(ctx, funder); err != nil {
		return common.Hash{}, err
	}
	return s.sendWalletTx(ctx, "approve_pool_funder", func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return s.delegateStakingFunding.ApprovePoolFunderTx(opts, opts.From, funder)
	})
}

// RevokePoolFunder stops draws from and top-ups by funder in the pool of the
// node wallet at once. It never blocks the funder's withdrawal.
func (s *BlockchainService) RevokePoolFunder(ctx context.Context, funder common.Address) (common.Hash, error) {
	if err := s.checkPoolFunder(ctx, funder); err != nil {
		return common.Hash{}, err
	}
	return s.sendWalletTx(ctx, "revoke_pool_funder", func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return s.delegateStakingFunding.RevokePoolFunderApprovalTx(opts, opts.From, funder)
	})
}

// checkPoolFunder refuses the node wallet as a funder of its own pool, see
// ErrPoolSelfFunder.
func (s *BlockchainService) checkPoolFunder(ctx context.Context, funder common.Address) error {
	addr, err := s.GetMyAddress(ctx)
	if err != nil {
		return lib.WrapError(ErrMyAddress, err)
	}
	if funder == addr {
		return ErrPoolSelfFunder
	}
	return nil
}

// ServePoolWithdrawals pays up to maxEntries (at most 32) queued withdrawals
// of the pool of hot. Anyone may send it; the money goes to the funders.
func (s *BlockchainService) ServePoolWithdrawals(ctx context.Context, hot common.Address, maxEntries uint8) (common.Hash, error) {
	hot, err := s.poolHot(ctx, hot)
	if err != nil {
		return common.Hash{}, err
	}
	return s.sendWalletTx(ctx, "serve_pool_withdrawals", func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return s.delegateStakingFunding.ServePoolWithdrawalsTx(opts, hot, maxEntries)
	})
}

// ReleasePoolLocks frees up to maxBuckets (at most 32) matured lock buckets
// of the pool of hot, then pays queued withdrawals with what it freed, up to
// the maxAutoService parameter. Anyone may send it.
func (s *BlockchainService) ReleasePoolLocks(ctx context.Context, hot common.Address, maxBuckets uint8) (common.Hash, error) {
	hot, err := s.poolHot(ctx, hot)
	if err != nil {
		return common.Hash{}, err
	}
	return s.sendWalletTx(ctx, "release_pool_locks", func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return s.delegateStakingFunding.ReleasePoolLocksTx(opts, hot, maxBuckets)
	})
}

// SettleExpiredSession settles a pool session nobody closed: the stake returns
// to the pool and the provider reward becomes debt. Anyone may send it once
// the settlement grace after endsAt has passed on chain. The preconditions
// are checked first so a session the contract would reject fails with a
// SettlePreconditionError instead of a revert.
func (s *BlockchainService) SettleExpiredSession(ctx context.Context, sessionID common.Hash) (common.Hash, error) {
	pool, err := s.delegateStakingFunding.GetPoolSession(ctx, sessionID)
	if err != nil {
		return common.Hash{}, err
	}
	if !pool.Exists {
		return common.Hash{}, &SettlePreconditionError{Reason: "not a pool session"}
	}

	session, err := s.sessionRouter.GetSession(ctx, sessionID)
	if err != nil {
		return common.Hash{}, err
	}
	params, err := s.delegateStakingFunding.GetDelegateStakingParameters(ctx, nil)
	if err != nil {
		return common.Hash{}, err
	}
	// Chain time rather than the local clock: the contract compares
	// block.timestamp, and a local chain can be far from wall time.
	head, err := s.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return common.Hash{}, err
	}
	now := new(big.Int).SetUint64(head.Time)
	if err := checkSettleable(pool.Settled, session.ClosedAt, session.EndsAt, params.SettlementGrace, now); err != nil {
		return common.Hash{}, err
	}

	return s.sendWalletTx(ctx, "settle", func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return s.delegateStaking.SettleExpiredSessionTx(opts, sessionID)
	})
}

// checkSettleable repeats the checks settleExpiredSession makes on a pool
// session: it is not closed, and chain time has reached endsAt + grace.
func checkSettleable(settled bool, closedAt, endsAt, grace, now *big.Int) error {
	if settled {
		return &SettlePreconditionError{Reason: "already settled"}
	}
	if closedAt.Sign() != 0 {
		return &SettlePreconditionError{Reason: "already closed"}
	}
	settleableAt := new(big.Int).Add(endsAt, grace)
	if now.Cmp(settleableAt) < 0 {
		return &SettlePreconditionError{Reason: fmt.Sprintf(
			"settleable from %s, chain time is %s",
			time.Unix(settleableAt.Int64(), 0).UTC().Format(time.RFC3339),
			time.Unix(now.Int64(), 0).UTC().Format(time.RFC3339),
		)}
	}
	return nil
}
