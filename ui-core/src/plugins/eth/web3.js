'use strict'

const logger = require('../../logger');

const Web3 = require('web3')
const { Web3Http } = require('./web3Http');


function createWeb3(config) {
  // debug.enabled = config.debug

  const web3 = new Web3Http(config.httpApiUrls)
  return web3
}

function createWeb3Subscribable(config, eventBus) {
  // debug.enabled = config.debug

  const options = {
    timeout: 1000 * 15, // ms
    // Enable auto reconnection
    reconnect: {
      auto: true,
      delay: 5000, // ms
      maxAttempts: false,
      onTimeout: false,
    },
  }

  const web3 = new Web3(
    new Web3.providers.WebsocketProvider(config.wsApiUrl, options)
  )

  web3.currentProvider.on('connect', function () {
    logger.debug('Web3 provider connected')
    eventBus.emit('web3-connection-status-changed', { connected: true })
  })
  web3.currentProvider.on('error', function (event) {
    logger.debug('Web3 provider connection error: ', event.type || event.message)
    eventBus.emit('web3-connection-status-changed', { connected: false })
  })
  web3.currentProvider.on('end', function (event) {
    logger.debug('Web3 provider connection ended: ', event.reason)
    eventBus.emit('web3-connection-status-changed', { connected: false })
  })

  return web3
}

function destroyWeb3(web3) {
  web3.currentProvider?.disconnect()
}

module.exports = {
  createWeb3,
  destroyWeb3,
  createWeb3Subscribable,
}
