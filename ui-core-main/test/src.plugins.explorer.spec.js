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

const {
  getEventDataCreators
} = require('../src/plugins/tokens/events')

const explorer = proxyquire('../src/plugins/explorer', {
  './indexer': () => ({ getTransactions: () => Promise.resolve([]) })
})()

const should = chai.use(chaiAsPromised).should()

const config = { debug: true, explorerDebounce: 50 }
const web3 = new Web3()

describe('Explorer plugin', function () {
  describe('refreshTransaction', function () {
    it('should refresh a single out ETH tx', function (done) {
      let stateChanged = false

      const end = once(function (err) {
        if (err) {
          done(err)
        }
        try {
          stateChanged.should.equal(true, 'State not changed')
          done()
        } catch (err) {
          done(err)
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
        blockNumber: 0,
        from: address,
        logs: [],
        to: randomAddress()
      }

      eventBus.on('wallet-error', function (err) {
        end(new Error(err.message))
      })
      eventBus.on('wallet-state-changed', function (data) {
        try {
          data.should.deep.equal({
            [walletId]: {
              addresses: {
                [address]: {
                  transactions: [{
                    meta: {
                      contractCallFailed: false
                    },
                    receipt,
                    transaction
                  }]
                }
              }
            }
          })
          stateChanged = true
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

      api.refreshTransaction(hash, address)
        .then(noop)
        .then(end)
        .catch(end)
    })

    it('should refresh a single in MET tx', function (done) {
      let stateChanged = false

      const end = once(function (err) {
        if (err) {
          done(err)
        }
        try {
          stateChanged.should.equal(true, 'State not changed')
          done()
        } catch (err) {
          done(err)
        }
        explorer.stop()
      })

      const eventBus = new EventEmitter()

      const address = randomAddress()
      const contractAddress = randomAddress()
      const fromAddress = randomAddress()
      const hash = randomTxId()
      const tokenValue = '1'
      const walletId = 1

      const transaction = {
        gasPrice: 0,
        hash,
        value: 0
      }
      const { eth } = web3
      const logs = [{
        transactionHash: hash,
        address: contractAddress,
        data: eth.abi.encodeParameters(['uint256'], [tokenValue]),
        topics: [
          eth.abi.encodeEventSignature('Transfer(address,address,uint256)'),
          eth.abi.encodeParameter('address', fromAddress),
          eth.abi.encodeParameter('address', address)
        ]
      }]
      const receipt = {
        blockNumber: 0,
        from: fromAddress,
        logs,
        to: contractAddress
      }

      eventBus.on('wallet-error', function (err) {
        end(new Error(err.message))
      })
      eventBus.on('wallet-state-changed', function (data) {
        try {
          data.should.deep.equal({
            [walletId]: {
              addresses: {
                [address]: {
                  transactions: [{
                    meta: {
                      contractCallFailed: false,
                      tokens: {
                        [contractAddress]: {
                          event: 'Transfer',
                          from: fromAddress,
                          to: address,
                          value: tokenValue,
                          processing: false
                        }
                      }
                    },
                    receipt,
                    transaction
                  }]
                }
              }
            }
          })
          stateChanged = true
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

      getEventDataCreators(contractAddress).map(api.registerEvent)

      api.refreshTransaction(hash, address)
        .then(noop)
        .then(end)
        .catch(end)
    })

    it('should skip refreshing an unconfirmed tx', function (done) {
      const end = once(function (err) {
        done(err)
        explorer.stop()
      })

      const eventBus = new EventEmitter()

      const address = randomAddress()
      const hash = randomTxId()

      const transaction = {
        gasPrice: 0,
        hash,
        value: 0
      }
      const receipt = null

      eventBus.on('wallet-error', function (err) {
        end(new Error(err.message))
      })
      eventBus.on('wallet-state-changed', function () {
        end(new Error('Should not have received an update'))
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

      api
        .refreshTransaction(hash, address)
        .then(noop)
        .then(end)
        .catch(end)
    })

    it('should reject on error during refresh', function (done) {
      let stateChanged = false
      let errorFired = false

      const end = once(function (err) {
        if (err) {
          done(err)
        }
        try {
          stateChanged.should.equal(false, 'State changed')
          errorFired.should.equal(true, 'Error not fired')
          done()
        } catch (err) {
          done(err)
        }
        explorer.stop()
      })

      const eventBus = new EventEmitter()

      const address = randomAddress()
      const contractAddress = randomAddress()
      const hash = randomTxId()
      const toAddress = randomAddress()
      const tokenValue = '1'

      const { eth } = web3
      const logs = [{
        transactionHash: hash,
        address: contractAddress,
        data: eth.abi.encodeParameters(['uint256'], [tokenValue]),
        topics: [
          eth.abi.encodeEventSignature('Transfer(address,address,uint256)'),
          eth.abi.encodeParameter('address', address),
          eth.abi.encodeParameter('address', toAddress)
        ]
      }]
      const receipt = {
        blockNumber: 0,
        from: address,
        logs,
        to: contractAddress
      }

      eventBus.on('wallet-error', function () {
        errorFired = true
      })
      eventBus.on('wallet-state-changed', function () {
        stateChanged = true
      })

      const responses = {
        eth_subscribe: () => ({ }),
        eth_getBlockByNumber: () => ({ number: 0 }),
        eth_getTransactionByHash (_hash) {
          _hash.should.equal(hash)
          throw new Error('Fake get transaction error')
        },
        eth_getTransactionReceipt (_hash) {
          _hash.should.equal(hash)
          return receipt
        }
      }
      const plugins = { eth: { web3Provider: new MockProvider(responses) } }

      const { api } = explorer.start({ config, eventBus, plugins })

      getEventDataCreators(contractAddress).map(api.registerEvent)

      api.refreshTransaction(hash, address).should.be.rejectedWith('Fake')
        .then(() => end())
        .catch(end)
    })
  })

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
