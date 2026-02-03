package lib

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Default escalation settings
const (
	DefaultCheckInterval = 20 * time.Second // How often to check if tx is mined
	DefaultMaxAttempts   = 5                // Max escalation attempts
	DefaultBumpPercent   = 15               // Gas price bump percentage (min 10% for node acceptance)
	DefaultMaxTotalTime  = 3 * time.Minute  // Max total time to wait for mining
)

// Errors
var (
	ErrEscalationFailed = errors.New("transaction not mined after all escalation attempts")
	ErrTxNotMinedYet    = errors.New("transaction not mined yet")
	ErrMaxGasExceeded   = errors.New("max gas price limit exceeded")
)

// EscalationConfig holds settings for transaction escalation
type EscalationConfig struct {
	CheckInterval time.Duration // How often to check if tx is mined
	MaxAttempts   int           // Max escalation attempts before giving up
	BumpPercent   int64         // Gas price bump percentage (must be >= 10%)
	MaxTotalTime  time.Duration // Total max time to wait
	MaxGasPrice   *big.Int      // Optional: max gas price to prevent excessive spending
}

// DefaultEscalationConfig returns sensible defaults for transaction escalation
func DefaultEscalationConfig() EscalationConfig {
	return EscalationConfig{
		CheckInterval: DefaultCheckInterval,
		MaxAttempts:   DefaultMaxAttempts,
		BumpPercent:   DefaultBumpPercent,
		MaxTotalTime:  DefaultMaxTotalTime,
		MaxGasPrice:   nil, // No limit by default
	}
}

// GasPrices holds gas pricing parameters for both legacy and EIP-1559 transactions
type GasPrices struct {
	// Legacy
	GasPrice *big.Int

	// EIP-1559
	GasTipCap *big.Int // maxPriorityFeePerGas
	GasFeeCap *big.Int // maxFeePerGas
}

// Clone creates a copy of GasPrices
func (g *GasPrices) Clone() *GasPrices {
	clone := &GasPrices{}
	if g.GasPrice != nil {
		clone.GasPrice = new(big.Int).Set(g.GasPrice)
	}
	if g.GasTipCap != nil {
		clone.GasTipCap = new(big.Int).Set(g.GasTipCap)
	}
	if g.GasFeeCap != nil {
		clone.GasFeeCap = new(big.Int).Set(g.GasFeeCap)
	}
	return clone
}

// Bump increases all gas prices by the given percentage
func (g *GasPrices) Bump(percent int64) {
	multiplier := big.NewInt(100 + percent)
	hundred := big.NewInt(100)

	if g.GasPrice != nil {
		g.GasPrice.Mul(g.GasPrice, multiplier)
		g.GasPrice.Div(g.GasPrice, hundred)
	}
	if g.GasTipCap != nil {
		g.GasTipCap.Mul(g.GasTipCap, multiplier)
		g.GasTipCap.Div(g.GasTipCap, hundred)
	}
	if g.GasFeeCap != nil {
		g.GasFeeCap.Mul(g.GasFeeCap, multiplier)
		g.GasFeeCap.Div(g.GasFeeCap, hundred)
	}
}

// MaxPrice returns the highest gas price (for limit checking)
func (g *GasPrices) MaxPrice() *big.Int {
	max := big.NewInt(0)
	if g.GasPrice != nil && g.GasPrice.Cmp(max) > 0 {
		max = g.GasPrice
	}
	if g.GasFeeCap != nil && g.GasFeeCap.Cmp(max) > 0 {
		max = g.GasFeeCap
	}
	return max
}

// TransactionEscalator handles resubmission of stuck transactions with higher gas
type TransactionEscalator struct {
	client EthClientForEscalation
	log    ILogger
	config EscalationConfig
}

