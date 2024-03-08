'use strict';

const debug = require('debug')('lmr-wallet:core:eth:web3');
const Web3 = require('web3');

function createWeb3 (config, eventBus) {
  debug.enabled = config.debug;

  const web3 = new Web3(new Web3.providers.WebsocketProvider(
    config.wsApiUrl,
    { autoReconnect: true, timeout: config.web3Timeout }
  ));

  web3.currentProvider.on('connect', function () {
    debug('Web3 provider connected');
    eventBus.emit('web3-connection-status-changed', { connected: true });
  });
  web3.currentProvider.on('error', function (event) {
    debug('Web3 provider connection error: ', event.type || event.message);
    eventBus.emit('web3-connection-status-changed', { connected: false });
  });
  web3.currentProvider.on('end', function (event) {
    debug('Web3 provider connection ended: ', event.reason);
    eventBus.emit('web3-connection-status-changed', { connected: false });
  });

  return web3;
}

function destroyWeb3 (web3) {
  web3.currentProvider.disconnect();
}

module.exports = {
  createWeb3,
  destroyWeb3
};
