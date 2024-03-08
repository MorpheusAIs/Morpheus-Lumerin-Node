const AxiosDigestAuth = require('@mhoc/axios-digest-auth')
const cheerio = require('cheerio')

const { ConfigurationStrategyInterface } = require('./strategy.interface')

/**
 * @class
 * @implements {ConfigurationStrategyInterface}
 */
class AntMinerStrategy {
  constructor(host) {
    this.host = host
    this.api = null
  }

  /**
   * @returns {Promise<Boolean>}
   */
  async isAvailable() {
    const api = await this.#createApi()
    return !!api
  }

  /**
   * Adds new pool to configuration and set highest priority
   *
   * @param {String} pool
   * @param {String} poolUser
   * @returns {Promise<void>}
   */
  async setPool(pool, poolUser) {
    try {
      if (!this.api) {
        throw new Error('No api available')
      }
      const config = await this.#getCurrentConfig()
      if (!config) {
        throw new Error('Cannot fetch current configuragion')
      }
      const result = await this.api.request({
        url: `http://${this.host}/cgi-bin/set_miner_conf.cgi`,
        method: 'POST',
        // We should send all fields otherwise config will be reset
        data: new URLSearchParams({
          ...config,
          _ant_pool1url: pool,
          _ant_pool1user: poolUser,
          _ant_pool1pw: '123',
        }),
      })
      if (result.status !== 200) {
        throw new Error(
          `Request failed with status: ${result.status}, data: ${result.data}`
        )
      }
      return
    } catch (err) {
      throw err;
    }
  }

  /**
   * @returns {Promise<Object | null>}
   */
  async #getCurrentConfig() {
    try {
      const result = await this.api.request({
        url: `http://${this.host}/cgi-bin/minerConfiguration.cgi`,
        method: 'GET',
      })
      const $ = cheerio.load(result.data)
      const data = JSON.parse(
        result.data.slice(
          result.data.indexOf('ant_data = {') + 11,
          result.data.indexOf('};') + 1
        )
      )
      return {
        _ant_pool1url: data.pools[0].url,
        _ant_pool1user: data.pools[0].user,
        _ant_pool1pw: data.pools[0].pass,
        _ant_pool2url: data.pools[1].url,
        _ant_pool2user: data.pools[1].user,
        _ant_pool2pw: data.pools[1].pass,
        _ant_pool3url: data.pools[2].url,
        _ant_pool3user: data.pools[2].user,
        _ant_pool3pw: data.pools[2].pass,
        _ant_nobeeper: $('#ant_beeper').attr('checked') === 'true',
        _ant_notempoverctrl: $('#ant_tempoverctrl').attr('checked') === 'true',
        _ant_fan_customize_switch:
          $('#ant_fan_customize_value').attr('checked') === 'true',
        _ant_fan_customize_value: data['bitmain-fan-pwm'] || '',
        _ant_freq: data['bitmain-freq'],
        _ant_voltage: data['bitmain-voltage'] || '0725',
      }
    } catch (err) {
      console.log(err)
      return null
    }
  }

  /**
   * @returns {Promise<AxiosDigestAuth.AxiosDigestAuth | null>}
   */
  async #createApi() {
    const defaultCredentials = [
      { username: 'root', password: 'root' },
      { username: 'admin', password: 'admin' },
    ]
    for (const credentials of defaultCredentials) {
      try {
        const api = new AxiosDigestAuth.default(credentials)
        const result = await api.request({
          url: `http://${this.host}/cgi-bin/minerConfiguration.cgi`,
          method: 'GET',
        })
        if (
          result.status === 200 &&
          result.headers['content-type'] === 'text/html'
        ) {
          this.api = api
          return this.api
        }
        return null
      } catch (err) {
        console.log(err)
        return null
      }
    }
  }
}

module.exports = { AntMinerStrategy }
