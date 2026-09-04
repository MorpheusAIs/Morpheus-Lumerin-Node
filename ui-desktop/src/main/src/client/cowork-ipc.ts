import { app, BrowserWindow, dialog, ipcMain, IpcMainInvokeEvent } from 'electron'
import { createHash } from 'node:crypto'
import path from 'node:path'
import { promises as fs } from 'node:fs'
import log from '../../logger'
import { getAuthHeaders, getActiveWallet, proxyFetch } from './subscriptions/handlers'
import {
  createProject,
  createTask,
  deleteProject,
  deleteTask,
  getProject,
  getTask,
  listProjects,
  listTaskSummaries,
  listTasks,
  recoverInterruptedTasks,
  updateProject
} from './cowork-store'
import {
  cancelCoworkRun,
  coworkRunActive,
  pauseCoworkRun,
  requestCoworkStart,
  resolveCoworkApproval,
  steerCoworkRun
} from './cowork-runner'
import { executeCoworkTool, openCoworkPath } from './cowork-tools'
import {
  CoworkModelOption,
  CoworkModelTarget,
  CoworkProject,
  CoworkTask,
  CoworkTaskEvent,
  CoworkToolCall
} from './cowork.types'
import {
  createCoworkSchedule,
  deleteCoworkSchedule,
  getCoworkSchedule,
  listCoworkSchedules,
  pauseCoworkSchedule,
  resumeCoworkSchedule,
  updateCoworkSchedule
} from './cowork-schedule-store'
import { createCoworkScheduler } from './cowork-scheduler'
import type { CoworkScheduleCadence } from './cowork-schedule.types'
import { runCoworkScheduledOccurrence } from './cowork-scheduled-occurrence'
import { discoverCoworkExtensionCatalog } from './cowork-extension-catalog'
import { assertTrustedRendererEvent } from '../../rendererTrust'
import {
  activeCoworkMarketplaceSessions,
  loadForStableCoworkWallet,
  normalizeCoworkWalletAddress
} from './cowork-marketplace-sessions'

const CHANNEL = {
  listProjects: 'cowork:list-projects',
  createProject: 'cowork:create-project',
  updateProject: 'cowork:update-project',
  deleteProject: 'cowork:delete-project',
  listTasks: 'cowork:list-tasks',
  getTask: 'cowork:get-task',
  createTask: 'cowork:create-task',
  startTask: 'cowork:start-task',
  steerTask: 'cowork:steer-task',
  cancelTask: 'cowork:cancel-task',
  pauseTask: 'cowork:pause-task',
  resolveApproval: 'cowork:resolve-approval',
  deleteTask: 'cowork:delete-task',
  listModelOptions: 'cowork:list-model-options',
  previewArtifact: 'cowork:preview-artifact',
  revealArtifact: 'cowork:reveal-artifact',
  listSchedules: 'cowork:list-schedules',
  createSchedule: 'cowork:create-schedule',
  updateSchedule: 'cowork:update-schedule',
  pauseSchedule: 'cowork:pause-schedule',
  resumeSchedule: 'cowork:resume-schedule',
  deleteSchedule: 'cowork:delete-schedule',
  runScheduleNow: 'cowork:run-schedule-now',
  listExtensions: 'cowork:list-extensions',
  configureExtensions: 'cowork:configure-extensions',
  event: 'cowork:event'
} as const

const ids = /^[a-zA-Z0-9-]{1,160}$/
const approvalModes = new Set(['manual', 'auto', 'skip'])
const MAX_SESSIONS = 1_000
const MODEL_OPTIONS_CACHE_MS = 5_000
const BINARY_ARTIFACT_EXTENSIONS = new Set(['.docx', '.xlsx', '.pptx', '.pdf'])
let registered = false
let recoveryBarrier: Promise<void> = Promise.resolve()
let quitting = false
type ResolvedModelOption = CoworkModelOption & { boundaryFingerprint: string }
const cachedModelOptions = new Map<string, { value: ResolvedModelOption[]; expiresAt: number }>()
const modelOptionsInFlight = new Map<string, Promise<ResolvedModelOption[]>>()

