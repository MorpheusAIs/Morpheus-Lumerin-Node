'use strict';

const logger = require('../../logger');

const Web3 = require('web3');
const { Lumerin } = require('contracts-js');

const createEventsRegistry = require('./events');
const { logTransaction } = require('./log-transaction');
const createQueue = require('./queue');
const createStream = require('./blocks-stream');
const createTransactionSyncer = require('./sync-transactions');
const tryParseEventLog = require('./parse-log');
const createExplorer = require('./explorer');

function createPlugin() {
  let blocksStream;
  let syncer;
  let interval;

  function start({ config, eventBus, plugins }) {
    // debug.enabled = config.debug;
    const { lmrTokenAddress } = config;

    const web3 = new Web3(plugins.eth.web3Provider);

    const web3Subscribable = new Web3(plugins.eth.web3SubscriptionProvider);

    const eventsRegistry = createEventsRegistry();
    const queue = createQueue(config, eventBus, web3);
    const lumerin = Lumerin(web3Subscribable, lmrTokenAddress);

    const explorer = createExplorer(config.explorerApiURLs, web3, lumerin, eventBus);

    syncer = createTransactionSyncer(
      config,
      eventBus,
      web3,
      queue,
      eventsRegistry,
      explorer
    );

    logger.debug('Initiating blocks stream');
    const streamData = createStream(web3, config.blocksUpdateMs);
    blocksStream = streamData.stream;
    interval = streamData.interval;

    blocksStream.on('data', function ({ hash, number, timestamp }) {
      logger.debug('New block', hash, number);
      eventBus.emit('coin-block', { hash, number, timestamp });
    });
    blocksStream.on('error', function (err) {
      logger.debug('Could not get latest block');
      eventBus.emit('wallet-error', {
        inner: err,
        message: 'Could not get latest block',
        meta: { plugin: 'explorer' }
      });
    });

    return {
      api: {
        logTransaction: logTransaction(queue),
        refreshAllTransactions: syncer.refreshAllTransactions,
        registerEvent: eventsRegistry.register,
        syncTransactions: syncer.syncTransactions,
        tryParseEventLog: tryParseEventLog(web3, eventsRegistry),
        getPastCoinTransactions: syncer.getPastCoinTransactions,
      },
      events: [
        'token-transactions-changed',
        'wallet-state-changed',
        'coin-block',
        'transactions-next-page',
        'indexer-connection-status-changed',
        'wallet-error'
      ],
      name: 'explorer'
    };
  }

  function stop() {
    // blocksStream.destroy();
    blocksStream.removeAllListeners();
    clearInterval(interval);
    syncer.stop();
  }

  return {
    start,
    stop
  };
}

module.exports = createPlugin
