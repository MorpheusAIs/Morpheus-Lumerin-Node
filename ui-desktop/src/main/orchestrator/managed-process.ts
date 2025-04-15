import { LogFunctions } from 'electron-log'
import { Pinger, Process, ProcessState, StateInfo } from './process'
import { ChildProcess } from 'node:child_process'
import { spawn } from 'child_process'
import net from 'node:net'
import fs from 'node:fs/promises'
import path from 'node:path'

export type ManagedProcessParams = {
  command: string
  args: string[]
  log: LogFunctions
  redirectProcessOutput?: boolean
  onStateChange?: (stateInfo: StateInfo) => void
  pinger?: Pinger
  ports?: number[]
}

export class ManagedProcess implements Process {
  private static readonly MAX_OUTPUT_LINES = 10

  private command: string
  private args: string[]
  private state: ProcessState = 'stopped'
  private process?: ChildProcess
  private error?: string
  private output: string[] = []
  private log: LogFunctions
  private redirectProcessOutput: boolean
  private onStateChange?: (stateInfo: StateInfo) => void
  private ports?: number[]
  private pinger?: Pinger

  constructor(params: ManagedProcessParams) {
    this.command = path.resolve(params.command)
    this.args = params.args
    this.log = params.log
    this.redirectProcessOutput = params.redirectProcessOutput ?? true
    this.onStateChange = params.onStateChange
    this.pinger = params.pinger
    this.ports = params.ports
  }

  async start(): Promise<void> {
    return new Promise(async (resolve, reject) => {
      if (this.state === 'running') {
        return resolve()
      }

      if (this.state === 'starting') {
        this.log.info('Starting process exists, stopping it before starting again')
        await this.stop().catch((err) => {
          this.log.error('Failed to stop process', err)
          return reject(err)
        })
      }

      try {
        // check if ports are available
        if (this.ports) {
          for (const port of this.ports) {
            const isAvailable = await isPortAvailable(port)
            if (!isAvailable) {
              throw new Error(`Port ${port} is not available`)
            }
          }
        }

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
          if (this.output.length > ManagedProcess.MAX_OUTPUT_LINES) {
            this.output.shift()
          }
        })
        child.stderr.on('data', (data: Buffer) => {
          const errorLine = data.toString('utf-8').trimEnd()

          if (this.redirectProcessOutput) {
            this.log.error('\n\t' + errorLine)
          }
          this.output.push(errorLine)
          if (this.output.length > ManagedProcess.MAX_OUTPUT_LINES) {
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

        resolve()
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
          return reject(err)
        }
      }, timeout)

      this.process.once('close', () => {
        clearTimeout(timeoutId)
        this.log.info('process stopped')
        this.setState('stopped', 'Process stopped')
        return resolve()
      })

      const res = this.process.kill('SIGTERM')
      if (!res) {
        const err = new Error(`process failed to stop`)
        this.log.error(err.message)
        this.setState('stopped', err.message)
        return reject(err)
      }
    })
  }

  async reset() {
    await this.stop()
    this.setState('pending', null)
  }

  async ping(timeoutArg?: number) {
    if (this.pinger) {
      try {
        await this.pinger.ping(timeoutArg)
      } catch (error) {
        await this.stop()
        this.setState('stopped', `Health check failed: ${(error as Error).message}`)
        throw error
      }
    }

    if (this.state !== 'running') {
      this.setState('running')
    }
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

  getState() {
    return this.state
  }

  getError() {
    return this.error
  }

  getOutput() {
    return this.output.join('\n')
  }

  isExternal(): boolean {
    return false
  }
}

async function isPortAvailable(port: number): Promise<boolean> {
  return new Promise((resolve) => {
    const server = net.createServer()

    server.once('error', () => {
      // Port is in use
      resolve(false)
    })

    server.once('listening', () => {
      // Port is available, now close the server
      server.close(() => {
        resolve(true)
      })
    })

    server.listen(port, '127.0.0.1')
  })
}