function record(value: unknown): Record<string, any> {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    throw new Error('Invalid Cowork request.')
  return value as Record<string, any>
}

function stringValue(value: unknown, label: string, max: number, allowEmpty = false): string {
  if (typeof value !== 'string') throw new Error(`${label} must be text.`)
  const result = value.trim()
  if (!allowEmpty && !result) throw new Error(`${label} is required.`)
  if (result.length > max) throw new Error(`${label} is too long.`)
  return result
}

function idValue(value: unknown, label = 'ID'): string {
  const result = stringValue(value, label, 160)
  if (!ids.test(result)) throw new Error(`Invalid ${label.toLowerCase()}.`)
  return result
}

function cadenceValue(value: unknown): CoworkScheduleCadence {
  const input = record(value)
  const kind = input.kind
  if (kind === 'manual') return { kind }
  const minute = Number(input.minute)
  if (kind === 'hourly') return { kind, minute }
  const hour = Number(input.hour)
  if (kind === 'daily' || kind === 'weekdays') return { kind, hour, minute }
  if (kind === 'weekly') return { kind, dayOfWeek: Number(input.dayOfWeek), hour, minute }
  throw new Error('Invalid schedule cadence.')
}

function publicProject(
  project: CoworkProject
): Omit<CoworkProject, 'rootPath'> & { folderName: string } {
  const { rootPath, ...visible } = project
  return { ...visible, folderName: path.basename(rootPath) }
}

function publicToolCall(call: CoworkToolCall): CoworkToolCall {
  if (call.function.name !== 'write_file') return call
  try {
    const parsed = JSON.parse(call.function.arguments || '{}')
    const length = typeof parsed.content === 'string' ? parsed.content.length : 0
    return {
      ...call,
      function: {
        ...call.function,
        arguments: JSON.stringify({
          ...parsed,
          content: `[${length} characters hidden from renderer]`
        })
      }
    }
  } catch {
    return { ...call, function: { ...call.function, arguments: '{}' } }
  }
}

function publicTask(
  task:
    | CoworkTask
    | (Omit<CoworkTask, 'agentMessages' | 'toolExecutions'> & {
        agentMessages?: undefined
        toolExecutions?: undefined
      })
): Omit<CoworkTask, 'agentMessages' | 'toolExecutions'> {
  const {
    agentMessages: _agentMessages,
    modelFingerprint: _modelFingerprint,
    toolProtocol: _toolProtocol,
    toolExecutions: _toolExecutions,
    dataAccessApprovedFingerprint: _dataAccessApprovedFingerprint,
    ...visible
  } = task
  if (!visible.pendingApproval) return visible
  return {
    ...visible,
    pendingApproval: {
      ...visible.pendingApproval,
      toolCall: publicToolCall(visible.pendingApproval.toolCall),
      remainingToolCalls: []
    }
  }
}

function visionCapability(model: any): CoworkModelOption['visionCapability'] {
  const tags = Array.isArray(model?.Tags)
    ? model.Tags.map((tag: unknown) => String(tag).toLowerCase().trim())
    : []
  if (tags.some((tag: string) => ['vision', 'multimodal', 'image', 'vlm'].includes(tag)))
    return 'declared'
  const hints = [
    'llava',
    'vision',
    'gpt-4o',
    'gpt-4-turbo',
    'claude-3',
    'claude-4',
    'claude-sonnet',
    'claude-opus',
    'gemini',
    'qwen-vl',
    'qwen2-vl',
    'qwen2.5-vl',
    'internvl',
    'minicpm-v',
    'pixtral',
    'molmo',
    'phi-3-vision',
    'phi-4-multimodal',
    'idefics',
    'cogvlm'
  ]
  const name = String(model?.Name ?? '').toLowerCase()
  return hints.some((hint) => name.includes(hint)) ? 'detected' : 'none'
}

function boundaryFingerprint(kind: string, ...values: unknown[]): string {
  return createHash('sha256')
    .update([kind, ...values.map((value) => String(value ?? ''))].join('\0'), 'utf8')
    .digest('hex')
}

