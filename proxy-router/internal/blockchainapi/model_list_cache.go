package blockchainapi

import (
	"context"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"golang.org/x/sync/singleflight"
)

const (
	// allModelsCacheTTL keeps the on-chain model registry responsive across
	// adjacent UI routes without hiding registry changes for long.
	allModelsCacheTTL = 30 * time.Second
	// A shared fetch is detached from any one HTTP request so a canceled first
	// waiter cannot fail every concurrent waiter. Still bound it to avoid a
	// stalled RPC living forever after all waiters leave.
	allModelsFetchTimeout = time.Minute
)

type modelListCache struct {
	mu         sync.Mutex
	group      singleflight.Group
	generation uint64
	models     []*structs.Model
	cachedAt   time.Time
	valid      bool

	// Tests can override these. Their zero values select production defaults.
	ttl time.Duration
	now func() time.Time
}

func (c *modelListCache) get(
	ctx context.Context,
	load func(context.Context) ([]*structs.Model, error),
) ([]*structs.Model, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	models, generation, ok := c.cached()
	if ok {
		return models, nil
	}

	result := c.group.DoChan(strconv.FormatUint(generation, 10), func() (any, error) {
		// Another caller may have populated this generation between our cache
		// check and joining the singleflight call.
		if models, currentGeneration, cacheHit := c.cached(); cacheHit && currentGeneration == generation {
			return models, nil
		}

		fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), allModelsFetchTimeout)
		defer cancel()

		models, err := load(fetchCtx)
		if err != nil {
			return nil, err
		}

		models = cloneModels(models)
		c.mu.Lock()
		if c.generation == generation {
			c.models = models
			c.cachedAt = c.currentTime()
			c.valid = true
		}
		c.mu.Unlock()

		return models, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case outcome := <-result:
		if outcome.Err != nil {
			return nil, outcome.Err
		}
		return cloneModels(outcome.Val.([]*structs.Model)), nil
	}
}

func (c *modelListCache) cached() ([]*structs.Model, uint64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	generation := c.generation
	if !c.valid || !c.currentTime().Before(c.cachedAt.Add(c.cacheTTL())) {
		return nil, generation, false
	}

	return cloneModels(c.models), generation, true
}

func (c *modelListCache) invalidate() {
	c.mu.Lock()
	c.generation++
	c.models = nil
	c.cachedAt = time.Time{}
	c.valid = false
	c.mu.Unlock()
}

func (c *modelListCache) cacheTTL() time.Duration {
	if c.ttl > 0 {
		return c.ttl
	}
	return allModelsCacheTTL
}

func (c *modelListCache) currentTime() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func cloneModels(models []*structs.Model) []*structs.Model {
	if models == nil {
		return nil
	}

	cloned := make([]*structs.Model, len(models))
	for index, model := range models {
		if model == nil {
			continue
		}

		copyOfModel := *model
		copyOfModel.Fee = cloneBigInt(model.Fee)
		copyOfModel.Stake = cloneBigInt(model.Stake)
		copyOfModel.CreatedAt = cloneBigInt(model.CreatedAt)
		copyOfModel.Tags = append([]string(nil), model.Tags...)
		cloned[index] = &copyOfModel
	}

	return cloned
}

func cloneBigInt(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
