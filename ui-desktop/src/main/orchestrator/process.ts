export type ProcessState = 'running' | 'stopped' | 'starting' | 'pending'

export interface ProcessInfo {
  state: ProcessState
  error?: string
  output?: string
}

export interface Process {
  start(): Promise<void>
  stop(): Promise<void>
  ping(timeoutMs?: number): Promise<void>
  getState(): ProcessState
  getError(): string | undefined
  getOutput(): string | undefined
  isExternal(): boolean
}

export interface Pinger {
  ping(timeoutMs?: number): Promise<void>
}

export interface StateInfo {
  state: ProcessState
  error?: string
  output?: string
}
