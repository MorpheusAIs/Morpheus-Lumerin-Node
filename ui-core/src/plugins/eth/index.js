'use strict';

const debug = require('debug')('lmr-wallet:core:eth');

const { createWeb3, destroyWeb3 } = require('./web3');
const checkChain = require('./check-chain');

function createPlugin () {
  let web3 = null;

  function start ({ config, eventBus }) {
    debug.enabled = config.debug;

    web3 = createWeb3(config, eventBus);

    checkChain(web3, config.chainId)
      .then(function () {
        debug('Chain ID is correct');
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
        web3Provider: web3.currentProvider
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
