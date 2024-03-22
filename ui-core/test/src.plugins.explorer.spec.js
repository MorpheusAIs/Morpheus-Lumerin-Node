'use strict'

const {
  noop,
  once
} = require('lodash')
const chai = require('chai')
const chaiAsPromised = require('chai-as-promised')
const EventEmitter = require('events')
// const LumerinContracts = require('metronome-contracts')
const LumerinContracts = require('lumerin-contracts')
const proxyquire = require('proxyquire').noPreserveCache().noCallThru()
const Web3 = require('web3')

const {
  randomAddress,
  randomTxId
} = require('./utils')
const MockProvider = require('./utils/mock-provider')

const explorer = proxyquire('../src/plugins/explorer', {
  './indexer': () => ({ getTransactions: () => Promise.resolve([]) })
})()

const should = chai.use(chaiAsPromised).should()

const config = { debug: true, explorerDebounce: 50 }
const web3 = new Web3()

describe('Explorer plugin', function () {
  describe('refreshAllTransactions', function () {
    it('should start from birthblock', function (done) {
      const chain = 'ropsten'
      const { birthblock } = LumerinContracts[chain].Auctions
      const latestBlock = birthblock + 10000

      let receivedFromBlock
      let receivedToBlock

      const responses = {
        eth_subscribe: () => ({ }),
        eth_getBlockByNumber: () => ({ number: latestBlock }),
        eth_getLogs ({ fromBlock, toBlock }) {
          if (receivedFromBlock === undefined) {
            receivedFromBlock = Web3.utils.hexToNumber(fromBlock)
          }
          receivedToBlock = Web3.utils.hexToNumber(toBlock)
          return []
        }
      }

      const eventBus = new EventEmitter()
      const plugins = { eth: { web3Provider: new MockProvider(responses) } }

      const { api } = explorer.start({ config, eventBus, plugins })

      getEventDataCreator(chain).map(api.registerEvent)

      const end = once(function (err) {
        if (err) {
          done(err)
          return
        }
        try {
          should.equal(receivedFromBlock, birthblock)
          should.equal(receivedToBlock, latestBlock)
          done()
        } catch (err) {
          done(err)
        }
        explorer.stop()
      })

      eventBus.on('wallet-error', function (err) {
        end(new Error(err.message))
      })

      eventBus.once('coin-block', function () {
        api.refreshAllTransactions(randomAddress())
          .then(noop)
          .then(end)
          .catch(end)
      })
    })
  })

  describe('logTransaction', function () {
    it('should queue a promise', function (done) {
      let receiptReceived = false
      let promiseResolved = false

      const end = once(function (err) {
        if (err) {
          done(err)
          return
        }
        try {
          receiptReceived.should.equal(true, 'Receipt not received')
          promiseResolved.should.equal(true, 'Promise not resolved')
          done()
        } catch (err) {
          // ignore error
        }
        explorer.stop()
      })

      const eventBus = new EventEmitter()

      const address = randomAddress()
      const hash = randomTxId()
      const walletId = 1

      const transaction = {
        gasPrice: 0,
        hash,
        value: 0
      }
      const receipt = {
        from: address,
        logs: [],
        to: randomAddress(),
        transactionHash: hash
      }

      const promise = Promise.resolve(receipt)

      eventBus.on('wallet-error', function (err) {
        end(new Error(err.message))
      })
      eventBus.on('wallet-state-changed', function (_data) {
        try {
          _data.should.deep.equal({
            [walletId]: {
              addresses: {
                [address]: {
                  transactions: [{
                    transaction,
                    receipt,
                    meta: {
                      contractCallFailed: false
                    }
                  }]
                }
              }
            }
          })
          receiptReceived = true
          end()
        } catch (err) {
          end(err)
        }
      })

      const responses = {
        eth_subscribe: () => ({ }),
        eth_getBlockByNumber: () => ({ number: 0 }),
        eth_getTransactionByHash (_hash) {
          _hash.should.equal(hash)
          return transaction
        },
        eth_getTransactionReceipt (_hash) {
          _hash.should.equal(hash)
          return receipt
        }
      }

      const plugins = { eth: { web3Provider: new MockProvider(responses) } }

      const { api } = explorer.start({ config, eventBus, plugins })

      eventBus.emit('open-wallets', { activeWallet: walletId })

      api.logTransaction(promise, address)
        .then(function ({ receipt: _receipt }) {
          _receipt.should.deep.equal(receipt)
          promiseResolved = true
        })
        .catch(end)
    })
  })

  it('should queue a PromiEvent', function (done) {
    let hashReceived = false
    let receiptReceived = false
    let promiseResolved = false

    const end = once(function (err) {
      if (err) {
        done(err)
        return
      }
      try {
        hashReceived.should.equal(true, 'Transaction hash not received')
        receiptReceived.should.equal(true, 'Receipt not received')
        promiseResolved.should.equal(true, 'Promise not resolved')
        done()
      } catch (err) {
        // ignore error
      }
      explorer.stop()
    })

    const eventBus = new EventEmitter()

    const address = randomAddress()
    const hash = randomTxId()
    const walletId = 1

    const transaction = {
      gasPrice: 0,
      hash,
      value: 0
    }
    const receipt = {
      from: address,
      logs: [],
      to: randomAddress(),
      transactionHash: hash
    }

    const promiEvent = new EventEmitter()

    eventBus.on('wallet-error', function (err) {
      end(new Error(err.message))
    })
    eventBus.on('wallet-state-changed', function (_data) {
      const data = {
        [walletId]: {
          addresses: {
            [address]: {
              transactions: [{
                transaction,
                receipt: null,
                meta: {}
              }]
            }
          }
        }
      }
      if (hashReceived) {
        data[walletId].addresses[address].transactions[0].receipt = receipt
        data[walletId].addresses[address].transactions[0].meta = {
          contractCallFailed: false
        }
      }
      try {
        _data.should.deep.equal(data)
        if (hashReceived) {
          receiptReceived = true
          end()
        } else {
          hashReceived = true
          promiEvent.emit('receipt', receipt)
        }
      } catch (err) {
        end(err)
      }
    })

    const responses = {
      eth_subscribe: () => ({ }),
      eth_getBlockByNumber: () => ({ number: 0 }),
      eth_getTransactionByHash (_hash) {
        _hash.should.equal(hash)
        return transaction
      },
      eth_getTransactionReceipt (_hash) {
        _hash.should.equal(hash)
        return hashReceived ? receipt : null
      }
    }

    const plugins = { eth: { web3Provider: new MockProvider(responses) } }

    const { api } = explorer.start({ config, eventBus, plugins })

    eventBus.emit('open-wallets', { activeWallet: walletId })

    api.logTransaction(promiEvent, address)
      .then(function ({ receipt: _receipt }) {
        _receipt.should.deep.equal(receipt)
        promiseResolved = true
      })
      .catch(end)

    promiEvent.emit('transactionHash', hash)
  })
})
