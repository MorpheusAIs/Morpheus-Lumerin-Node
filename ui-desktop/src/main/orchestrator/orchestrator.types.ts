export type DownloadStatus = 'pending' | 'downloading' | 'unzipping' | 'success' | 'error'
export type StartupStatus = 'pending' | 'starting' | 'running' | 'stopped'
export type OrchestratorStatus = 'initializing' | 'ready' | 'error'

export type DownloadItem = {
  name: string
  progress: number
  status: DownloadStatus
  error?: string
}

export interface StartupItem {
  id: keyof OrchestratorConfig
  name: string
  status: StartupStatus
  error?: string
  stderrOutput?: string
  ports?: number[]
  isExternal?: boolean // undefined if the process management is not determined yet
}

export type LoadingState = {
  download: DownloadItem[]
  startup: StartupItem[]
  orchestratorStatus: OrchestratorStatus
}

export type ProbeConfig = {
  url: string
  method?: 'GET' | 'POST'
  interval?: number
  timeout?: number
}

type ExternalServiceConfig = {
  downloadUrl: string
  probe: ProbeConfig
}

type ServiceConfig<T extends {} = {}> = {
  downloadUrl: string
  fileName: string
  extractPath?: string
  ports: number[] // ports exposed by the service
  runPath: string
  runArgs?: string[]
  env?: Record<string, string>
  probe: ProbeConfig
} & T

type ArtifactConfig = {
  downloadUrl: string
  fileName: string
}

export type OrchestratorConfig = {
  proxyRouter: ServiceConfig<{
    modelsConfig: string
    ratingConfig: string
  }>
  aiRuntime: ServiceConfig
  aiModel: ArtifactConfig
  ipfs: ServiceConfig
  containerRuntime: ExternalServiceConfig
}
