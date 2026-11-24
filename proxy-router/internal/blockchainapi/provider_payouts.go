package blockchainapi

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

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