// EthClientForEscalation is the interface needed for transaction escalation
type EthClientForEscalation interface {
	bind.DeployBackend
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

// NewTransactionEscalator creates a new escalator with the given configuration
func NewTransactionEscalator(client EthClientForEscalation, log ILogger, config EscalationConfig) *TransactionEscalator {
	// Ensure minimum bump percent (nodes require at least 10%)
	if config.BumpPercent < 10 {
		config.BumpPercent = 10
	}
	return &TransactionEscalator{
		client: client,
		log:    log,
		config: config,
	}
}

// TxBuilder is a function that builds a transaction with the given gas prices and nonce.
// The opts.NoSend should be true, so the transaction is built but not sent.
// Returns the transaction to be sent.
type TxBuilder func(opts *bind.TransactOpts) (*types.Transaction, error)

// SendWithEscalation sends a transaction using the builder and escalates gas if not mined.
// The builder function should use opts.NoSend = true to return the transaction without sending.
// This function handles nonce management and gas price escalation automatically.
func (e *TransactionEscalator) SendWithEscalation(
	ctx context.Context,
	baseOpts *bind.TransactOpts,
	builder TxBuilder,
	isLegacy bool,
) (*types.Receipt, error) {
	// Create timeout context for the entire escalation process
	timeoutCtx, cancel := context.WithTimeout(ctx, e.config.MaxTotalTime)
	defer cancel()

	// Get initial gas prices
	gasPrices, err := e.getInitialGasPrices(ctx, isLegacy)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial gas prices: %w", err)
	}

	// Get nonce once - will be reused for all escalation attempts
	nonce, err := e.client.PendingNonceAt(ctx, baseOpts.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Track all submitted transactions (any of them could get mined)
	var submittedTxs []*types.Transaction
	var lastErr error

	for attempt := 0; attempt < e.config.MaxAttempts; attempt++ {
		// Check max gas limit
		if e.config.MaxGasPrice != nil && gasPrices.MaxPrice().Cmp(e.config.MaxGasPrice) > 0 {
			e.log.Warnf("Gas price %v exceeds max limit %v, stopping escalation",
				gasPrices.MaxPrice(), e.config.MaxGasPrice)
			return nil, WrapError(ErrMaxGasExceeded, lastErr)
		}

		// Prepare opts for this attempt
		opts := e.prepareOpts(baseOpts, gasPrices, nonce, isLegacy)
		opts.NoSend = true // Build transaction without sending

		// Build transaction
		tx, err := builder(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to build transaction: %w", err)
		}

		// Send transaction
		err = e.client.SendTransaction(ctx, tx)
		if err != nil {
			// Check if this is a "nonce too low" error - previous tx may have been mined
			if IsNonceTooLowError(err) && len(submittedTxs) > 0 {
				e.log.Infof("Nonce too low - checking if previous tx was mined")
				// Check if any of our submitted txs got mined
				if receipt := e.checkSubmittedTxs(ctx, submittedTxs); receipt != nil {
					e.log.Infof("Previous tx was mined: %s", receipt.TxHash.Hex())
					return receipt, nil
				}
			}
			// Check if this is a replacement issue (need to bump more)
			if IsReplacementError(err) && attempt < e.config.MaxAttempts-1 {
				e.log.Warnf("Tx rejected (replacement underpriced), bumping gas: %v", err)
				gasPrices.Bump(e.config.BumpPercent)
				continue
			}
			// For other errors, fail immediately
			return nil, fmt.Errorf("failed to send transaction: %w", err)
		}

		submittedTxs = append(submittedTxs, tx)
		e.log.Infof("Tx submitted (attempt %d/%d): %s, gasPrice: %v",
			attempt+1, e.config.MaxAttempts, tx.Hash().Hex(), gasPrices.MaxPrice())

		// Wait for mining with check interval
		receipt, err := e.waitForMining(timeoutCtx, tx, e.config.CheckInterval)
		if err == nil {
			e.log.Infof("Tx mined successfully: %s (attempt %d)", tx.Hash().Hex(), attempt+1)
			return receipt, nil
		}

		lastErr = err

		// Check if we timed out completely
		if errors.Is(err, context.DeadlineExceeded) {
			break
		}

		// Before escalating, check if any previous tx got mined in the meantime
		if receipt := e.checkSubmittedTxs(ctx, submittedTxs); receipt != nil {
			e.log.Infof("Previous tx was mined during escalation: %s", receipt.TxHash.Hex())
			return receipt, nil
		}

		// Not mined yet, escalate if we have attempts left
		if attempt < e.config.MaxAttempts-1 {
			e.log.Warnf("Tx %s not mined after %v, escalating gas (attempt %d/%d)",
				tx.Hash().Hex(), e.config.CheckInterval, attempt+1, e.config.MaxAttempts)
			gasPrices.Bump(e.config.BumpPercent)
		}
	}

	// Final check - maybe a tx got mined while we were processing
	if receipt := e.checkSubmittedTxs(ctx, submittedTxs); receipt != nil {
		e.log.Infof("Tx was mined on final check: %s", receipt.TxHash.Hex())
		return receipt, nil
	}

	if len(submittedTxs) > 0 {
		return nil, fmt.Errorf("%w: last tx %s", ErrEscalationFailed, submittedTxs[len(submittedTxs)-1].Hash().Hex())
	}
	return nil, ErrEscalationFailed
}

// checkSubmittedTxs checks if any of the submitted transactions have been mined
func (e *TransactionEscalator) checkSubmittedTxs(ctx context.Context, txs []*types.Transaction) *types.Receipt {
	for _, tx := range txs {
		receipt, err := e.client.TransactionReceipt(ctx, tx.Hash())
		if err == nil && receipt != nil {
			return receipt
		}
	}
	return nil
}

// IsNonceTooLowError checks if the error indicates the nonce is too low (tx already mined)
func IsNonceTooLowError(err error) bool {
	if err == nil {
		return false
	}
	errStr := toLower(err.Error())
	return contains(errStr, "nonce too low")
}

// getInitialGasPrices fetches current gas prices from the network
func (e *TransactionEscalator) getInitialGasPrices(ctx context.Context, isLegacy bool) (*GasPrices, error) {
	prices := &GasPrices{}

	if isLegacy {
		gasPrice, err := e.client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
		prices.GasPrice = gasPrice
	} else {
		gasTipCap, err := e.client.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, err
		}
		prices.GasTipCap = gasTipCap

		head, err := e.client.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, err
		}

		// gasFeeCap = gasTipCap + 2 * baseFee (buffer for fluctuation)
		prices.GasFeeCap = new(big.Int).Add(
			gasTipCap,
			new(big.Int).Mul(head.BaseFee, big.NewInt(basefeeWiggleMultiplier)),
		)
	}

	return prices, nil
}

