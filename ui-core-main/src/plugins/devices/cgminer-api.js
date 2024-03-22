//@ts-check
const { Socket } = require('net')
const { AbortController, AbortSignal } = require('@azure/abort-controller')
const { Status } = require('./consts')

const MINER_API_PORT = 4028
const CGMINER_API_CONNECTION_TIMEOUT = 10000
const CGMINER_API_REQUEST_TIMEOUT = 10000

class CGMinerApi {
  /**
   * @param {Socket} [socket]
   */
  constructor(socket) {
    /**
     * @type {Socket}
     * @public
     */
    this.socket = socket || new Socket({ allowHalfOpen: true })
    /**
     * @type {Object}
     * @private
     */
    this._data = null
  }

  /**
   * @returns {Boolean}
   */
  isSocketOpen() {
    return this.socket.readyState === 'open'
  }

  /**
   * Performs connection over the socket with initial data
   * @returns {Promise<void>}
   */
  async reconnect() {
    this.socket = new Socket({ allowHalfOpen: true })
    return this.connect(this._data)
  }

  /**
   * Performs connection over the socket and if successful calls miner API.
   * Calls onUpdate after each event with received data, allowing to pass
   * events immediately after connection. Emits data.isDone on the last request
   * @param {Object} data
   * @param {String} data.host
   * @param {AbortSignal} data.abort // controls cancellation
   * @returns {Promise<void>}
   */
  connect(data) {
    this._data = data
    return new Promise((resolve, reject) => {
      const { signal } = new AbortController(
        data.abort,
        AbortController.timeout(CGMINER_API_CONNECTION_TIMEOUT)
      )

      const onConnect = () => {
        removeListeners()
        resolve()
      }

      const onError = (err) => {
        removeListeners()
        reject(err)
      }

      const onAbort = () => {
        removeListeners()
        this.socket.destroy()
        reject(new Error('Aborted'))
      }

      const removeListeners = () => {
        this.socket.removeListener('connect', onConnect)
        this.socket.removeListener('error', onError)
        signal.removeEventListener('abort', onAbort)
      }

      if (signal.aborted) {
        return onAbort()
      }

      this.socket.connect(MINER_API_PORT, data.host)
      this.socket.once('connect', onConnect)
      this.socket.once('error', onError)
      signal.onabort = () => onAbort()
    })
  }

  /**
   *
   * @returns {void}
   */
  disconnect() {
    this.socket.destroy()
  }

  /**
   *
   * @returns {Promise<{
   *  deviceModel: String,
   *  hashRateGHS: Number,
   *  poolAddress: String,
   *  poolUser: String
   * }>}
   */
  async getMinerData() {
    const data = await this.batchRequest(['version', 'summary', 'pools'])
    const summary = data['summary'][0].SUMMARY[0]
    const pools = data['pools'][0].POOLS.sort(
      (p1, p2) => p1.Priority - p2.Priority
    )
    return {
      deviceModel: data['version'][0].VERSION[0].Type || 'Unknown',
      hashRateGHS: this.getHashRateGHS(summary),
      poolAddress: pools[0].URL,
      poolUser: pools[0].User,
    }
  }

  /**
   *  Returns true if privileged api is available
   * @returns {Promise<Boolean>}
   */
  async hasPrivilegedAccess() {
    try {
      const data = await this.request(
        {
          command: 'privileged',
          parameter: '0',
        },
        AbortController.timeout(CGMINER_API_REQUEST_TIMEOUT)
      )
      return data.STATUS[0].STATUS === Status.Success
    } catch (err) {
      console.log(err)
      return false
    }
  }

  /**
   *  Adds new pool to the miner configuration
   * @returns {Promise<{ id: string }>} Returns object with id of added pool
   */
  async addPool(pool, poolUser, password) {
    try {
      const data = await this.request(
        {
          command: 'addpool',
          parameter: `${pool},${poolUser},${password}`,
        },
        AbortController.timeout(CGMINER_API_REQUEST_TIMEOUT)
      )
      if (!data || data.STATUS[0].STATUS !== Status.Success) {
        throw new Error(`Cannot add new pool: ${JSON.stringify(data)}`)
      }
      return {
        id: data.id,
      }
    } catch (err) {
      throw err
    }
  }

  /**
   *   Switching pool N to the highest priority (the pool is also enabled)
   * @returns {Promise<void>}
   */
  async switchPool(n) {
    try {
      const data = await this.request(
        {
          command: 'switchpool',
          parameter: n,
        },
        AbortController.timeout(CGMINER_API_REQUEST_TIMEOUT)
      )
      if (!data || data.STATUS[0].STATUS !== Status.Success) {
        throw new Error(`Cannot switch ${n} pool to the highest priority: ${JSON.stringify(data)}`)
      }
      return
    } catch (err) {
      throw err
    }
  }

  getHashRateGHS(stats) {
    const ghs = stats['GHS av']
    if (ghs !== undefined) {
      return ghs
    }

    const mhs = stats['MHS av']
    if (mhs != undefined) {
      return mhs / 1000
    }

    console.log('Cannot detect hashrate', stats)
    return null
  }

  /**
   *
   * @param {String[]} commands
   * @returns {Promise<{data: Record<string, [any]>}>}
   */
  async batchRequest(commands) {
    const command = commands.join('+')
    // TODO: parse response code for each command
    return this.request(
      { command },
      AbortController.timeout(CGMINER_API_REQUEST_TIMEOUT)
    )
  }

  /**
   * Generic miner api request
   * @param {Object} param0
   * @param {String} param0.command rpc to execute
   * @param {String} [param0.parameter] command parameter
   * @param {AbortSignal} abort abort signal
   */
  request({ command, parameter }, abort) {
    return new Promise((resolve, reject) => {
      let responseBuffer = Buffer.from('')

      const onData = (buffer) => {
        if (abort.aborted) {
          return onAbort()
        }
        responseBuffer = Buffer.concat([responseBuffer, buffer])
      }

      const onEnd = () => {
        const data = this.parseJSON(responseBuffer)
        removeListeners()
        resolve(data)
      }

      const onError = (err) => {
        removeListeners()
        reject(err)
      }

      const onAbort = () => {
        removeListeners()
        this.socket.destroy()
        console.log('REQUEST_ABORTED', this.socket.remoteAddress)
        reject(new Error('Aborted'))
      }

      const removeListeners = () => {
        this.socket.removeListener('data', onData)
        this.socket.removeListener('end', onEnd)
        this.socket.removeListener('error', onError)
      }

      if (abort.aborted) {
        return onAbort()
      }

      this.socket.on('data', onData)
      this.socket.once('end', onEnd)
      this.socket.once('error', onError)

      abort.onabort = () => onAbort()

      const request = JSON.stringify({
        command,
        parameter,
      })

      this.socket.write(Buffer.from(request))
    })
  }

  /**
   * Normalizes and parses JSON response
   * @param {Buffer} buffer
   * @returns {Object}
   */
  parseJSON(buffer) {
    const dataString = buffer
      .toString()
      .replace(/\0[\s\S]*$/g, '') // removes null characters at the end
      // removes invalid json (prepended array element for stats command
      // response without comma between array elemets "}{" in BMMiner=2.0.0,
      // API=3.1, Miner=30.0.1.3, Type=Antminer S9i)
      .replace(/\{[^\{\}]*?\}\{/g, '{')
    let data

    try {
      data = JSON.parse(dataString)
      return data
    } catch (err) {
      throw new Error(`Cannot parse miner API response: ${err.message}`)
    }
  }
}

module.exports = { CGMinerApi }
