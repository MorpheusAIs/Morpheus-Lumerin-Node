'use strict'

const { once } = require('lodash')
const chai = require('chai')
const EventEmitter = require('events')
const nock = require('nock')
const proxyquire = require('proxyquire')

chai.should()

describe('Rates', function () {
  before(function () {
    nock.disableNetConnect()
  })

  it('should emit coin price', function (done) {
    const symbol = 'ETH'
    const price = 200

    const createPlugin = proxyquire('../src/plugins/rates', {
      'safe-exchange-rate': {
        getExchangeRate (pair) {
          pair.should.equal(`${symbol}:USD`)
          return Promise.resolve(price)
        }
      }
    })

    const plugin = createPlugin()

    const eventBus = new EventEmitter()

    const end = once(function (err) {
      plugin.stop()
      eventBus.removeAllListeners()
      done(err)
    })

    eventBus.on('wallet-error', function ({ inner }) {
      done(inner)
    })

    const start = Date.now()
    let times = 2

    eventBus.on('coin-price-updated', function (priceData) {
      try {
        priceData.should.deep.equal({ token: symbol, currency: 'USD', price })
        times -= 1
        if (!times) {
          const elapsed = Date.now() - start
          elapsed.should.be.closeTo(100, 32)
          end()
        }
      } catch (err) {
        end(err)
      }
    })

    const config = { debug: true, ratesUpdateMs: 100, symbol }

    plugin.start({ config, eventBus })
  })

  after(function () {
    nock.enableNetConnect()
  })
})
