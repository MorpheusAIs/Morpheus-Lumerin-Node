package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/Lumerin-protocol/contracts-go/clonefactory"
	"github.com/Lumerin-protocol/contracts-go/implementation"
	"github.com/Lumerin-protocol/contracts-go/lumerintoken"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/repositories/contracts"
)

const (
	LUMERIN_ADDR      = "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	CLONEFACTORY_ADDR = "0xa513E6E4b8f2a923D98304ec87F64353C4D5C853"
	ETH_NODE_ADDR     = "ws://localhost:8545"
	PRIVATE_KEY       = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	PRIVATE_KEY_2     = "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	CONTRACT_ID       = "0x9bd03768a7DCc129555dE410FF8E85528A4F88b5"
)

func TestGetContracts(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(ETH_NODE_ADDR)
	require.NoError(t, err)
	ethGateway := contracts.NewHashrateEthereum(common.HexToAddress(CLONEFACTORY_ADDR), client, &lib.LoggerMock{})

	ids, err := ethGateway.GetContractsIDs(ctx)
	require.NoError(t, err)
	fmt.Println(ids)
}

func TestGetContract(t *testing.T) {
	ctx := context.Background()
	ethGateway := makeEthGateway(t, makeEthClient(t))

	contract, err := ethGateway.GetContract(ctx, CONTRACT_ID)
	require.NoError(t, err)
	fmt.Printf("%+v\n", contract)
}

func TestPurchaseContract(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(ETH_NODE_ADDR)
	require.NoError(t, err)

	toAddr := lib.MustPrivKeyStringToAddr(PRIVATE_KEY_2)
	transferLMR(t, client, PRIVATE_KEY, toAddr, big.NewInt(100*1e8))

	err = PurchaseContract(ctx, client, CONTRACT_ID, PRIVATE_KEY_2)
	require.NoError(t, err)
}

func TestCloseContract(t *testing.T) {
	ctx := context.Background()
	ethGateway := makeEthGateway(t, makeEthClient(t))

	err := ethGateway.CloseContract(ctx, CONTRACT_ID, 0, PRIVATE_KEY_2)
	require.NoError(t, err)
}

func TestWatchClonefactoryContractPurchased(t *testing.T) {
	ctx := context.Background()
	ethClient := makeEthClient(t)
	ethGateway := makeEthGateway(t, ethClient)

	sub, err := ethGateway.CreateCloneFactorySubscription(ctx, common.HexToAddress(CLONEFACTORY_ADDR))
	require.NoError(t, err)
	defer sub.Unsubscribe()

	errCh := make(chan error)

	go func() {
		errCh <- PurchaseContract(ctx, ethClient, CONTRACT_ID, PRIVATE_KEY_2)
	}()

	select {
	case event := <-sub.Events():
		purchasedEvent, ok := event.(*clonefactory.ClonefactoryClonefactoryContractPurchased)
		require.True(t, ok)
		require.Equal(t, common.HexToAddress(CONTRACT_ID), purchasedEvent.Address)
	case err := <-sub.Err():
		require.NoError(t, err)
	case err := <-errCh:
		require.NoErrorf(t, err, "error while purchasing contract: %s", err)
	}
}

func TestImplementationContractClosed(t *testing.T) {
	ctx := context.Background()
	ethGateway := makeEthGateway(t, makeEthClient(t))

	sub, err := ethGateway.CreateImplementationSubscription(ctx, common.HexToAddress(CONTRACT_ID))
	require.NoError(t, err)
	defer sub.Unsubscribe()

	errCh := make(chan error)
	go func() {
		errCh <- ethGateway.CloseContract(ctx, CONTRACT_ID, 0, PRIVATE_KEY_2)
	}()

	select {
	case event := <-sub.Events():
		closedEvent, ok := event.(*implementation.ImplementationContractClosed)
		require.True(t, ok)
		require.Equal(t, lib.MustPrivKeyStringToAddr(PRIVATE_KEY_2), closedEvent.Buyer)
	case err := <-sub.Err():
		require.NoError(t, err)
	case err := <-errCh:
		require.NoErrorf(t, err, "error while closing contract: %s", err)
	}
}

func transferLMR(t *testing.T, client contracts.EthereumClient, fromPrivateKey string, toAddr common.Address, lmrAmount *big.Int) {
	ctx := context.Background()

	chainID, err := client.ChainID(ctx)
	require.NoError(t, err)

	ethGateway := contracts.NewHashrateEthereum(common.HexToAddress(CLONEFACTORY_ADDR), client, &lib.LoggerMock{})
	ethGateway.SetLegacyTx(true)

	lumerin, err := lumerintoken.NewLumerintoken(common.HexToAddress(LUMERIN_ADDR), client)
	require.NoError(t, err)

	opts, err := privateKeyToTransactOpts(ctx, fromPrivateKey, chainID)
	require.NoError(t, err)

	gasPrice, err := client.SuggestGasPrice(ctx)
	require.NoError(t, err)
	opts.GasPrice = gasPrice

	_, err = lumerin.Transfer(opts, toAddr, lmrAmount)
	require.NoError(t, err)
}

func makeEthClient(t *testing.T) *ethclient.Client {
	client, err := ethclient.Dial(ETH_NODE_ADDR)
	require.NoError(t, err)
	return client
}

func makeEthGateway(t *testing.T, client *ethclient.Client) *contracts.HashrateEthereum {
	ethGateway := contracts.NewHashrateEthereum(common.HexToAddress(CLONEFACTORY_ADDR), client, &lib.LoggerMock{})
	ethGateway.SetLegacyTx(true)
	return ethGateway
}

func privateKeyToTransactOpts(ctx context.Context, privKey string, chainID *big.Int) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return nil, err
	}

	return bind.NewKeyedTransactorWithChainID(privateKey, chainID)
}

func PurchaseContract(ctx context.Context, client contracts.EthereumClient, contractID string, privKey string) error {
	lumerin, err := lumerintoken.NewLumerintoken(common.HexToAddress(LUMERIN_ADDR), client)
	if err != nil {
		return err
	}

	chainId, err := client.ChainID(ctx)
	if err != nil {
		return err
	}

	opts, err := privateKeyToTransactOpts(ctx, PRIVATE_KEY_2, chainId)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	opts.GasPrice = gasPrice

	_, err = lumerin.Approve(opts, common.HexToAddress(CLONEFACTORY_ADDR), big.NewInt(5*1e8))
	if err != nil {
		return err
	}

	cloneFactory, err := clonefactory.NewClonefactory(common.HexToAddress(CLONEFACTORY_ADDR), client)
	if err != nil {
		return err
	}

	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return err
	}

	transactOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return err
	}

	// TODO: deal with likely gasPrice issue so our transaction processes before another pending nonce.
	gasPrice, err = client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	transactOpts.GasPrice = gasPrice

	transactOpts.GasLimit = uint64(1_000_000)
	transactOpts.Value = big.NewInt(0)
	transactOpts.Context = ctx

	watchOpts := &bind.WatchOpts{
		Context: ctx,
	}
	sink := make(chan *clonefactory.ClonefactoryClonefactoryContractPurchased)
	sub, err := cloneFactory.WatchClonefactoryContractPurchased(watchOpts, sink, []common.Address{})
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	_, err = cloneFactory.SetPurchaseRentalContract(transactOpts, common.HexToAddress(contractID), "", 0)
	if err != nil {
		return err
	}

	select {
	case <-sink:
		return nil
	case err := <-sub.Err():
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