async function remoteModelOptions(walletAddress: string): Promise<ResolvedModelOption[]> {
  const modelsResponse = await proxyFetch<{ models?: any[] }>(
    '/blockchain/models',
    {},
    'Cowork models'
  )
  const sessions: any[] = []
  const limit = 50
  for (let offset = 0; offset < MAX_SESSIONS; offset += limit) {
    const page = await proxyFetch<{ sessions?: any[] }>(
      `/blockchain/sessions/user?user=${encodeURIComponent(walletAddress)}&offset=${offset}&limit=${limit}&order=desc`,
      {},
      'Cowork sessions'
    )
    const values = Array.isArray(page.sessions) ? page.sessions : []
    sessions.push(...values)
    if (values.length < limit) break
  }

  const models = Array.isArray(modelsResponse.models) ? modelsResponse.models : []
  return activeCoworkMarketplaceSessions(sessions, models, Date.now(), walletAddress).map(
    ({ session, model, endsAt }): ResolvedModelOption => ({
      modelId: String(model.Id),
      modelName: String(model.Name || model.Id),
      isLocal: false,
      sessionId: String(session.Id),
      sessionEndsAt: endsAt,
      source: 'marketplace',
      dataBoundary: 'independent-provider',
      visionCapability: visionCapability(model),
      boundaryFingerprint: boundaryFingerprint('marketplace-session', session.Id, model.Id)
    })
  )
}

async function loadModelOptions(walletAddress: string): Promise<ResolvedModelOption[]> {
  return loadForStableCoworkWallet(
    walletAddress,
    () => remoteModelOptions(walletAddress),
    async () => (await getActiveWallet())?.address
  )
}

async function modelOptions(force = false): Promise<ResolvedModelOption[]> {
  const wallet = await getActiveWallet()
  const walletAddress = normalizeCoworkWalletAddress(wallet?.address)
  if (!walletAddress) return []
  const now = Date.now()
  const cached = cachedModelOptions.get(walletAddress)
  if (!force && cached && cached.expiresAt > now) {
    return cached.value
  }
  const existing = modelOptionsInFlight.get(walletAddress)
  if (existing && !force) return existing
  let request: Promise<ResolvedModelOption[]>
  request = loadModelOptions(walletAddress)
    .then((value) => {
      cachedModelOptions.set(walletAddress, {
        value,
        expiresAt: Date.now() + MODEL_OPTIONS_CACHE_MS
      })
      return value
    })
    .finally(() => {
      if (modelOptionsInFlight.get(walletAddress) === request) {
        modelOptionsInFlight.delete(walletAddress)
      }
    })
  modelOptionsInFlight.set(walletAddress, request)
  return request
}

async function requireActiveCoworkSession(): Promise<ResolvedModelOption> {
  const option = (await modelOptions(true)).find(
    (candidate) =>
      !candidate.isLocal &&
      candidate.source === 'marketplace' &&
      Boolean(candidate.sessionId) &&
      Boolean(candidate.sessionEndsAt && candidate.sessionEndsAt > Date.now())
  )
  if (!option) {
    throw new Error(
      'Cowork requires an active Morpheus marketplace session. Choose a model and open a session in Chat first.'
    )
  }
  return option
}

function modelTarget(option: ResolvedModelOption): CoworkModelTarget {
  return {
    modelId: option.modelId,
    modelName: option.modelName,
    isLocal: option.isLocal,
    dataBoundary: option.dataBoundary,
    ...(option.sessionId ? { sessionId: option.sessionId } : {}),
    ...(option.sessionEndsAt ? { sessionEndsAt: option.sessionEndsAt } : {})
  }
}

function publicModelOption(option: ResolvedModelOption): CoworkModelOption {
  const { boundaryFingerprint: _boundaryFingerprint, ...visible } = option
  return visible
}

function emit(event: CoworkTaskEvent): void {
  for (const window of BrowserWindow.getAllWindows()) {
    if (!window.isDestroyed()) window.webContents.send(CHANNEL.event, event)
  }
}

