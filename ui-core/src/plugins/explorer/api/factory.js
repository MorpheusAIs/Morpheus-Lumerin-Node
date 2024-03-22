const { BlockscoutApi } = require('./blockscout-api')
const { EtherscanApi } = require('./etherscan-api')

/**
 * @param {string[]} explorerApiURLs 
 */
const createExplorerApis = (explorerApiURLs) => {
  return explorerApiURLs.map(baseURL => {
    const isBlockscoutApi = baseURL.includes('blockscout')
    return isBlockscoutApi ? new BlockscoutApi({ baseURL }) : new EtherscanApi({ baseURL })
  })
}

module.exports = { createExplorerApis }
