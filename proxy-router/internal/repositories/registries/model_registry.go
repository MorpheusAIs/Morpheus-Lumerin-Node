package registries

import (
	"context"
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/modelregistry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ModelRegistry struct {
	// config
	modelRegistryAddr common.Address

	// state
	nonce uint64

	// deps
	modelRegistry *modelregistry.ModelRegistry
	client        *ethclient.Client
	log           lib.ILogger
}

func NewModelRegistry(modelRegistryAddr common.Address, client *ethclient.Client, log lib.ILogger) *ModelRegistry {
	mr, err := modelregistry.NewModelRegistry(modelRegistryAddr, client)
	if err != nil {
		panic("invalid model registry ABI")
	}
	return &ModelRegistry{
		modelRegistry:     mr,
		modelRegistryAddr: modelRegistryAddr,
		client:            client,
		log:               log,
	}
}

func (g *ModelRegistry) GetAllModels(ctx context.Context) ([][32]byte, []modelregistry.Model, error) {
	adresses, models, err := g.modelRegistry.ModelGetAll(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, nil, err
	}

	return adresses, models, nil
}

func (g *ModelRegistry) CreateNewModel(ctx *bind.TransactOpts, modelId common.Hash, ipfsID common.Hash, fee *lib.BigInt, stake *lib.BigInt, owner common.Address, name string, tags []string) error {
	tx, err := g.modelRegistry.ModelRegister(ctx, modelId, ipfsID, &fee.Int, &stake.Int, owner, name, tags)
	if err != nil {
		return lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(context.Background(), g.client, tx)
	if err != nil {
		return lib.TryConvertGethError(err)
	}

	// Find the event log
	for _, log := range receipt.Logs {
		_, err := g.modelRegistry.ParseModelRegisteredUpdated(*log)

		if err != nil {
			continue // not our event, skip it
		}

		return nil
	}

	return fmt.Errorf("ModelRegistered event not found in transaction logs")
}

func (g *ModelRegistry) DeregisterModel(ctx *bind.TransactOpts, modelId common.Hash) (common.Hash, error) {
	tx, err := g.modelRegistry.ModelDeregister(ctx, modelId)

	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(context.Background(), g.client, tx)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Find the event log
	for _, log := range receipt.Logs {
		_, err := g.modelRegistry.ParseModelDeregistered(*log)

		if err != nil {
			continue // not our event, skip it
		}

		return tx.Hash(), nil
	}

	return common.Hash{}, fmt.Errorf("ModelDeregistered event not found in transaction logs")
}

func (g *ModelRegistry) GetModelById(ctx context.Context, modelId common.Hash) (*modelregistry.Model, error) {
	model, err := g.modelRegistry.ModelMap(&bind.CallOpts{Context: ctx}, modelId)
	if err != nil {
		return nil, err
	}
	return &model, nil
}