// prepareOpts creates a copy of baseOpts with the given gas prices and nonce
func (e *TransactionEscalator) prepareOpts(
	baseOpts *bind.TransactOpts,
	gasPrices *GasPrices,
	nonce uint64,
	isLegacy bool,
) *bind.TransactOpts {
	opts := *baseOpts // Copy base opts
	opts.Nonce = big.NewInt(int64(nonce))

	if isLegacy {
		opts.GasPrice = gasPrices.GasPrice
	} else {
		opts.GasTipCap = gasPrices.GasTipCap
		opts.GasFeeCap = gasPrices.GasFeeCap
	}

	return &opts
}

// waitForMining waits for a transaction to be mined for the given interval
func (e *TransactionEscalator) waitForMining(
	ctx context.Context,
	tx *types.Transaction,
	interval time.Duration,
) (*types.Receipt, error) {
	checkCtx, cancel := context.WithTimeout(ctx, interval)
	defer cancel()

	receipt, err := bind.WaitMined(checkCtx, e.client, tx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrTxNotMinedYet
		}
		return nil, err
	}
	return receipt, nil
}

// IsReplacementError checks if an error indicates the tx was rejected
// due to insufficient gas price (needs bump and retry)
func IsReplacementError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	patterns := []string{
		"replacement transaction underpriced",
		"already known",
		"transaction underpriced",
		"max fee per gas less than block base fee", // EIP-1559: gasFeeCap < baseFee
		"max priority fee per gas higher than max fee per gas",
		"insufficient funds for gas",
	}
	for _, pattern := range patterns {
		if contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && containsLower(toLower(s), toLower(substr)))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// basefeeWiggleMultiplier is a multiplier for the basefee to set the maxFeePerGas
const basefeeWiggleMultiplier = 2
