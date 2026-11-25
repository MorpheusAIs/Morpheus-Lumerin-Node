package registries

import (
	"context"
	"math/big"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerpayouts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ProviderPayouts covers where provider rewards are paid and what is still
// owed. The stake and emission views this facet also serves are read through
// SessionRouter.
//
// Its views read at the given block, or at the latest one when it is nil, so
// a caller combining several of them can read them all at once.
type ProviderPayouts struct {
	// deps
	providerPayouts *providerpayouts.ProviderPayouts
	log             lib.ILogger
}

func NewProviderPayouts(diamondAddr common.Address, client i.ContractBackend, log lib.ILogger) *ProviderPayouts {
	pp, err := providerpayouts.NewProviderPayouts(diamondAddr, client)
	if err != nil {
		panic("invalid provider payouts ABI")
	}
	return &ProviderPayouts{
		providerPayouts: pp,
		log:             log,
	}
}

// GetPayoutTarget is the address rewards are paid to: the provider itself
// unless it set another target.
func (g *ProviderPayouts) GetPayoutTarget(ctx context.Context, provider common.Address, block *big.Int) (common.Address, error) {
	target, err := g.providerPayouts.GetPayoutTarget(&bind.CallOpts{Context: ctx, BlockNumber: block}, provider)
	if err != nil {
		return common.Address{}, stage1Error(err)
	}
	return target, nil
}

// GetProviderOwed is reward the funding account could not pay when it accrued.
// A later claim pays it once the funding account covers it.
func (g *ProviderPayouts) GetProviderOwed(ctx context.Context, provider common.Address, block *big.Int) (*big.Int, error) {
	owed, err := g.providerPayouts.GetProviderOwed(&bind.CallOpts{Context: ctx, BlockNumber: block}, provider)
	if err != nil {
		return nil, stage1Error(err)
	}
	return owed, nil
}

// GetPendingProviderClaims is the reward owed to all providers together.
func (g *ProviderPayouts) GetPendingProviderClaims(ctx context.Context, block *big.Int) (*big.Int, error) {
	pending, err := g.providerPayouts.GetPendingProviderClaims(&bind.CallOpts{Context: ctx, BlockNumber: block})
	if err != nil {
		return nil, stage1Error(err)
	}
	return pending, nil
}

// GetProviderLifetimeClaimed is what the provider was paid since the counter
// started at the since timestamp.
func (g *ProviderPayouts) GetProviderLifetimeClaimed(ctx context.Context, provider common.Address, block *big.Int) (claimed *big.Int, since *big.Int, err error) {
	claimed, since, err = g.providerPayouts.GetProviderLifetimeClaimed(&bind.CallOpts{Context: ctx, BlockNumber: block}, provider)
	if err != nil {
		return nil, nil, stage1Error(err)
	}
	return claimed, since, nil
}

// SetPayoutTargetTx builds a setPayoutTarget transaction without sending it.
// Use with opts.NoSend = true for escalation support. It redirects the
// provider's rewards; a zero target or the provider itself resets it. The
// contract rejects the diamond address and any sender other than the provider
// or its payout delegate.
func (g *ProviderPayouts) SetPayoutTargetTx(opts *bind.TransactOpts, provider common.Address, target common.Address) (*types.Transaction, error) {
	tx, err := g.providerPayouts.SetPayoutTarget(opts, provider, target)
	if err != nil {
		return nil, stage1Error(err)
	}
	return tx, nil
}
