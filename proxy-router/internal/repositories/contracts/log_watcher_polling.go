package contracts

import (
	"context"
	"errors"
	"math/big"
	"time"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	SubClosedError = errors.New("subscription closed")
)

type LogWatcherPolling struct {
	// config
	maxReconnects int
	pollInterval  time.Duration

	// deps
	client i.EthClient
	log    lib.ILogger
}

func NewLogWatcherPolling(client i.EthClient, pollInterval time.Duration, maxReconnects int, log lib.ILogger) *LogWatcherPolling {
	return &LogWatcherPolling{
		client:        client,
		pollInterval:  pollInterval,
		maxReconnects: maxReconnects,
		log:           log,
	}
}

func (w *LogWatcherPolling) Watch(ctx context.Context, contractAddr common.Address, mapper EventMapper, fromBlock *big.Int) (*lib.Subscription, error) {
	var nextFromBlock *big.Int

	if fromBlock != nil {
		nextFromBlock = new(big.Int).Set(fromBlock)
	}

	sink := make(chan interface{})
	return lib.NewSubscription(func(quit <-chan struct{}) error {
		defer close(sink)

		for {
			currentBlock, err := w.client.HeaderByNumber(ctx, nil)
			if err != nil {
				return err
			}

			if nextFromBlock == nil {
				nextFromBlock = new(big.Int).Set(currentBlock.Number)
			}

			// if we poll too often, we might be behind the chain, so we wait for the next block
			if currentBlock.Number.Cmp(nextFromBlock) < 0 {
				select {
				case <-quit:
					return SubClosedError
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(w.pollInterval):
				}
				continue
			}

			query := ethereum.FilterQuery{
				Addresses: []common.Address{contractAddr},
				FromBlock: nextFromBlock,
				ToBlock:   currentBlock.Number,
			}
			sub, err := w.filterLogsRetry(ctx, query, quit)
			if err != nil {
				return err
			}

			for _, log := range sub {
				if log.Removed {
					continue
				}
				event, err := mapper(log)
				if err != nil {
					w.log.Debugf("error mapping event: %s", err)
					continue // mapper error, retry won't help, but we can continue
				}

				select {
				case <-quit:
					return SubClosedError
				case <-ctx.Done():
					return ctx.Err()
				case sink <- event:
				}
			}

			nextFromBlock.Add(currentBlock.Number, big.NewInt(1))

			select {
			case <-quit:
				return SubClosedError
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(w.pollInterval):
			}
		}
	}, sink), nil
}

func (w *LogWatcherPolling) filterLogsRetry(ctx context.Context, query ethereum.FilterQuery, quit <-chan struct{}) ([]types.Log, error) {
	var lastErr error

	for attempts := 0; attempts < w.maxReconnects; attempts++ {
		logs, err := w.client.FilterLogs(ctx, query)
		if err == nil {
			if attempts > 0 {
				w.log.Warnf("subscription successfully reconnected after error: %s", lastErr)
			}

			return logs, nil
		}

		w.log.Debugf("subscription error: %s, retrying in %s", err, w.pollInterval.String())
		lastErr = err

		select {
		case <-quit:
			return nil, SubClosedError
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(w.pollInterval):
		}
	}

	return nil, lastErr
}