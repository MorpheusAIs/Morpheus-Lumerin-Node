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
  /**
   * True when we're tracking an instance we did not spawn (adopted because it
   * was already listening and healthy). We have no child handle for it, so
   * stop() cannot signal it — surfaced honestly rather than silently no-oping.
   */
  private adopted = false

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
        // Check ports before spawning. A busy port has two very different
        // causes and the old code conflated them into one opaque
        // "Port X is not available":
        //
        //   1. A healthy instance of *this* service is already listening —
        //      typically one we spawned and lost the handle to, or one left
        //      over from a previous run. Adopt it instead of failing.
        //   2. Something else owns the port. Say so, with the port number, so
        //      the user has a chance of acting on it.
        if (this.ports) {
          const busyPorts: number[] = []
          for (const port of this.ports) {
            if (!(await isPortAvailable(port))) {
              busyPorts.push(port)
            }
          }

          if (busyPorts.length) {
            const healthy = await this.isHealthy()
            if (healthy) {
              this.log.info(
                `port(s) ${busyPorts.join(', ')} already serving a healthy instance; adopting it`
              )
              this.adopted = true
              this.setState('running', null)
              return resolve()
            }

            throw new Error(
              `Port ${busyPorts.join(', ')} is already in use by another process and is not ` +
                `responding as a healthy ${path.basename(this.command)}. ` +
                `Close whatever is using it, then restart this service from Settings.`
            )
          }
          this.adopted = false
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

          // Capture the pre-transition state FIRST. The previous version called
          // setState('stopped', ...) and *then* tested `this.state === 'starting'`
          // — which setState had just overwritten, so the branch was unreachable
          // and the reject never fired. A process that died during startup left
          // start() hanging in the health-check poll instead of failing fast.
          // (The reject string was also single-quoted, so `${code}` was literal.)
          const wasStarting = this.state === 'starting'
          this.setState('stopped', errMessage)

          if (wasStarting) {
            return reject(new Error(`Process exited during startup (code ${code})`))
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

  /** Best-effort health probe that never throws. */
  private async isHealthy(timeoutMs = 1500): Promise<boolean> {
    if (!this.pinger) {
      return false
    }
    return this.pinger
      .ping(timeoutMs)
      .then(() => true)
      .catch(() => false)
  }

  async stop(): Promise<void> {
    this.log.info('stopping process started')

    if (this.adopted && !this.process) {
      // Adopted instance: no child handle, so there is nothing to signal.
      // Drop our claim on it so a subsequent start() re-evaluates the port
      // rather than assuming the old state still holds.
      this.log.warn(
        'this instance was adopted (not spawned by the app) and cannot be stopped from here'
      )
      this.adopted = false
      this.setState('stopped', 'Process was started outside the app')
      return
    }

    if (!this.process || this.state === 'stopped') {
      this.log.info('stopping process which already stopped')
      return
    }

    const timeout = 5000

    // Deliberately never rejects: stop() is called from quit and restart paths
    // where a failure to stop must not abort the rest of the shutdown sequence.
    return new Promise((resolve) => {
      if (!this.process) {
        this.log.info('attempt to stop process which never started')
        return resolve()
      }

      if (this.state === 'stopped') {
        this.log.info('attempt to stop process which already stopped')
        return resolve()
      }

      let forceResolveId: NodeJS.Timeout | undefined

      // Escalation ladder: SIGTERM -> SIGKILL. The previous code escalated to
      // SIGINT, which is *weaker* than the SIGTERM already sent, so a wedged
      // proxy-router was never actually killed. It survived app quit, kept
      // holding port 8082, and the next launch either failed the port check or
      // (worse) detected the zombie as an "external" process it could not
      // manage. That is the root of the "restart does nothing" failure loop.
      const killTimeoutId = setTimeout(() => {
        if (!this.process) {
          return resolve()
        }
        this.log.warn(`shutdown timed out after ${timeout}ms, sending SIGKILL`)
        this.process.kill('SIGKILL')

        // SIGKILL cannot be trapped, but the 'close' event still has to make it
        // back to us. If it doesn't, don't hang the quit path forever.
        forceResolveId = setTimeout(() => {
          this.log.error('process did not exit after SIGKILL; giving up on it')
          this.setState('stopped', 'Process did not exit after SIGKILL')
          resolve()
        }, 3000)
      }, timeout)

      this.process.once('close', () => {
        clearTimeout(killTimeoutId)
        if (forceResolveId) {
          clearTimeout(forceResolveId)
        }
        this.log.info('process stopped')
        this.setState('stopped', 'Process stopped')
        return resolve()
      })

      const res = this.process.kill('SIGTERM')
      if (!res) {
        // kill() returning false means the signal could not be delivered — the
        // process is already gone. Treat that as a successful stop rather than
        // rejecting, otherwise stopAll() aborts partway through and leaves the
        // remaining services running after the window closes.
        clearTimeout(killTimeoutId)
        this.log.warn('SIGTERM could not be delivered; process already exited')
        this.setState('stopped', 'Process stopped')
        return resolve()
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
