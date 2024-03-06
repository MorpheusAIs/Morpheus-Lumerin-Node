'use strict'

const axios = require('axios')
const EventEmitter = require('events')

const createExplorer = (chainId, web3, lumerin) => {
  let baseURL;
  switch (chainId.toString()) {
    case 'mainnet':
    case '1':
      baseURL = 'https://api.etherscan.io/api'
      break
    case 'goerli':
    case '5':
      baseURL = 'https://api-goerli.etherscan.io/api'
      break
    default:
      throw new Error(`Unsupported chain ${chainId}`)
  }
  const api = axios.create({
    baseURL,
  });

  return new Explorer({ api, lumerin, web3 });
}

class Explorer {
  constructor({ api, lumerin, web3 }) {
    this.api = api;
    this.lumerin = lumerin;
    this.web3 = web3;
  }

  async getTransactions(from, to, address) {
    const params = {
      module: 'account',
      action: 'tokentx',
      sort: 'desc',
      contractaddress: this.lumerin.address,
      startBlock: from,
      endBlock: to,
      address,
    }

    const { data } = await this.api.get('/', { params })
    const { status, message, result } = data
    if (status !== '1' && message !== 'No transactions found') {
      throw new Error(result)
    }
    return result.map((transaction) => transaction.hash)
  }

  /**
   * Create a stream that will emit an event each time a transaction for the
   * specified address is indexed.
   *
   * The stream will emit `data` for each transaction. If the connection is lost
   * or an error occurs, an `error` event will be emitted.
   *
   * @param {string} address The address.
   * @returns {object} The event emitter.
   */
  getTransactionStream = (address) => {
    const stream = new EventEmitter()

    this.lumerin.events
      .Transfer({
        filter: {
          to: address,
        },
      })
      .on('data', (data) => {
        const { transactionHash } = data
        stream.emit('data', transactionHash)
      })
      .on('error', (err) => {
        stream.emit('error', err)
      })
    return stream
  }

  getLatestBlock() {
    return this.web3.eth.getBlock('latest').then((block) => {
      return {
        number: block.number,
        hash: block.hash,
        totalDifficulty: block.totalDifficulty,
      }
    })
  }
}

module.exports = createExplorer
