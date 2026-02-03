package registries

import (
	"context"
	"fmt"
	"math/big"
	"time"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/morpheustoken"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type MorToken struct {
	// config
	morTokenAddr common.Address

	// state
	nonce  uint64
	morABI *abi.ABI

	// deps
	mor    *morpheustoken.MorpheusToken
	client i.ContractBackend
	log    lib.ILogger
}

func NewMorToken(morTokenAddr common.Address, client i.ContractBackend, log lib.ILogger) *MorToken {
	mor, err := morpheustoken.NewMorpheusToken(morTokenAddr, client)
	if err != nil {
		panic("invalid mor ABI")
	}
	return &MorToken{
		mor:          mor,
		morTokenAddr: morTokenAddr,
		client:       client,
		log:          log,
	}
}

func (g *MorToken) GetBalance(ctx context.Context, account common.Address) (*big.Int, error) {
	return g.mor.BalanceOf(&bind.CallOpts{Context: ctx}, account)
}

func (g *MorToken) GetAllowance(ctx context.Context, owner common.Address, spender common.Address) (*big.Int, error) {
	return g.mor.Allowance(&bind.CallOpts{Context: ctx}, owner, spender)
}

func (g *MorToken) Approve(ctx *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, *types.Receipt, error) {
	tx, err := g.mor.Approve(ctx, spender, amount)
	if err != nil {
		return nil, nil, lib.TryConvertGethError(err)
	}
	// Wait for the transaction receipt with timeout to prevent infinite polling
	receipt, err := lib.WaitMinedWithTimeout(ctx.Context, g.client, tx, lib.DefaultTxMineTimeout)
	if err != nil {
		return nil, nil, err
	}

	if receipt.Status != 1 {
		return nil, nil, fmt.Errorf("transaction failed with status %d", receipt.Status)
	}

	err = g.waitForConfirmations(ctx.Context, receipt, 1)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to wait for confirmations %s", err)
	}

	return tx, receipt, nil
}

// ApproveTx builds an approval transaction without sending it.
// Use with opts.NoSend = true for escalation support.
func (g *MorToken) ApproveTx(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	tx, err := g.mor.Approve(opts, spender, amount)
	if err != nil {
		return nil, lib.TryConvertGethError(err)
	}
	return tx, nil
}

func (g *MorToken) GetTotalSupply(ctx context.Context) (*big.Int, error) {
	supply, err := g.mor.TotalSupply(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, lib.TryConvertGethError(err)
	}
	return supply, nil
}

func (g *MorToken) Transfer(ctx *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, *types.Receipt, error) {
	tx, err := g.mor.Transfer(ctx, to, value)
	if err != nil {
		return nil, nil, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt with timeout
	receipt, err := lib.WaitMinedWithTimeout(ctx.Context, g.client, tx, lib.DefaultTxMineTimeout)
	if err != nil {
		return nil, nil, err
	}
	return tx, receipt, nil
}

func (g *MorToken) waitForConfirmations(ctx context.Context, receipt *types.Receipt, confirmations uint64) error {
	targetBlock := receipt.BlockNumber.Uint64() + confirmations
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			header, err := g.client.HeaderByNumber(ctx, nil)
			if err != nil {
				return err
			}
			if header.Number.Uint64() >= targetBlock {
				return nil
			}
		}
	}
}
