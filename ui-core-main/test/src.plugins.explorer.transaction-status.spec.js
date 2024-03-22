'use strict'

const chai = require('chai')

const getStatus = require('../src/plugins/explorer/transaction-status')

const should = chai.should()

describe('Transaction status', function () {
  it('should report success status', function () {
    getStatus({}, { blockNumber: 0, status: true }).should.equal(true)
  })

  it('should report failure status', function () {
    getStatus({}, { blockNumber: 0, status: false }).should.equal(false)
  })

  it('should report success status on pre-Byzantine fork', function () {
    const blockNumber = 0
    const gasUsed = 1
    const status = null

    let gas, input, logs

    // not a contract call
    input = '0x'
    gas = 1
    logs = []
    getStatus({ input, gas }, { blockNumber, gasUsed, logs, status })
      .should.equal(true)

    // there are logs
    input = '0x00'
    gas = 1
    logs = [{}]
    getStatus({ input, gas }, { blockNumber, gasUsed, logs, status })
      .should.equal(true)

    // not all gas was used
    input = '0x00'
    gas = 2
    logs = []
    getStatus({ input, gas }, { blockNumber, gasUsed, logs, status })
      .should.equal(true)
  })

  it('should report failure status on pre-Byzantine fork', function () {
    const transaction = { input: '0x00', gas: 1 }
    const receipt = { blockNumber: 0, gasUsed: 1, logs: [], status: null }
    getStatus(transaction, receipt)
      .should.equal(false)
  })

  it('should throw if no receipt is provided', function () {
    should.Throw(() => getStatus(), 'No transaction receipt')
  })
})
