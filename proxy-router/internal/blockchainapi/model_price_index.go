package blockchainapi

import (
	"context"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	r "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/singleflight"
)

const (
	// modelPricesCacheTTL is longer than the model list's own TTL because this
	// is a far more expensive answer to rebuild: one bid read per registered
	// model. Bid prices are posted by providers and do not move minute to
	// minute, so a minute of staleness costs a user nothing while rebuilding on
	// every request would hammer the RPC endpoint.
	modelPricesCacheTTL = time.Minute
	// The whole sweep is bounded so a slow RPC cannot leave the goroutines
	// running long after the last waiter has gone.
	modelPricesFetchTimeout = 2 * time.Minute
	// How many bid reads are in flight at once. Sequential would take a model
	// count's worth of round trips; unbounded would open hundreds of concurrent
	// requests and get the node rate-limited, which fails the sweep entirely.
	modelPricesConcurrency = 16
	// Bids are read one page deep. A page holds 255 live bids for a single
	// model, which no model on this marketplace approaches; paginating further
	// would multiply the request count of the sweep to guard against a case
	// that does not occur, and the consequence if it ever did is a price range
	// drawn from the first 255 providers rather than all of them.
	modelPricesBidPageSize uint8 = 255
)

// modelPriceIndexCache is the same shape as modelListCache and separate from it
// on purpose: prices and the registry change at different rates and one must
// not invalidate the other. Registering a model does not change any price, and
// a price sweep going stale is no reason to re-read the registry.
type modelPriceIndexCache struct {
	mu         sync.Mutex
	group      singleflight.Group
	generation uint64
	result     *structs.ModelPricesRes
	cachedAt   time.Time
	valid      bool

	// Tests can override these. Their zero values select production defaults.
	ttl time.Duration
	now func() time.Time
}

func (c *modelPriceIndexCache) get(
	ctx context.Context,
	load func(context.Context) (*structs.ModelPricesRes, error),
) (*structs.ModelPricesRes, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cached, generation, ok := c.cached()
	if ok {
		return cached, nil
	}

	result := c.group.DoChan(strconv.FormatUint(generation, 10), func() (any, error) {
		if cached, currentGeneration, hit := c.cached(); hit && currentGeneration == generation {
			return cached, nil
		}

		// Detached from the requesting context for the same reason the model
		// list cache detaches: one client hanging up must not fail the sweep
		// every other waiter is sharing.
		fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), modelPricesFetchTimeout)
		defer cancel()

		loaded, err := load(fetchCtx)
		if err != nil {
			return nil, err
		}

		loaded = cloneModelPrices(loaded)
		c.mu.Lock()
		if c.generation == generation {
			c.result = loaded
			c.cachedAt = c.currentTime()
			c.valid = true
		}
		c.mu.Unlock()

		return loaded, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case outcome := <-result:
		if outcome.Err != nil {
			return nil, outcome.Err
		}
		return cloneModelPrices(outcome.Val.(*structs.ModelPricesRes)), nil
	}
}

func (c *modelPriceIndexCache) cached() (*structs.ModelPricesRes, uint64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	generation := c.generation
	if !c.valid || !c.currentTime().Before(c.cachedAt.Add(c.cacheTTL())) {
		return nil, generation, false
	}

	return cloneModelPrices(c.result), generation, true
}

func (c *modelPriceIndexCache) cacheTTL() time.Duration {
	if c.ttl > 0 {
		return c.ttl
	}
	return modelPricesCacheTTL
}

func (c *modelPriceIndexCache) currentTime() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

// cloneModelPrices hands every caller its own copy. The cached value outlives
// the request that produced it, and a handler that sorted the shared slice in
// place would reorder it for everyone else.
func cloneModelPrices(source *structs.ModelPricesRes) *structs.ModelPricesRes {
	if source == nil {
		return nil
	}

	cloned := &structs.ModelPricesRes{
		// Always an array in JSON, never null: a client iterating the result
		// should not have to special-case an empty marketplace.
		Prices:         make([]structs.ModelPrice, len(source.Prices)),
		FailedModelIDs: make([]string, len(source.FailedModelIDs)),
	}
	copy(cloned.Prices, source.Prices)
	copy(cloned.FailedModelIDs, source.FailedModelIDs)
	return cloned
}

