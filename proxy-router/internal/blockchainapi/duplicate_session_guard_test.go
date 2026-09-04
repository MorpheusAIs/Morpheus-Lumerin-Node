package blockchainapi

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	sessionrouter "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestScanLiveSessionPagesChecksEveryPage(t *testing.T) {
	want := &structs.Session{Id: common.HexToHash("0x99").Hex()}
	var offsets []int64

	got, err := scanLiveSessionPages(context.Background(), func(_ context.Context, offset *big.Int) (*structs.Session, int, error) {
		offsets = append(offsets, offset.Int64())
		switch offset.Int64() {
		case 0:
			return nil, int(guardedSessionScanPageSize), nil
		case int64(guardedSessionScanPageSize):
			return want, 1, nil
		default:
			t.Fatalf("unexpected offset %s", offset)
			return nil, 0, nil
		}
	})

	require.NoError(t, err)
	require.Same(t, want, got)
	require.Equal(t, []int64{0, int64(guardedSessionScanPageSize)}, offsets)
}

func TestScanLiveSessionPagesKeepsLatestMatch(t *testing.T) {
	older := &structs.Session{Id: common.HexToHash("0x41").Hex()}
	newer := &structs.Session{Id: common.HexToHash("0x42").Hex()}
	var offsets []int64

	got, err := scanLiveSessionPages(context.Background(), func(_ context.Context, offset *big.Int) (*structs.Session, int, error) {
		offsets = append(offsets, offset.Int64())
		switch offset.Int64() {
		case 0:
			return older, int(guardedSessionScanPageSize), nil
		case int64(guardedSessionScanPageSize):
			return newer, 1, nil
		default:
			t.Fatalf("unexpected offset %s", offset)
			return nil, 0, nil
		}
	})

	require.NoError(t, err)
	require.Same(t, newer, got)
	require.Equal(t, []int64{0, int64(guardedSessionScanPageSize)}, offsets)
}

func TestScanLiveSessionPagesFailsClosedOnQueryError(t *testing.T) {
	wantErr := errors.New("rpc unavailable")
	got, err := scanLiveSessionPages(context.Background(), func(context.Context, *big.Int) (*structs.Session, int, error) {
		return nil, 0, wantErr
	})

	require.Nil(t, got)
	require.ErrorIs(t, err, wantErr)
}

func TestIsLiveUserSession(t *testing.T) {
	user := common.HexToAddress("0x1234")
	otherUser := common.HexToAddress("0x5678")
	now := big.NewInt(1_000)

	tests := []struct {
		name    string
		session sessionrouter.ISessionStorageSession
		want    bool
		wantErr bool
	}{
		{
			name: "live",
			session: sessionrouter.ISessionStorageSession{
				User: user, ClosedAt: big.NewInt(0), EndsAt: big.NewInt(1_001),
			},
			want: true,
		},
		{
			name: "closed on chain",
			session: sessionrouter.ISessionStorageSession{
				User: user, ClosedAt: big.NewInt(999), EndsAt: big.NewInt(1_001),
			},
		},
		{
			name: "expired",
			session: sessionrouter.ISessionStorageSession{
				User: user, ClosedAt: big.NewInt(0), EndsAt: big.NewInt(1_000),
			},
		},
		{
			name: "different user",
			session: sessionrouter.ISessionStorageSession{
				User: otherUser, ClosedAt: big.NewInt(0), EndsAt: big.NewInt(1_001),
			},
			wantErr: true,
		},
		{
			name: "incomplete chain response",
			session: sessionrouter.ISessionStorageSession{
				User: user, ClosedAt: nil, EndsAt: big.NewInt(1_001),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isLiveUserSession(tt.session, user, now)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestKeyedLockSetSerializesOnlyMatchingKeys(t *testing.T) {
	var locks keyedLockSet
	unlockFirst := locks.lock("wallet:model-a")

	sameKeyAcquired := make(chan func(), 1)
	go func() {
		sameKeyAcquired <- locks.lock("wallet:model-a")
	}()

	select {
	case unlock := <-sameKeyAcquired:
		unlock()
		t.Fatal("matching key acquired before the first holder released it")
	case <-time.After(25 * time.Millisecond):
	}

	differentKeyAcquired := make(chan func(), 1)
	go func() {
		differentKeyAcquired <- locks.lock("wallet:model-b")
	}()
	select {
	case unlock := <-differentKeyAcquired:
		unlock()
	case <-time.After(time.Second):
		t.Fatal("unrelated model was blocked by another key")
	}

	unlockFirst()
	select {
	case unlock := <-sameKeyAcquired:
		unlock()
	case <-time.After(time.Second):
		t.Fatal("matching key did not acquire after release")
	}
}

func TestWriteOpenSessionByModelErrorReturnsStructuredConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	sessionID := common.HexToHash("0x42").Hex()

	writeOpenSessionByModelError(ctx, &ExistingSessionError{SessionID: sessionID})

	require.Equal(t, http.StatusConflict, recorder.Code)
	var body structs.ExistingSessionRes
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, sessionID, body.ExistingSessionID)
}