async function extensionCatalog(project: CoworkProject) {
  const userConfigRoot = path.join(app.getPath('userData'), 'CoworkExtensions')
  await fs.mkdir(userConfigRoot, { recursive: true, mode: 0o700 })
  await fs.chmod(userConfigRoot, 0o700)
  return discoverCoworkExtensionCatalog({ projectRoot: project.rootPath, userConfigRoot })
}

function publicExtensionCatalog(
  catalog: Awaited<ReturnType<typeof extensionCatalog>>,
  project: CoworkProject
) {
  const settings = project.extensionSettings ?? {
    folderInstructionsEnabled: false,
    enabledSkillIds: []
  }
  let projectInstructions
  if (catalog.projectInstructions) {
    const { content: _content, ...metadata } = catalog.projectInstructions
    projectInstructions = {
      ...metadata,
      enabled:
        settings.folderInstructionsEnabled &&
        settings.folderInstructionsHash === catalog.projectInstructions.contentHash
    }
  }
  return {
    ...catalog,
    ...(projectInstructions ? { projectInstructions } : {}),
    skills: catalog.skills.map(({ instructions: _instructions, ...skill }) => ({
      ...skill,
      enabled:
        settings.enabledSkillIds.includes(skill.id) &&
        settings.skillInstructionHashes?.[skill.id] === skill.instructionsHash
    }))
  }
}

const scheduler = createCoworkScheduler(
  (occurrence) =>
    runCoworkScheduledOccurrence(occurrence, {
      pauseSchedule: pauseCoworkSchedule,
      refreshModelTarget,
      createTask,
      getTask,
      deleteTask,
      startTask: (taskId) => requestCoworkStart(taskId, getAuthHeaders, emit, refreshModelTarget),
      cancelTask: (taskId) => cancelCoworkRun(taskId, emit)
    }),
  { onError: (error) => log.error(`Cowork schedule failed: ${error.message}`) }
)

async function selectedModel(value: unknown): Promise<CoworkModelTarget> {
  const input = record(value)
  const modelId = stringValue(input.modelId, 'Model ID', 512)
  const sessionId =
    input.sessionId === undefined ? undefined : stringValue(input.sessionId, 'Session ID', 512)
  const option = (await modelOptions(true)).find(
    (candidate) => candidate.modelId === modelId && candidate.sessionId === sessionId
  )
  if (!option) throw new Error('The selected model or session is no longer available.')
  if (option.isLocal || option.source !== 'marketplace' || !option.sessionId) {
    throw new Error('Cowork requires an active Morpheus marketplace session.')
  }
  if (!option.sessionEndsAt || option.sessionEndsAt <= Date.now()) {
    throw new Error('The selected marketplace session has expired.')
  }
  return modelTarget(option)
}

async function refreshModelTarget(
  current: CoworkModelTarget
): Promise<{ model: CoworkModelTarget; fingerprint: string }> {
  const option = (await modelOptions(true)).find(
    (candidate) =>
      candidate.modelId === current.modelId && candidate.sessionId === current.sessionId
  )
  if (!option) throw new Error('The selected model or session is no longer available.')
  if (option.isLocal || option.source !== 'marketplace' || !option.sessionId) {
    throw new Error('Cowork requires an active Morpheus marketplace session.')
  }
  if (!option.sessionEndsAt || option.sessionEndsAt <= Date.now()) {
    throw new Error('The selected marketplace session has expired.')
  }
  return { model: modelTarget(option), fingerprint: option.boundaryFingerprint }
}

function handle(
  channel: string,
  handler: (input: unknown, event: IpcMainInvokeEvent) => unknown
): void {
  ipcMain.handle(channel, async (event, input) => {
    assertTrustedRendererEvent(event)
    await recoveryBarrier
    return handler(input, event)
  })
}

