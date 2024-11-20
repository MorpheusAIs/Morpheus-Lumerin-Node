package contracts

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
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

		for { // infinite polling loop
			nextFrom, err := w.pollReconnect(ctx, quit, nextFromBlock, contractAddr, mapper, sink)
			if err != nil {
				return err
			}
			nextFromBlock = nextFrom
		}
	}, sink), nil
}

func (w *LogWatcherPolling) pollReconnect(ctx context.Context, quit <-chan struct{}, nextFromBlock *big.Int, contractAddr common.Address, mapper EventMapper, sink chan interface{}) (*big.Int, error) {
	var lastErr error
	for i := 0; i < w.maxReconnects || w.maxReconnects == 0; i++ {
		// for any of those cases, we should stop retrying
		select {
		case <-quit:
			return nextFromBlock, SubClosedError
		case <-ctx.Done():
			return nextFromBlock, ctx.Err()
		default:
		}
		newNextFromBlock, err := w.pollChanges(ctx, nextFromBlock, quit, contractAddr, mapper, sink)
		if err == nil {
			return newNextFromBlock, nil
		}
		maxReconnects := fmt.Sprintf("%d", w.maxReconnects)
		if w.maxReconnects == 0 {
			maxReconnects = "∞"
		}
		w.log.Warnf("polling error, retrying (%d/%s): %s", i, maxReconnects, err)
		lastErr = err

		// retry delay
		select {
		case <-quit:
			return nextFromBlock, SubClosedError
		case <-ctx.Done():
			return nextFromBlock, ctx.Err()
		case <-time.After(w.pollInterval):
		}
	}

	err := fmt.Errorf("polling error, retries exhausted (%d), stopping: %s", w.maxReconnects, lastErr)
	w.log.Warnf(err.Error())
	return nextFromBlock, err
}

func (w *LogWatcherPolling) pollChanges(ctx context.Context, nextFromBlock *big.Int, quit <-chan struct{}, contractAddr common.Address, mapper EventMapper, sink chan interface{}) (*big.Int, error) {
	currentBlock, err := w.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	if nextFromBlock == nil {
		nextFromBlock = new(big.Int).Set(currentBlock.Number)
	}

	// if we poll too often, we might be behind the chain, so we wait for the next block
	// mapper error, retry won't help, but we can continue
	if currentBlock.Number.Cmp(nextFromBlock) < 0 {
		select {
		case <-quit:
			return nextFromBlock, SubClosedError
		case <-ctx.Done():
			return nextFromBlock, ctx.Err()
		case <-time.After(w.pollInterval):
		}
		return nextFromBlock, nil
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
		FromBlock: nextFromBlock,
		ToBlock:   currentBlock.Number,
	}

	w.log.Debugf("=====> calling poll from %s to %s", query.FromBlock.String(), query.ToBlock.String())
	sub, err := w.client.FilterLogs(ctx, query)
	if err != nil {
		return nextFromBlock, err
	}

	for _, log := range sub {
		if log.Removed {
			continue
		}
		event, err := mapper(log)
		if err != nil {
			w.log.Debugf("error mapping event, skipping: %s", err)
			continue
		}

		select {
		case <-quit:
			return nextFromBlock, SubClosedError
		case <-ctx.Done():
			return nextFromBlock, ctx.Err()
		case sink <- event:
		}
	}

	nextFromBlock.Add(currentBlock.Number, big.NewInt(1))

	select {
	case <-quit:
		return nextFromBlock, SubClosedError
	case <-ctx.Done():
		return nextFromBlock, ctx.Err()
	case <-time.After(w.pollInterval):
	}

	return nextFromBlock, nil
}
