/**
 * @interface
 */
class ConfigurationStrategyInterface {
  /**
   * @returns {Promise<Boolean>}
   */
  async isAvailable() {
    throw new Error('Not implemented')
  }

  /**
   * Adds new pool to configuration and set highest priority
   *
   * @param {String} pool
   * @param {String} poolUser
   * @returns {Promise<void>} Returns true if successfully updated configuration
   */
  async setPool(pool, poolUser) {
    throw new Error('Not implemented')
  }
}

module.exports = { ConfigurationStrategyInterface }
