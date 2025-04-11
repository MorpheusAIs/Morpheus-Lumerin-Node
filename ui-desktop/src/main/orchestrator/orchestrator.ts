import { app } from 'electron'
import fs from 'fs-extra'
import path from 'node:path'
import { BackgroundProcess } from './runner'
import { downloadFile } from './downloader'
import logger from '../logger'
import { extractFile } from './unzipper'
import {
  DownloadItem,
  LoadingState,
  OrchestratorConfig,
  OrchestratorStatus
} from './orchestrator.types'

console.log('Process cwd', process.cwd())
console.log('App path', resolveAppDataPath(''))

export class Orchestrator {
  private proxyRouterProcess?: BackgroundProcess
  private aiRuntimeProcess?: BackgroundProcess
  private ipfsProcess?: BackgroundProcess
  private onStateUpdate: (state: LoadingState) => void
  private cfg: OrchestratorConfig
  private log: typeof logger

  private proxyDownloadState: DownloadItem = {
    name: 'Proxy Router',
    status: 'pending',
    progress: 0
  }

  private aiRuntimeDownloadState: DownloadItem = {
    name: 'AI Runtime',
    status: 'pending',
    progress: 0
  }

  private aiModelDownloadState: DownloadItem = {
    name: 'AI Model',
    status: 'pending',
    progress: 0
  }

  private ipfsDownloadState: DownloadItem = {
    name: 'IPFS',
    status: 'pending',
    progress: 0
  }

  constructor(
    cfg: OrchestratorConfig,
    onStateUpdate: (state: LoadingState) => void,
    log: typeof logger
  ) {
    this.cfg = cfg
    this.log = log
    this.onStateUpdate = onStateUpdate
  }

