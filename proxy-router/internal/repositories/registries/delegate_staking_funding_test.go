package registries

import (
	"context"
	"errors"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces/mocks"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/delegatestakingfunding"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// revertError is how a node answers a reverted eth_call: a message, and the
// revert data that geth exposes through ErrorData.
type revertError struct {
	message string
	data    []byte
}

func (e revertError) Error() string          { return e.message }
func (e revertError) ErrorData() interface{} { return hexutil.Encode(e.data) }

// Every close and failover asks IsPoolSession. A diamond without the Stage 1
// cut has no pool sessions, also when the node leaves the revert reason out
// of the message, but a failed read must never pass for a regular session.
func TestIsPoolSession(t *testing.T) {
	stringType, err := abi.NewType("string", "", nil)
	require.NoError(t, err)
	reason, err := abi.Arguments{{Type: stringType}}.Pack("Diamond: selector is not registered")
	require.NoError(t, err)
	notRegistered := append(crypto.Keccak256([]byte("Error(string)"))[:4], reason...)

	contractABI, err := delegatestakingfunding.DelegateStakingFundingMetaData.GetAbi()
	require.NoError(t, err)
	poolSession, err := contractABI.Methods["getPoolSession"].Outputs.Pack(delegatestakingfunding.IDelegateStakingStoragePoolSession{Exists: true})
	require.NoError(t, err)

	cases := []struct {
		name    string
		ret     []byte
		callErr error
		want    bool
		wantErr bool
	}{
		{"diamond without the Stage 1 cut", nil, revertError{"execution reverted", notRegistered}, false, false},
		{"pool session", poolSession, nil, true, false},
		{"network error", nil, errors.New("connection refused"), false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := mocks.NewEthClientMock(t)
			client.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(tc.ret, tc.callErr)

			got, err := NewDelegateStakingFunding(common.Address{}, client, lib.NewTestLogger()).IsPoolSession(context.Background(), common.Hash{})

			require.Equal(t, tc.want, got)
			require.Equal(t, tc.wantErr, err != nil, "err: %v", err)
		})
	}
}
