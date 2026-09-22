// Gateway operation journal, stake guard, and process-local wallet lock.
package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const GatewayProgressKey = "morpheus_gateway_progress_v1"
const GatewayMaxStakeKey = "morpheus_gateway_max_stake_v1"

// StageNeedsReconcile is set when a transaction hash was recorded but the
// durable journal write failed afterward. Operators must reconcile from the
// in-memory/HTTP progress payload — never treat this as a silent open-failed.
const StageNeedsReconcile = "needs_reconcile"

var gatewayOperationIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

type GatewayTransaction struct {
	Kind string      `json:"kind"`
	Hash common.Hash `json:"hash"`
}

type GatewayProgress struct {
	Stage             string               `json:"stage"`
	Completed         bool                 `json:"completed"`
	HTTPStatus        int                  `json:"http_status"`
	SessionID         common.Hash          `json:"sessionID"`
	Transactions      []GatewayTransaction `json:"transactions"`
	UpdatedAt         int64                `json:"updated_at"`
	JournalWriteError string               `json:"journal_write_error,omitempty"`
	path              string
}

func BeginGatewayOperation(id string) (*GatewayProgress, error) {
	// Unmanaged / stock callers omit X-Gateway-Operation. Return nil so
	// omitempty drops progress from response bodies on unmanaged nodes.
	if id == "" {
		return nil, nil
	}
	p := &GatewayProgress{Stage: "not_submitted", Transactions: []GatewayTransaction{}}
	if !gatewayOperationIDPattern.MatchString(id) {
		return nil, fmt.Errorf("invalid gateway operation ID")
	}
	root := os.Getenv("GATEWAY_JOURNAL_PATH")
	if root == "" {
		return nil, fmt.Errorf("gateway operation journal is not configured")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	p.path = filepath.Join(root, id+".json")
	// Never execute the same operation ID twice, including after a restart.
	file, err := os.OpenFile(p.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("gateway operation already exists or journal unavailable")
	}
	file.Close()
	if err := p.persist(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *GatewayProgress) persist() error {
	if p == nil || p.path == "" {
		return nil
	}
	p.UpdatedAt = time.Now().Unix()
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	temp, err := os.OpenFile(p.path+".tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err = temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if err = os.Rename(p.path+".tmp", p.path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(p.path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func GatewayProgressFor(ctx context.Context) *GatewayProgress {
	p, _ := ctx.Value(GatewayProgressKey).(*GatewayProgress)
	return p
}

func GatewayAttempt(ctx context.Context, kind string) error {
	if p := GatewayProgressFor(ctx); p != nil {
		previous := p.Stage
		p.Stage = "submitting_" + kind
		if err := p.persist(); err != nil {
			p.Stage = previous
			return err
		}
		return nil
	}
	return nil
}

// GatewayRecord appends a transaction hash after the builder produced a tx.
// Persist failure after a hash is known marks needs_reconcile and keeps the
// hash in memory so HTTP responses remain operator-visible (H1b). It does not
// clear the stage back to not_submitted (H1a).
func GatewayRecord(ctx context.Context, kind string, tx *types.Transaction) {
	p := GatewayProgressFor(ctx)
	if p == nil || tx == nil {
		return
	}
	p.Transactions = append(p.Transactions, GatewayTransaction{Kind: kind, Hash: tx.Hash()})
	p.Stage = "submitted_" + kind
	if err := p.persist(); err != nil {
		p.Stage = StageNeedsReconcile
		p.JournalWriteError = err.Error()
		// Best-effort second persist is intentionally skipped — disk may be gone.
		// In-memory progress (including txs) must still be returned to the caller.
	}
}

func GatewaySession(ctx context.Context, id common.Hash) {
	if p := GatewayProgressFor(ctx); p != nil {
		p.SessionID = id
		if err := p.persist(); err != nil {
			p.Stage = StageNeedsReconcile
			p.JournalWriteError = err.Error()
		}
	}
}

// Finish marks the operation completed and persists the journal.
// Controllers defer Finish after ctx.JSON, so HTTP response bodies always
// show pre-completion progress (completed:false, http_status:0). The
// durable journal under GATEWAY_JOURNAL_PATH is the source of truth for
// completed / http_status after the handler returns.
func (p *GatewayProgress) Finish(status int) {
	if p == nil {
		return
	}
	p.Completed = true
	p.HTTPStatus = status
	_ = p.persist()
}

// NeedsReconcile reports whether chain progress may have advanced while the
// durable journal could not be updated.
func (p *GatewayProgress) NeedsReconcile() bool {
	return p != nil && (p.Stage == StageNeedsReconcile || p.JournalWriteError != "")
}

func CheckGatewayStake(ctx context.Context, amount *big.Int) error {
	limit, _ := ctx.Value(GatewayMaxStakeKey).(*big.Int)
	if limit != nil && (limit.Sign() <= 0 || amount.Cmp(limit) > 0) {
		return fmt.Errorf("gateway stake limit exceeded")
	}
	return nil
}

var gatewayWallet = make(chan struct{}, 1)

// GatewayManaged is true when an operation journal path is configured.
// This enables journal headers, stake-limit context, and the process-local
// wallet lock. It does NOT by itself claim exclusive expiry cleanup (see H2 /
// GatewayOwnsCleanup).
func GatewayManaged() bool {
	return os.Getenv("GATEWAY_JOURNAL_PATH") != ""
}

// GatewayOwnsCleanup is true only when managed mode is on AND an explicit
// companion signal proves the gateway owns durable expiry cleanup.
// Without GATEWAY_OWN_CLEANUP=1 the native expiry loop MUST keep running (H2).
func GatewayOwnsCleanup() bool {
	return GatewayManaged() && os.Getenv("GATEWAY_OWN_CLEANUP") == "1"
}

// GatewayWalletLock serializes wallet-bearing mutations in this process only.
// It does not coordinate across replicas (H3).
func GatewayWalletLock(ctx context.Context) (func(), error) {
	if !GatewayManaged() {
		return func() {}, nil
	}
	select {
	case gatewayWallet <- struct{}{}:
		return func() { <-gatewayWallet }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GatewayBuildFailure resets stage only when no transaction of the given kind
// was recorded. If a prior hash exists the overall outcome is ambiguous — do
// not authorize a second session for the same operation (H1a).
func GatewayBuildFailure(ctx context.Context, kind string) {
	p := GatewayProgressFor(ctx)
	if p == nil {
		return
	}
	for _, tx := range p.Transactions {
		if tx.Kind == kind {
			return
		}
	}
	p.Stage = "not_submitted"
	_ = p.persist()
}
