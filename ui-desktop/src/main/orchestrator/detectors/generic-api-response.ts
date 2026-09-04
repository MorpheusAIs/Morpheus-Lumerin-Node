import { LogFunctions } from 'electron-log'
import Axios, { AxiosRequestConfig, AxiosRequestHeaders, InternalAxiosRequestConfig } from 'axios'

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'
type Params = {
  url: string
  method?: HttpMethod
  responseRegexp?: string
  timeout?: number
  pollInterval?: number
  log?: LogFunctions
  expectedHeader?: { name: string; value: string }
}

const DEFAULT_TIMEOUT = 10000
const DEFAULT_POLL_INTERVAL = 1000
/** Per-attempt HTTP timeout. See the note in `request()`. */
const REQUEST_TIMEOUT_MS = 5000

export class GenericApiResponseDetector {
  private url: string
  private responseRegexp: RegExp | null
  private method: HttpMethod
  private timeout: number
  private pollInterval: number
  private log: LogFunctions | null
  private expectedHeader?: { name: string; value: string }

  constructor(params: Params) {
    this.url = params.url
    this.method = params.method ?? 'GET'
    this.responseRegexp = params.responseRegexp ? new RegExp(params.responseRegexp) : null
    this.timeout = params.timeout ?? DEFAULT_TIMEOUT
    this.pollInterval = params.pollInterval ?? DEFAULT_POLL_INTERVAL
    this.log = params.log ?? null
    this.expectedHeader = params.expectedHeader
  }

  setExpectedHeader(expectedHeader?: { name: string; value: string }) {
    this.expectedHeader = expectedHeader
  }

  async ping(timeoutMs?: number): Promise<void> {
    const timeout = timeoutMs ?? this.timeout
    const startTime = Date.now()
    const pollInterval = this.pollInterval

    while (Date.now() - startTime < timeout) {
      try {
        const res = await this.request(this.url, this.method)

        if (this.expectedHeader) {
          const actualValue = res.headers[this.expectedHeader.name.toLowerCase()]
          if (actualValue !== this.expectedHeader.value) {
            throw new Error('Health response did not contain the expected process identity')
          }
        }

        if (this.responseRegexp) {
          const isMatch = this.responseRegexp.test(res.data)
          if (!isMatch) {
            throw new Error(`Response body expected ${this.responseRegexp.source}, got ${res.data}`)
          }
        }
        this.log?.info('Service health check passed')
        return
      } catch (error: any) {
        this.log?.info('Ping attempt failed, retrying...', this.url, error?.message)
      }

      // Wait before next attempt
      this.log?.info(`waiting ${pollInterval}ms before next attempt`)
      await new Promise((resolve) => setTimeout(resolve, pollInterval))
    }

    this.log?.info('Service health check timed out')
    throw new Error('Service health check timed out')
  }

  request(uri: string, method: HttpMethod) {
    return Axios.request({
      url: uri,
      method,
      // This detector runs in Electron's main process. Pin the Node adapter so
      // test/browser globals cannot switch it to XHR and apply renderer CORS.
      adapter: 'http',
      transformRequest: function (data, headers) {
        return unixNpipeProtocolTransform(this, data, headers)
      },
      transformResponse: (data) => data,
      // Per-request timeout used to be `this.pollInterval` (1s by default),
      // which meant any service that took longer than a second to answer could
      // *never* pass a health check, no matter how long the overall timeout
      // was. A proxy-router doing initial chain sync routinely exceeds that,
      // so startup would fail and the process got stopped again. Give each
      // attempt a realistic budget, capped by the overall deadline.
      timeout: Math.max(this.pollInterval, REQUEST_TIMEOUT_MS)
    })
  }
}

function unixNpipeProtocolTransform(
  config: InternalAxiosRequestConfig,
  data: any,
  _: AxiosRequestHeaders
): AxiosRequestConfig {
  const [proto, pathname] = config.url?.split('://') ?? []
  if (proto === 'unix' || proto === 'npipe') {
    const [socketPath, apiPath] = pathname.split(':')

    config.socketPath = socketPath
    config.baseURL = 'http://localhost'
    config.url = apiPath
  }

  return data
}
