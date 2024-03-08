'use strict'

const axios = require('axios').default;
const { getExchangeRate } = require('safe-exchange-rate')
const debug = require('debug')('lmr-wallet:core:rates')

const { getLmrRate } = require('./getLmrRate');

const createStream = require('./stream')

/**
 * Create a plugin instance.
 *
 * @returns {({ start: Function, stop: () => void})} The plugin instance.
 */
function createPlugin() {
  let dataStream

  /**
   * Start the plugin instance.
   *
   * @param {object} options Start options.
   * @returns {{ events: string[] }} The instance details.
   */
  function start({ config, eventBus }) {
    debug.enabled = debug.enabled || config.debug

    debug('Plugin starting')

    const { ratesUpdateMs, symbol } = config

    const getRate = () =>
      {
        return symbol === 'LMR' ? getLmrRate() : getExchangeRate(`${symbol}:USD`).then(function (rate) {
          if (typeof rate !== 'number') {
            throw new Error(`No exchange rate retrieved for ${symbol}`)
          }
          return rate
        })
      }

    dataStream = createStream(getRate, ratesUpdateMs)

    dataStream.on('data', function (price) {
      debug('Coin price received')

      const priceData = { token: symbol, currency: 'USD', price }
      eventBus.emit('coin-price-updated', priceData)
    })

    dataStream.on('error', function (err) {
      debug('Data stream error')

      eventBus.emit('wallet-error', {
        inner: err,
        message: `Could not get exchange rate for ${symbol}`,
        meta: { plugin: 'rates' },
      })
    })

    return {
      events: ['coin-price-updated', 'wallet-error'],
    }
  }

  /**
   * Stop the plugin instance.
   */
  function stop() {
    debug('Plugin stopping')

    dataStream.destroy()
  }

  return {
    start,
    stop,
  }
}

module.exports = createPlugin
