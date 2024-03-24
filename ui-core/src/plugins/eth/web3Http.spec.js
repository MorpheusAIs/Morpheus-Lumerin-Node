//@ts-check

const { expect } = require('chai')
const { Web3Http } = require('./web3Http')
const { Lumerin, CloneFactory } = require('contracts-js')

const invalidNode = 'https://arbitrum.llamarpc.com_INVALID'
const validNode = 'https://arbitrum.blockpi.network/v1/rpc/public'
const providerList = [
  invalidNode,
  validNode,
  'https://rpc.ankr.com/arbitrum',
  'https://arbitrum.api.onfinality.io/public',
  'https://arb1-mainnet-public.unifra.io',
  'https://arbitrum-one.public.blastapi.io',
  'https://endpoints.omniatech.io/v1/arbitrum/one/public',
  'https://1rpc.io/arb',
]

describe('Web3 multiple nodes integration tests', () => {
  const lumerinAddress = '0x0FC0c323Cf76E188654D63D62e668caBeC7a525b'
  const cloneFactoryAddress = '0x05C9F9E9041EeBCD060df2aee107C66516E2b9bA'

  it('should work with simple blockchain query', async () => {
    const web3 = new Web3Http(providerList)

    const result = await web3.eth.getBlockNumber()
    expect(typeof result).eq('number')
    expect(web3.currentIndex).eq(1)
  })

  it('should iterate all nodes', async () => {
    const web3 = new Web3Http([
      invalidNode,
      invalidNode,
      invalidNode,
      invalidNode,
      validNode,
    ])

    const result = await web3.eth.getBlockNumber()
    expect(typeof result).eq('number')
    expect(web3.currentIndex).eq(4)
  })

  it('should work with Contract.call()', async () => {
    const web3 = new Web3Http(providerList)
    const lumerin = Lumerin(web3, lumerinAddress)

    const result = await lumerin.methods
      .balanceOf('0x0000000000000000000000000000000000000000')
      .call()
    expect(typeof result).eq('string')
    expect(web3.currentIndex).eq(1)
  })

  it('should work with Contract.send()', async () => {
    const web3 = new Web3Http(providerList)
    const cf = CloneFactory(web3, cloneFactoryAddress)

    try {
      await cf.methods
        .setCreateNewRentalContract(
          '0',
          '0',
          '0',
          '0',
          '0x0000000000000000000000000000000000000000',
          '0'
        )
        .send({
          from: '0x0000000000000000000000000000000000000000',
        })
      expect(1).eq(0)
    } catch (err) {
      expect(err.message.includes('unknown account')).eq(true)
    }
  })

  it('should work with Contract.estimateGas()', async () => {
    const web3 = new Web3Http(providerList)
    const cf = CloneFactory(web3, cloneFactoryAddress)

    try {
      await cf.methods
        .setCreateNewRentalContract(
          '0',
          '0',
          '0',
          '0',
          '0x0000000000000000000000000000000000000000',
          '0'
        )
        .estimateGas({
          from: '0x0000000000000000000000000000000000000000',
        })
      expect(1).eq(0)
    } catch (err) {
      expect(err.message.includes('execution reverted')).eq(true)
      expect(web3.currentIndex).eq(1)
    }
  })

  it('should not iterate if request if invalid/reverted', async () => {
    const web3 = new Web3Http([validNode, validNode, validNode])
    const cf = CloneFactory(web3, cloneFactoryAddress)

    try {
      await cf.methods
        .setCreateNewRentalContract(
          '0',
          '0',
          '0',
          '0',
          '0x0000000000000000000000000000000000000000',
          '0'
        )
        .send({
          from: '0x0000000000000000000000000000000000000000',
        })
      expect(1).eq(0)
    } catch (err) {
      expect(err.message.includes('unknown account')).eq(true)
      expect(web3.currentIndex).eq(0)
    }
  })

  it('should not loop forever', async () => {
    const web3 = new Web3Http([invalidNode, invalidNode, invalidNode])

    try {
      await web3.eth.getBlockNumber()
      expect(1).eq(0)
    } catch (err) {
      expect(web3.retryCount).eq(0)
      expect(web3.currentIndex).eq(0)
    }
  })
})