  async startAll() {
    app.on('quit', () => {
      this.log.warn('Quit event received')
      return this.stopAll()
    })

    this.log.info('Orchestrator started')
    this.emitStateUpdate()

    if (this.cfg.proxyRouter.downloadUrl) {
      await downloadFile(
        this.cfg.proxyRouter.downloadUrl,
        resolveAppDataPath(this.cfg.proxyRouter.fileName),
        (progress) => {
          this.proxyDownloadState.status = 'downloading'
          this.proxyDownloadState.progress = progress.bytesDownloaded / progress.totalBytes!
          this.proxyDownloadState.error = progress.error
          this.emitStateUpdate()
          this.log.info(`Downloading proxy-router: ${progress.bytesDownloaded} bytes`)
        },
        this.log.scope('Proxy-router download')
      )
    }

    this.proxyDownloadState.status = 'success'
    this.emitStateUpdate()

    if (this.cfg.aiRuntime.downloadUrl) {
      if (fs.existsSync(resolveAppDataPath(this.cfg.aiRuntime.extractPath))) {
        this.log.info(
          'AI runtime already exists, skipping download',
          resolveAppDataPath(this.cfg.aiRuntime.extractPath)
        )
        this.aiRuntimeDownloadState.status = 'success'
        this.emitStateUpdate()
      } else {
        await downloadFile(
          this.cfg.aiRuntime.downloadUrl,
          resolveAppDataPath(this.cfg.aiRuntime.fileName),
          (progress) => {
            this.aiRuntimeDownloadState.status = 'downloading'
            this.aiRuntimeDownloadState.progress = progress.bytesDownloaded / progress.totalBytes!
            this.aiRuntimeDownloadState.error = progress.error
            this.emitStateUpdate()
            this.log.info(`Downloading ai-runtime: ${progress.bytesDownloaded} bytes`)
          },
          this.log.scope('Ai-runtime download')
        )

        this.log.info(`unzipping ai runtime`)

        await extractFile(
          resolveAppDataPath(this.cfg.aiRuntime.fileName),
          resolveAppDataPath(this.cfg.aiRuntime.extractPath),
          (progress) => {
            this.aiRuntimeDownloadState.status = progress.status === 'error' ? 'error' : 'unzipping'
            this.aiRuntimeDownloadState.progress = progress.progress
            this.aiRuntimeDownloadState.error = progress.error
            this.emitStateUpdate()
            this.log.info(`Extracting ai-runtime`, progress)
          }
        )
      }
    }

    this.aiRuntimeDownloadState.status = 'success'
    this.emitStateUpdate()

    if (this.cfg.aiModel.downloadUrl) {
      await downloadFile(
        this.cfg.aiModel.downloadUrl,
        resolveAppDataPath(this.cfg.aiModel.fileName),
        (progress) => {
          this.aiModelDownloadState.status = 'downloading'
          this.aiModelDownloadState.progress = progress.bytesDownloaded / progress.totalBytes! || 0
          this.emitStateUpdate()
          this.log.info(`Downloading ai-model: ${progress.bytesDownloaded} bytes`)
        },
        this.log.scope('Ai-model download')
      )
    }
    this.aiModelDownloadState.status = 'success'
    this.emitStateUpdate()

    if (
      this.cfg.ipfs.downloadUrl &&
      !fs.existsSync(resolveAppDataPath(this.cfg.ipfs.extractPath))
    ) {
      await downloadFile(
        this.cfg.ipfs.downloadUrl,
        resolveAppDataPath(this.cfg.ipfs.fileName),
        (progress) => {
          this.ipfsDownloadState.status = 'downloading'
          this.ipfsDownloadState.progress = progress.bytesDownloaded / progress.totalBytes! || 0
          this.emitStateUpdate()
          this.log.info(`Downloading ipfs: ${progress.bytesDownloaded} bytes`)
        },
        this.log.scope('IPFS node download')
      )

      this.log.info(`unzipping ipfs`)

      await extractFile(
        resolveAppDataPath(this.cfg.ipfs.fileName),
        resolveAppDataPath(this.cfg.ipfs.extractPath),
        (progress) => {
          this.ipfsDownloadState.status = progress.status === 'error' ? 'error' : 'unzipping'
          this.ipfsDownloadState.progress = progress.progress
          this.ipfsDownloadState.error = progress.error
          this.emitStateUpdate()
          this.log.info(`Extracting ipfs: ${progress.status} ${progress.progress}`)
        }
      )
    }
    this.ipfsDownloadState.status = 'success'
    this.emitStateUpdate()

    if (!this.ipfsProcess) {
      this.ipfsProcess = new BackgroundProcess(
        resolveAppDataPath(this.cfg.ipfs.runPath),
        this.cfg.ipfs.runArgs,
        {
          log: this.log.scope('IPFS'),
          redirectProcessOutput: true
        },
        () => this.emitStateUpdate(),
        this.cfg.ipfs.probe
      )
    }
    await this.ipfsProcess.start()
    this.emitStateUpdate()

    if (!this.aiRuntimeProcess) {
      this.aiRuntimeProcess = new BackgroundProcess(
        resolveAppDataPath(this.cfg.aiRuntime.runPath),
        this.cfg.aiRuntime.runArgs,
        {
          log: this.log.scope('Ai-runtime'),
          redirectProcessOutput: false
        },
        () => this.emitStateUpdate(),
        this.cfg.aiRuntime.probe
      )
    }
    await this.aiRuntimeProcess.start()
    this.emitStateUpdate()

    const proxyFolder = path.dirname(resolveAppDataPath(this.cfg.proxyRouter.runPath))

    // writting local config files if not exist
    await this.writeEnvFile(path.join(proxyFolder, '.env'), this.cfg.proxyRouter.env)
    await this.writeLocalConfigFile(
      path.join(proxyFolder, 'models-config.json'),
      this.cfg.proxyRouter.modelsConfig
    )
    await this.writeLocalConfigFile(
      path.join(proxyFolder, 'rating-config.json'),
      this.cfg.proxyRouter.ratingConfig
    )

    if (!this.proxyRouterProcess) {
      this.proxyRouterProcess = new BackgroundProcess(
        resolveAppDataPath(this.cfg.proxyRouter.runPath),
        this.cfg.proxyRouter.runArgs || [],
        {
          log: this.log.scope('Proxy-router'),
          redirectProcessOutput: false
        },
        () => this.emitStateUpdate(),
        this.cfg.proxyRouter.probe
      )
    }
    await this.proxyRouterProcess.start()
    this.emitStateUpdate()
  }

  async stopAll() {
    this.log.info('Orchestrator shutting down')

    await this.proxyRouterProcess?.stop()
    this.emitStateUpdate()

    await this.aiRuntimeProcess?.stop()
    this.emitStateUpdate()

    await this.ipfsProcess?.stop()
    this.emitStateUpdate()
  }

