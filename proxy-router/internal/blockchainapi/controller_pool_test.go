package blockchainapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// stage1ServiceStub records what the keeper and payout routes send. The other
// methods of the embedded interfaces are not called by these tests.
type stage1ServiceStub struct {
	poolService
	providerPayoutsService
	batches []uint8
	targets []common.Address
}

func (s *stage1ServiceStub) ServePoolWithdrawals(_ context.Context, _ common.Address, maxEntries uint8) (common.Hash, error) {
	s.batches = append(s.batches, maxEntries)
	return common.Hash{}, nil
}

func (s *stage1ServiceStub) ReleasePoolLocks(_ context.Context, _ common.Address, maxBuckets uint8) (common.Hash, error) {
	s.batches = append(s.batches, maxBuckets)
	return common.Hash{}, nil
}

func (s *stage1ServiceStub) SetPayoutTarget(_ context.Context, _ common.Address, target common.Address) (common.Hash, error) {
	s.targets = append(s.targets, target)
	return common.Hash{}, nil
}

// A keeper call sends a batch the contract accepts, 1 through 32, and 32 when
// the request leaves it out. Anything else is refused before a transaction.
func TestPoolKeeperBatchSize(t *testing.T) {
	serve, release := (*BlockchainController).servePoolWithdrawals, (*BlockchainController).releasePoolLocks
	cases := []struct {
		name     string
		handle   func(*BlockchainController, *gin.Context)
		body     string
		wantCode int
		wantSent []uint8
	}{
		{"size left out", serve, `{}`, http.StatusOK, []uint8{32}},
		{"the maximum", serve, `{"maxEntries":32}`, http.StatusOK, []uint8{32}},
		{"zero", serve, `{"maxEntries":0}`, http.StatusBadRequest, nil},
		{"above the maximum", serve, `{"maxEntries":33}`, http.StatusBadRequest, nil},
		{"release without a body", release, ``, http.StatusOK, []uint8{32}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stage1ServiceStub{}
			controller := &BlockchainController{pool: stub, log: lib.NewTestLogger()}

			response := performUserStakesRequest(t, http.MethodPost, "/", tc.body, func(ctx *gin.Context) { tc.handle(controller, ctx) })

			require.Equal(t, tc.wantCode, response.Code)
			require.Equal(t, tc.wantSent, stub.batches)
		})
	}
}

// The payout target receives the provider's income, so it changes only when a
// full-access caller names it. The zero address is an explicit reset, never
// what a request that left the target out gets.
func TestSetPayoutTargetNeedsFullAccessAndExplicitTarget(t *testing.T) {
	target := common.HexToAddress("0x2222222222222222222222222222222222222222")
	authConf := system.HTTPAuthConfig{
		AuthEntries: map[string]*system.HTTPAuthEntry{"admin": {}, "agent": {}},
		Whitelists:  map[string][]string{"admin": {"*"}, "agent": {"set_payout_target"}},
	}
	cases := []struct {
		name     string
		username string
		body     string
		wantCode int
		wantSent []common.Address
	}{
		{"full access", "admin", `{"target":"` + target.Hex() + `"}`, http.StatusOK, []common.Address{target}},
		{"reset by the zero address", "admin", `{"target":"0x0000000000000000000000000000000000000000"}`, http.StatusOK, []common.Address{{}}},
		{"permission without full access", "agent", `{"target":"` + target.Hex() + `"}`, http.StatusForbidden, nil},
		{"target left out", "admin", `{}`, http.StatusBadRequest, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stage1ServiceStub{}
			controller := &BlockchainController{payouts: stub, authConf: authConf, log: lib.NewTestLogger()}

			response := performUserStakesRequest(t, http.MethodPost, "/", tc.body, func(ctx *gin.Context) {
				ctx.Params = gin.Params{{Key: "id", Value: "0x1111111111111111111111111111111111111111"}}
				ctx.Set("username", tc.username)
				controller.setPayoutTarget(ctx)
			})

			require.Equal(t, tc.wantCode, response.Code)
			require.Equal(t, tc.wantSent, stub.targets)
		})
	}
}
