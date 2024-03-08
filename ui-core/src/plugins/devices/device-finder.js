//@ts-check
const { Socket } = require('net')
const { AbortSignal } = require('@azure/abort-controller')
const { DeviceType } = require('./consts')
const { CGMinerApi } = require('./cgminer-api')
const { getIPsWithinRange, getHostsInLocalSubnets } = require('./ip-range')
const { ConfigurationStrategyFactory } = require('./configuration-strategies/factory')

/**
 * @typedef {Object} UpdateEventData
 * @property {String} host - IP address of the server
 * @property {Boolean} isDone - true when this is the last event for this host
 * @property {Boolean} isHostUp - true when connection established
 * @property {Boolean} [isApiAvailable] - true when miner API is responding
 * @property {("miner"|"unknown")} [deviceType] - detected device type
 * @property {String} [deviceModel] - detected device model
 * @property {Number} [hashRateGHS] - hash rate in GH/s
 * @property {String} [poolAddress] - url address of the destination pool
 * @property {String} [poolUser] - username for the destination pool
 * @property {Boolean} [isPrivilegedApiAvailable] - true when local network has privileged access
 */

/**
 * Detects miners by IP range
 * @param {[string, string]} range
 * @param {AbortSignal} abortSignal
 * @param {(data: UpdateEventData)=>void} onUpdate
 * @returns {Promise<void>} resolves when all IP addresses are scanned
 */
function detectMinersByRange(range, abortSignal, onUpdate) {
  const addresses = range
    ? getIPsWithinRange(range[0], range[1])
    : getHostsInLocalSubnets()

  return detectIPBatch(addresses, abortSignal, onUpdate)
}

/**
 * Performs batch of detecMiners requests
 * @param {Array.<String>} addresses array of addresses to scan
 * @param {AbortSignal} batchAbort signal to cancel all of the requests
 * @param {(data: UpdateEventData)=>void} onUpdate called when new device found or existing device state updated
 * @return {Promise<void>} resolved when all addresses were scanned
 */
async function detectIPBatch(addresses, batchAbort, onUpdate) {
  const promises = addresses.map(
    (host) =>
      /** @type {Promise<void>} */ (
        new Promise((resolve) => {
          detectMinerWrapped(host, batchAbort, (data) => {
            if (data.isHostUp) {
              onUpdate(data)
            }
            if (data.isDone) {
              resolve()
            }
          }).catch(() => {})
        })
      )
  )

  await Promise.all(promises)
}

/**
 * Wraps detectMiner with timeout and batch abort
 * @param {string} host
 * @param {AbortSignal} batchAbort
 * @param {(data: UpdateEventData)=>void} onUpdate
 * @returns {Promise} resolves when
 */
async function detectMinerWrapped(host, batchAbort, onUpdate) {
  if (batchAbort.aborted) {
    onUpdate({ host, isDone: true, isHostUp: false })
    return
  }

  return detectMiner({ host, abort: batchAbort }, onUpdate)
}

/**
 * Performs connection over the socket and if successful calls miner API.
 * Calls onUpdate after each event with received data, allowing to pass
 * events immediately after connection. Emits data.isDone on the last request
 * @param {Object} data
 * @param {String} data.host
 * @param {AbortSignal} data.abort // controls cancellation
 * @param {()=>Socket} [data.SocketFactory] // for tests
 * @param {(data: UpdateEventData) => void} onUpdate // is called when new data returned
 * @returns {Promise<void>}
 */
const detectMiner = async (data, onUpdate) => {
  const cgMinerApi = new CGMinerApi()

  if (data.abort.aborted) {
    cgMinerApi.disconnect()
    return onUpdate({
      host: data.host,
      isHostUp: false,
      isDone: true,
    })
  }

  await cgMinerApi
    .connect({ host: data.host, abort: data.abort })
    .catch((err) => {
      onUpdate({
        host: data.host,
        isHostUp: false,
        isDone: true,
      })
      cgMinerApi.disconnect()
      throw err
    })

  onUpdate({
    host: data.host,
    isHostUp: true,
    isDone: false,
  })

  const res = await cgMinerApi.getMinerData().catch((err) => {
    onUpdate({
      host: data.host,
      isDone: true,
      isHostUp: true,
      isApiAvailable: false,
      deviceType: DeviceType.Unknown,
      isPrivilegedApiAvailable: false,
    })
    cgMinerApi.disconnect()
    throw err
  })

  const configurationStrategy = await ConfigurationStrategyFactory.createStrategy(data.host, data.abort);
  const isPrivilegedApiAvailable = !!configurationStrategy;

  onUpdate({
    host: data.host,
    isHostUp: true,
    isDone: true,
    isApiAvailable: true,
    deviceType: DeviceType.Miner,
    deviceModel: res.deviceModel,
    hashRateGHS: res.hashRateGHS,
    poolAddress: res.poolAddress,
    poolUser: res.poolUser,
    isPrivilegedApiAvailable,
  })
  cgMinerApi.disconnect()
}

module.exports = {
  detectMinersByRange,
  detectIPBatch,
  detectMiner,
}
