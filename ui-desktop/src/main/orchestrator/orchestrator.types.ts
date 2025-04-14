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
}

export type LoadingState = {
  download: DownloadItem[]
  startup: StartupItem[]
  orchestratorStatus: OrchestratorStatus
}

type ProbeConfig = {
  url: string
  method?: 'GET' | 'POST'
  interval?: number
  timeout?: number
}

export type OrchestratorConfig = {
  proxyRouter: {
    downloadUrl: string | null
    fileName: string
    runPath: string
    ports: number[] // ports exposed by the service
    runArgs?: string[]
    env: Record<string, string>
    modelsConfig: string
    ratingConfig: string
    probe: ProbeConfig
  }
  aiRuntime: {
    downloadUrl: string
    fileName: string
    extractPath: string
    runPath: string
    ports: number[] // ports exposed by the service
    runArgs: string[]
    probe: ProbeConfig
  }
  aiModel: {
    downloadUrl: string
    fileName: string
  }
  ipfs: {
    downloadUrl: string
    fileName: string
    extractPath: string
    runPath: string
    runArgs: string[]
    ports: number[] // ports exposed by the service
    probe: ProbeConfig
  }
}
