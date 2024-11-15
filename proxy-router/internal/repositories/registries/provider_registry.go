package registries

import (
	"context"
	"fmt"
	"math/big"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	mc "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/multicall"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ProviderRegistry struct {
	// config
	providerRegistryAddr common.Address

	// state
	nonce uint64

	// deps
	providerRegistry    *providerregistry.ProviderRegistry
	providerRegistryAbi *abi.ABI
	client              i.ContractBackend
	multicall           mc.MulticallBackend
	log                 lib.ILogger
}

func NewProviderRegistry(providerRegistryAddr common.Address, client i.ContractBackend, multicall mc.MulticallBackend, log lib.ILogger) *ProviderRegistry {
	pr, err := providerregistry.NewProviderRegistry(providerRegistryAddr, client)
	if err != nil {
		panic("invalid provider registry ABI")
	}
	providerRegistryAbi, err := providerregistry.ProviderRegistryMetaData.GetAbi()
	if err != nil {
		panic("invalid provider registry ABI")
	}
	return &ProviderRegistry{
		providerRegistry:     pr,
		providerRegistryAddr: providerRegistryAddr,
		providerRegistryAbi:  providerRegistryAbi,
		multicall:            multicall,
		client:               client,
		log:                  log,
	}
}

func (g *ProviderRegistry) GetAllProviders(ctx context.Context) ([]common.Address, []providerregistry.IProviderStorageProvider, error) {
	batchSize := 100
	offset := big.NewInt(0)
	var allIDs []common.Address
	var allProviders []providerregistry.IProviderStorageProvider
	for {
		ids, providers, err := g.GetProviders(ctx, offset, uint8(batchSize), OrderASC)
		if err != nil {
			return nil, nil, err
		}
		if len(ids) == 0 {
			break
		}
		allProviders = append(allProviders, providers...)
		allIDs = append(allIDs, ids...)
		if len(ids) < batchSize {
			break
		}
		offset.Add(offset, big.NewInt(int64(batchSize)))
	}
	return allIDs, allProviders, nil
}

func (g *ProviderRegistry) GetProviders(ctx context.Context, offset *big.Int, limit uint8, order Order) ([]common.Address, []providerregistry.IProviderStorageProvider, error) {
	_, len, err := g.providerRegistry.GetActiveProviders(&bind.CallOpts{Context: ctx}, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return nil, nil, err
	}

	_offset, _limit := adjustPagination(order, len, offset, limit)
	ids, _, err := g.providerRegistry.GetActiveProviders(&bind.CallOpts{Context: ctx}, _offset, _limit)
	if err != nil {
		return nil, nil, err
	}

	adjustOrder(order, ids)
	return g.getMultipleProviders(ctx, ids)
}

func (g *ProviderRegistry) CreateNewProvider(opts *bind.TransactOpts, addStake *lib.BigInt, endpoint string) error {
	providerTx, err := g.providerRegistry.ProviderRegister(opts, opts.From, &addStake.Int, endpoint)

	if err != nil {
		return lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(opts.Context, g.client, providerTx)
	if err != nil {
		return lib.TryConvertGethError(err)
	}

	if receipt.Status != 1 {
		return fmt.Errorf("Transaction failed with status %d", receipt.Status)
	}

	return nil
}

func (g *ProviderRegistry) DeregisterProvider(opts *bind.TransactOpts) (common.Hash, error) {
	providerTx, err := g.providerRegistry.ProviderDeregister(opts, opts.From)

	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(opts.Context, g.client, providerTx)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	if receipt.Status != 1 {
		return receipt.TxHash, fmt.Errorf("Transaction failed with status %d", receipt.Status)
	}

	return receipt.TxHash, nil
}

func (g *ProviderRegistry) GetProviderById(ctx context.Context, id common.Address) (*providerregistry.IProviderStorageProvider, error) {
	provider, err := g.providerRegistry.GetProvider(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return nil, err
	}

	return &provider, nil
}

func (g *ProviderRegistry) getMultipleProviders(ctx context.Context, IDs []common.Address) ([]common.Address, []providerregistry.IProviderStorageProvider, error) {
	args := make([][]interface{}, len(IDs))
	for i, id := range IDs {
		args[i] = []interface{}{id}
	}
	providers, err := mc.Batch[providerregistry.IProviderStorageProvider](ctx, g.multicall, g.providerRegistryAbi, g.providerRegistryAddr, "getProvider", args)
	if err != nil {
		return nil, nil, err
	}
	return IDs, providers, nil
}
