package lib

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// fakeEscalationClient is a minimal in-memory chain for the escalator: it tracks
// the sender's pending nonce, the highest fee seen per nonce, and which tx hashes
// are "mined". Mining is driven by the test via mine(hash).
type fakeEscalationClient struct {
	mu          sync.Mutex
	nonce       uint64
	pendingFee  map[uint64]*big.Int // highest fee cap accepted per nonce
	mined       map[common.Hash]*types.Receipt
	sent        []*types.Transaction
	mineOnSend  bool // if true, every send is mined immediately (nonce advances)
	minedCalls  int
	receiptCall int
}

func newFakeClient() *fakeEscalationClient {
	return &fakeEscalationClient{
		pendingFee: make(map[uint64]*big.Int),
		mined:      make(map[common.Hash]*types.Receipt),
	}
}

func (c *fakeEscalationClient) TransactionReceipt(_ context.Context, h common.Hash) (*types.Receipt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.receiptCall++
	if r, ok := c.mined[h]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

func (c *fakeEscalationClient) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return nil, nil
}

func (c *fakeEscalationClient) SendTransaction(_ context.Context, tx *types.Transaction) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := tx.Nonce()
	if n < c.nonce {
		return errors.New("nonce too low")
	}
	if prev, ok := c.pendingFee[n]; ok {
		// Node rule: a replacement must be >= 10% above the pending fee.
		min := new(big.Int).Mul(prev, big.NewInt(110))
		min.Div(min, big.NewInt(100))
		if tx.GasFeeCap().Cmp(min) < 0 {
			return errors.New("replacement transaction underpriced")
		}
	}
	c.pendingFee[n] = new(big.Int).Set(tx.GasFeeCap())
	c.sent = append(c.sent, tx)
	if c.mineOnSend {
		c.mineLocked(tx)
	}
	return nil
}

func (c *fakeEscalationClient) mineLocked(tx *types.Transaction) {
	c.mined[tx.Hash()] = &types.Receipt{Status: 1, TxHash: tx.Hash(), BlockNumber: big.NewInt(1)}
	if tx.Nonce() >= c.nonce {
		c.nonce = tx.Nonce() + 1
	}
	delete(c.pendingFee, tx.Nonce())
	c.minedCalls++
}

func (c *fakeEscalationClient) mine(tx *types.Transaction) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mineLocked(tx)
}

func (c *fakeEscalationClient) lastSent() *types.Transaction {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sent[len(c.sent)-1]
}

func (c *fakeEscalationClient) SuggestGasPrice(context.Context) (*big.Int, error) {
	return big.NewInt(1_000_000), nil
}
func (c *fakeEscalationClient) SuggestGasTipCap(context.Context) (*big.Int, error) {
	return big.NewInt(1_000_000), nil // 0.001 gwei
}
func (c *fakeEscalationClient) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	return &types.Header{BaseFee: big.NewInt(5_000_000)}, nil // 0.005 gwei
}
func (c *fakeEscalationClient) PendingNonceAt(context.Context, common.Address) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.nonce, nil
}

var testFrom = common.HexToAddress("0x1111111111111111111111111111111111111111")

func testBuilder(opts *bind.TransactOpts) (*types.Transaction, error) {
	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(8453),
		Nonce:     opts.Nonce.Uint64(),
		GasTipCap: opts.GasTipCap,
		GasFeeCap: opts.GasFeeCap,
		Gas:       21000,
		To:        &testFrom,
	}), nil
}

func fastConfig() EscalationConfig {
	return EscalationConfig{
		CheckInterval: 30 * time.Millisecond,
		MaxAttempts:   3,
		BumpPercent:   15,
		MaxTotalTime:  2 * time.Second,
	}
}

func newTestEscalator(c *fakeEscalationClient, cfg EscalationConfig) *TransactionEscalator {
	return NewTransactionEscalator(c, NewTestLogger(), cfg)
}

