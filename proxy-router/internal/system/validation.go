package system

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type EthConnectionValidator struct {
	chainId big.Int
}

func NewEthConnectionValidator(chainId big.Int) *EthConnectionValidator {
	return &EthConnectionValidator{
		chainId: chainId,
	}
}

func (v *EthConnectionValidator) ValidateEthResourse(ctx context.Context, url string, timeout time.Duration) error {
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
