import { LogFunctions } from 'electron-log'
import { Pinger, Process, ProcessState, StateInfo } from './process'

/**
 * ExternalProcess represents a process that is managed outside of our application.
 * It can only monitor the process state through HTTP probes and cannot control its lifecycle.
 */

export type ExternalProcessParams = {
  pinger?: Pinger
  onStateChange?: (stateInfo: StateInfo) => void
  healthCheckIntervalMs?: number
  log?: LogFunctions
}

export class ExternalProcess implements Process {
  private state: ProcessState = 'pending' // represents the state of the process as seen by the orchestrator
  private monitoringState: 'running' | 'stopped' = 'stopped' // represents the state of the process monitoring
  private healthCheckIntervalMs: number = 5000
  private pinger?: Pinger
  private error: string | null = null
  private log?: LogFunctions
  private onStateChange?: (stateInfo: StateInfo) => void
  private healthCheckTimer: NodeJS.Timeout | null = null

  constructor(params: ExternalProcessParams) {
    this.pinger = params.pinger
    this.log = params.log
    this.onStateChange = params.onStateChange
    if (params.healthCheckIntervalMs) {
      this.healthCheckIntervalMs = params.healthCheckIntervalMs
    }
  }

  /**
   * For external processes, start() is a no-op since we cannot control the process lifecycle.
   * It only performs an initial ping to check if the process is available.
   */
  async start(): Promise<void> {
    this.log?.info('starting external process monitoring')
    if (this.monitoringState === 'running') {
      this.log?.info('external process monitoring was already running')
      return
    }

    this.setState('starting', null)
    // Just do an initial ping to check if the process is available
    await this.ping()

    // Start regular health checks
    this.startHealthChecks()

    this.log?.info('external process monitoring started')
  }

  /**
   * For external processes, stop() is a no-op since we cannot control the process lifecycle.
   * However, we can stop the health check timer.
   */
  async stop(): Promise<void> {
    this.stopHealthChecks()
    this.log?.info('external process monitoring stopped')
    this.setState('stopped', 'Monitoring stopped')
    this.monitoringState = 'stopped'
  }

  async ping(timeoutMs: number = 3000): Promise<void> {
    try {
      await this.pinger?.ping(timeoutMs)
      this.setState('running', null)
    } catch (err) {
      this.setState('stopped', (err as Error).message)
      throw err
    }
  }

  private setState(newState?: ProcessState, newError?: string | null) {
    let changed = false
    if (newState && this.state !== newState) {
      this.state = newState
      changed = true
    }

    if (newError !== undefined && this.error !== newError) {
      this.error = newError
      changed = true
    }

    if (changed) {
      this.onStateChange?.({ state: this.state, error: this.error ?? undefined })
    }
  }

  private startHealthChecks(): void {
    this.stopHealthChecks() // Clear any existing timer

    this.healthCheckTimer = setInterval(async () => {
      try {
        await this.ping()
      } catch (err) {
        this.log?.error('Health check failed:', err)
      }
    }, this.healthCheckIntervalMs)
    this.monitoringState = 'running'

    this.log?.info(`Started health checks every ${this.healthCheckIntervalMs}ms`)
  }

  private stopHealthChecks(): void {
    if (this.healthCheckTimer) {
      clearInterval(this.healthCheckTimer)
      this.healthCheckTimer = null
      this.log?.info('Stopped health checks')
      this.monitoringState = 'stopped'
    }
  }

  getState(): ProcessState {
    return this.state
  }

  getError(): string | undefined {
    return this.error ?? undefined
  }

  getOutput(): string | undefined {
    return undefined
  }

  isExternal(): boolean {
    return true
  }
}
