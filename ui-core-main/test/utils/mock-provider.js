'use strict'

function MockProvider(responses, delay = 50) {
  this._delay = delay
  this._responses = responses
}

MockProvider.prototype.send = function (payload, callback) {
  // eslint-disable-next-line arrow-body-style
  setTimeout(() => {
    try {
      callback(null, {
        id: payload.id,
        jsonrpc: '2.0',
        result: this._responses[payload.method](...payload.params)
      })
    } catch (err) {
      callback(err)
    }
  }, this._delay)
}

MockProvider.prototype.on = function () {
  // as of web3@1.3.6, this method is required to be implemented on a provider
  // but we don't really use it in the tests (so far)
  // eslint-disable-next-line no-console
  console.warn('Suscriptions mocked but not implemented.')
}

module.exports = MockProvider
