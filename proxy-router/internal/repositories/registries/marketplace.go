package registries

import (
	"context"
	"fmt"
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/marketplace"
	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Marketplace struct {
	// config
	marketplaceAddr common.Address

	// deps
	marketplace *marketplace.Marketplace
	client      i.ContractBackend
	log         lib.ILogger
}

func NewMarketplace(marketplaceAddr common.Address, client i.ContractBackend, log lib.ILogger) *Marketplace {
	mp, err := marketplace.NewMarketplace(marketplaceAddr, client)
	if err != nil {
		panic("invalid marketplace ABI")
	}
	return &Marketplace{
		marketplace:     mp,
		marketplaceAddr: marketplaceAddr,
		client:          client,
		log:             log,
	}
}

func (g *Marketplace) PostModelBid(opts *bind.TransactOpts, provider common.Address, model common.Hash, pricePerSecond *big.Int) error {
	tx, err := g.marketplace.PostModelBid(opts, provider, model, pricePerSecond)
	if err != nil {
		return lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(opts.Context, g.client, tx)

	if err != nil {
		return lib.TryConvertGethError(err)
	}

	// Find the event log
	for _, log := range receipt.Logs {
		// Check if the log belongs to the OpenSession event
		_, err := g.marketplace.ParseBidPosted(*log)

		if err != nil {
			continue // not our event, skip it
		}

		return nil
	}

	return fmt.Errorf("PostModelBid event not found in transaction logs")
}

func (g *Marketplace) DeleteBid(opts *bind.TransactOpts, bidId common.Hash) (common.Hash, error) {
	tx, err := g.marketplace.DeleteModelBid(opts, bidId)
	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(opts.Context, g.client, tx)

	if err != nil {
		return common.Hash{}, lib.TryConvertGethError(err)
	}

	// Find the event log
	for _, log := range receipt.Logs {
		// Check if the log belongs to the OpenSession event
		_, err := g.marketplace.ParseBidDeleted(*log)

		if err != nil {
			continue // not our event, skip it
		}

		return tx.Hash(), nil
	}

	return common.Hash{}, fmt.Errorf("BidDeleted event not found in transaction logs")
}

func (g *Marketplace) GetBidById(ctx context.Context, bidId [32]byte) (*marketplace.IBidStorageBid, error) {
	bid, err := g.marketplace.Bids(&bind.CallOpts{Context: ctx}, bidId)
	if err != nil {
		return nil, err
	}
	return &bid, nil
}

func (g *Marketplace) GetBestBidByModelId(ctx context.Context, modelId common.Hash) (common.Hash, *marketplace.IBidStorageBid, error) {
	limit := big.NewInt(100)
	offset := big.NewInt(0)

	bidIDs, err := g.marketplace.ModelActiveBids(&bind.CallOpts{Context: ctx}, modelId, offset, limit)
	if err != nil {
		return common.Hash{}, nil, err
	}

	var cheapestBidID common.Hash
	var cheapestBid *marketplace.IBidStorageBid

	for _, bidID := range bidIDs {
		bid, err := g.marketplace.Bids(&bind.CallOpts{Context: ctx}, bidID)
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

func (g *Marketplace) GetBidsByProvider(ctx context.Context, provider common.Address, offset *big.Int, limit uint8) ([][32]byte, []marketplace.IBidStorageBid, error) {
	bidIDs, err := g.marketplace.ProviderBids(&bind.CallOpts{Context: ctx}, provider, offset, big.NewInt(int64(limit)))
	if err != nil {
		return nil, nil, err
	}

	return g.getMultipleBids(ctx, bidIDs)
}

func (g *Marketplace) GetBidsByModelAgent(ctx context.Context, modelAgentId common.Hash, offset *big.Int, limit uint8) ([][32]byte, []marketplace.IBidStorageBid, error) {
	bidIDs, err := g.marketplace.ModelBids(&bind.CallOpts{Context: ctx}, modelAgentId, offset, big.NewInt(int64(limit)))
	if err != nil {
		return nil, nil, err
	}
	return g.getMultipleBids(ctx, bidIDs)
}

func (g *Marketplace) GetActiveBidsByProvider(ctx context.Context, provider common.Address, offset *big.Int, limit uint8) ([][32]byte, []marketplace.IBidStorageBid, error) {
	bidIDs, err := g.marketplace.ProviderActiveBids(&bind.CallOpts{Context: ctx}, provider, offset, big.NewInt(int64(limit)))
	if err != nil {
		return nil, nil, err
	}
	return g.getMultipleBids(ctx, bidIDs)
}

func (g *Marketplace) GetActiveBidsByModel(ctx context.Context, modelAgentId common.Hash, offset *big.Int, limit uint8) ([][32]byte, []marketplace.IBidStorageBid, error) {
	bidIDs, err := g.marketplace.ModelActiveBids(&bind.CallOpts{Context: ctx}, modelAgentId, offset, big.NewInt(int64(limit)))
	if err != nil {
		return nil, nil, err
	}
	return g.getMultipleBids(ctx, bidIDs)
}

func (g *Marketplace) getMultipleBids(ctx context.Context, IDs [][32]byte) ([][32]byte, []marketplace.IBidStorageBid, error) {
	// todo: replace with multicall
	bids := make([]marketplace.IBidStorageBid, len(IDs))
	for i, id := range IDs {
		bid, err := g.marketplace.Bids(&bind.CallOpts{Context: ctx}, id)
		if err != nil {
			return nil, nil, err
		}
		bids[i] = bid
	}
	return IDs, bids, nil
}
