package system

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Validator struct {
	chainId big.Int
}

func NewValidator(chainId big.Int) *Validator {
	return &Validator{
		chainId: chainId,
	}
}

func (v *Validator) ValidateEthResourse(ctx context.Context, url string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ethClient, err := ethclient.DialContext(timeoutCtx, url)
	if err != nil {
		return err
	}

	urlChainId, chainIdError := ethClient.ChainID(timeoutCtx)
	if chainIdError != nil {
		return chainIdError
	}

	if v.chainId.Cmp(urlChainId) != 0 {
		return fmt.Errorf("invalid chain id %s, expected: %s", urlChainId, &v.chainId)
	}

	return nil
}
