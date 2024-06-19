package registries

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/contracts/marketplace"
	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Marketplace struct {
	// config
	marketplaceAddr common.Address

	// state
	nonce uint64
	mutex lib.Mutex
	mpABI *abi.ABI

	// deps
	marketplace *marketplace.Marketplace
	client      *ethclient.Client
	log         interfaces.ILogger
}

func NewMarketplace(marketplaceAddr common.Address, client *ethclient.Client, log interfaces.ILogger) *Marketplace {
	mp, err := marketplace.NewMarketplace(marketplaceAddr, client)
	if err != nil {
		panic("invalid marketplace ABI")
	}
	mpABI, err := marketplace.MarketplaceMetaData.GetAbi()
	if err != nil {
		panic("invalid marketplace ABI: " + err.Error())
	}
	return &Marketplace{
		marketplace:     mp,
		marketplaceAddr: marketplaceAddr,
		client:          client,
		mpABI:           mpABI,
		mutex:           lib.NewMutex(),
		log:             log,
	}
}

func (g *Marketplace) PostModelBid(ctx *bind.TransactOpts, provider string, model [32]byte, pricePerSecond *big.Int) error {
	fmt.Println("PostModelBid", provider, model, pricePerSecond)
	tx, err := g.marketplace.PostModelBid(ctx, common.HexToAddress(provider), model, pricePerSecond)
	if err != nil {
		return lib.TryConvertGethError(err, marketplace.MarketplaceMetaData)
	}

	// Wait for the transaction receipt
	receipt, err := bind.WaitMined(context.Background(), g.client, tx)

	if err != nil {
		return lib.TryConvertGethError(err, marketplace.MarketplaceMetaData)
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

func (g *Marketplace) GetBidById(ctx context.Context, bidId [32]byte) (*marketplace.Bid, error) {
	bid, err := g.marketplace.BidMap(&bind.CallOpts{Context: ctx}, bidId)
	if err != nil {
		return nil, err
	}
	return &bid, nil
}

func (g *Marketplace) GetBidsByProvider(ctx context.Context, provider common.Address, offset *big.Int, limit uint8) ([][32]byte, []marketplace.Bid, error) {
	adresses, bids, err := g.marketplace.GetBidsByProvider(&bind.CallOpts{Context: ctx}, provider, offset, limit)
	if err != nil {
		return nil, nil, err
	}

	return adresses, bids, nil
}

func (g *Marketplace) GetBidsByModelAgent(ctx context.Context, modelAgentId [32]byte, offset *big.Int, limit uint8) ([][32]byte, []marketplace.Bid, error) {
	addresses, bids, err := g.marketplace.GetBidsByModelAgent(&bind.CallOpts{Context: ctx}, modelAgentId, offset, limit)
	if err != nil {
		return nil, nil, err
	}

	return addresses, bids, nil
}
