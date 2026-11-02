package lib

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
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

	// lingerPollDivisor sets how often submitted-but-unmined txs are re-checked once
	// escalation attempts are exhausted but MaxTotalTime budget remains: CheckInterval / divisor.
	lingerPollDivisor = 4
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

// pendingTx remembers the last transaction this escalator submitted for a sender
// that it did not see mined. A later SendWithEscalation for the same sender that
// lands on the same nonce is a replacement and must be priced above it, not from a
// fresh network estimate (see #1011).
type pendingTx struct {
	nonce  uint64
	prices *GasPrices
}

// TransactionEscalator handles resubmission of stuck transactions with higher gas
type TransactionEscalator struct {
	client EthClientForEscalation
	log    ILogger
	config EscalationConfig

	mu      sync.Mutex
	pending map[common.Address]pendingTx
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
		client:  client,
		log:     log,
		config:  config,
		pending: make(map[common.Address]pendingTx),
	}
}

// rememberPending records the gas prices of the last submitted, unmined tx for a sender.
func (e *TransactionEscalator) rememberPending(from common.Address, nonce uint64, prices *GasPrices) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pending[from] = pendingTx{nonce: nonce, prices: prices.Clone()}
}

// forgetPending clears the remembered pending tx for a sender once its nonce is consumed.
func (e *TransactionEscalator) forgetPending(from common.Address) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.pending, from)
}

