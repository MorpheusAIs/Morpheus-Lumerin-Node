export type CoworkApprovalMode = 'manual' | 'auto' | 'skip'

export type CoworkTaskStatus =
  | 'draft'
  | 'queued'
  | 'running'
  | 'waiting_approval'
  | 'paused'
  | 'completed'
  | 'failed'
  | 'cancelled'

export type CoworkPlanStepStatus = 'pending' | 'in_progress' | 'completed'

export interface CoworkProject {
  schemaVersion: 1
  id: string
  name: string
  rootPath: string
  instructions: string
  approvalMode: CoworkApprovalMode
  extensionSettings?: {
    folderInstructionsEnabled: boolean
    enabledSkillIds: string[]
    /** Hashes bind activation to the exact untrusted guidance the user reviewed. */
    folderInstructionsHash?: string
    skillInstructionHashes?: Record<string, string>
  }
  createdAt: number
  updatedAt: number
  archivedAt?: number
}

export interface CoworkModelTarget {
  modelId: string
  modelName: string
  isLocal: boolean
  sessionId?: string
  sessionEndsAt?: number
  dataBoundary?: 'on-device' | 'configured-endpoint' | 'independent-provider'
}

export interface CoworkModelOption extends CoworkModelTarget {
  source: 'local' | 'marketplace'
  dataBoundary: 'on-device' | 'configured-endpoint' | 'independent-provider'
  visionCapability: 'declared' | 'detected' | 'none'
}

export interface CoworkModelBinding extends CoworkModelTarget {
  boundAt: number
  unboundAt?: number
}

export interface CoworkPlanStep {
  id: string
  title: string
  status: CoworkPlanStepStatus
  note?: string
}

export interface CoworkArtifact {
  path: string
  name: string
  kind: 'file' | 'folder'
  createdAt: number
  updatedAt: number
}

export interface CoworkDisplayMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  createdAt: number
  /** Stable per-task ordering used by the durable, paginated transcript. */
  sequence?: number
  author?: {
    kind: 'workspace' | 'model'
    modelId?: string
    modelName?: string
    sessionId?: string
  }
}

export interface CoworkHandoffContext {
  createdAt: number
  previousModelName: string
  previousSessionId?: string
  goal: string
  status: CoworkTaskStatus
  summary?: string
  plan: CoworkPlanStep[]
  artifacts: Array<Pick<CoworkArtifact, 'path' | 'name' | 'kind' | 'updatedAt'>>
  recentMessages: Array<Pick<CoworkDisplayMessage, 'role' | 'content' | 'createdAt'>>
}

export interface CoworkRunSafety {
  /** A run epoch starts with a new user instruction and survives approvals/resumes. */
  id: string
  startedAt: number
  modelSteps: number
  mutations: number
  loopGuard: {
    version: 1
    unverifiedMutations: number
    lastMutationHash?: string
    equivalentMutationStreak: number
    lastDestinationHash?: string
    destinationMutationStreak: number
    recentMutationHashes: string[]
    contentDestinationsByHash: Record<string, string[]>
  }
}

export interface CoworkActivity {
  id: string
  type: 'plan' | 'tool' | 'approval' | 'system'
  label: string
  detail?: string
  status: 'running' | 'success' | 'error' | 'waiting'
  createdAt: number
}

export interface CoworkToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string
  }
}

export interface CoworkPendingApproval {
  id: string
  toolCall: CoworkToolCall
  remainingToolCalls: CoworkToolCall[]
  reason: string
  risk: 'read' | 'write' | 'overwrite' | 'delete' | 'network'
  createdAt: number
}

export interface CoworkAgentMessage {
  role: 'user' | 'assistant' | 'tool'
  content?: string | null
  tool_calls?: CoworkToolCall[]
  tool_call_id?: string
}

export interface CoworkToolExecution {
  toolCallId: string
  toolName: string
  status: 'prepared' | 'succeeded' | 'failed' | 'ambiguous'
  argumentsHash: string
  /** User-instruction safety epoch; absent only on records created before this field existed. */
  instructionId?: string
  resultMessage?: string
  preparedAt: number
  completedAt?: number
}

export interface CoworkTask {
  schemaVersion: 1
  revision: number
  id: string
  projectId: string
  title: string
  goal: string
  status: CoworkTaskStatus
  model: CoworkModelTarget
  /** Provenance for every compute binding used by this durable task. */
  modelBindings?: CoworkModelBinding[]
  /** First model-protocol message that may be sent to the current binding. */
  modelContextStart?: number
  /** Bounded, explicitly labelled context for a replacement model/session. */
  handoff?: CoworkHandoffContext
  plan: CoworkPlanStep[]
  messages: CoworkDisplayMessage[]
  /** Highest display-message sequence copied into the durable transcript store. */
  messagesPersistedThrough?: number
  /** The task payload contains a recent window; older turns remain pageable. */
  hasEarlierMessages?: boolean
  agentMessages: CoworkAgentMessage[]
  activities: CoworkActivity[]
  artifacts: CoworkArtifact[]
  pendingApproval?: CoworkPendingApproval
  /** Internal identity of the endpoint/session resolved immediately before a run. */
  modelFingerprint?: string
  /** Session-scoped fallback for providers that reject native OpenAI tool fields. */
  toolProtocol?: 'text-v1'
  /** Durable mutation journal used to suppress duplicate or ambiguous file actions. */
  toolExecutions?: CoworkToolExecution[]
  dataAccessApproved?: boolean
  /** Prevents an approval from silently carrying across an endpoint change. */
  dataAccessApprovedFingerprint?: string
  summary?: string
  error?: string
  pauseReason?: 'repetition_guard'
  runSafety?: CoworkRunSafety
  createdAt: number
  updatedAt: number
  startedAt?: number
  completedAt?: number
}

/** Minimal rail payload; full transcripts and tool history stay in the main process. */
export type CoworkTaskSummary = Pick<
  CoworkTask,
  'id' | 'projectId' | 'title' | 'status' | 'createdAt' | 'updatedAt' | 'startedAt' | 'completedAt'
>

export interface CoworkTaskEvent {
  taskId: string
  projectId: string
  title: string
  status: CoworkTaskStatus
  createdAt: number
  updatedAt: number
  startedAt?: number
  completedAt?: number
}
