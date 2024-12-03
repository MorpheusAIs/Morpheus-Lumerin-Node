package registries

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/multicall"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestGetSessions(t *testing.T) {
	t.Skip()
	ethNodeAddr := ""
	ethClient, err := ethclient.Dial(ethNodeAddr)
	require.NoError(t, err)

	mc := multicall.NewMulticall3(ethClient)

	diamondAddr := common.HexToAddress("0xb8C55cD613af947E73E262F0d3C54b7211Af16CF")
	sr := NewSessionRouter(diamondAddr, ethClient, mc, lib.NewTestLogger())
	sessionIDs, err := sr.GetSessionsIDsByProvider(context.Background(), common.HexToAddress("0x1441Bc52156Cf18c12cde6A92aE6BDE8B7f775D4"), big.NewInt(0), 2, OrderASC)
	require.NoError(t, err)
	for _, sessionID := range sessionIDs {
		fmt.Printf("sessionID: %v\n", common.Hash(sessionID).Hex())
	}

	_, sessions, err := sr.getMultipleSessions(context.Background(), sessionIDs)
	require.NoError(t, err)

	fmt.Printf("sessions: %v\n", sessions[0].Stake)
}
