package registries

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/contracts/modelregistry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ModelRegistry struct {
	// config
	modelRegistryAddr common.Address

	// state
	nonce uint64
	mutex lib.Mutex
	mrABI *abi.ABI

	// deps
	modelRegistry *modelregistry.ModelRegistry
	client        *ethclient.Client
	log           interfaces.ILogger
}

func NewModelRegistry(modelRegistryAddr common.Address, client *ethclient.Client, log interfaces.ILogger) *ModelRegistry {
	mr, err := modelregistry.NewModelRegistry(modelRegistryAddr, client)
	if err != nil {
		panic("invalid model registry ABI")
	}
	mrABI, err := modelregistry.ModelRegistryMetaData.GetAbi()
	if err != nil {
		panic("invalid model registry ABI: " + err.Error())
	}
	return &ModelRegistry{
		modelRegistry:     mr,
		modelRegistryAddr: modelRegistryAddr,
		client:            client,
		mrABI:             mrABI,
		mutex:             lib.NewMutex(),
		log:               log,
	}
}

func (g *ModelRegistry) GetAllModels(ctx context.Context) ([]modelregistry.Model, error) {
	models, err := g.modelRegistry.ModelGetAll(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, err
	}

	return models, nil
}
