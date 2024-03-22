'use strict';

const logger = require('../../logger');

// eslint-disable-next-line max-params
function createSyncer (config, eventBus, web3, queue, eventsRegistry, indexer) {
  // debug.enabled = config.debug;

  let bestBlock;
  const gotBestBlockPromise = new Promise(function (resolve) {
    eventBus.once('coin-block', function (header) {
      bestBlock = header.number;
      logger.debug('Got best block', bestBlock);
      resolve(bestBlock);
    });
  })

  function subscribeCoinTransactions (fromBlock, address) {
    let shallResync = false;
    let resyncing = false;
    let bestSyncBlock = fromBlock;

    const { symbol, displayName } = config;

    // LMR transactions
    indexer.getTransactionStream(address)
      .on('data', (data)=>{
        queue.addTx(address, null)(mapApiResponseToTrxReceipt(data))
      })
      .on('resync', function () {
        logger.debug(`Shall resync ${symbol} transactions on next block`)
        shallResync = true;
      })
      .on('error', function (err) {
        logger.debug(`Shall resync ${symbol} transactions on next block`)
        shallResync = true;
        eventBus.emit('wallet-error', {
          inner: err,
          message: `Failed to sync ${displayName} transactions`,
          meta: { plugin: 'explorer' }
        });
      });

    // ETH transactions
    // Check if shall resync when a new block is seen, as that is the
    // indication of proper reconnection to the Ethereum node.
    eventBus.on('coin-block', function ({ number }) {
      if (shallResync && !resyncing) {
        resyncing = true;
        shallResync = false;
        // eslint-disable-next-line promise/catch-or-return
        indexer.getETHTransactions(bestSyncBlock, number, address)
          .then(function (transactions) {
            const { length } = transactions;
            logger.debug(`${length} past ETH transactions retrieved`)
            const txs = transactions.map(mapApiResponseToTrxReceipt)
            queue.addTxs(address,null)(txs)
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
        bestBlock = number;
      }
    })
  }

  function mapApiResponseToTrxReceipt(trx){
    const transaction = {
      from: trx.from,
      to: trx.to,
      value: trx.value,
      input: trx.input,
      gas: trx.gas,
      gasPrice: trx.gasPrice,
      hash: trx.hash,
      nonce: trx.nonce,
      logIndex: trx.logIndex, // emitted only in events, used to differentiate between LMR transfers within one transaction 
      // maxFeePerGas: params.maxFeePerGas,
      // maxPriorityFeePerGas: params.maxPriorityFeePerGas,
    }

    if (trx.returnValues){
      transaction.from = trx.returnValues.from;
      transaction.to = trx.returnValues.to;
      transaction.value = trx.returnValues.value;
      transaction.hash = trx.transactionHash;
    }

    const receipt = {
      transactionHash: trx.hash,
      transactionIndex: trx.transactionIndex,
      blockHash: trx.blockHash,
      blockNumber: trx.blockNumber,
      from: trx.from,
      to: trx.to,
      value: trx.value,
      contractAddress: trx.contractAddress,
      cumulativeGasUsed: trx.cumulativeGasUsed,
      gasUsed: trx.gasUsed,
      tokenSymbol: trx.tokenSymbol,
    }

    if (trx.returnValues){
      receipt.from = trx.returnValues.from;
      receipt.to = trx.returnValues.to;
      receipt.value = trx.returnValues.value;
      receipt.transactionHash = trx.transactionHash;
      receipt.tokenSymbol = trx.address === config.chain.lmrTokenAddress ? 'LMR' : undefined;
    }

    return {transaction, receipt}
  }

  /**
   * @param {string} fromBlock 
   * @param {string} toBlock 
   * @param {string} address 
   * @returns {Promise<string>} lastSyncedBlock
   */
  async function getPastCoinTransactions (fromBlock, toBlock, address, page, pageSize) {
    const { symbol } = config;

    const transactions = await indexer.getTransactions(fromBlock, toBlock || bestBlock, address, page, pageSize)
    logger.debug(`${transactions.length} past ${symbol} transactions retrieved`);

    queue.addTxs(address, null)(transactions.map(mapApiResponseToTrxReceipt))

    return toBlock;
  }

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
        logger.error('Could not subscribe: event not found', eventName);
        return;
      }

      // Get past events and subscribe to incoming events
      const emitter = contract.events[eventName]({ fromBlock, filter })
        .on('data', queue.addEvent(address, metaParser))
        .on('changed', queue.addEvent(address, metaParser))
        .on('error', function (err) {
          logger.error('Shall resync events on next block');
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

  const syncTransactions = (fromBlock, address, onProgress, page, pageSize) =>
    gotBestBlockPromise
      .then(function () {
        logger.debug('Syncing', fromBlock, bestBlock);
        subscribeCoinTransactions(bestBlock, address);
        subscribeEvents(bestBlock, address);
        return getPastCoinTransactions(fromBlock, bestBlock, address, page, pageSize)
      })
      .then(function (syncedBlock) {
        bestBlock = syncedBlock;
        return syncedBlock;
      });

  const refreshAllTransactions = address =>
    gotBestBlockPromise
      .then(() => {
        return getPastCoinTransactions(0, bestBlock, address)
          .then(function ([syncedBlock]) {
            bestBlock = syncedBlock;
            return syncedBlock;
          })
        });

  function stop () {
    subscriptions.forEach(function (subscription) {
      subscription.unsubscribe(function (err) {
        if (err) {
          logger.error('Could not unsubscribe from event', err.message);
        }
      });
    });
  }

  return {
    getPastCoinTransactions,
    refreshAllTransactions,
    stop,
    syncTransactions
  };
}

module.exports = createSyncer;
