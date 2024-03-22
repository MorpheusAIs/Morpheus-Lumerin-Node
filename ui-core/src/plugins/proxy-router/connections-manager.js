'use strict'

const { create: createAxios } = require('axios')
const axios = require('axios')
const logger = require('../../logger')
const EventEmitter = require('events')
const killer = require('cross-port-killer')

/**
 * Create an object to interact with the Lumerin indexer.
 *
 * @param {object} config The configuration object.
 * @param {object} eventBus The corss-plugin event bus.
 * @returns {object} The exposed indexer API.
 */
function createConnectionsManager(config, eventBus) {
  const {
    debug: enableDebug,
    proxyRouterUrl,
    ipLookupUrl,
    portCheckerUrl,
  } = config
  const pollingInterval = 5000

  // debug.enabled = enableDebug

  let interval

  const getConnections = async (proxyUrl) => {
    const getMiners = async (url) => {
      return (await createAxios({ baseURL: url })('/miners')).data?.Miners
    }

    if (proxyUrl) {
      const sellerMiners = await getMiners(proxyUrl)
      return [...sellerMiners]
    }

    return await getMiners(proxyRouterUrl)
  }

  const healthCheck = async (url) => {
    return createAxios({ baseURL: url })('/healthcheck')
  }

  const kill = (port) => {
    return killer.kill(port)
  }

  /**
   * Create a stream that will emit an event each time a connection is published to the proxy-router
   *
   * The stream will emit `data` for each connection. If the proxy-router connection is lost
   * or an error occurs, an `error` event will be emitted. In addition, when the
   * connection is restablished, a `resync` will be emitted.
   *
   * @param {string} [url] Overrides url from config
   *
   * @returns {object} The event emitter.
   */
  function getConnectionsStream(proxyUrl) {
    const stream = new EventEmitter()

    let isConnected = false

    disconnect()
    interval = setInterval(async () => {
      try {
        const connections = await getConnections(proxyUrl)

        if (!isConnected) {
          isConnected = true
          logger.debug('emit proxy-router-status-changed')
          eventBus.emit('proxy-router-status-changed', {
            isConnected,
            syncStatus: 'synced',
          })
        }

        stream.emit('data', {
          connections,
        })
      } catch (err) {
        isConnected = false
        eventBus.emit('proxy-router-status-changed', {
          isConnected,
          syncStatus: 'syncing',
        })
        stream.emit('error', err)
      }
    }, pollingInterval)

    return stream
  }

  /**
   * Disconnects.
   */
  function disconnect() {
    if (interval) {
      clearInterval(interval)
    }
  }

  /**
   *
   * @returns {string|null}
   */
  const getLocalIp = async () => {
    const baseURL = ipLookupUrl || 'https://ifconfig.io/ip'
    const { data } = await createAxios({ baseURL })()
    const stringData = typeof data === 'string' ? data : JSON.stringify(data)

    const ipRegex =
      /\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b/
    const [ip] = stringData.match(ipRegex)
    return ip
  }

  /**
   *
   * @returns {Promise<boolean>}
   */
  const isProxyPortPublic = async (payload) => {
    const { ip, port } = payload;
    const baseURL = portCheckerUrl || 'https://portchecker.io/api/v1/query'
    const { data } = await axios.post(baseURL, {
      host: ip,
      ports: [port],
    })

    return !!data.check?.find((c) => c.port === port && c.status === true)
  }

  return {
    disconnect,
    getConnectionsStream,
    getLocalIp,
    healthCheck,
    kill,
    isProxyPortPublic,
  }
}

module.exports = createConnectionsManager
