'use strict';

const { identity } = require('lodash');
const debug = require('debug')('lmr-wallet:core:explorer:syncer');
const pAll = require('p-all');
const pWhilst = require('p-whilst');
const pTimeout = require('p-timeout');
const noop = require('lodash/noop');

// eslint-disable-next-line max-params
function createSyncer (config, eventBus, web3, queue, eventsRegistry, indexer) {
  debug.enabled = config.debug;

  let bestBlock;
  const gotBestBlockPromise = new Promise(function (resolve) {
    eventBus.once('coin-block', function (header) {
      bestBlock = header.number;
      debug('Got best block', bestBlock);
      resolve(bestBlock);
    });
  })

  function subscribeCoinTransactions (fromBlock, address) {
    let shallResync = false;
    let resyncing = false;
    let bestSyncBlock = fromBlock;

    const { symbol, displayName } = config;

    indexer.getTransactionStream(address)
      .on('data', queue.addTransaction(address))
      .on('resync', function () {
        debug(`Shall resync ${symbol} transactions on next block`)
        shallResync = true;
      })
      .on('error', function (err) {
        debug(`Shall resync ${symbol} transactions on next block`)
        shallResync = true;
        eventBus.emit('wallet-error', {
          inner: err,
          message: `Failed to sync ${displayName} transactions`,
          meta: { plugin: 'explorer' }
        });
      });

    // Check if shall resync when a new block is seen, as that is the
    // indication of proper reconnection to the Ethereum node.
    eventBus.on('coin-block', function ({ number }) {
      if (shallResync && !resyncing) {
        resyncing = true;
        shallResync = false;
        // eslint-disable-next-line promise/catch-or-return
        indexer.getTransactions(bestSyncBlock, number, address)
          .then(function (transactions) {
            const { length } = transactions;
            debug(`${length} past ${symbol} transactions retrieved`)
            transactions.forEach(queue.addTransaction(address));
            bestSyncBlock = number;
          })
          .catch(function (err) {
            shallResync = true;
            eventBus.emit('wallet-error', {
              inner: err,
              message: 'Failed to resync transactions',
              meta: { plugin: 'explorer' }
            });
          })
          .then(function () {
            resyncing = false;
          })
      } else if (!resyncing) {
        bestSyncBlock = number;
        bestBlock = number;
      }
    })
  }

  function getPastCoinTransactions (fromBlock, toBlock, address) {
    const { symbol } = config;

    return indexer.getTransactions(fromBlock, toBlock, address)
      .then(function (transactions) {
        debug(`${transactions.length} past ${symbol} transactions retrieved`);
        return Promise.all(transactions.map(queue.addTransaction(address)))
          .then(() => toBlock);
      });
  }

  function getPastEventsWithChunks (options) {
    const CHUNK_SIZE = 4000;
    const {
      address,
      contract,
      eventName,
      fromBlock,
      toBlock,
      filter,
      metaParser,
      minBlock = 0,
      onProgress = noop
    } = options;
    const baseFromBlock = Math.max(fromBlock, minBlock);
    debug('Retrieving from %s to %s in chunks', baseFromBlock, toBlock);
    let chunkIndex = 0;
    const getNewFromBlock = index => baseFromBlock + CHUNK_SIZE * index;
    const getNewToBlock = function (index) {
      const istLastRequest = baseFromBlock + CHUNK_SIZE * (index + 1) === toBlock;
      return Math.max(Math.min(baseFromBlock + CHUNK_SIZE * (index + 1) - (istLastRequest ? 0 : 1), toBlock), minBlock);
    };
    return pWhilst(
      function () {
        const newFromBlock = getNewFromBlock(chunkIndex);
        const newToBlock = getNewToBlock(chunkIndex);
        return newFromBlock < newToBlock;
      },
      function () {
        const newFromBlock = getNewFromBlock(chunkIndex);
        const newToBlock = getNewToBlock(chunkIndex);
        debug('Retrieving from %s to %s for event %s', newFromBlock, newToBlock, eventName);
        return pTimeout(
          contract
            .getPastEvents(eventName, {
              fromBlock: newFromBlock,
              toBlock: newToBlock,
              filter
            })
            .then(function (events) {
              debug(`${events.length} past ${eventName} events retrieved`)
              return Promise.all(
                events.map(queue.addEvent(address, metaParser))
              );
            })
            .then(function () {
              debug('Retrieved from %s to %s for event %s', newFromBlock, newToBlock, eventName);
              chunkIndex++;
              return onProgress(newToBlock);
            })
          ,
          1000 * 60 * 2
        );
      });
  }

  const getPastEvents = (fromBlock, toBlock, address, onProgress) =>
    pAll(
      eventsRegistry
        .getAll()
        .map(function (registration) {
          const {
            contractAddress,
            abi,
            eventName,
            filter,
            metaParser,
            minBlock = 0
          } = registration(address);

          const contract = new web3.eth.Contract(abi, contractAddress);

          // Ignore missing events
          if (!contract.events[eventName]) {
            debug(`Could not get past events for ${eventName}`);
            return null;
          }
          return () =>
            getPastEventsWithChunks({
              address,
              contract,
              eventName,
              fromBlock,
              toBlock,
              filter,
              minBlock,
              onProgress,
              metaParser
            })
        })
        .filter(identity),
      { concurrency: 3 }
    );

  const subscriptions = [];

  function subscribeEvents (fromBlock, address) {
    eventsRegistry.getAll().forEach(function (registration) {
      let shallResync = false;
      let resyncing = false;
      let bestSyncBlock = fromBlock;

      const {
        contractAddress,
        abi,
        eventName,
        filter,
        metaParser
      } = registration(address);

      const contract = new web3.eth.Contract(abi, contractAddress);

      // Ignore missing events
      if (!contract.events[eventName]) {
        debug('Could not subscribe: event not found', eventName);
        return;
      }

      // Get past events and subscribe to incoming events
      const emitter = contract.events[eventName]({ fromBlock, filter })
        .on('data', queue.addEvent(address, metaParser))
        .on('changed', queue.addEvent(address, metaParser))
        .on('error', function (err) {
          debug('Shall resync events on next block');
          shallResync = true;
          eventBus.emit('wallet-error', {
            inner: err,
            message: `Subscription to event ${eventName} failed`,
            meta: { plugin: 'explorer' }
          })
        });
      subscriptions.push(emitter);

      // Resync on new block or save it as best sync block
      eventBus.on('coin-block', function ({ number }) {
        if (shallResync && !resyncing) {
          resyncing = true;
          shallResync = false;
          // eslint-disable-next-line promise/catch-or-return
          getPastEventsWithChunks({
            address,
            contract,
            eventName,
            fromBlock: bestSyncBlock,
            toBlock: number,
            filter,
            metaParser
          })
            .catch(function (err) {
              shallResync = true
              eventBus.emit('wallet-error', {
                inner: err,
                message: `Failed to resync event ${eventName}`,
                meta: { plugin: 'explorer' }
              })
            })
            .then(function () {
              resyncing = false
            });
        } else if (!resyncing) {
          bestSyncBlock = number;
          bestBlock = number;
        }
      });
    });
  }

  const syncTransactions = (fromBlock, address, onProgress) =>
    gotBestBlockPromise
      .then(function () {
        debug('Syncing', fromBlock, bestBlock);
        subscribeCoinTransactions(bestBlock, address);
        subscribeEvents(bestBlock, address);
        return Promise.all([
          getPastCoinTransactions(fromBlock, bestBlock, address),
          getPastEvents(fromBlock, bestBlock, address, onProgress)
        ]);
      })
      .then(function ([syncedBlock]) {
        bestBlock = syncedBlock;
        return syncedBlock;
      });

  const refreshAllTransactions = address =>
    gotBestBlockPromise
      .then(() => {
        return Promise.all([
          getPastCoinTransactions(0, bestBlock, address),
          getPastEvents(0, bestBlock, address, function (syncedBlock) { bestBlock = syncedBlock })
        ])
          .then(function ([syncedBlock]) {
            bestBlock = syncedBlock;

            return syncedBlock;
          })
        });

  function stop () {
    subscriptions.forEach(function (subscription) {
      subscription.unsubscribe(function (err) {
        if (err) {
          debug('Could not unsubscribe from event', err.message);
        }
      });
    });
  }

  return {
    getPastCoinTransactions,
    getPastEvents,
    refreshAllTransactions,
    stop,
    syncTransactions
  };
}

module.exports = createSyncer;
