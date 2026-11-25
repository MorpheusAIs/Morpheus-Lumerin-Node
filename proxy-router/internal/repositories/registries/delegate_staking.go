package registries

import (
	"math/big"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/delegatestaking"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// DelegateStaking opens and closes sessions whose stake is drawn from a hot
// wallet's delegate-staking pool instead of the wallet's own MOR.
type DelegateStaking struct {
	// deps
	delegateStaking *delegatestaking.DelegateStaking
	log             lib.ILogger
}

func NewDelegateStaking(diamondAddr common.Address, client i.ContractBackend, log lib.ILogger) *DelegateStaking {
	ds, err := delegatestaking.NewDelegateStaking(diamondAddr, client)
	if err != nil {
		panic("invalid delegate staking ABI")
	}
	return &DelegateStaking{
		delegateStaking: ds,
		log:             log,
	}
}

// OpenSessionFromPoolTx builds an openSessionFromPool transaction without sending it.
// Use with opts.NoSend = true for escalation support. The receipt carries the
// same SessionOpened event as openSession, so SessionRouter.ParseOpenSessionReceipt
// reads it.
func (g *DelegateStaking) OpenSessionFromPoolTx(opts *bind.TransactOpts, hot common.Address, stake *big.Int, approval []byte, approvalSig []byte) (*types.Transaction, error) {
	tx, err := g.delegateStaking.OpenSessionFromPool(opts, hot, stake, approval, approvalSig)
	if err != nil {
		return nil, stage1Error(err)
	}
	return tx, nil
}

// ClosePoolSessionTx builds a closePoolSession transaction without sending it.
// Use with opts.NoSend = true for escalation support. SessionRouter.closeSession
// rejects pool sessions.
func (g *DelegateStaking) ClosePoolSessionTx(opts *bind.TransactOpts, report []byte, signedReport []byte) (*types.Transaction, error) {
	tx, err := g.delegateStaking.ClosePoolSession(opts, report, signedReport)
	if err != nil {
		return nil, stage1Error(err)
	}
	return tx, nil
}

// SettleExpiredSessionTx builds a settleExpiredSession transaction without
// sending it. Use with opts.NoSend = true for escalation support. It closes a
// pool session nobody closed once endsAt plus the settlement grace has passed.
// Anyone may send it: the stake returns to the pool and the provider reward
// becomes debt.
func (g *DelegateStaking) SettleExpiredSessionTx(opts *bind.TransactOpts, sessionID common.Hash) (*types.Transaction, error) {
	tx, err := g.delegateStaking.SettleExpiredSession(opts, sessionID)
	if err != nil {
		return nil, stage1Error(err)
	}
	return tx, nil
}
