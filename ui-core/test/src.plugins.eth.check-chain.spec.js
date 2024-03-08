'use strict'

const chai = require('chai')
const chaiAsPromised = require('chai-as-promised')

const checkChain = require('../src/plugins/eth/check-chain')

chai.use(chaiAsPromised).should()

const mockWeb3 = ({ id }) => ({
  eth: {
    getChainId: () => Promise.resolve(id)
  }
})

describe('Chain checker', function () {
  it('should resolve if the chain is correct', function () {
    return checkChain(mockWeb3({ id: 1 }), 1)
  })

  it('should reject if the chain does not match', function () {
    return checkChain(mockWeb3({ id: 1 }), 2)
      .should.be.rejectedWith('Wrong chain')
  })
})
