'use strict'

const EventEmitter = require('events')
const pRetry = require('p-retry');
const { createExplorerApis } = require('./api/factory');

/**
 * 
 * @param {*} config 
 * @param {*} web3 
 * @param {*} eventBus 
 * @returns 
 */
const createExplorer = (config, web3, eventBus) => {
  const explorerApiURLs = config.explorerApiURLs;
  const apis = createExplorerApis(explorerApiURLs);
  return new Explorer({ apis, web3, eventBus, config })
}

class Explorer {
  /** @type {import('contracts-js').LumerinContext} */
  lumerin = null;

  constructor({ apis, web3, eventBus, config }) {
    this.apis = apis
    this.web3 = web3
    this.eventBus = eventBus
    this.config = config
  }

  /**
   * Returns list of transactions for ETH and LMR token
   * @returns {Promise<any[]>}
   */
  async getTransactions(page, pageSize) {
    try {
      const transactions = await this.getTransactionsGateway(page, pageSize);

      // OLD CALL TO BLOCKCHAIN
      // const lmrTransactions = await this.invoke('getTokenTransactions', from, to, address, this.lumerin._address, page, pageSize)
      // const ethTransactions = await this.invoke('getEthTransactions', from, to, address, page, pageSize)
  
      if (page && pageSize) {
        const hasNextPage = transactions.length;
        this.eventBus.emit('transactions-next-page', {
          hasNextPage: Boolean(hasNextPage),
          page: page + 1,
        })
      }
      return [...transactions]
    }
    catch(e) {
      console.log("Error", e);
      return [];
    }
  }

 async getTransactionsGateway(page = 1, size = 15) {
    try {
        const path = `${this.config.chain.localProxyRouterUrl}/blockchain/transactions?page=${page}&limit=${size}`
        const response = await fetch(path);
        const data = await response.json();
        return data?.transactions || [];
    }
    catch (e) {
        console.log("Error", e)
        return [];
    }
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