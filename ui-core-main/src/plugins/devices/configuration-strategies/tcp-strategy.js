const { CGMinerApi } = require('../cgminer-api')
const { Status } = require('../consts')

const { ConfigurationStrategyInterface } = require('./strategy.interface')

/**
 * @class
 * @implements {ConfigurationStrategyInterface}
 */
class TcpConfigurationStrategy {
  constructor(host, abort) {
    this.host = host
    this.abort = abort
  }

  /**
   * @returns {Promise<Boolean>}
   */
  async isAvailable() {
    const api = new CGMinerApi()
    await api.connect({ host: this.host, abort: this.abort })
    return api.hasPrivilegedAccess()
  }

  /**
   * Adds new pool to configuration and set highest priority
   *
   * @param {String} pool
   * @param {String} poolUser
   * @returns {Promise<void>} Returns true if successfully updated configuration
   */
  async setPool(pool, poolUser) {
    try {
      const api = new CGMinerApi()
      await api.connect({ host: this.host, abort: this.abort })

      const { id } = await api.addPool(pool, poolUser)
      if (!api.isSocketOpen()) {
        await api.reconnect()
      }
      await api.switchPool(id)
      return
    } catch (err) {
      throw err;
    }
  }
}

module.exports = { TcpConfigurationStrategy }
