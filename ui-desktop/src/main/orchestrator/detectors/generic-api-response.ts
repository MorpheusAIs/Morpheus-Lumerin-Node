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
}

const DEFAULT_TIMEOUT = 10000
const DEFAULT_POLL_INTERVAL = 1000

export class GenericApiResponseDetector {
  private url: string
  private responseRegexp: RegExp | null
  private method: HttpMethod
  private timeout: number
  private pollInterval: number
  private log: LogFunctions | null

  constructor(params: Params) {
    this.url = params.url
    this.method = params.method ?? 'GET'
    this.responseRegexp = params.responseRegexp ? new RegExp(params.responseRegexp) : null
    this.timeout = params.timeout ?? DEFAULT_TIMEOUT
    this.pollInterval = params.pollInterval ?? DEFAULT_POLL_INTERVAL
    this.log = params.log ?? null
  }

  async ping(timeoutMs?: number): Promise<void> {
    const timeout = timeoutMs ?? this.timeout
    const startTime = Date.now()
    const pollInterval = this.pollInterval

    while (Date.now() - startTime < timeout) {
      try {
        const res = await this.request(this.url, this.method)

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
      transformRequest: function (data, headers) {
        return unixNpipeProtocolTransform(this, data, headers)
      },
      transformResponse: (data) => data,
      timeout: this.pollInterval
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
