'use strict'

const proxyquire = require('proxyquire')
  .noPreserveCache()
  .noCallThru()
require('chai').should()

const burnHashes = new Array(32)
  .fill(null)
  .map((_, i) => i)
  .reduce(
    (all, i) =>
      Object.assign(all, { [i]: `0x${i < 16 ? '0' : ''}${i.toString(16)}` }),
    {}
  )

const MockLumerinContracts = function () {
  this.TokenPorter = {
    methods: {
      exportedBurns: seq => ({
        call: () => Promise.resolve(burnHashes[seq])
      })
    }
  }
}

const porterApi = proxyquire('../src/plugins/lumerin/porter-api', {
  // 'metronome-contracts': MockLumerinContracts
  'lumerin-contracts': MockLumerinContracts
})

const getMerkleRoot = porterApi.getMerkleRoot({}, 'chain')

describe('TokenPorter API', function () {
  it('should return the root of the last 16 burn hashes', function () {
    return getMerkleRoot('24').then(function (root) {
      root.should.equal(
        '0x4742e6b7570e0e65d74d9de33bff85dc6c9c61614e2d099cf8596df4d550a45a'
      )
    })
  })

  it('should return the root of the last 10 burn hashes', function () {
    return getMerkleRoot('8').then(function (root) {
      root.should.equal(
        '0xa57cc928900b85a888948eb49b669d4a3e8d2b4f7ad47f96c6a941e3b7886c7e'
      )
    })
  })
})
