package lib

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

// Default timeout for waiting for transactions to be mined
// Set to 1 minute to accommodate slow networks and gas price fluctuations
const DefaultTxMineTimeout = 1 * time.Minute

// ErrTxTimeout is returned when a transaction times out waiting to be mined
var ErrTxTimeout = errors.New("transaction timeout: tx not mined within timeout period")

// TransactionBackend is the interface for transaction-related operations
type TransactionBackend interface {
	bind.DeployBackend
}

// WaitMinedWithTimeout waits for a transaction to be mined with a timeout.
// If the transaction is not mined within the timeout, it returns ErrTxTimeout.
// This prevents infinite polling when transactions are stuck in the mempool.
func WaitMinedWithTimeout(ctx context.Context, backend TransactionBackend, tx *types.Transaction, timeout time.Duration) (*types.Receipt, error) {
	if timeout == 0 {
		timeout = DefaultTxMineTimeout
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	receipt, err := bind.WaitMined(timeoutCtx, backend, tx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, WrapError(ErrTxTimeout, err)
		}
		return nil, err
	}
	return receipt, nil
}

// IsNonceError checks if an error is related to nonce issues.
// Common nonce errors from various Ethereum clients:
// - "nonce too low" - transaction nonce is lower than expected
// - "nonce too high" - transaction nonce is higher than expected (gap)
// - "replacement transaction underpriced" - same nonce but gas too low to replace
// - "already known" - transaction with same nonce already in mempool
// - "transaction underpriced" - gas price too low (related to replacement)
func IsNonceError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	nonceErrorPatterns := []string{
		"nonce too low",
		"nonce too high",
		"replacement transaction underpriced",
		"already known",
		"transaction underpriced",
		"invalid nonce",
		"incorrect nonce",
	}
	for _, pattern := range nonceErrorPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}
