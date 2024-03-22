'use strict'

const chai = require('chai')
const chaiAsPromised = require('chai-as-promised')
const nock = require('nock')

const createIndexer = require('../src/plugins/explorer/indexer')

const { randomAddress, randomTxId } = require('./utils')

chai.use(chaiAsPromised).should()

describe('Indexer', function () {
  before(function () {
    nock.disableNetConnect()
  })

  beforeEach(function () {
    nock.cleanAll()
  })

  it('should query the Lumerin indexer for transactions', function () {
    const config = {
      chainId: 1,
      indexerUrl: 'http://localhost:3005',
      useNativeCookieJar: true
    }
    const eventBus = null
    const indexer = createIndexer(config, eventBus)

    const address = randomAddress()
    const transactions = [randomTxId()]

    const scope = nock(config.indexerUrl)
      .get(`/addresses/${address}/transactions`)
      .query(() => true)
      .reply(200, transactions)

    return indexer.getTransactions(0, 1, address).then(function (list) {
      list.should.deep.equals(transactions)
      scope.done()
    })
  })

  it('should query BlockScout for ETC mainnet transactions', function () {
    const config = {
      chainId: 61,
      useNativeCookieJar: true
    }
    const eventBus = null
    const indexer = createIndexer(config, eventBus)

    const address = randomAddress()
    const transactions = [randomTxId()]

    const scope = nock('https://blockscout.com')
      .get('/etc/mainnet/api/')
      .query(q => q.address === address)
      .reply(200, { status: '1', result: transactions.map(hash => ({ hash })) })

    return indexer.getTransactions(0, 1, address).then(function (list) {
      list.should.deep.equals(transactions)
      scope.done()
    })
  })

  it('should query BlockScout and parse errors', function () {
    const config = {
      chainId: 61,
      useNativeCookieJar: true
    }
    const eventBus = null
    const indexer = createIndexer(config, eventBus)

    const address = randomAddress()
    const message = 'Error message'

    const scope = nock('https://blockscout.com')
      .get('/etc/mainnet/api/')
      .query(q => q.address === address)
      .reply(200, { status: '0', result: [], message })

    return indexer
      .getTransactions(0, 1, address)
      .should.be.rejectedWith(message)
      .then(function () {
        scope.done()
      })
  })

  after(function () {
    nock.enableNetConnect()
  })
})
