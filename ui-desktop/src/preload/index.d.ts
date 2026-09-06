type IpcUnsubscribe = () => void

export {}

type IpcRendererBridge = {
  send: (eventName: string, payload?: unknown) => void
  on: (
    eventName: string,
    listener: (payload: any, unsubscribe: IpcUnsubscribe) => void
  ) => IpcUnsubscribe
}

type ChatStreamEvent =
  | { requestId: string; kind: 'chunk'; dataBase64: string }
  | { requestId: string; kind: 'end' }
  | { requestId: string; kind: 'error'; message: string }

type ChatStreamApi = {
  start: (
    requestId: string,
    payload: unknown
  ) => Promise<{ ok: boolean; status: number; contentType: string }>
  cancel: (requestId: string) => void
  onEvent: (requestId: string, listener: (event: ChatStreamEvent) => void) => IpcUnsubscribe
}

type IpfsDownloadProgress = {
  status: 'downloading' | 'completed'
  downloaded: number
  total: number
  percentage: number
  timeUpdated: number
}

type IpfsDownloadApi = {
  selectFolder: () => Promise<{
    canceled: boolean
    folderToken?: string
  }>
  start: (requestId: string, folderToken: string, cidHash: string) => Promise<{ accepted: true }>
  cancel: (requestId: string) => void
  onEvent: (
    requestId: string,
    listener: (
      event:
        | { requestId: string; kind: 'progress'; progress: IpfsDownloadProgress }
        | { requestId: string; kind: 'error'; message: string }
    ) => void
  ) => IpcUnsubscribe
}

type CoworkApprovalMode = 'manual' | 'auto' | 'skip'
type CoworkApprovalPolicyView = {
  schemaVersion: 1
  id: 'workspace'
  mode: CoworkApprovalMode
  revision: number
  updatedAt: number
}
type CoworkTaskStatus =
  | 'draft'
  | 'queued'
  | 'running'
  | 'waiting_approval'
  | 'paused'
  | 'completed'
  | 'failed'
  | 'cancelled'

type CoworkProjectView = {
  schemaVersion: 1
  id: string
  name: string
  folderName: string
  instructions: string
  approvalMode: CoworkApprovalMode
  extensionSettings?: {
    folderInstructionsEnabled: boolean
    enabledSkillIds: string[]
    folderInstructionsHash?: string
    skillInstructionHashes?: Record<string, string>
  }
  createdAt: number
  updatedAt: number
  archivedAt?: number
}

type CoworkModelOptionView = {
  modelId: string
  modelName: string
  isLocal: boolean
  sessionId?: string
  sessionEndsAt?: number
  source: 'local' | 'marketplace'
  dataBoundary: 'on-device' | 'configured-endpoint' | 'independent-provider'
  visionCapability: 'verified' | 'declared' | 'detected' | 'none'
  visionProbedAt?: number
}

type CoworkToolCallView = {
  id: string
  type: 'function'
  function: { name: string; arguments: string }
}

type CoworkDisplayMessageView = {
  id: string
  role: 'user' | 'assistant'
  content: string
  createdAt: number
  sequence?: number
  author?: {
    kind: 'workspace' | 'model'
    modelId?: string
    modelName?: string
    sessionId?: string
  }
}

type CoworkTaskView = {
  schemaVersion: 1
  revision: number
  id: string
  projectId: string
  title: string
  goal: string
  status: CoworkTaskStatus
  model: Pick<
    CoworkModelOptionView,
    'modelId' | 'modelName' | 'isLocal' | 'sessionId' | 'sessionEndsAt' | 'dataBoundary'
  >
  plan: Array<{
    id: string
    title: string
    status: 'pending' | 'in_progress' | 'completed'
    note?: string
  }>
  messages: CoworkDisplayMessageView[]
  hasEarlierMessages?: boolean
  activities: Array<{
    id: string
    type: 'plan' | 'tool' | 'approval' | 'system'
    label: string
    detail?: string
    status: 'running' | 'success' | 'error' | 'waiting'
    createdAt: number
  }>
  artifacts: Array<{
    path: string
    name: string
    kind: 'file' | 'folder'
    createdAt: number
    updatedAt: number
  }>
  pendingApproval?: {
    id: string
    toolCall: CoworkToolCallView
    remainingToolCalls: CoworkToolCallView[]
    reason: string
    risk: 'read' | 'write' | 'overwrite' | 'delete' | 'network'
    createdAt: number
  }
  dataAccessApproved?: boolean
  summary?: string
  error?: string
  createdAt: number
  updatedAt: number
  startedAt?: number
  completedAt?: number
}

type CoworkTaskSummaryView = Pick<
  CoworkTaskView,
  'id' | 'projectId' | 'title' | 'status' | 'createdAt' | 'updatedAt' | 'startedAt' | 'completedAt'
>