// Issue #1011 case 2: a prior call left an escalated tx pending on the nonce. A new
// call must price above it instead of restarting from the fresh network estimate.
func TestSendWithEscalation_PricesAbovePendingFromEarlierCall(t *testing.T) {
	c := newFakeClient()
	cfg := fastConfig()
	cfg.MaxTotalTime = 300 * time.Millisecond
	e := newTestEscalator(c, cfg)
	opts := &bind.TransactOpts{From: testFrom}

	// Call 1: nothing mines; escalation runs out with the txs still pending.
	_, err := e.SendWithEscalation(context.Background(), opts, testBuilder, false)
	require.ErrorIs(t, err, ErrEscalationFailed)
	require.Contains(t, err.Error(), "still pending")
	stuckFee := c.lastSent().GasFeeCap()
	require.Len(t, c.sent, cfg.MaxAttempts)

	// Call 2 on the same (still pending) nonce: previously died on
	// "replacement transaction underpriced" on every attempt. Now it must
	// out-price the stuck tx on the first send.
	c.mu.Lock()
	c.mineOnSend = true
	c.mu.Unlock()
	receipt, err := e.SendWithEscalation(context.Background(), opts, testBuilder, false)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	require.Len(t, c.sent, cfg.MaxAttempts+1, "second call should succeed on its first send")

	replacement := c.lastSent()
	require.Equal(t, uint64(0), replacement.Nonce())
	min := new(big.Int).Mul(stuckFee, big.NewInt(110))
	min.Div(min, big.NewInt(100))
	require.True(t, replacement.GasFeeCap().Cmp(min) >= 0,
		"replacement fee %s must be >= 110%% of pending %s", replacement.GasFeeCap(), stuckFee)

	// Pending memory is cleared once the nonce is consumed.
	e.mu.Lock()
	_, stillPending := e.pending[testFrom]
	e.mu.Unlock()
	require.False(t, stillPending)
}

// Issue #1011 case 1: attempts are exhausted but a submitted tx mines a little
// later, inside MaxTotalTime. That must be reported as success, not failure.
func TestSendWithEscalation_LingersForLateInclusion(t *testing.T) {
	c := newFakeClient()
	cfg := fastConfig()
	e := newTestEscalator(c, cfg)
	opts := &bind.TransactOpts{From: testFrom}

	// Mine the first submitted tx shortly after all attempts have gone out.
	go func() {
		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			c.mu.Lock()
			n := len(c.sent)
			c.mu.Unlock()
			if n >= cfg.MaxAttempts {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		time.Sleep(3 * cfg.CheckInterval) // past the last waitForMining window
		c.mu.Lock()
		first := c.sent[0]
		c.mu.Unlock()
		c.mine(first)
	}()

	receipt, err := e.SendWithEscalation(context.Background(), opts, testBuilder, false)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	require.Equal(t, c.sent[0].Hash(), receipt.TxHash)
}

// A fresh nonce (nothing pending) must not be touched by the pending memory.
func TestSendWithEscalation_FreshNonceUsesNetworkEstimate(t *testing.T) {
	c := newFakeClient()
	c.mineOnSend = true
	e := newTestEscalator(c, fastConfig())
	opts := &bind.TransactOpts{From: testFrom}

	_, err := e.SendWithEscalation(context.Background(), opts, testBuilder, false)
	require.NoError(t, err)
	// gasFeeCap = tip + 2*baseFee from the fake client.
	require.Equal(t, int64(1_000_000+2*5_000_000), c.lastSent().GasFeeCap().Int64())

	// Second call lands on the next nonce; still the plain estimate.
	_, err = e.SendWithEscalation(context.Background(), opts, testBuilder, false)
	require.NoError(t, err)
	require.Equal(t, uint64(1), c.lastSent().Nonce())
	require.Equal(t, int64(1_000_000+2*5_000_000), c.lastSent().GasFeeCap().Int64())
}
