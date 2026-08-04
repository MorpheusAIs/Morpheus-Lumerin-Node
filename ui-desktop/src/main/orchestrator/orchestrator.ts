import { app } from 'electron'
import fs from 'fs-extra'
import path from 'node:path'
import { downloadFile } from './downloader'
import logger from '../logger'
import { extractFile } from './unzipper'
import {
  DownloadItem,
  LoadingState,
  OrchestratorConfig,
  OrchestratorStatus
} from './orchestrator.types'
import { Process } from './process'
import { ProcessFactory } from './process-factory'

console.log('Process cwd', process.cwd())
console.log('App path', resolveAppDataPath(''))

export class Orchestrator {
  private proxyRouterProcess?: Process
  private aiRuntimeProcess?: Process
  private ipfsProcess?: Process
  private containerRuntimeProcess?: Process
  private onStateUpdate: (state: LoadingState) => void
  private cfg: OrchestratorConfig
  private log: typeof logger

  // Mutex-style guard: ensures only one startAll pipeline is in flight at a
  // time. Re-entrant callers (e.g. an auto-resume after a successful
  // restartService while the initial startAll hasn't returned yet) await the
  // existing promise instead of racing a second pipeline.
  private startInProgress: Promise<void> | null = null

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
    app.on('quit', () => {
      this.log.warn('Quit event received')
      return this.stopAll()
    })
  }

  async startAll(): Promise<void> {
    if (this.startInProgress) {
      this.log.info('startAll already in progress; awaiting existing run')
      return this.startInProgress
    }
    this.startInProgress = this.runStartupPipeline()
    try {
      await this.startInProgress
    } finally {
      this.startInProgress = null
    }
  }

  private async runStartupPipeline() {
    this.log.info('Orchestrator started')
    await this.resetState()
    this.emitStateUpdate()

    // --- Downloads: proxy-router is required; AI / IPFS are best-effort ---
    await this.downloadProxyRouter()
    await this.downloadOptionalAiRuntime()
    await this.downloadOptionalAiModel()
    await this.downloadOptionalIpfs()

    // --- Startup: proxy-router first (required for UI/onboarding) ---
    // Local AI, IPFS, and Docker are optional. The proxy only *points at*
    // localhost AI/IPFS in config; it does not need them running to boot,
    // open sessions, or serve remote Morpheus models.
    await this.startProxyRouter()
    this.emitStateUpdate()

    await this.startOptionalService('ipfs', () => this.ensureIpfsProcess())
    await this.startOptionalService('aiRuntime', () => this.ensureAiRuntimeProcess())
    await this.startOptionalService('containerRuntime', () => this.ensureContainerRuntimeProcess())
  }

  private async downloadProxyRouter() {
    if (this.cfg.proxyRouter.downloadUrl) {
      await downloadFile(
        this.cfg.proxyRouter.downloadUrl,
        resolveAppDataPath(this.cfg.proxyRouter.fileName),
        (progress) => {
          this.proxyDownloadState.status = progress.status
          this.proxyDownloadState.progress = progress.progress
          this.proxyDownloadState.error = progress.error
          this.emitStateUpdate()
          this.log.info(`Downloading proxy-router: ${progress.bytesDownloaded} bytes`)
        },
        this.log.scope('Proxy-router download')
      )
    }
    this.proxyDownloadState.status = 'success'
    this.emitStateUpdate()
  }

  private async downloadOptionalAiRuntime() {
    try {
      if (this.cfg.aiRuntime.downloadUrl && this.cfg.aiRuntime.extractPath) {
        if (fs.existsSync(resolveAppDataPath(this.cfg.aiRuntime.extractPath))) {
          this.log.info(
            'AI runtime already exists, skipping download',
            resolveAppDataPath(this.cfg.aiRuntime.extractPath)
          )
        } else {
          await downloadFile(
            this.cfg.aiRuntime.downloadUrl,
            resolveAppDataPath(this.cfg.aiRuntime.fileName),
            (progress) => {
              this.aiRuntimeDownloadState.status = progress.status
              this.aiRuntimeDownloadState.progress = progress.progress
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
              this.aiRuntimeDownloadState.status =
                progress.status === 'error' ? 'error' : 'unzipping'
              this.aiRuntimeDownloadState.progress = progress.progress
              this.aiRuntimeDownloadState.error = progress.error
              this.emitStateUpdate()
              this.log.info(`Extracting ai-runtime`, progress)
            }
          )
        }
      }
      this.aiRuntimeDownloadState.status = 'success'
    } catch (err) {
      this.log.error('Optional AI runtime download failed; continuing', err)
      this.aiRuntimeDownloadState.status = 'error'
      this.aiRuntimeDownloadState.error = (err as Error)?.message ?? String(err)
    }
    this.emitStateUpdate()
  }

  private async downloadOptionalAiModel() {
    try {
      if (this.cfg.aiModel.downloadUrl) {
        await downloadFile(
          this.cfg.aiModel.downloadUrl,
          resolveAppDataPath(this.cfg.aiModel.fileName),
          (progress) => {
            this.aiModelDownloadState.status = progress.status
            this.aiModelDownloadState.progress = progress.progress
            this.aiModelDownloadState.error = progress.error
            this.emitStateUpdate()
            this.log.info(`Downloading ai-model: ${progress.bytesDownloaded} bytes`)
          },
          this.log.scope('Ai-model download')
        )
      }
      this.aiModelDownloadState.status = 'success'
    } catch (err) {
      this.log.error('Optional AI model download failed; continuing', err)
      this.aiModelDownloadState.status = 'error'
      this.aiModelDownloadState.error = (err as Error)?.message ?? String(err)
    }
    this.emitStateUpdate()
  }

  private async downloadOptionalIpfs() {
    try {
      if (
        this.cfg.ipfs.downloadUrl &&
        this.cfg.ipfs.extractPath &&
        !fs.existsSync(resolveAppDataPath(this.cfg.ipfs.extractPath))
      ) {
        await downloadFile(
          this.cfg.ipfs.downloadUrl,
          resolveAppDataPath(this.cfg.ipfs.fileName),
          (progress) => {
            this.ipfsDownloadState.status = progress.status
            this.ipfsDownloadState.progress = progress.progress
            this.ipfsDownloadState.error = progress.error
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
    } catch (err) {
      this.log.error('Optional IPFS download failed; continuing', err)
      this.ipfsDownloadState.status = 'error'
      this.ipfsDownloadState.error = (err as Error)?.message ?? String(err)
    }
    this.emitStateUpdate()
  }

  private async startProxyRouter() {
    const proxyFolder = path.dirname(resolveAppDataPath(this.cfg.proxyRouter.runPath))

    await this.writeEnvFile(path.join(proxyFolder, '.env'), this.cfg.proxyRouter.env ?? {})
    await this.writeLocalConfigFile(
      path.join(proxyFolder, 'models-config.json'),
      this.cfg.proxyRouter.modelsConfig
    )
    await this.writeLocalConfigFile(
      path.join(proxyFolder, 'rating-config.json'),
      this.cfg.proxyRouter.ratingConfig
    )

    await this.ensureProxyRouterProcess()
    await this.proxyRouterProcess!.start()
  }

  private async ensureProxyRouterProcess() {
    if (!this.proxyRouterProcess) {
      this.proxyRouterProcess = await ProcessFactory({
        command: resolveAppDataPath(this.cfg.proxyRouter.runPath),
        args: this.cfg.proxyRouter.runArgs || [],
        log: this.log.scope('Proxy-router'),
        redirectProcessOutput: false,
        probe: this.cfg.proxyRouter.probe,
        ports: this.cfg.proxyRouter.ports,
        onStateChange: () => this.emitStateUpdate()
      })
    }
  }

  private async ensureIpfsProcess() {
    if (!this.ipfsProcess) {
      this.ipfsProcess = await ProcessFactory({
        command: resolveAppDataPath(this.cfg.ipfs.runPath),
        args: this.cfg.ipfs.runArgs,
        log: this.log.scope('IPFS'),
        redirectProcessOutput: true,
        probe: this.cfg.ipfs.probe,
        ports: this.cfg.ipfs.ports,
        onStateChange: () => this.emitStateUpdate()
      })
    }
  }

  private async ensureAiRuntimeProcess() {
    if (!this.aiRuntimeProcess) {
      this.aiRuntimeProcess = await ProcessFactory({
        command: resolveAppDataPath(this.cfg.aiRuntime.runPath),
        args: this.cfg.aiRuntime.runArgs,
        log: this.log.scope('Ai-runtime'),
        redirectProcessOutput: false,
        probe: this.cfg.aiRuntime.probe,
        ports: this.cfg.aiRuntime.ports,
        onStateChange: () => this.emitStateUpdate()
      })
    }
  }

  private async ensureContainerRuntimeProcess() {
    if (!this.containerRuntimeProcess) {
      this.containerRuntimeProcess = await ProcessFactory({
        probe: this.cfg.containerRuntime.probe,
        onStateChange: () => this.emitStateUpdate(),
        log: this.log.scope('Container-runtime')
      })
    }
  }

  private async startOptionalService(
    name: string,
    ensure: () => Promise<void>
  ): Promise<void> {
    try {
      await ensure()
      const processMap: Record<string, Process | undefined> = {
        ipfs: this.ipfsProcess,
        aiRuntime: this.aiRuntimeProcess,
        containerRuntime: this.containerRuntimeProcess
      }
      await processMap[name]?.start()
    } catch (err) {
      this.log.error(`Optional service ${name} failed to start; continuing`, err)
    }
    this.emitStateUpdate()
  }

  async stopAll() {
    this.log.info('Orchestrator shutting down')

    // Only stop managed processes
    await this.proxyRouterProcess?.stop()
    this.emitStateUpdate()

    await this.aiRuntimeProcess?.stop()
    this.emitStateUpdate()

    await this.ipfsProcess?.stop()
    this.emitStateUpdate()
  }

  public async restartService(service: keyof OrchestratorConfig) {
    // Ensure the process object exists even if the prior pipeline never
    // reached this service (e.g. aborted before proxy-router was created).
    try {
      if (service === 'proxyRouter') await this.ensureProxyRouterProcess()
      else if (service === 'aiRuntime') await this.ensureAiRuntimeProcess()
      else if (service === 'ipfs') await this.ensureIpfsProcess()
      else if (service === 'containerRuntime') await this.ensureContainerRuntimeProcess()
      else {
        this.log.error(`Service ${service} is not restartable`)
        return
      }
    } catch (err) {
      this.log.error(`Failed to prepare service ${service} for restart`, err)
      this.emitStateUpdate()
      return
    }

    const processMap: Partial<Record<keyof OrchestratorConfig, Process | undefined>> = {
      proxyRouter: this.proxyRouterProcess,
      aiRuntime: this.aiRuntimeProcess,
      ipfs: this.ipfsProcess,
      containerRuntime: this.containerRuntimeProcess
    }
    const process = processMap[service]
    if (!process) {
      this.log.error(`Service ${service} not found`)
      return
    }

    // Only restart managed processes (container runtime is probe-only / external)
    if (process.isExternal()) {
      this.log.warn(`Cannot restart external service ${service}`)
      return
    }

    await process.stop()
    this.emitStateUpdate()

    try {
      await process.start()
    } finally {
      this.emitStateUpdate()
    }

    // Resume optional services that may still be pending after a proxy restart.
    // startAll is idempotent: downloads check fs.exists, and each process's
    // start() short-circuits when already running.
    if (process.getState() === 'running' && !this.requiredServicesRunning()) {
      this.log.info(
        `Service ${service} restarted; resuming startup pipeline for any remaining services`
      )
      try {
        await this.startAll()
      } catch (err) {
        this.log.error('Resume after restart failed', err)
      }
    }
  }

  /** Proxy-router is the only service required for the UI / onboarding to work. */
  private requiredServicesRunning(): boolean {
    return this.proxyRouterProcess?.getState() === 'running'
  }

  async ping(service: keyof OrchestratorConfig): Promise<boolean> {
    const processMap = {
      proxyRouter: this.proxyRouterProcess,
      aiRuntime: this.aiRuntimeProcess,
      ipfs: this.ipfsProcess,
      containerRuntime: this.containerRuntimeProcess
    }

    const process: Process | undefined = processMap[service]
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
          id: 'proxyRouter',
          name: 'Proxy Router',
          status: this.proxyRouterProcess?.getState() ?? 'pending',
          error: this.proxyRouterProcess?.getError(),
          stderrOutput: this.proxyRouterProcess?.getOutput(),
          ports: this.cfg.proxyRouter.ports,
          isExternal: this.proxyRouterProcess?.isExternal()
        },
        {
          id: 'ipfs',
          name: 'IPFS',
          status: this.ipfsProcess?.getState() ?? 'pending',
          error: this.ipfsProcess?.getError(),
          stderrOutput: this.ipfsProcess?.getOutput(),
          ports: this.cfg.ipfs.ports,
          isExternal: this.ipfsProcess?.isExternal()
        },
        {
          id: 'aiRuntime',
          name: 'AI Runtime',
          status: this.aiRuntimeProcess?.getState() ?? 'pending',
          error: this.aiRuntimeProcess?.getError(),
          stderrOutput: this.aiRuntimeProcess?.getOutput(),
          ports: this.cfg.aiRuntime.ports,
          isExternal: this.aiRuntimeProcess?.isExternal()
        },
        {
          id: 'containerRuntime',
          name: 'Container Runtime',
          status: this.containerRuntimeProcess?.getState() ?? 'pending',
          error: this.containerRuntimeProcess?.getError(),
          stderrOutput: this.containerRuntimeProcess?.getOutput(),
          isExternal: this.containerRuntimeProcess?.isExternal()
        }
      ],
      orchestratorStatus
    })
  }

  private calculateOrchestratorStatus(): OrchestratorStatus {
    // Proxy-router download/start is required. Optional service failures must
    // not block the UI — local AI / IPFS / Docker are best-effort.
    if (this.proxyDownloadState.status === 'error' || this.proxyRouterProcess?.getError()) {
      return 'error'
    }

    if (this.proxyDownloadState.status === 'success' && this.requiredServicesRunning()) {
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

  private async resetState() {
    // Only reset processes that aren't already running. This preserves the
    // healthy ones across a resume (e.g. when restartService kicks off a
    // pipeline re-run, we don't want to flap IPFS / AI Runtime).
    this.proxyRouterProcess?.getState() !== 'running' && (await this.proxyRouterProcess?.reset())
    this.aiRuntimeProcess?.getState() !== 'running' && (await this.aiRuntimeProcess?.reset())
    this.ipfsProcess?.getState() !== 'running' && (await this.ipfsProcess?.reset())
    this.containerRuntimeProcess?.getState() !== 'running' &&
      (await this.containerRuntimeProcess?.reset())

    // Preserve `success` download statuses across resume so the UI doesn't
    // flash completed bars back to pending. Clear errors regardless — they
    // belong to the previous attempt.
    for (const dl of [
      this.proxyDownloadState,
      this.aiRuntimeDownloadState,
      this.aiModelDownloadState,
      this.ipfsDownloadState
    ]) {
      dl.error = undefined
      if (dl.status !== 'success') {
        dl.status = 'pending'
      }
    }
  }
}

function resolveAppDataPath(subPath: string) {
  return path.join(app.getPath('userData'), subPath)
}
