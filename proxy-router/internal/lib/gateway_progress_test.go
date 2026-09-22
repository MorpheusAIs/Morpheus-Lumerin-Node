package lib

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestGatewayStakeGuard(t *testing.T) {
	ctx := context.WithValue(context.Background(), GatewayMaxStakeKey, big.NewInt(5))
	if CheckGatewayStake(ctx, big.NewInt(6)) == nil {
		t.Fatal("over-limit stake accepted")
	}
	if CheckGatewayStake(ctx, big.NewInt(5)) != nil {
		t.Fatal("exact limit rejected")
	}
}

func TestGatewayJournalPersistsAndRejectsDuplicate(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GATEWAY_JOURNAL_PATH", root)
	id := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	p, err := BeginGatewayOperation(id)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), GatewayProgressKey, p)
	if err := GatewayAttempt(ctx, "open"); err != nil {
		t.Fatal(err)
	}
	p.Finish(500)
	data, err := os.ReadFile(filepath.Join(root, id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var got GatewayProgress
	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Completed || got.Stage != "submitting_open" || got.HTTPStatus != 500 {
		t.Fatal(string(data))
	}
	if _, err := BeginGatewayOperation(id); err == nil {
		t.Fatal("duplicate operation accepted")
	}
}

func TestGatewayWalletLockCancellation(t *testing.T) {
	t.Setenv("GATEWAY_JOURNAL_PATH", t.TempDir())
	unlock, err := GatewayWalletLock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GatewayWalletLock(ctx); err == nil {
		t.Fatal("concurrent wallet mutation allowed")
	}
}

func TestGatewayJournalFailsClosedWhenDiskUnavailable(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GATEWAY_JOURNAL_PATH", root)
	p, err := BeginGatewayOperation("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), GatewayProgressKey, p)
	if GatewayAttempt(ctx, "open") == nil {
		t.Fatal("submission authorized without durable intent")
	}
}

// H1(a): ambiguous prior broadcast must not release the operation back to
// not_submitted (would invite a second session under the same op id / retry).
func TestGatewayBuildFailureOnlyReleasesUnsubmittedAttempt(t *testing.T) {
	p := &GatewayProgress{Stage: "submitting_open"}
	ctx := context.WithValue(context.Background(), GatewayProgressKey, p)
	GatewayBuildFailure(ctx, "open")
	if p.Stage != "not_submitted" {
		t.Fatal("proven pre-broadcast failure blocked")
	}
	p.Stage = "submitting_open"
	p.Transactions = append(p.Transactions, GatewayTransaction{Kind: "open"})
	GatewayBuildFailure(ctx, "open")
	if p.Stage == "not_submitted" {
		t.Fatal("previous unknown transaction forgotten")
	}
}

// H1(a) integration-style: exclusive journal rejects a second Begin for the
// same operation ID even after an ambiguous submitted_* stage.
func TestGatewayAmbiguousBroadcastDoesNotAllowSecondSession(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GATEWAY_JOURNAL_PATH", root)
	id := "cccccccccccccccccccccccccccccccc"
	p, err := BeginGatewayOperation(id)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), GatewayProgressKey, p)
	if err := GatewayAttempt(ctx, "open"); err != nil {
		t.Fatal(err)
	}
	// Simulate builder returning a tx whose broadcast outcome is unknown.
	fakeTx := types.NewTx(&types.LegacyTx{Nonce: 1, GasPrice: big.NewInt(1), Gas: 21000, To: &common.Address{}, Value: big.NewInt(0)})
	GatewayRecord(ctx, "open", fakeTx)
	GatewayBuildFailure(ctx, "open")
	if p.Stage == "not_submitted" {
		t.Fatal("H1a: ambiguous broadcast released stage to not_submitted")
	}
	if len(p.Transactions) == 0 {
		t.Fatal("H1a: transaction hash lost")
	}
	// Same operation ID must not start a second session attempt.
	if _, err := BeginGatewayOperation(id); err == nil {
		t.Fatal("H1a: second BeginGatewayOperation accepted for ambiguous op")
	}
}

// H1(b): journal write failure after a tx hash is known must leave durable
// operator-visible reconcile state (needs_reconcile + txs), not look like a
// silent open-failed / not_submitted.
func TestGatewayJournalWriteFailureAfterBroadcastNeedsReconcile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GATEWAY_JOURNAL_PATH", root)
	id := "dddddddddddddddddddddddddddddddd"
	p, err := BeginGatewayOperation(id)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), GatewayProgressKey, p)
	if err := GatewayAttempt(ctx, "open"); err != nil {
		t.Fatal(err)
	}
	// Destroy journal storage after intent was written, then record a tx.
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	fakeTx := types.NewTx(&types.LegacyTx{Nonce: 2, GasPrice: big.NewInt(1), Gas: 21000, To: &common.Address{}, Value: big.NewInt(0)})
	GatewayRecord(ctx, "open", fakeTx)
	if !p.NeedsReconcile() {
		t.Fatal("H1b: expected needs_reconcile after journal write failure")
	}
	if p.Stage == "not_submitted" {
		t.Fatal("H1b: looked like silent open-failed / not_submitted")
	}
	if len(p.Transactions) != 1 {
		t.Fatal("H1b: tx hash not retained for operator reconcile")
	}
	// Finish still surfaces completed + progress for HTTP (operator-visible).
	p.Finish(500)
	if !p.Completed || p.HTTPStatus != 500 {
		t.Fatal("H1b: Finish did not preserve operator-visible completion state")
	}
	if len(p.Transactions) != 1 {
		t.Fatal("H1b: Finish cleared transaction hashes")
	}
}

func TestGatewayOwnsCleanupRequiresExplicitCompanionSignal(t *testing.T) {
	t.Setenv("GATEWAY_JOURNAL_PATH", "")
	t.Setenv("GATEWAY_OWN_CLEANUP", "")
	if GatewayManaged() || GatewayOwnsCleanup() {
		t.Fatal("expected unmanaged without journal path")
	}
	t.Setenv("GATEWAY_JOURNAL_PATH", t.TempDir())
	if !GatewayManaged() {
		t.Fatal("expected managed with journal path")
	}
	if GatewayOwnsCleanup() {
		t.Fatal("H2: journal path alone must not claim exclusive cleanup")
	}
	t.Setenv("GATEWAY_OWN_CLEANUP", "1")
	if !GatewayOwnsCleanup() {
		t.Fatal("H2: companion signal should enable owns-cleanup")
	}
}

// Journal and progress structs must never carry wallet material fields (H4).
func TestGatewayProgressJSONHasNoWalletMaterial(t *testing.T) {
	p := &GatewayProgress{
		Stage:        "submitted_open",
		Transactions: []GatewayTransaction{{Kind: "open", Hash: common.HexToHash("0x1")}},
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, banned := range []string{"private", "PrivateKey", "mnemonic", "seed", "wallet_key", "WALLET"} {
		if containsFold(s, banned) {
			t.Fatalf("H4: journal JSON must not contain %q: %s", banned, s)
		}
	}
}

func containsFold(s, sub string) bool {
	return len(sub) > 0 && (len(s) >= len(sub)) && (stringIndexFold(s, sub) >= 0)
}

func stringIndexFold(s, sub string) int {
	ls, lsub := len(s), len(sub)
	for i := 0; i+lsub <= ls; i++ {
		ok := true
		for j := 0; j < lsub; j++ {
			a, b := s[i+j], sub[j]
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	return -1
}
