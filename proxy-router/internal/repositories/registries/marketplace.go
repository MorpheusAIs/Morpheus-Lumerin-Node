package registries

import (
	"context"
	"fmt"
	"math/big"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/marketplace"
	mc "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/multicall"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Marketplace struct {
	// config
	marketplaceAddr common.Address

	// deps
	marketplace    *marketplace.Marketplace
	multicall      mc.MulticallBackend
	marketplaceABI *abi.ABI
	client         i.ContractBackend
	log            lib.ILogger
}

func NewMarketplace(marketplaceAddr common.Address, client i.ContractBackend, multicall mc.MulticallBackend, log lib.ILogger) *Marketplace {
	mp, err := marketplace.NewMarketplace(marketplaceAddr, client)
	if err != nil {
		panic("invalid marketplace ABI")
	}
	marketplaceABI, err := marketplace.MarketplaceMetaData.GetAbi()
	if err != nil {
		panic("invalid marketplace ABI: " + err.Error())
	}

	return &Marketplace{
		marketplace:     mp,
		marketplaceAddr: marketplaceAddr,
		marketplaceABI:  marketplaceABI,
		multicall:       multicall,
		client:          client,
		log:             log,
	}
}

func (g *Marketplace) PostModelBid(opts *bind.TransactOpts, model common.Hash, pricePerSecond *big.Int) (common.Hash, error) {
	tx, err := g.marketplace.PostModelBid(opts, opts.From, model, pricePerSecond)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(opts.Context, g.client, tx)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	for _, log := range receipt.Logs {
		event, err := g.marketplace.ParseMarketplaceBidPosted(*log)
		if err == nil {
			bidId, errBid := g.marketplace.GetBidId(&bind.CallOpts{Context: opts.Context}, event.Provider, event.ModelId, event.Nonce)
			if errBid == nil {
				return bidId, nil
			}
		}
	}

	return common.Hash{}, nil
}

func (g *Marketplace) DeleteBid(opts *bind.TransactOpts, bidID common.Hash) (common.Hash, error) {
	tx, err := g.marketplace.DeleteModelBid(opts, bidID)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(opts.Context, g.client, tx)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	if receipt.Status != 1 {
		return receipt.TxHash, fmt.Errorf("Transaction failed")
	}

	return receipt.TxHash, nil
}

func (g *Marketplace) GetBidById(ctx context.Context, bidID common.Hash) (*marketplace.IBidStorageBid, error) {
	bid, err := g.marketplace.GetBid(&bind.CallOpts{Context: ctx}, bidID)
	if err != nil {
		return nil, err
	}
	return &bid, nil
}

func (g *Marketplace) GetBestBidByModelId(ctx context.Context, modelID common.Hash) (common.Hash, *marketplace.IBidStorageBid, error) {
	limit := big.NewInt(100)
	offset := big.NewInt(0)

	bidIDs, _, err := g.marketplace.GetModelActiveBids(&bind.CallOpts{Context: ctx}, modelID, offset, limit)
	if err != nil {
		return common.Hash{}, nil, err
	}

	var cheapestBidID common.Hash
	var cheapestBid *marketplace.IBidStorageBid

	for _, bidID := range bidIDs {
		bid, err := g.marketplace.GetBid(&bind.CallOpts{Context: ctx}, bidID)
		if err != nil {
			return common.Hash{}, nil, err
		}
		if cheapestBid == nil {
			cheapestBid = &bid
			cheapestBidID = bidID

		} else if bid.PricePerSecond.Cmp(cheapestBid.PricePerSecond) < 0 {
			cheapestBid = &bid
			cheapestBidID = bidID
		}
	}

	return cheapestBidID, cheapestBid, nil
}

func (g *Marketplace) GetBidsByProvider(ctx context.Context, provider common.Address, offset *big.Int, limit uint8, order Order) ([][32]byte, []marketplace.IBidStorageBid, error) {
	_, len, err := g.marketplace.GetProviderBids(&bind.CallOpts{Context: ctx}, provider, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return nil, nil, err
	}

	_offset, _limit := adjustPagination(order, len, offset, limit)
	bidIDs, _, err := g.marketplace.GetProviderBids(&bind.CallOpts{Context: ctx}, provider, _offset, _limit)
	if err != nil {
		return nil, nil, err
	}

	adjustOrder(order, bidIDs)
	return g.GetMultipleBids(ctx, bidIDs)
}

func (g *Marketplace) GetBidsByModelAgent(ctx context.Context, modelAgentId common.Hash, offset *big.Int, limit uint8, order Order) ([][32]byte, []marketplace.IBidStorageBid, error) {
	_, len, err := g.marketplace.GetModelBids(&bind.CallOpts{Context: ctx}, modelAgentId, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return nil, nil, err
	}
	_offset, _limit := adjustPagination(order, len, offset, limit)
	bidIDs, _, err := g.marketplace.GetModelBids(&bind.CallOpts{Context: ctx}, modelAgentId, _offset, _limit)
	if err != nil {
		return nil, nil, err
	}
	adjustOrder(order, bidIDs)
	return g.GetMultipleBids(ctx, bidIDs)
}

func (g *Marketplace) GetActiveBidsByProvider(ctx context.Context, provider common.Address, offset *big.Int, limit uint8, order Order) ([][32]byte, []marketplace.IBidStorageBid, error) {
	_, len, err := g.marketplace.GetProviderActiveBids(&bind.CallOpts{Context: ctx}, provider, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return nil, nil, err
	}

	_offset, _limit := adjustPagination(order, len, offset, limit)
	bidIDs, _, err := g.marketplace.GetProviderActiveBids(&bind.CallOpts{Context: ctx}, provider, _offset, _limit)
	if err != nil {
		return nil, nil, err
	}

	adjustOrder(order, bidIDs)
	return g.GetMultipleBids(ctx, bidIDs)
}

func (g *Marketplace) GetActiveBidsByModel(ctx context.Context, modelAgentId common.Hash, offset *big.Int, limit uint8, order Order) ([][32]byte, []marketplace.IBidStorageBid, error) {
	_, len, err := g.marketplace.GetModelActiveBids(&bind.CallOpts{Context: ctx}, modelAgentId, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return nil, nil, err
	}
	_offset, _limit := adjustPagination(order, len, offset, limit)
	bidIDs, _, err := g.marketplace.GetModelActiveBids(&bind.CallOpts{Context: ctx}, modelAgentId, _offset, _limit)
	if err != nil {
		return nil, nil, err
	}
	adjustOrder(order, bidIDs)
	return g.GetMultipleBids(ctx, bidIDs)
}

func (g *Marketplace) GetMultipleBids(ctx context.Context, IDs [][32]byte) ([][32]byte, []marketplace.IBidStorageBid, error) {
	args := make([][]interface{}, len(IDs))
	for i, id := range IDs {
		args[i] = []interface{}{id}
	}
	bids, err := mc.Batch[marketplace.IBidStorageBid](ctx, g.multicall, g.marketplaceABI, g.marketplaceAddr, "getBid", args)
	if err != nil {
		return nil, nil, err
	}
	return IDs, bids, nil
}
