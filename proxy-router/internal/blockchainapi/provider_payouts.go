package blockchainapi

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	pr "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	sr "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const secondsInDay = 24 * 60 * 60

var errSessionNotEnded = errors.New("session not ended or does not exist")

// GetProviderClaimableBalance is what a claim for the session would accrue now.
// Reward the provider is already owed as debt is not in it; GetProviderPayouts
// reports that. The session and the provider are read at one block, so a claim
// landing between the two reads cannot show its amount again.
func (s *BlockchainService) GetProviderClaimableBalance(ctx context.Context, sessionID common.Hash) (*big.Int, error) {
	block, err := s.headBlock(ctx)
	if err != nil {
		return nil, err
	}
	session, err := s.sessionRouter.GetSessionAt(ctx, sessionID, block)
	if err != nil {
		return nil, err
	}
	bid, err := s.marketplace.GetBidById(ctx, session.BidId)
	if err != nil {
		return nil, err
	}
	provider, err := s.providerRegistry.GetProviderByIdAt(ctx, bid.Provider, block)
	if err != nil {
		return nil, err
	}
	return providerClaimable(session, bid.PricePerSecond, provider, time.Now().Unix())
}

// providerClaimable mirrors SessionRouter.claimForProvider and the accrual in
// SessionSettlement: the reward up to the earlier of closedAt and endsAt that
// the session has not accrued yet, less the closing day's share until that UTC
// day ends, capped by what the provider's reward limiter still allows.
func providerClaimable(session *sr.ISessionStorageSession, pricePerSecond *big.Int, provider *pr.IProviderStorageProvider, now int64) (*big.Int, error) {
	end := session.EndsAt
	if session.ClosedAt.Sign() != 0 && session.ClosedAt.Cmp(end) < 0 {
		end = session.ClosedAt
	}
	if end.Int64() > now {
		return nil, errSessionNotEnded
	}

	amount := new(big.Int).Sub(end, session.OpenedAt)
	amount.Mul(amount, pricePerSecond)
	amount.Sub(amount, session.ProviderWithdrawnAmount)

	if closedAt := session.ClosedAt.Int64(); closedAt != 0 {
		dayStart := closedAt - closedAt%secondsInDay
		if now < dayStart+secondsInDay {
			heldFrom := max(dayStart, session.OpenedAt.Int64())
			amount.Sub(amount, new(big.Int).Mul(big.NewInt(closedAt-heldFrom), pricePerSecond))
		}
	}

	earned := provider.LimitPeriodEarned
	if now > provider.LimitPeriodEnd.Int64() {
		earned = new(big.Int)
	}
	if limit := new(big.Int).Sub(provider.Stake, earned); amount.Cmp(limit) > 0 {
		amount = limit
	}
	if amount.Sign() < 0 {
		amount.SetInt64(0)
	}
	return amount, nil
}

// GetProviderPayouts reports where the provider's rewards are paid and the
// reward still owed to it. Debt is kept per provider, not per target. Every
// view is read at one block, so a claim landing between two reads cannot show
// in lifetimeClaimed and still in owed.
func (s *BlockchainService) GetProviderPayouts(ctx context.Context, provider common.Address) (*structs.ProviderPayoutsRes, error) {
	block, err := s.headBlock(ctx)
	if err != nil {
		return nil, err
	}
	target, err := s.providerPayouts.GetPayoutTarget(ctx, provider, block)
	if err != nil {
		return nil, err
	}
	owed, err := s.providerPayouts.GetProviderOwed(ctx, provider, block)
	if err != nil {
		return nil, err
	}
	claimed, since, err := s.providerPayouts.GetProviderLifetimeClaimed(ctx, provider, block)
	if err != nil {
		return nil, err
	}
	pending, err := s.providerPayouts.GetPendingProviderClaims(ctx, block)
	if err != nil {
		return nil, err
	}

	return &structs.ProviderPayoutsRes{
		Provider:              provider,
		PayoutTarget:          target,
		Owed:                  &lib.BigInt{Int: *owed},
		LifetimeClaimed:       &lib.BigInt{Int: *claimed},
		LifetimeClaimedSince:  &lib.BigInt{Int: *since},
		PendingProviderClaims: &lib.BigInt{Int: *pending},
	}, nil
}

// SetPayoutTarget sends the provider's future payouts to target; the zero
// address or the provider itself resets it. The contract accepts it only
// from the provider or a delegate holding its payout right, and rejects the
// diamond as a target.
func (s *BlockchainService) SetPayoutTarget(ctx context.Context, provider common.Address, target common.Address) (common.Hash, error) {
	return s.sendWalletTx(ctx, "set_payout_target", func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return s.providerPayouts.SetPayoutTargetTx(opts, provider, target)
	})
}
