package registries

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/providerregistry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ProviderRegistry struct {
	// config
	providerRegistryAddr common.Address

	// state
	nonce uint64
	mutex lib.Mutex
	prABI *abi.ABI

	// deps
	providerRegistry *providerregistry.ProviderRegistry
	client           *ethclient.Client
	log              interfaces.ILogger
}

func NewProviderRegistry(providerRegistryAddr common.Address, client *ethclient.Client, log interfaces.ILogger) *ProviderRegistry {
	pr, err := providerregistry.NewProviderRegistry(providerRegistryAddr, client)
	if err != nil {
		panic("invalid provider registry ABI")
	}
	prABI, err := providerregistry.ProviderRegistryMetaData.GetAbi()
	if err != nil {
		panic("invalid provider registry ABI: " + err.Error())
	}
	return &ProviderRegistry{
		providerRegistry:     pr,
		providerRegistryAddr: providerRegistryAddr,
		client:               client,
		prABI:                prABI,
		mutex:                lib.NewMutex(),
		log:                  log,
	}
}

func (g *ProviderRegistry) GetAllProviders(ctx context.Context) ([]string, []providerregistry.Provider, error) {
	providerAddrs, providers, err := g.providerRegistry.ProviderGetAll(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, nil, err
	}

	addresses := make([]string, len(providerAddrs))
	for i, address := range providerAddrs {
		addresses[i] = address.Hex()
	}

	return addresses, providers, nil
}
