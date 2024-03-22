'use strict'

const logger = require('../../logger');

const createConnectionManager = require('./connections-manager')

function createPlugin() {
  let connectionManager

  function start({ config, eventBus }) {
    // debug.enabled = config.debug

    logger.debug('Initiating proxy-router connections stream')
    connectionManager = createConnectionManager(config, eventBus)

    const refreshConnectionsStream = (data) =>
      connectionManager
        .getConnectionsStream(data.proxyNodeUrl)
        .on('data', (data) => {
          eventBus.emit('proxy-router-connections-changed', {
            connections: data.connections,
          })
        })
        .on('error', (err) => {
          eventBus.emit('wallet-error', {
            inner: err,
            message: `Proxy router connection error`,
            meta: { plugin: 'proxy-router' },
          })
        })

    return {
      api: {
        refreshConnectionsStream: refreshConnectionsStream,
        getLocalIp: connectionManager.getLocalIp,
        healthCheck: connectionManager.healthCheck,
        kill: connectionManager.kill,
        isProxyPortPublic: connectionManager.isProxyPortPublic,
      },
      events: [
        'proxy-router-connections-changed',
        'proxy-router-status-changed',
        'proxy-router-error',
      ],
      name: 'proxy-router',
    }
  }

  function stop() {
    connectionManager.disconnect()
  }

  return {
    start,
    stop,
  }
}

module.exports = createPlugin
