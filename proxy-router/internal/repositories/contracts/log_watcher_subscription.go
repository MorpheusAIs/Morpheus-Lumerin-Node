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
	"github.com/ethereum/go-ethereum/core/types"
)

// Sentinel used to distinguish "the caller asked us to stop" from a real error
// while unwinding the per-attempt closure.
var errWatchStopped = errors.New("watch stopped")

// Reconnect backoff bounds. The retry loop previously had no delay at all, so a
// dead RPC endpoint turned into a hot spin loop: it burned CPU and hammered the
// provider hard enough to get the user rate-limited or banned.
const (
	reconnectBaseDelay = 500 * time.Millisecond
	reconnectMaxDelay  = 30 * time.Second
)

type LogWatcherSubscription struct {
	// config
	maxReconnects int

	// deps
	client i.EthClient
	log    lib.ILogger
}

// NewLogWatcherSubscription creates a new log subscription using websocket
// TODO: if it is going to be primary implementation we should rewrite it so it doesn't skip events in case of temporary downtime
func NewLogWatcherSubscription(client i.EthClient, maxReconnects int, log lib.ILogger) *LogWatcherSubscription {
	return &LogWatcherSubscription{
		maxReconnects: maxReconnects,
		client:        client,
		log:           log,
	}
}

func (w *LogWatcherSubscription) Watch(ctx context.Context, contractAddr common.Address, mapper EventMapper, fromBlock *big.Int) (*lib.Subscription, error) {
	sink := make(chan interface{})

	return lib.NewSubscription(func(quit <-chan struct{}) error {
		defer close(sink)

		// `fromBlock` was previously accepted and then never used: the query was
		// built with Addresses only, so every (re)subscription resumed from the
		// chain head. Any event emitted while the websocket was down — a session
		// opening or closing during a brief RPC blip — was lost permanently and
		// never showed up in the UI.
		//
		// We now track the last block we actually observed and resume from it on
		// reconnect, so the gap is replayed instead of skipped.
		query := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddr},
			FromBlock: fromBlock,
		}
		in := make(chan types.Log)
		defer close(in)

		var lastSeenBlock uint64
		if fromBlock != nil && fromBlock.IsUint64() {
			lastSeenBlock = fromBlock.Uint64()
		}

		for {
			sub, err := w.subscribeFilterLogsRetry(ctx, query, in)
			if err != nil {
				w.log.Errorf("failed to subscribe to logs: %s", err)
				return err
			}

			// Unsubscribe as soon as THIS attempt ends. The previous code used a
			// bare `defer` inside the reconnect loop, so every reconnect stacked
			// another deferred call that only ran when Watch finally returned —
			// leaking a subscription (and its goroutine) per reconnect for the
			// lifetime of the daemon.
			err = func() error {
				defer sub.Unsubscribe()

				for {
					select {
					case log := <-in:
						// Track progress so a reconnect resumes from here rather
						// than from the head.
						if log.BlockNumber > lastSeenBlock {
							lastSeenBlock = log.BlockNumber
						}

						event, mapErr := mapper(log)
						if mapErr != nil {
							w.log.Debugf("failed to map event: %s", mapErr)
							// mapper error, retry won't help, continue to next event
							continue
						}

						select {
						case sink <- event:
						case subErr := <-sub.Err():
							w.log.Debugf("subscription error: %s", subErr)
							return nil
						case <-quit:
							return errWatchStopped
						case <-ctx.Done():
							return ctx.Err()
						}
					case subErr := <-sub.Err():
						w.log.Debugf("subscription error: %s", subErr)
						return nil
					case <-quit:
						return errWatchStopped
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}()

			if err != nil {
				if errors.Is(err, errWatchStopped) {
					return nil
				}
				return err
			}

			// Reconnecting: resume one block after the last one we processed.
			if lastSeenBlock > 0 {
				query.FromBlock = new(big.Int).SetUint64(lastSeenBlock + 1)
				w.log.Infof("resubscribing from block %d", lastSeenBlock+1)
			}
		}
	}, sink), nil
}

func (w *LogWatcherSubscription) subscribeFilterLogsRetry(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	var lastErr error

	for attempts := 0; attempts < w.maxReconnects || w.maxReconnects == -1; attempts++ {
		sub, err := w.client.SubscribeFilterLogs(ctx, query, ch)
		if err != nil {
			maxReconnects := fmt.Sprintf("%d", w.maxReconnects)
			if w.maxReconnects == -1 {
				maxReconnects = "∞"
			}

			// Exponential backoff, capped. Without this the loop retried as fast
			// as the CPU allowed.
			delay := reconnectBaseDelay << min(attempts, 6)
			if delay > reconnectMaxDelay {
				delay = reconnectMaxDelay
			}

			w.log.Warnf("subscription error, retrying in %s (%d/%s): %s", delay, attempts, maxReconnects, err)
			lastErr = err

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}
		if attempts > 0 {
			w.log.Warnf("subscription successfully reconnected after error: %s", lastErr)
		}

		return sub, nil
	}

	err := fmt.Errorf("subscription error, retries exhausted (%d), stopping: %s", w.maxReconnects, lastErr)
	w.log.Warnf(err.Error())

	return nil, err
}
