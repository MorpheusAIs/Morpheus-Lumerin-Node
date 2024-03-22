'use strict'

const EventEmitter = require('events')
const pRetry = require('p-retry');
const { createExplorerApis } = require('./api/factory');

/**
 * 
 * @param {string[]} explorerApiURLs 
 * @param {*} web3 
 * @param {*} lumerin 
 * @param {*} eventBus 
 * @returns 
 */
const createExplorer = (explorerApiURLs, web3, lumerin, eventBus) => {
  const apis = createExplorerApis(explorerApiURLs);
  return new Explorer({ apis, lumerin, web3, eventBus })
}

class Explorer {
  /** @type {import('contracts-js').LumerinContext} */
  lumerin = null;

  constructor({ apis, lumerin, web3, eventBus }) {
    this.apis = apis
    this.lumerin = lumerin
    this.web3 = web3
    this.eventBus = eventBus
  }

  /**
   * Returns list of transactions for ETH and LMR token
   * @param {string} from start block
   * @param {string} to end block
   * @param {string} address wallet address
   * @returns {Promise<any[]>}
   */
  async getTransactions(from, to, address, page, pageSize) {
    const lmrTransactions = await this.invoke('getTokenTransactions', from, to, address, this.lumerin._address, page, pageSize)
    const ethTransactions = await this.invoke('getEthTransactions', from, to, address, page, pageSize)

    if (page && pageSize) {
      const hasNextPage = lmrTransactions.length || ethTransactions.length;
      this.eventBus.emit('transactions-next-page', {
        hasNextPage: Boolean(hasNextPage),
        page: page + 1,
      })
    }
    return [...lmrTransactions, ...ethTransactions]
  }

  /**
   * Returns list of transactions for LMR token
   * @param {string} from start block
   * @param {string} to end block
   * @param {string} address wallet address
   * @returns {Promise<any[]>}
   */
  async getETHTransactions(from, to, address) {
    return await this.invoke('getEthTransactions', from, to, address)
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
        stream.emit('data', data)
      })
      .on('error', (err) => {
        stream.emit('error', err)
      })

    this.lumerin.events
      .Transfer({
        filter: {
          from: address,
        },
      })
      .on('data', (data) => {
        stream.emit('data', data)
      })
      .on('error', (err) => {
        stream.emit('error', err)
      })

    setInterval(() => {
      stream.emit('resync')
    }, 60000)

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

  /**
   * Helper method that attempts to make a function call for multiple providers
   * @param {string} methodName 
   * @param  {...any} args 
   * @returns {Promise<any>}
   */
  async invoke(methodName, ...args) {
    return await pRetry(async () => {
      let lastErr

      for (const api of this.apis) {
        try {
          return await api[methodName](...args)
        } catch (err) {
          lastErr = err
        }
      }

      throw new Error(`Explorer error, tried all of the providers without success, ${lastErr}`)
    }, { minTimeout: 5000, retries: 5 })
  }
}

module.exports = createExplorer