  public async restartService(service: keyof OrchestratorConfig) {
    const processMap = {
      proxyRouter: this.proxyRouterProcess,
      aiRuntime: this.aiRuntimeProcess,
      ipfs: this.ipfsProcess
    }
    const process: BackgroundProcess | undefined = processMap[service]
    if (!process) {
      this.log.error(`Service ${service} not found`)
      return
    }
    await process.stop()
    this.emitStateUpdate()
    await process.start()
    this.emitStateUpdate()
  }

  async ping(service: keyof OrchestratorConfig): Promise<boolean> {
    const processMap = {
      proxyRouter: this.proxyRouterProcess,
      aiRuntime: this.aiRuntimeProcess,
      ipfs: this.ipfsProcess
    }

    const process: BackgroundProcess | undefined = processMap[service]
    if (!process) {
      const error = `Service ${service} not found`
      this.log.error(error)
      throw new Error(error)
    }
    try {
      await process.ping(3000)
      this.emitStateUpdate()
      return true
    } catch (error) {
      this.log.error(`Service ${service} ping failed`, error)
      this.emitStateUpdate()
      return false
    }
  }

  private emitStateUpdate() {
    const orchestratorStatus = this.calculateOrchestratorStatus()
    this.onStateUpdate({
      download: [
        this.proxyDownloadState,
        this.aiRuntimeDownloadState,
        this.aiModelDownloadState,
        this.ipfsDownloadState
      ],
      startup: [
        {
          id: 'ipfs',
          name: 'IPFS',
          status: this.ipfsProcess?.getState() ?? 'pending',
          error: this.ipfsProcess?.getError(),
          stderrOutput: this.ipfsProcess?.getOutput(),
          ports: this.cfg.ipfs.ports
        },
        {
          id: 'aiRuntime',
          name: 'AI Runtime',
          status: this.aiRuntimeProcess?.getState() ?? 'pending',
          error: this.aiRuntimeProcess?.getError(),
          stderrOutput: this.aiRuntimeProcess?.getOutput(),
          ports: this.cfg.aiRuntime.ports
        },
        {
          id: 'proxyRouter',
          name: 'Proxy Router',
          status: this.proxyRouterProcess?.getState() ?? 'pending',
          error: this.proxyRouterProcess?.getError(),
          stderrOutput: this.proxyRouterProcess?.getOutput(),
          ports: this.cfg.proxyRouter.ports
        }
      ],
      orchestratorStatus
    })
  }

  private calculateOrchestratorStatus(): OrchestratorStatus {
    // Check for any errors in downloads
    const hasDownloadErrors = [
      this.proxyDownloadState,
      this.aiRuntimeDownloadState,
      this.aiModelDownloadState,
      this.ipfsDownloadState
    ].some((item) => item.status === 'error')

    // Check for any errors in startup processes
    const hasStartupErrors = [
      this.ipfsProcess,
      this.aiRuntimeProcess,
      this.proxyRouterProcess
    ].some((process) => process?.getError())

    if (hasDownloadErrors || hasStartupErrors) {
      return 'error'
    }

    // Check if all downloads are complete
    const allDownloadsComplete = [
      this.proxyDownloadState,
      this.aiRuntimeDownloadState,
      this.aiModelDownloadState,
      this.ipfsDownloadState
    ].every((item) => item.status === 'success')

    // Check if all processes are running
    const allProcessesRunning = [
      this.ipfsProcess,
      this.aiRuntimeProcess,
      this.proxyRouterProcess
    ].every((process) => process?.getState() === 'running')

    if (allDownloadsComplete && allProcessesRunning) {
      return 'ready'
    }

    return 'initializing'
  }

  private async writeEnvFile(path: string, env: Record<string, string>) {
    // check if the file exists
    if (fs.existsSync(path)) {
      this.log.info(`Env file already exists: ${path}`)
      return
    }

    const envString = Object.entries(env)
      .map(([key, value]) => `${key}=${value}`)
      .join('\n')
    await fs.writeFile(path, envString)
    this.log.info(`Created env file: ${path}`)
  }

  private async writeLocalConfigFile(filepath: string, content: string) {
    // check if the file exists
    if (fs.existsSync(filepath)) {
      this.log.info(`Config file already exists: ${filepath}`)
      return
    }

    await fs.writeFile(filepath, content)
    this.log.info(`Created config file: ${filepath}`)
  }
}

function resolveAppDataPath(subPath: string) {
  return path.join(app.getPath('userData'), subPath)
}