export function registerCoworkIpc(): void {
  if (registered) return
  registered = true

  recoveryBarrier = recoverInterruptedTasks()
    .then((count) => {
      if (count) log.warn(`Paused ${count} interrupted Cowork task(s) after restart.`)
    })
    .catch((error) => {
      log.error(`Could not recover Cowork tasks: ${error.message}`)
      throw error
    })
  void recoveryBarrier
    .then(() => (quitting ? undefined : scheduler.start()))
    .catch((error) => log.error(`Could not initialize Cowork safely: ${error.message}`))
  app.once('before-quit', () => {
    quitting = true
    scheduler.stop()
  })

  handle(CHANNEL.listProjects, async () => (await listProjects()).map(publicProject))

  handle(CHANNEL.createProject, async (value, event) => {
    await requireActiveCoworkSession()
    const input = record(value)
    const owner = BrowserWindow.fromWebContents(event.sender) ?? undefined
    const selection = owner
      ? await dialog.showOpenDialog(owner, {
          title: 'Connect a folder to Cowork',
          properties: ['openDirectory', 'createDirectory']
        })
      : await dialog.showOpenDialog({
          title: 'Connect a folder to Cowork',
          properties: ['openDirectory', 'createDirectory']
        })
    if (selection.canceled || !selection.filePaths[0]) return null
    const selectedRoot = await fs.realpath(selection.filePaths[0])
    const userDataRoot = await fs
      .realpath(app.getPath('userData'))
      .catch(() => app.getPath('userData'))
    if (
      selectedRoot === userDataRoot ||
      selectedRoot.startsWith(userDataRoot + path.sep) ||
      userDataRoot.startsWith(selectedRoot + path.sep)
    ) {
      throw new Error('The app data directory cannot be connected as a Cowork project.')
    }
    const mode = input.approvalMode ?? 'manual'
    if (!approvalModes.has(mode)) throw new Error('Invalid approval mode.')
    const project = await createProject({
      name: stringValue(input.name, 'Project name', 120),
      rootPath: selectedRoot,
      instructions:
        input.instructions === undefined
          ? ''
          : stringValue(input.instructions, 'Instructions', 20_000, true),
      approvalMode: mode
    })
    return publicProject(project)
  })

  handle(CHANNEL.updateProject, async (value) => {
    await requireActiveCoworkSession()
    const input = record(value)
    const id = idValue(input.id, 'Project ID')
    const patch: Parameters<typeof updateProject>[1] = {}
    if (input.name !== undefined) patch.name = stringValue(input.name, 'Project name', 120)
    if (input.instructions !== undefined)
      patch.instructions = stringValue(input.instructions, 'Instructions', 20_000, true)
    if (input.approvalMode !== undefined) {
      if (!approvalModes.has(input.approvalMode)) throw new Error('Invalid approval mode.')
      patch.approvalMode = input.approvalMode
    }
    return publicProject(await updateProject(id, patch))
  })

  handle(CHANNEL.deleteProject, async (value) => {
    await requireActiveCoworkSession()
    const id = idValue(record(value).id, 'Project ID')
    // Persist the gate first. New tasks and schedule callbacks re-check it, so
    // nothing can enter the project while existing work is being stopped.
    await deleteProject(id)
    const projectTasks = await listTasks(id)
    for (const task of projectTasks) {
      if (['queued', 'running', 'waiting_approval'].includes(task.status)) {
        await cancelCoworkRun(task.id, emit)
      }
    }
    for (const schedule of await listCoworkSchedules(id)) {
      scheduler.cancel(schedule.id)
      if (schedule.status === 'active') {
        await pauseCoworkSchedule(schedule.id)
      }
    }
    return true
  })

  handle(CHANNEL.listTasks, async (value) => {
    const input = record(value)
    const projectId =
      input.projectId === undefined ? undefined : idValue(input.projectId, 'Project ID')
    return listTaskSummaries(projectId)
  })

  handle(CHANNEL.getTask, async (value) => {
    const task = await getTask(idValue(record(value).id, 'Task ID'))
    return task ? publicTask(task) : null
  })

  handle(CHANNEL.createTask, async (value) => {
    const input = record(value)
    const task = await createTask({
      projectId: idValue(input.projectId, 'Project ID'),
      title: input.title === undefined ? '' : stringValue(input.title, 'Task title', 160, true),
      goal: stringValue(input.goal, 'Task goal', 20_000),
      model: await selectedModel(input.model)
    })
    return publicTask(task)
  })

  handle(CHANNEL.startTask, async (value) => {
    const id = idValue(record(value).id, 'Task ID')
    return publicTask(await requestCoworkStart(id, getAuthHeaders, emit, refreshModelTarget, true))
  })

  handle(CHANNEL.steerTask, async (value) => {
    const input = record(value)
    const id = idValue(input.id, 'Task ID')
    const task = await getTask(id)
    if (!task) throw new Error('Cowork task not found.')
    const project = await getProject(task.projectId)
    if (!project || project.archivedAt) throw new Error('This Cowork project is archived.')
    // steerCoworkRun persists the instruction before its shared start path
    // refreshes the model. Close that mutation window at the IPC boundary.
    await refreshModelTarget(task.model)
    return publicTask(
      await steerCoworkRun(
        id,
        stringValue(input.content, 'Instruction', 20_000),
        getAuthHeaders,
        emit,
        refreshModelTarget
      )
    )
  })

  handle(CHANNEL.cancelTask, async (value) =>
    publicTask(await cancelCoworkRun(idValue(record(value).id, 'Task ID'), emit))
  )

  handle(CHANNEL.pauseTask, async (value) =>
    publicTask(await pauseCoworkRun(idValue(record(value).id, 'Task ID'), emit))
  )

  handle(CHANNEL.resolveApproval, async (value) => {
    const input = record(value)
    if (typeof input.approved !== 'boolean') throw new Error('Approval decision is required.')
    const taskId = idValue(input.taskId, 'Task ID')
    const task = await getTask(taskId)
    if (!task) throw new Error('Cowork task not found.')
    const project = await getProject(task.projectId)
    if (!project || project.archivedAt) throw new Error('This Cowork project is archived.')
    // Denial remains available as a safe recovery action. Approval can read or
    // mutate project files, so it requires the task's exact current-wallet
    // marketplace session to still be active.
    if (input.approved) await refreshModelTarget(task.model)
    return publicTask(
      await resolveCoworkApproval(
        taskId,
        idValue(input.approvalId, 'Approval ID'),
        input.approved,
        getAuthHeaders,
        emit,
        refreshModelTarget
      )
    )
  })

  handle(CHANNEL.deleteTask, async (value) => {
    const id = idValue(record(value).id, 'Task ID')
    if (coworkRunActive(id)) await cancelCoworkRun(id, emit)
    await deleteTask(id)
    return true
  })

  handle(CHANNEL.listModelOptions, async (value) => {
    const input = record(value)
    if (typeof input.force !== 'boolean') throw new Error('Model refresh choice is required.')
    return (await modelOptions(input.force)).map(publicModelOption)
  })

  handle(CHANNEL.previewArtifact, async (value) => {
    await requireActiveCoworkSession()
    const input = record(value)
    const task = await getTask(idValue(input.taskId, 'Task ID'))
    if (!task) throw new Error('Cowork task not found.')
    const requestedPath = stringValue(input.path, 'Artifact path', 2_000)
    const artifact = task.artifacts.find((item) => item.path === requestedPath)
    if (!artifact) throw new Error('Artifact is not part of this task.')
    const project = await getProject(task.projectId)
    if (!project) throw new Error('Cowork project not found.')
    if (artifact.kind === 'folder') {
      const output = await executeCoworkTool(project, 'list_files', {
        path: artifact.path,
        maxDepth: 1
      })
      return { ...artifact, content: JSON.stringify(output.result, null, 2), truncated: false }
    }
    const extension = path.extname(artifact.path).toLowerCase()
    if (BINARY_ARTIFACT_EXTENSIONS.has(extension)) {
      const output = await executeCoworkTool(project, 'read_document', { path: artifact.path })
      const extracted = output.result as {
        format: string
        text: string
        metadata: {
          inputBytes: number
          pageCount?: number
          sheetCount?: number
          slideCount?: number
        }
        empty: boolean
        truncated: boolean
        warnings: string[]
      }
      const sectionCount =
        extracted.metadata.pageCount ??
        extracted.metadata.sheetCount ??
        extracted.metadata.slideCount
      const details = [
        `${extracted.format.toUpperCase()} text preview · ${extracted.metadata.inputBytes.toLocaleString()} bytes`,
        sectionCount === undefined ? undefined : `${sectionCount.toLocaleString()} sections`,
        ...extracted.warnings.map((warning) => `Warning: ${warning}`)
      ].filter((line): line is string => Boolean(line))
      return {
        ...artifact,
        content: `${details.join(' · ')}\n\n${
          extracted.empty ? '(No extractable text found.)' : extracted.text
        }`,
        truncated: extracted.truncated
      }
    }
    const output = await executeCoworkTool(project, 'read_file', {
      path: artifact.path,
      startLine: 1,
      endLine: 400
    })
    const result = output.result as { content: string; endLine: number; totalLines: number }
    return { ...artifact, content: result.content, truncated: result.endLine < result.totalLines }
  })

  handle(CHANNEL.revealArtifact, async (value) => {
    await requireActiveCoworkSession()
    const input = record(value)
    const task = await getTask(idValue(input.taskId, 'Task ID'))
    if (!task) throw new Error('Cowork task not found.')
    const requestedPath = stringValue(input.path, 'Artifact path', 2_000)
    if (!task.artifacts.some((item) => item.path === requestedPath))
      throw new Error('Artifact is not part of this task.')
    const project = await getProject(task.projectId)
    if (!project) throw new Error('Cowork project not found.')
    await openCoworkPath(project, requestedPath)
    return true
  })

  handle(CHANNEL.listSchedules, async (value) => {
    const input = record(value)
    const projectId =
      input.projectId === undefined ? undefined : idValue(input.projectId, 'Project ID')
    return listCoworkSchedules(projectId)
  })

  handle(CHANNEL.createSchedule, async (value) => {
    const input = record(value)
    const projectId = idValue(input.projectId, 'Project ID')
    const project = await getProject(projectId)
    if (!project || project.archivedAt) throw new Error('Cowork project not found.')
    if (input.status !== undefined && input.status !== 'active' && input.status !== 'paused') {
      throw new Error('Invalid schedule status.')
    }
    const schedule = await createCoworkSchedule({
      projectId,
      name: stringValue(input.name, 'Schedule name', 120),
      task: {
        title: input.title === undefined ? '' : stringValue(input.title, 'Task title', 160, true),
        goal: stringValue(input.goal, 'Task goal', 20_000),
        model: await selectedModel(input.model)
      },
      cadence: cadenceValue(input.cadence),
      timeZone:
        input.timeZone === undefined ? undefined : stringValue(input.timeZone, 'Time zone', 100),
      status: input.status ?? 'active'
    })
    scheduler.wake()
    return schedule
  })

  handle(CHANNEL.updateSchedule, async (value) => {
    await requireActiveCoworkSession()
    const input = record(value)
    const id = idValue(input.id, 'Schedule ID')
    const current = await getCoworkSchedule(id)
    if (!current) throw new Error('Cowork schedule not found.')
    const project = await getProject(current.projectId)
    if (!project || project.archivedAt) throw new Error('This Cowork project is archived.')
    const patch: Parameters<typeof updateCoworkSchedule>[1] = {}
    if (input.name !== undefined) patch.name = stringValue(input.name, 'Schedule name', 120)
    if (input.cadence !== undefined) patch.cadence = cadenceValue(input.cadence)
    if (input.timeZone !== undefined) patch.timeZone = stringValue(input.timeZone, 'Time zone', 100)
    if (input.task !== undefined) {
      const task = record(input.task)
      patch.task = {
        title: task.title === undefined ? '' : stringValue(task.title, 'Task title', 160, true),
        goal: stringValue(task.goal, 'Task goal', 20_000),
        model: await selectedModel(task.model)
      }
    }
    const schedule = await updateCoworkSchedule(id, patch)
    scheduler.wake()
    return schedule
  })

  handle(CHANNEL.pauseSchedule, async (value) => {
    const id = idValue(record(value).id, 'Schedule ID')
    scheduler.cancel(id)
    return pauseCoworkSchedule(id)
  })

  handle(CHANNEL.resumeSchedule, async (value) => {
    const id = idValue(record(value).id, 'Schedule ID')
    const current = await getCoworkSchedule(id)
    if (!current) throw new Error('Cowork schedule not found.')
    const project = await getProject(current.projectId)
    if (!project || project.archivedAt) throw new Error('This Cowork project is archived.')
    await refreshModelTarget(current.task.model)
    const schedule = await resumeCoworkSchedule(id)
    scheduler.wake()
    return schedule
  })

  handle(CHANNEL.deleteSchedule, async (value) => {
    const id = idValue(record(value).id, 'Schedule ID')
    scheduler.cancel(id)
    await deleteCoworkSchedule(id)
    return true
  })

  handle(CHANNEL.runScheduleNow, async (value) => {
    const id = idValue(record(value).id, 'Schedule ID')
    const current = await getCoworkSchedule(id)
    if (!current) throw new Error('Cowork schedule not found.')
    const project = await getProject(current.projectId)
    if (!project || project.archivedAt) throw new Error('This Cowork project is archived.')
    return scheduler.runNow(id)
  })

  handle(CHANNEL.listExtensions, async (value) => {
    await requireActiveCoworkSession()
    const project = await getProject(idValue(record(value).projectId, 'Project ID'))
    if (!project || project.archivedAt) throw new Error('Cowork project not found.')
    return publicExtensionCatalog(await extensionCatalog(project), project)
  })

  handle(CHANNEL.configureExtensions, async (value) => {
    await requireActiveCoworkSession()
    const input = record(value)
    const projectId = idValue(input.projectId, 'Project ID')
    let project = await getProject(projectId)
    if (!project || project.archivedAt) throw new Error('Cowork project not found.')
    if (typeof input.folderInstructionsEnabled !== 'boolean') {
      throw new Error('Folder-instruction choice is required.')
    }
    if (!Array.isArray(input.enabledSkillIds) || input.enabledSkillIds.length > 8) {
      throw new Error('Choose at most eight project skills.')
    }
    const enabledSkillIds = input.enabledSkillIds.map((id: unknown) => idValue(id, 'Skill ID'))
    if (new Set(enabledSkillIds).size !== enabledSkillIds.length) {
      throw new Error('A project skill can only be enabled once.')
    }
    const catalog = await extensionCatalog(project)
    if (input.folderInstructionsEnabled && !catalog.projectInstructions) {
      throw new Error('No project instruction file was discovered.')
    }
    const enabledSkills = enabledSkillIds.map((id) => {
      const skill = catalog.skills.find((candidate) => candidate.id === id)
      if (!skill) throw new Error(`Project skill not found: ${id}`)
      return skill
    })
    const instructionBytes =
      (input.folderInstructionsEnabled ? (catalog.projectInstructions?.bytes ?? 0) : 0) +
      enabledSkills.reduce(
        (total, skill) => total + Buffer.byteLength(skill.instructions, 'utf8'),
        0
      )
    if (instructionBytes > 128 * 1024) {
      throw new Error('Enabled project guidance exceeds the 128 KB task-context limit.')
    }
    project = await updateProject(projectId, {
      extensionSettings: {
        folderInstructionsEnabled: input.folderInstructionsEnabled,
        enabledSkillIds,
        ...(input.folderInstructionsEnabled && catalog.projectInstructions
          ? { folderInstructionsHash: catalog.projectInstructions.contentHash }
          : {}),
        skillInstructionHashes: Object.fromEntries(
          enabledSkills.map((skill) => [skill.id, skill.instructionsHash])
        )
      }
    })
    return publicExtensionCatalog(catalog, project)
  })
}
