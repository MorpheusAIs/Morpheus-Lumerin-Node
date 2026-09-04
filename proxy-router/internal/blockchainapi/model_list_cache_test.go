package blockchainapi

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestModelListCacheCoalescesConcurrentFetches(t *testing.T) {
	var cache modelListCache
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	load := func(context.Context) ([]*structs.Model, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-release
		return []*structs.Model{{Id: common.HexToHash("0x1"), Name: "cached"}}, nil
	}

	const waiterCount = 8
	results := make(chan []*structs.Model, waiterCount)
	errs := make(chan error, waiterCount)
	var waiters sync.WaitGroup
	waiters.Add(waiterCount)
	for range waiterCount {
		go func() {
			defer waiters.Done()
			models, err := cache.get(context.Background(), load)
			results <- models
			errs <- err
		}()
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("model fetch did not start")
	}
	close(release)
	waiters.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	for models := range results {
		require.Len(t, models, 1)
		require.Equal(t, "cached", models[0].Name)
	}
	require.EqualValues(t, 1, calls.Load())
}

func TestModelListCacheReturnsDefensiveCopies(t *testing.T) {
	var cache modelListCache
	var calls atomic.Int32
	load := func(context.Context) ([]*structs.Model, error) {
		calls.Add(1)
		return []*structs.Model{{
			Name:      "original",
			Fee:       big.NewInt(10),
			Stake:     big.NewInt(20),
			CreatedAt: big.NewInt(30),
			Tags:      []string{"llm", "tools"},
		}}, nil
	}

	first, err := cache.get(context.Background(), load)
	require.NoError(t, err)
	first[0].Name = "mutated"
	first[0].Fee.SetInt64(100)
	first[0].Stake.SetInt64(200)
	first[0].CreatedAt.SetInt64(300)
	first[0].Tags[0] = "changed"

	second, err := cache.get(context.Background(), load)
	require.NoError(t, err)
	require.Equal(t, "original", second[0].Name)
	require.Equal(t, int64(10), second[0].Fee.Int64())
	require.Equal(t, int64(20), second[0].Stake.Int64())
	require.Equal(t, int64(30), second[0].CreatedAt.Int64())
	require.Equal(t, []string{"llm", "tools"}, second[0].Tags)
	require.EqualValues(t, 1, calls.Load())
}

func TestModelListCacheSerializesTagsAsJSONArrays(t *testing.T) {
	var cache modelListCache
	models, err := cache.get(context.Background(), func(context.Context) ([]*structs.Model, error) {
		return []*structs.Model{
			{Name: "empty-tags", Tags: []string{}},
			{Name: "nil-tags", Tags: nil},
		}, nil
	})
	require.NoError(t, err)
	for _, model := range models {
		require.NotNil(t, model.Tags)
		require.Empty(t, model.Tags)
	}

	cached, err := cache.get(context.Background(), func(context.Context) ([]*structs.Model, error) {
		t.Fatal("loader must not run on a cache hit")
		return nil, nil
	})
	require.NoError(t, err)
	for _, model := range cached {
		require.NotNil(t, model.Tags)
		require.Empty(t, model.Tags)
	}

	encoded, err := json.Marshal(cached)
	require.NoError(t, err)
	require.Equal(t, 2, strings.Count(string(encoded), `"Tags":[]`))
}

func TestModelListCacheInvalidationStartsANewGeneration(t *testing.T) {
	var cache modelListCache
	oldStarted := make(chan struct{})
	releaseOld := make(chan struct{})
	var calls atomic.Int32
	load := func(context.Context) ([]*structs.Model, error) {
		call := calls.Add(1)
		if call == 1 {
			close(oldStarted)
			<-releaseOld
			return []*structs.Model{{Name: "old"}}, nil
		}
		return []*structs.Model{{Name: "fresh"}}, nil
	}

	oldResult := make(chan []*structs.Model, 1)
	go func() {
		models, _ := cache.get(context.Background(), load)
		oldResult <- models
	}()
	select {
	case <-oldStarted:
	case <-time.After(time.Second):
		t.Fatal("old model fetch did not start")
	}

	cache.invalidate()
	fresh, err := cache.get(context.Background(), load)
	require.NoError(t, err)
	require.Equal(t, "fresh", fresh[0].Name)

	close(releaseOld)
	require.Equal(t, "old", (<-oldResult)[0].Name)

	cached, err := cache.get(context.Background(), load)
	require.NoError(t, err)
	require.Equal(t, "fresh", cached[0].Name)
	require.EqualValues(t, 2, calls.Load())
}

func TestModelListCacheCanceledWaiterDoesNotCancelSharedFetch(t *testing.T) {
	var cache modelListCache
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	load := func(ctx context.Context) ([]*structs.Model, error) {
		calls.Add(1)
		close(started)
		select {
		case <-release:
			return []*structs.Model{{Name: "available"}}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	firstCtx, cancelFirst := context.WithCancel(context.Background())
	firstErr := make(chan error, 1)
	go func() {
		_, err := cache.get(firstCtx, load)
		firstErr <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("shared model fetch did not start")
	}
	cancelFirst()
	require.ErrorIs(t, <-firstErr, context.Canceled)

	secondResult := make(chan []*structs.Model, 1)
	secondErr := make(chan error, 1)
	go func() {
		models, err := cache.get(context.Background(), load)
		secondResult <- models
		secondErr <- err
	}()
	close(release)

	require.NoError(t, <-secondErr)
	require.Equal(t, "available", (<-secondResult)[0].Name)
	require.EqualValues(t, 1, calls.Load())
}

func TestModelListCacheExpiresAfterTTL(t *testing.T) {
	now := time.Unix(1_000, 0)
	cache := modelListCache{
		ttl: 10 * time.Second,
		now: func() time.Time { return now },
	}
	var calls atomic.Int32
	load := func(context.Context) ([]*structs.Model, error) {
		call := calls.Add(1)
		return []*structs.Model{{Name: "version-" + big.NewInt(int64(call)).String()}}, nil
	}

	first, err := cache.get(context.Background(), load)
	require.NoError(t, err)
	require.Equal(t, "version-1", first[0].Name)

	now = now.Add(9 * time.Second)
	stillCached, err := cache.get(context.Background(), load)
	require.NoError(t, err)
	require.Equal(t, "version-1", stillCached[0].Name)

	now = now.Add(time.Second)
	refreshed, err := cache.get(context.Background(), load)
	require.NoError(t, err)
	require.Equal(t, "version-2", refreshed[0].Name)
	require.EqualValues(t, 2, calls.Load())
}

func TestModelListCacheDoesNotCacheFetchErrors(t *testing.T) {
	var cache modelListCache
	wantErr := errors.New("registry unavailable")
	var calls atomic.Int32
	load := func(context.Context) ([]*structs.Model, error) {
		if calls.Add(1) == 1 {
			return nil, wantErr
		}
		return []*structs.Model{{Name: "recovered"}}, nil
	}

	_, err := cache.get(context.Background(), load)
	require.ErrorIs(t, err, wantErr)

	models, err := cache.get(context.Background(), load)
	require.NoError(t, err)
	require.Equal(t, "recovered", models[0].Name)
	require.EqualValues(t, 2, calls.Load())
}
