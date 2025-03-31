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
  private output?: string
  private log: LogFunctions
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
    log: LogFunctions,
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
    this.log = log
    this.onStateChange = onStateChange
    this.healthCheckConfig = healthCheckConfig
  }

  private setState(newState?: ProcessState, error?: string, output?: string) {
    if (newState !== undefined) {
      this.state = newState
    }
    if (error !== undefined) {
      this.error = error
    }
    if (output !== undefined) {
      this.output = output
    }
    this.onStateChange?.({ state: this.state, error: this.error, output: this.output })
  }

  async start() {
    return new Promise(async (resolve, reject) => {
      try {
        const cwd = path.resolve(path.dirname(this.command))

        this.setState('starting')
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

        let outputLines: string[] = []

        // log the stdout and stderr
        child.stdout.on('data', (data: Buffer) => {
          const outputLine = data.toString('utf-8').trimEnd()
          this.log.info('\n\t' + outputLine)
          if (outputLines.length < BackgroundProcess.MAX_OUTPUT_LINES) {
            outputLines.push(outputLine)
          }
        })
        child.stderr.on('data', (data: Buffer) => {
          const errorLine = data.toString('utf-8').trimEnd()
          this.log.error('\n\t' + errorLine)
          if (outputLines.length < BackgroundProcess.MAX_OUTPUT_LINES) {
            outputLines.push(errorLine)
          }
        })

        child.on('close', (code) => {
          const errMessage = `Process closed with code ${code}`
          this.log.info(errMessage)
          this.setState('stopped', errMessage, this.output)
          if (this.state === 'starting') {
            return reject('closed with code ${code}')
          }
        })

        child.on('error', (error) => {
          this.log.error(error.message)
          this.setState(undefined, error.message)
        })

        // Perform health check if configured
        if (this.healthCheckConfig) {
          try {
            const startTime = Date.now()
            const pollInterval =
              this.healthCheckConfig.pollInterval ??
              BackgroundProcess.DEFAULT_HEALTH_CHECK_POLL_INTERVAL
            const timeout =
              this.healthCheckConfig.timeout ?? BackgroundProcess.DEFAULT_HEALTH_CHECK_TIMEOUT
            let isAvailable = false

            while (!isAvailable && this.state === 'starting' && Date.now() - startTime < timeout) {
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

            if (!isAvailable && this.state === 'starting') {
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

        resolve(child)
      } catch (err) {
        this.setState('stopped', (err as Error)?.message)
        return reject(err)
      }
    })
  }

  async stop(): Promise<void> {
    if (!this.process || this.state === 'stopped') {
      return
    }

    const timeout = 5000

    return new Promise((resolve, reject) => {
      if (!this.process) {
        return
      }

      const timeoutId = setTimeout(() => {
        if (!this.process) {
          return
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
        this.log.info('stopped')
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
    })
  }

  getState() {
    return this.state
  }

  getError() {
    return this.error
  }

  getOutput() {
    return this.output
  }
}
