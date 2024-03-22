'use strict';

const logger = require('../../logger');

const { createWeb3, destroyWeb3, createWeb3Subscribable } = require('./web3');
const checkChain = require('./check-chain');

function createPlugin () {
  let web3 = null;
  let web3SubscribAble = null;

  function start ({ config, eventBus }) {
    // debug.enabled = config.debug;
    
    web3 = createWeb3(config, eventBus);
    web3SubscribAble = createWeb3Subscribable(config, eventBus);

    checkChain(web3, config.chainId)
      .then(function () {
        logger.debug('Chain ID is correct');
      })
      .catch(function (err) {
        eventBus.emit('wallet-error', {
          inner: err,
          message: 'Could not check chain ID',
          meta: { plugin: 'eth' }
        });
      });

    return {
      api: {
        web3,
        web3Provider: web3.currentProvider,
        web3SubscriptionProvider: web3SubscribAble.currentProvider,
      },
      events: [
        'wallet-error',
        'web3-connection-status-changed'
      ],
      name: 'eth'
    };
  }

  function stop () {
    destroyWeb3(web3);
    web3 = null;
  }

  return {
    start,
    stop
  };
}

module.exports = createPlugin
