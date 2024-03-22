'use strict'

const debug = require('debug')('lmr-wallet:core:devices')
const { AbortController } = require('@azure/abort-controller')
const {
  EVENT_DEVICES_STATE_UPDATED,
  EVENT_DEVICES_DEVICE_UPDATED,
} = require('./consts')
const { detectMinersByRange } = require('./device-finder')
const { setPool } = require('./miner-configurator')

function createPlugin() {
  /**
   * @param {Object} params
   * @param {Object} params.config - configuration of the module
   * @param {boolean} params.config.debug - is debug messages enabled
   * @param {Object} params.eventBus - instance of global event bus
   * @returns Module
   */
  function start({ config, eventBus }) {
    debug.enabled = config.debug
    debug('Initiating devices discovery module')
    let abort
    let isDiscovering = false

    return {
      api: {
        startDiscovery: async (data) => {
          if (isDiscovering) {
            return
          }

          abort = new AbortController()
          isDiscovering = true
          eventBus.emit(EVENT_DEVICES_STATE_UPDATED, { isDiscovering })

          const range =
            data.fromIp && data.toIp ? [data.fromIp, data.toIp] : null

          // intentionally not awaiting, as it is a lengthy process
          // frontend will receive response as events separately from promise
          detectMinersByRange(range, abort.signal, (update) => {
            eventBus.emit(EVENT_DEVICES_DEVICE_UPDATED, update)
          })
            .then(() => {
              isDiscovering = false
              eventBus.emit(EVENT_DEVICES_STATE_UPDATED, { isDiscovering })
            })
            .catch((err) => {
              debug('Device discovery error', err)
            })
        },
        stopDiscovery: async () => {
          abort?.abort()
          isDiscovering = false
          eventBus.emit(EVENT_DEVICES_STATE_UPDATED, { isDiscovering })
        },
        setMinerPool: async (data) => {
          const { host, pool } = data
          eventBus.emit(EVENT_DEVICES_DEVICE_UPDATED, { isDone: false, host })

          setPool(host, pool, abort.signal, (update) => {
            eventBus.emit(EVENT_DEVICES_DEVICE_UPDATED, update)
          }).catch((err) => {
            debug('Set pool error:', err)
            eventBus.emit('wallet-error', {
              inner: err,
              message: `Failed to update miner configuration: ${host}`,
              meta: { plugin: 'devices' },
            })
            eventBus.emit(EVENT_DEVICES_DEVICE_UPDATED, {
              host,
              isDone: true,
            })
          })
        },
      },
      events: [EVENT_DEVICES_DEVICE_UPDATED, EVENT_DEVICES_STATE_UPDATED],
      name: 'devices',
    }
  }

  function stop() {}

  return {
    start,
    stop,
  }
}

module.exports = createPlugin
