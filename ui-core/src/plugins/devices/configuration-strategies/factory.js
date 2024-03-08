const { AbortSignal } = require('@azure/abort-controller')

const { AntMinerStrategy } = require('./ant-miner-strategy')
const { TcpConfigurationStrategy } = require('./tcp-strategy')
const { ConfigurationStrategyInterface } = require('./strategy.interface');

class ConfigurationStrategyFactory {
  /**
   * @param {String} host Miner's IP address
   * @param {AbortSignal} abort 
   * @returns {Promise<ConfigurationStrategyInterface|null>}
   */
  static async createStrategy(host, abort) {
    const strategies = [TcpConfigurationStrategy, AntMinerStrategy]
    for (const Strategy of strategies) {
      try {
        const strategy = new Strategy(host, abort)
        const isAvailable = await strategy.isAvailable()
        if (isAvailable) {
          return strategy
        }
      } catch (err) {
        console.log(err)
      }
    }
    return null
  }
}

module.exports = { ConfigurationStrategyFactory }
