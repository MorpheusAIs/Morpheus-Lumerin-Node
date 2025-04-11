import { ChildProcess, spawn } from 'node:child_process'
import { LogFunctions } from 'electron-log'
import path from 'node:path'
import fs from 'node:fs/promises'

type ProcessState = 'running' | 'stopped' | 'starting'

interface StateInfo {
  state: ProcessState
  error?: string
  output?: string
}

export class BackgroundProcess {
  private static readonly MAX_OUTPUT_LINES = 10
  private static readonly DEFAULT_HEALTH_CHECK_TIMEOUT = 30000 // 30 seconds
  private static readonly DEFAULT_HEALTH_CHECK_POLL_INTERVAL = 1000 // 1 second

  private command: string
  private args: string[]
  private state: ProcessState = 'stopped'
  private process?: ChildProcess
  private error?: string
  private output: string[] = []
  private log: LogFunctions
  private redirectProcessOutput: boolean
  private onStateChange?: (stateInfo: StateInfo) => void
  private healthCheckConfig?: {
    url: string
    method?: 'GET' | 'POST'
    timeout?: number
    pollInterval?: number
  }

  constructor(
    command: string,
    args: string[],
    logConfig: {
      log: LogFunctions
      redirectProcessOutput?: boolean
    },
    onStateChange?: (stateInfo: StateInfo) => void,
    healthCheckConfig?: {
      url: string
      method?: 'GET' | 'POST'
      timeout?: number
      pollInterval?: number
    }
  ) {
    this.command = path.resolve(command)
    this.args = args
    this.log = logConfig.log
    this.redirectProcessOutput = logConfig.redirectProcessOutput ?? true
    this.onStateChange = onStateChange
    this.healthCheckConfig = healthCheckConfig
  }

  private setState(newState?: ProcessState, error?: string | null) {
    if (newState !== undefined) {
      this.state = newState
    }
    if (error === null) {
      this.error = undefined
    } else if (error !== undefined) {
      this.error = error
    }

    this.onStateChange?.({ state: this.state, error: this.error, output: this.output.join('\n') })
  }

  async start() {
    return new Promise(async (resolve, reject) => {
      if (this.state === 'running') {
        return resolve(this.process)
      }

      if (this.state === 'starting') {
        this.log.info('Starting process exists, stopping it before starting again')
        await this.stop().catch((err) => {
          this.log.error('Failed to stop process', err)
          return reject(err)
        })
      }

      try {
        const cwd = path.resolve(path.dirname(this.command))

        this.setState('starting', null)
        this.log.info('process starting')

        try {
          // Check if file exists and is executable
          await fs.access(this.command, fs.constants.X_OK)
        } catch (err) {
          // If not executable, change permissions
          this.log.info(`Setting executable permissions for ${this.command}`)
          await fs.chmod(this.command, 0o755) // rwxr-xr-x
        }

        const child = spawn(this.command, this.args, { stdio: 'pipe', cwd })
        this.process = child

        // log the stdout and stderr
        child.stdout.on('data', (data: Buffer) => {
          const outputLine = data.toString('utf-8').trimEnd()
          if (this.redirectProcessOutput) {
            this.log.info('\n\t' + outputLine)
          }
          this.output.push(outputLine)
          if (this.output.length > BackgroundProcess.MAX_OUTPUT_LINES) {
            this.output.shift()
          }
        })
        child.stderr.on('data', (data: Buffer) => {
          const errorLine = data.toString('utf-8').trimEnd()

          if (this.redirectProcessOutput) {
            this.log.error('\n\t' + errorLine)
          }
          this.output.push(errorLine)
          if (this.output.length > BackgroundProcess.MAX_OUTPUT_LINES) {
            this.output.shift()
          }
        })

        child.on('close', (code) => {
          const errMessage = `Process closed with code ${code}`
          this.log.info(errMessage)
          this.setState('stopped', errMessage)
          if (this.state === 'starting') {
            return reject('closed with code ${code}')
          }
        })

        child.on('error', (error) => {
          this.log.error(error.message)
          this.setState(undefined, error.message)
        })

        // Perform health check if configured
        await this.ping()

        resolve(child)
      } catch (err) {
        this.setState('stopped', (err as Error)?.message)
        return reject(err)
      }
    })
  }

  async stop(): Promise<void> {
    this.log.info('stopping process started')
    if (!this.process || this.state === 'stopped') {
      this.log.info('stopping process which already stopped')
      return
    }

    const timeout = 5000

    return new Promise((resolve, reject) => {
      if (!this.process) {
        this.log.info('attempt to stop process which never started')
        return resolve()
      }

      if (this.state === 'stopped') {
        this.log.info('attempt to stop process which already stopped')
        return resolve()
      }

      const timeoutId = setTimeout(() => {
        if (!this.process) {
          this.log.info('attempt to stop process which never started')
          return resolve()
        }
        this.log.warn(`shutdown timed out after ${timeout}ms, killing process`)
        if (!this.process.kill('SIGINT')) {
          const err = new Error(`failed to kill process`)
          this.log.error(err)
          this.setState('stopped', err.message)
          reject(err)
        }
        this.log.info('process killed')
        this.setState('stopped')
        resolve()
      }, timeout)

      this.process.once('close', () => {
        clearTimeout(timeoutId)
        this.log.info('process stopped')
        this.setState('stopped')
        resolve()
      })

      const res = this.process.kill('SIGINT')
      if (!res) {
        const err = new Error(`[${name}] failed to stop`)
        this.log.error(err.message)
        this.setState('stopped', err.message)
        reject(err)
      }
      resolve()
    })
  }

  async ping(timeoutArg?: number) {
    if (this.healthCheckConfig) {
      const timeout =
        timeoutArg ??
        this.healthCheckConfig.timeout ??
        BackgroundProcess.DEFAULT_HEALTH_CHECK_TIMEOUT
      try {
        const startTime = Date.now()
        const pollInterval =
          this.healthCheckConfig.pollInterval ??
          BackgroundProcess.DEFAULT_HEALTH_CHECK_POLL_INTERVAL
        let isAvailable = false

        while (!isAvailable && Date.now() - startTime < timeout) {
          try {
            const response = await fetch(this.healthCheckConfig!.url, {
              method: this.healthCheckConfig!.method || 'GET'
            })
            if (response.ok) {
              isAvailable = true
              this.log.info('Service health check passed')
              this.setState('running')
              break
            }
            const resBody = await response.text()
            this.log.info(
              `Health check attempt failed with status ${response.status}, body "${resBody}". Retrying...`
            )
          } catch (error) {
            this.log.info(
              'Health check attempt failed, retrying...',
              this.healthCheckConfig!.url,
              error
            )
          }

          // Wait before next attempt
          await new Promise((resolve) => setTimeout(resolve, pollInterval))
        }

        if (!isAvailable) {
          throw new Error(
            `Health check failed after ${timeout}ms - service did not become available`
          )
        }
      } catch (error: any) {
        this.log.error('Health check failed:', error)
        // Only set error if there isn't one already
        if (!this.error) {
          this.setState('stopped', `Health check failed: ${error?.message}`)
        }
        await this.stop()
        throw error
      }
    } else {
      // If no health check is configured, set state to running immediately
      this.setState('running')
    }
  }

  getState() {
    return this.state
  }

  getError() {
    return this.error
  }

  getOutput() {
    return this.output.join('\n')
  }
}