type CoworkApi = {
  getApprovalPolicy: () => Promise<CoworkApprovalPolicyView>
  updateApprovalPolicy: (
    mode: CoworkApprovalMode,
    expectedRevision: number
  ) => Promise<CoworkApprovalPolicyView>
  listProjects: () => Promise<CoworkProjectView[]>
  createProject: (input: {
    name: string
    instructions?: string
  }) => Promise<CoworkProjectView | null>
  updateProject: (input: {
    id: string
    name?: string
    instructions?: string
  }) => Promise<CoworkProjectView>
  deleteProject: (id: string) => Promise<boolean>
  listTasks: (projectId?: string) => Promise<CoworkTaskSummaryView[]>
  getTask: (id: string) => Promise<CoworkTaskView | null>
  listTaskMessages: (
    taskId: string,
    beforeSequence?: number,
    limit?: number
  ) => Promise<{
    messages: CoworkDisplayMessageView[]
    hasMore: boolean
    nextBeforeSequence?: number
  }>
  createTask: (input: {
    projectId: string
    title?: string
    goal: string
    model: CoworkModelOptionView
  }) => Promise<CoworkTaskView>
  startTask: (id: string) => Promise<CoworkTaskView>
  steerTask: (id: string, content: string) => Promise<CoworkTaskView>
  cancelTask: (id: string) => Promise<CoworkTaskView>
  pauseTask: (id: string) => Promise<CoworkTaskView>
  rebindTask: (id: string, model: CoworkModelOptionView) => Promise<CoworkTaskView>
  resolveApproval: (
    taskId: string,
    approvalId: string,
    approved: boolean
  ) => Promise<CoworkTaskView>
  deleteTask: (id: string) => Promise<boolean>
  listModelOptions: (force?: boolean) => Promise<CoworkModelOptionView[]>
  previewArtifact: (
    taskId: string,
    path: string
  ) => Promise<{
    path: string
    name: string
    kind: 'file' | 'folder'
    createdAt: number
    updatedAt: number
    content: string
    truncated: boolean
  }>
  revealArtifact: (taskId: string, path: string) => Promise<boolean>
  listSchedules: (projectId?: string) => Promise<CoworkScheduleView[]>
  createSchedule: (input: {
    projectId: string
    name: string
    title?: string
    goal: string
    model: CoworkModelOptionView
    cadence: CoworkScheduleCadenceView
    timeZone?: string
    status?: 'active' | 'paused'
  }) => Promise<CoworkScheduleView>
  updateSchedule: (input: {
    id: string
    name?: string
    task?: { title?: string; goal: string; model: CoworkModelOptionView }
    cadence?: CoworkScheduleCadenceView
    timeZone?: string
  }) => Promise<CoworkScheduleView>
  pauseSchedule: (id: string) => Promise<CoworkScheduleView>
  resumeSchedule: (id: string) => Promise<CoworkScheduleView>
  deleteSchedule: (id: string) => Promise<boolean>
  runScheduleNow: (id: string) => Promise<CoworkScheduleView>
  listExtensions: (projectId: string) => Promise<CoworkExtensionCatalogView>
  configureExtensions: (
    projectId: string,
    settings: { folderInstructionsEnabled: boolean; enabledSkillIds: string[] }
  ) => Promise<CoworkExtensionCatalogView>
  onTaskEvent: (
    listener: (event: {
      taskId: string
      projectId: string
      title: string
      status: CoworkTaskStatus
      createdAt: number
      updatedAt: number
      startedAt?: number
      completedAt?: number
    }) => void
  ) => IpcUnsubscribe
}

type CoworkScheduleCadenceView =
  | { kind: 'manual' }
  | { kind: 'hourly'; minute: number }
  | { kind: 'daily'; hour: number; minute: number }
  | { kind: 'weekly'; dayOfWeek: number; hour: number; minute: number }
  | { kind: 'weekdays'; hour: number; minute: number }

type CoworkScheduleView = {
  schemaVersion: 1
  revision: number
  id: string
  projectId: string
  name: string
  task: {
    title: string
    goal: string
    model: CoworkTaskView['model']
  }
  cadence: CoworkScheduleCadenceView
  timeZone: string
  status: 'active' | 'paused'
  createdAt: number
  updatedAt: number
  nextRunAt?: number
  lastScheduledFor?: number
  lastRunAt?: number
  lastTaskId?: string
  lastError?: string
  runningSince?: number
}

type CoworkExtensionCatalogView = {
  schemaVersion: 1
  projectInstructions?: {
    kind: 'project-instructions'
    contentHash: string
    source: string
    bytes: number
    enabled: boolean
    trust: 'untrusted-project-content'
  }
  skills: Array<{
    kind: 'skill'
    schemaVersion: 1
    id: string
    name: string
    description: string
    instructionsHash: string
    source: string
    declaredCapabilities: string[]
    activationRequested: boolean
    enabled: boolean
    executionMode: 'instructions-only'
    trust: 'untrusted-project-content'
  }>
  connectors: Array<{
    kind: 'remote-mcp-connector'
    schemaVersion: 1
    id: string
    name: string
    description: string
    transport: 'streamable-http' | 'sse'
    url: string
    authentication: 'none' | 'oauth2'
    declaredCapabilities: string[]
    activationRequested: boolean
    enabled: false
    connectionState: 'not-connected'
    requiresEndpointVerification: true
    trust: 'unverified-user-configuration'
    source: string
  }>
  issues: Array<{
    severity: 'warning' | 'error'
    code: string
    source: string
    message: string
  }>
  executionAvailable: false
  networkAccessPerformed: false
}

declare global {
  interface Window {
    ipcRenderer: IpcRendererBridge
    openLink: (url: string) => Promise<void>
    getAppVersion: () => string
    copyToClipboard: (text: string) => Promise<void>
    cowork: CoworkApi
    chatStream: ChatStreamApi
    ipfsDownload: IpfsDownloadApi
  }
}
