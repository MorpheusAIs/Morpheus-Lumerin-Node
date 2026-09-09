package blockchainapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type userStakesServiceStub struct {
	available          *big.Int
	hold               *big.Int
	txHash             common.Hash
	getErr             error
	withdrawErr        error
	getIterations      []uint8
	withdrawIterations []uint8
}

func (s *userStakesServiceStub) GetUserStakesOnHold(_ context.Context, iterations uint8) (*big.Int, *big.Int, error) {
	s.getIterations = append(s.getIterations, iterations)
	return s.available, s.hold, s.getErr
}

func (s *userStakesServiceStub) WithdrawUserStakes(_ context.Context, iterations uint8) (common.Hash, error) {
	s.withdrawIterations = append(s.withdrawIterations, iterations)
	return s.txHash, s.withdrawErr
}

func newUserStakesController(stub *userStakesServiceStub) *BlockchainController {
	return &BlockchainController{
		userStakes: stub,
		log:        lib.NewTestLogger(),
	}
}

func performUserStakesRequest(t *testing.T, method, target, body string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		ctx.Request.Header.Set("Content-Type", "application/json")
	}

	handler(ctx)
	return recorder
}

func TestGetUserStakesOnHoldUsesDefaultIterations(t *testing.T) {
	stub := &userStakesServiceStub{available: big.NewInt(100), hold: big.NewInt(200)}
	controller := newUserStakesController(stub)

	response := performUserStakesRequest(t, http.MethodGet, "/blockchain/stakes/onhold", "", controller.getUserStakesOnHold)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, []uint8{structs.DefaultUserStakeIterations}, stub.getIterations)
	require.JSONEq(t, `{"available":"100","hold":"200"}`, response.Body.String())
}

func TestGetUserStakesOnHoldAcceptsCustomIterations(t *testing.T) {
	stub := &userStakesServiceStub{available: big.NewInt(1), hold: big.NewInt(2)}
	controller := newUserStakesController(stub)

	response := performUserStakesRequest(t, http.MethodGet, "/blockchain/stakes/onhold?iterations=20", "", controller.getUserStakesOnHold)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, []uint8{20}, stub.getIterations)
}

func TestGetUserStakesOnHoldRejectsInvalidIterations(t *testing.T) {
	stub := &userStakesServiceStub{}
	controller := newUserStakesController(stub)

	for _, target := range []string{
		"/blockchain/stakes/onhold?iterations=0",
		"/blockchain/stakes/onhold?iterations=256",
		"/blockchain/stakes/onhold?iterations=invalid",
	} {
		t.Run(target, func(t *testing.T) {
			response := performUserStakesRequest(t, http.MethodGet, target, "", controller.getUserStakesOnHold)
			require.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
	require.Empty(t, stub.getIterations)
}

func TestGetUserStakesOnHoldReturnsServiceError(t *testing.T) {
	stub := &userStakesServiceStub{getErr: errors.New("contract read failed")}
	controller := newUserStakesController(stub)

	response := performUserStakesRequest(t, http.MethodGet, "/blockchain/stakes/onhold", "", controller.getUserStakesOnHold)

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"error":"contract read failed"}`, response.Body.String())
}

func TestWithdrawUserStakesUsesDefaultIterationsWithoutBody(t *testing.T) {
	txHash := common.HexToHash("0x1234")
	stub := &userStakesServiceStub{txHash: txHash}
	controller := newUserStakesController(stub)

	response := performUserStakesRequest(t, http.MethodPost, "/blockchain/stakes/withdraw", "", controller.withdrawUserStakes)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, []uint8{structs.DefaultUserStakeIterations}, stub.withdrawIterations)

	var body structs.TxRes
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, txHash, body.Tx)
}

func TestWithdrawUserStakesAcceptsCustomIterations(t *testing.T) {
	stub := &userStakesServiceStub{}
	controller := newUserStakesController(stub)

	response := performUserStakesRequest(t, http.MethodPost, "/blockchain/stakes/withdraw", `{"iterations":20}`, controller.withdrawUserStakes)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, []uint8{20}, stub.withdrawIterations)
}

func TestWithdrawUserStakesRejectsInvalidBody(t *testing.T) {
	tests := map[string]string{
		"zero":      `{"iterations":0}`,
		"too large": `{"iterations":256}`,
		"malformed": `{"iterations":`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			stub := &userStakesServiceStub{}
			controller := newUserStakesController(stub)

			response := performUserStakesRequest(t, http.MethodPost, "/blockchain/stakes/withdraw", body, controller.withdrawUserStakes)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Empty(t, stub.withdrawIterations)
		})
	}
}

func TestWithdrawUserStakesReturnsServiceError(t *testing.T) {
	stub := &userStakesServiceStub{withdrawErr: errors.New("transaction failed")}
	controller := newUserStakesController(stub)

	response := performUserStakesRequest(t, http.MethodPost, "/blockchain/stakes/withdraw", `{}`, controller.withdrawUserStakes)

	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.JSONEq(t, `{"error":"transaction failed"}`, response.Body.String())
}