// priceAbovePending raises gasPrices so the tx out-prices a remembered pending tx
// from an earlier call on the same nonce. Returns true if prices were raised.
func (e *TransactionEscalator) priceAbovePending(from common.Address, nonce uint64, gasPrices *GasPrices) bool {
	e.mu.Lock()
	p, ok := e.pending[from]
	e.mu.Unlock()
	if !ok || p.nonce != nonce || p.prices == nil {
		return false
	}
	if gasPrices.MaxPrice().Cmp(p.prices.MaxPrice()) > 0 {
		return false
	}
	raised := p.prices.Clone()
	raised.Bump(e.config.BumpPercent)
	*gasPrices = *raised
	return true
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
	log := e.log
	if id := RequestIDFromContext(ctx); id != "" {
		log = log.With("request_id", id)
	}

	// Create timeout context for the entire escalation process
	timeoutCtx, cancel := context.WithTimeout(ctx, e.config.MaxTotalTime)
	defer cancel()

	// Get initial gas prices
	gasPrices, err := e.getInitialGasPrices(ctx, isLegacy)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial gas prices: %w", err)
	}

	// Track all submitted transactions (any of them could get mined)
	var submittedTxs []*types.Transaction
	var lastErr error
	var nonceRefreshCount int
	const maxNonceRefreshes = 3 // Max times to refresh nonce on conflicts

	for attempt := 0; attempt < e.config.MaxAttempts; attempt++ {
		// Get fresh nonce for this attempt
		// We fetch nonce inside the loop to handle concurrent request conflicts
		nonce, err := e.client.PendingNonceAt(ctx, baseOpts.From)
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}

		// If an earlier call left a tx pending on this nonce, this tx is a replacement
		// and must be priced above it; a fresh estimate would be rejected as underpriced.
		if e.priceAbovePending(baseOpts.From, nonce, gasPrices) {
			log.Warnf("Nonce %d still has a pending tx from an earlier call, pricing above it: %v", nonce, gasPrices.MaxPrice())
		}

		// Check max gas limit
		if e.config.MaxGasPrice != nil && gasPrices.MaxPrice().Cmp(e.config.MaxGasPrice) > 0 {
			log.Warnf("Gas price %v exceeds max limit %v, stopping escalation",
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
			// Check if this is a "nonce too low" error
			if IsNonceTooLowError(err) {
				// If we have previously submitted txs, check if one was mined
				if len(submittedTxs) > 0 {
					log.Infof("Nonce too low - checking if previous tx was mined")
					if receipt := e.checkSubmittedTxs(ctx, submittedTxs); receipt != nil {
						log.Infof("Previous tx was mined: %s", receipt.TxHash.Hex())
						e.forgetPending(baseOpts.From)
						return receipt, nil
					}
				}
				// Nonce conflict from concurrent request - refresh and retry
				if nonceRefreshCount < maxNonceRefreshes {
					nonceRefreshCount++
					log.Warnf("Nonce conflict (attempt %d), refreshing nonce and retrying: %v", attempt+1, err)
					// Don't count this as an escalation attempt, just retry with fresh nonce
					attempt--
					continue
				}
				log.Errorf("Max nonce refreshes exceeded, giving up")
			}
			// Check if this is a replacement issue (need to bump more)
			if IsReplacementError(err) && attempt < e.config.MaxAttempts-1 {
				log.Warnf("Tx rejected (replacement underpriced), bumping gas: %v", err)
				gasPrices.Bump(e.config.BumpPercent)
				continue
			}
			// For other errors, fail immediately
			return nil, fmt.Errorf("failed to send transaction: %w", err)
		}

		// Reset nonce refresh count on successful send
		nonceRefreshCount = 0

		submittedTxs = append(submittedTxs, tx)
		e.rememberPending(baseOpts.From, nonce, gasPrices)
		log.Infof("Tx submitted (attempt %d/%d): %s, gasPrice: %v",
			attempt+1, e.config.MaxAttempts, tx.Hash().Hex(), gasPrices.MaxPrice())

		// Wait for mining with check interval
		receipt, err := e.waitForMining(timeoutCtx, tx, e.config.CheckInterval)
		if err == nil {
			log.Infof("Tx mined successfully: %s (attempt %d)", tx.Hash().Hex(), attempt+1)
			e.forgetPending(baseOpts.From)
			return receipt, nil
		}

		lastErr = err

		// Check if we timed out completely
		if errors.Is(err, context.DeadlineExceeded) {
			break
		}

		// Before escalating, check if any previous tx got mined in the meantime
		if receipt := e.checkSubmittedTxs(ctx, submittedTxs); receipt != nil {
			log.Infof("Previous tx was mined during escalation: %s", receipt.TxHash.Hex())
			e.forgetPending(baseOpts.From)
			return receipt, nil
		}

		// Not mined yet, escalate if we have attempts left
		if attempt < e.config.MaxAttempts-1 {
			log.Warnf("Tx %s not mined after %v, escalating gas (attempt %d/%d)",
				tx.Hash().Hex(), e.config.CheckInterval, attempt+1, e.config.MaxAttempts)
			gasPrices.Bump(e.config.BumpPercent)
		}
	}

	// Final check - maybe a tx got mined while we were processing
	if receipt := e.checkSubmittedTxs(ctx, submittedTxs); receipt != nil {
		log.Infof("Tx was mined on final check: %s", receipt.TxHash.Hex())
		e.forgetPending(baseOpts.From)
		return receipt, nil
	}

	// Attempts are exhausted but the submitted txs are still in the pool and the
	// caller's MaxTotalTime budget may not be. Keep watching them until it is, so a
	// slow-but-successful inclusion is reported as success rather than failure.
	if len(submittedTxs) > 0 {
		if receipt := e.lingerForSubmitted(timeoutCtx, log, submittedTxs); receipt != nil {
			e.forgetPending(baseOpts.From)
			return receipt, nil
		}
		hashes := make([]string, 0, len(submittedTxs))
		for _, tx := range submittedTxs {
			hashes = append(hashes, tx.Hash().Hex())
		}
		return nil, fmt.Errorf("%w: %d tx(s) still pending and may yet be mined: %s",
			ErrEscalationFailed, len(submittedTxs), strings.Join(hashes, ", "))
	}
	return nil, ErrEscalationFailed
}

// lingerForSubmitted polls the submitted txs for a receipt until ctx expires.
func (e *TransactionEscalator) lingerForSubmitted(ctx context.Context, log ILogger, txs []*types.Transaction) *types.Receipt {
	if ctx.Err() != nil {
		return nil
	}
	log.Warnf("Escalation attempts exhausted, %d tx(s) still pending; watching until the wait window closes", len(txs))
	poll := e.config.CheckInterval / lingerPollDivisor
	if poll <= 0 {
		poll = time.Second
	}
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if receipt := e.checkSubmittedTxs(ctx, txs); receipt != nil {
				log.Infof("Pending tx mined after escalation attempts: %s", receipt.TxHash.Hex())
				return receipt
			}
		}
	}
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
