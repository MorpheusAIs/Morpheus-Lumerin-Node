package registries

import (
	"context"
	"fmt"
	"math/big"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/modelregistry"
	mc "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/multicall"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ModelRegistry struct {
	// config
	modelRegistryAddr common.Address

	// state
	nonce uint64

	// deps
	modelRegistry    *modelregistry.ModelRegistry
	modelRegistryAbi *abi.ABI
	multicall        mc.MulticallBackend
	client           i.ContractBackend
	log              lib.ILogger
}

func NewModelRegistry(modelRegistryAddr common.Address, client i.ContractBackend, multicall mc.MulticallBackend, log lib.ILogger) *ModelRegistry {
	mr, err := modelregistry.NewModelRegistry(modelRegistryAddr, client)
	if err != nil {
		panic("invalid model registry ABI")
	}
	mrAbi, err := modelregistry.ModelRegistryMetaData.GetAbi()
	if err != nil {
		panic("invalid model registry ABI")
	}
	return &ModelRegistry{
		modelRegistry:     mr,
		modelRegistryAddr: modelRegistryAddr,
		modelRegistryAbi:  mrAbi,
		multicall:         multicall,
		client:            client,
		log:               log,
	}
}

func (g *ModelRegistry) GetAllModels(ctx context.Context) ([][32]byte, []modelregistry.IModelStorageModel, error) {
	batchSize := 100
	offset := big.NewInt(0)
	var allIDs [][32]byte
	var allModels []modelregistry.IModelStorageModel
	for {
		ids, providers, err := g.GetModels(ctx, offset, uint8(batchSize), OrderASC)
		if err != nil {
			return nil, nil, err
		}
		if len(ids) == 0 {
			break
		}
		allModels = append(allModels, providers...)
		allIDs = append(allIDs, ids...)
		if len(ids) < batchSize {
			break
		}
		offset.Add(offset, big.NewInt(int64(batchSize)))
	}
	return allIDs, allModels, nil
}

func (g *ModelRegistry) GetModels(ctx context.Context, offset *big.Int, limit uint8, order Order) ([][32]byte, []modelregistry.IModelStorageModel, error) {
	_, len, err := g.modelRegistry.GetActiveModelIds(&bind.CallOpts{Context: ctx}, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return nil, nil, err
	}
	_offset, _limit := adjustPagination(order, len, offset, limit)
	ids, _, err := g.modelRegistry.GetActiveModelIds(&bind.CallOpts{Context: ctx}, _offset, _limit)
	if err != nil {
		return nil, nil, err
	}
	adjustOrder(order, ids)
	return g.getMultipleModels(ctx, ids)
}

func (g *ModelRegistry) CreateNewModel(opts *bind.TransactOpts, modelId common.Hash, ipfsID common.Hash, fee *lib.BigInt, stake *lib.BigInt, name string, tags []string) error {
	tx, err := g.modelRegistry.ModelRegister(opts, opts.From, modelId, ipfsID, &fee.Int, &stake.Int, name, tags)
	if err != nil {
		return lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt with timeout
	receipt, err := lib.WaitMinedWithTimeout(opts.Context, g.client, tx, lib.DefaultTxMineTimeout)
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

func (g *ModelRegistry) DeregisterModel(opts *bind.TransactOpts, modelId common.Hash) (common.Hash, error) {
	tx, err := g.modelRegistry.ModelDeregister(opts, modelId)

	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt with timeout
	receipt, err := lib.WaitMinedWithTimeout(opts.Context, g.client, tx, lib.DefaultTxMineTimeout)
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

func (g *ModelRegistry) GetModelById(ctx context.Context, modelId common.Hash) (*modelregistry.IModelStorageModel, error) {
	model, err := g.modelRegistry.GetModel(&bind.CallOpts{Context: ctx}, modelId)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (g *ModelRegistry) GetModelId(ctx context.Context, accountAddr common.Address, baseModelId common.Hash) (common.Hash, error) {
	id, err := g.modelRegistry.GetModelId(&bind.CallOpts{Context: ctx}, accountAddr, baseModelId)
	if err != nil {
		return common.Hash{}, err
	}
	return id, nil
}

func (g *ModelRegistry) getMultipleModels(ctx context.Context, IDs [][32]byte) ([][32]byte, []modelregistry.IModelStorageModel, error) {
	args := make([][]interface{}, len(IDs))
	for i, id := range IDs {
		args[i] = []interface{}{id}
	}
	models, err := mc.Batch[modelregistry.IModelStorageModel](ctx, g.multicall, g.modelRegistryAbi, g.modelRegistryAddr, "getModel", args)
	if err != nil {
		return nil, nil, err
	}
	return IDs, models, nil
}