// GetModelPrices reports what every registered model currently costs per second
// across its live bids.
//
// The router's own bids are left out, matching every other path that offers a
// provider to this wallet: a user cannot open a session against their own bid,
// so a price they can never pay has no business setting the bottom of a range
// labelled "cheapest".
func (s *BlockchainService) GetModelPrices(ctx context.Context) (*structs.ModelPricesRes, error) {
	return s.modelPricesCache.get(ctx, func(fetchCtx context.Context) (*structs.ModelPricesRes, error) {
		models, err := s.GetAllModels(fetchCtx)
		if err != nil {
			return nil, err
		}

		// A wallet failure is not fatal here. Without an address the sweep
		// simply cannot exclude the router's own bids, which is a far smaller
		// problem than returning no prices at all.
		var ownAddress common.Address
		if address, addressErr := s.GetMyAddress(fetchCtx); addressErr == nil {
			ownAddress = address
		} else {
			s.log.Warnf("model price sweep could not read the wallet address, own bids will be included: %s", addressErr)
		}

		return s.sweepModelPrices(fetchCtx, models, ownAddress), nil
	})
}

func (s *BlockchainService) sweepModelPrices(
	ctx context.Context,
	models []*structs.Model,
	ownAddress common.Address,
) *structs.ModelPricesRes {
	type outcome struct {
		price  structs.ModelPrice
		failed bool
	}

	outcomes := make([]outcome, len(models))
	// A buffered channel as a counting semaphore: cheaper than a pool and it
	// keeps the result indexed by position, so the output order is the model
	// list's order and does not depend on which reads finish first.
	slots := make(chan struct{}, modelPricesConcurrency)
	var wait sync.WaitGroup

	for index, model := range models {
		if model == nil || model.Id == (common.Hash{}) {
			continue
		}

		wait.Add(1)
		slots <- struct{}{}
		go func(index int, modelID common.Hash) {
			defer wait.Done()
			defer func() { <-slots }()

			price, err := s.priceForModel(ctx, modelID, ownAddress)
			if err != nil {
				outcomes[index] = outcome{
					price:  structs.ModelPrice{ModelID: modelID.Hex()},
					failed: true,
				}
				return
			}
			outcomes[index] = outcome{price: price}
		}(index, model.Id)
	}

	wait.Wait()

	result := &structs.ModelPricesRes{
		Prices:         make([]structs.ModelPrice, 0, len(models)),
		FailedModelIDs: make([]string, 0),
	}
	for _, entry := range outcomes {
		if entry.price.ModelID == "" {
			continue
		}
		if entry.failed {
			result.FailedModelIDs = append(result.FailedModelIDs, entry.price.ModelID)
			continue
		}
		result.Prices = append(result.Prices, entry.price)
	}

	return result
}

func (s *BlockchainService) priceForModel(
	ctx context.Context,
	modelID common.Hash,
	ownAddress common.Address,
) (structs.ModelPrice, error) {
	price := structs.ModelPrice{ModelID: modelID.Hex()}

	bids, err := s.GetActiveBidsByModel(ctx, modelID, big.NewInt(0), modelPricesBidPageSize, r.OrderASC)
	if err != nil {
		return structs.ModelPrice{}, err
	}

	var min, max *big.Int
	for _, bid := range bids {
		if bid == nil || bid.PricePerSecond == nil {
			continue
		}
		if ownAddress != (common.Address{}) && bid.Provider == ownAddress {
			continue
		}
		// A zero price is not a bargain, it is a malformed bid: getSessionEnd
		// divides by it. Letting one in would sort that model to the top of
		// "cheapest" and then fail at the point of opening.
		value := bid.PricePerSecond.Unpack()
		if value == nil || value.Sign() <= 0 {
			continue
		}

		price.BidCount++
		if min == nil || value.Cmp(min) < 0 {
			min = value
		}
		if max == nil || value.Cmp(max) > 0 {
			max = value
		}
	}

	if min != nil {
		price.MinPricePerSecondWei = min.String()
		price.MaxPricePerSecondWei = max.String()
	}

	return price, nil
}

// SortModelPricesAscending orders an index cheapest first, with models that
// have no live provider last regardless of direction. It is exported for tests;
// clients do their own ordering.
func SortModelPricesAscending(prices []structs.ModelPrice) {
	sort.SliceStable(prices, func(i, j int) bool {
		left, leftOK := new(big.Int).SetString(prices[i].MinPricePerSecondWei, 10)
		right, rightOK := new(big.Int).SetString(prices[j].MinPricePerSecondWei, 10)
		if leftOK != rightOK {
			return leftOK
		}
		if !leftOK {
			return false
		}
		return left.Cmp(right) < 0
	})
}
