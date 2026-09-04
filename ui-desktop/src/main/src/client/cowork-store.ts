import { randomUUID } from 'node:crypto'
import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import {
  withCoworkApprovalPolicyReadLock,
  withCoworkApprovalPolicyWriteLock
} from './cowork-approval-policy-lock'
import { coworkCollection } from './cowork-database'
import { executionMatchesToolCall } from './cowork-mutation-journal'
import { createCoworkLoopGuardState } from './cowork-loop-guard'
import {
  deleteCoworkMessages,
  listCoworkMessages,
  persistCoworkMessages
} from './cowork-message-store'
import {
  CoworkActivity,
  CoworkAgentMessage,
  CoworkApprovalMode,
  CoworkApprovalPolicy,
  CoworkArtifact,
  CoworkDisplayMessage,
  CoworkModelTarget,
  CoworkPlanStep,
  CoworkProject,
  CoworkTask,
  CoworkTaskSummary,
  CoworkTaskStatus,
  CoworkToolCall,
  CoworkToolExecution
} from './cowork.types'

const projects = () => coworkCollection('projects')
const tasks = () => coworkCollection('tasks')
const preferences = () => coworkCollection('preferences')
const APPROVAL_POLICY_ID = 'workspace'
const MAX_TASKS_PER_PROJECT = 500
const MAX_DISPLAY_HISTORY_BYTES = 2 * 1024 * 1024
const MAX_RECENT_DISPLAY_MESSAGES = 100
const MAX_AGENT_HISTORY_BYTES = 12 * 1024 * 1024
const LARGE_ARGUMENT_TOOLS = new Set([
  'write_file',
  'create_docx',
  'create_xlsx',
  'create_pptx',
  'create_pdf'
])

const clean = <T>(value: T): T => JSON.parse(JSON.stringify(value))
const historyByteSizes = new WeakMap<unknown[], number>()
const taskMutationTails = new Map<string, Promise<void>>()
let approvalPolicyMutationTail: Promise<void> = Promise.resolve()
let cachedApprovalPolicy: CoworkApprovalPolicy | null = null

const isApprovalMode = (value: unknown): value is CoworkApprovalMode =>
  value === 'manual' || value === 'auto' || value === 'skip'

const approvalStrictness = (mode: CoworkApprovalMode): number =>
  mode === 'manual' ? 2 : mode === 'auto' ? 1 : 0

const syncLegacyProjectApprovalMode = async (mode: CoworkApprovalMode): Promise<void> => {
  await projects().updateAsync(
    { approvalMode: { $ne: mode } },
    { $set: { approvalMode: mode } },
    { multi: true }
  )
}

const normalizedApprovalPolicy = (value: unknown): CoworkApprovalPolicy | null => {
  const stored = value as Partial<CoworkApprovalPolicy> | null
  if (
    stored?.schemaVersion !== 1 ||
    stored.id !== APPROVAL_POLICY_ID ||
    !isApprovalMode(stored.mode) ||
    !Number.isSafeInteger(stored.revision) ||
    stored.revision! < 1 ||
    !Number.isSafeInteger(stored.updatedAt) ||
    stored.updatedAt! < 0
  ) {
    return null
  }
  return {
    schemaVersion: 1,
    id: APPROVAL_POLICY_ID,
    mode: stored.mode,
    revision: stored.revision!,
    updatedAt: stored.updatedAt!
  }
}

async function withApprovalPolicyMutationLock<T>(operation: () => Promise<T>): Promise<T> {
  const previous = approvalPolicyMutationTail
  let release!: () => void
  const gate = new Promise<void>((resolve) => {
    release = resolve
  })
  const tail = previous.catch(() => undefined).then(() => gate)
  approvalPolicyMutationTail = tail
  await previous.catch(() => undefined)
  try {
    return await operation()
  } finally {
    release()
    if (approvalPolicyMutationTail === tail) approvalPolicyMutationTail = Promise.resolve()
  }
}

async function readApprovalPolicyUnlocked(): Promise<CoworkApprovalPolicy> {
  if (cachedApprovalPolicy) return clean(cachedApprovalPolicy)
  const storedRows = (await preferences().findAsync({ id: APPROVAL_POLICY_ID })) as unknown[]
  const futurePolicy = storedRows.find((value) => {
    const schemaVersion = (value as { schemaVersion?: unknown } | null)?.schemaVersion
    return Number.isSafeInteger(schemaVersion) && Number(schemaVersion) > 1
  })
  if (futurePolicy) {
    throw new Error(
      'Workspace approval settings were created by a newer app version. Update Morpheus before running Workspace tasks.'
    )
  }
  const highestObservedRevision = storedRows.reduce<number>((highest, value) => {
    const revision = (value as { revision?: unknown } | null)?.revision
    return Number.isSafeInteger(revision) && Number(revision) >= 1
      ? Math.max(highest, Number(revision))
      : highest
  }, 0)
  const validPolicies = storedRows
    .map(normalizedApprovalPolicy)
    .filter((policy): policy is CoworkApprovalPolicy => Boolean(policy))
    .sort(
      (left, right) =>
        right.revision - left.revision ||
        approvalStrictness(right.mode) - approvalStrictness(left.mode) ||
        right.updatedAt - left.updatedAt
    )
  const normalized = validPolicies[0]
  const malformedAtOrAboveWinner = Boolean(
    normalized &&
    storedRows.some((value) => {
      if (normalizedApprovalPolicy(value)) return false
      const revision = (value as { revision?: unknown } | null)?.revision
      return Number.isSafeInteger(revision) && Number(revision) >= normalized.revision
    })
  )
  if (normalized && !malformedAtOrAboveWinner) {
    if (storedRows.length !== 1) {
      await preferences().removeAsync({ id: APPROVAL_POLICY_ID }, { multi: true })
      await preferences().insertAsync(normalized)
    }
    await syncLegacyProjectApprovalMode(normalized.mode)
    cachedApprovalPolicy = normalized
    return clean(normalized)
  }

  // Project-scoped modes from older builds cannot safely be promoted across
  // unrelated folders. Migrate to the least-privileged global default and let
  // the user make one explicit Workspace-wide choice.
  if (highestObservedRevision >= Number.MAX_SAFE_INTEGER) {
    throw new Error(
      'Workspace approval settings are invalid and cannot be repaired safely. Update Morpheus or restore the Workspace preferences file.'
    )
  }
  const fallback: CoworkApprovalPolicy = {
    schemaVersion: 1,
    id: APPROVAL_POLICY_ID,
    mode: 'manual',
    revision: highestObservedRevision + 1,
    updatedAt: Date.now()
  }
  // Tighten legacy project rows before publishing the new canonical default so
  // a rollback cannot briefly reactivate an old per-project Skip value.
  await syncLegacyProjectApprovalMode(fallback.mode)
  await preferences().removeAsync({ id: APPROVAL_POLICY_ID }, { multi: true })
  await preferences().insertAsync(fallback)
  cachedApprovalPolicy = fallback
  return clean(fallback)
}

export const getCoworkApprovalPolicy = (): Promise<CoworkApprovalPolicy> =>
  withApprovalPolicyMutationLock(readApprovalPolicyUnlocked)

export const updateCoworkApprovalPolicy = (
  mode: CoworkApprovalMode,
  expectedRevision: number
): Promise<CoworkApprovalPolicy> =>
  withCoworkApprovalPolicyWriteLock(() =>
    withApprovalPolicyMutationLock(async () => {
      if (!isApprovalMode(mode)) throw new Error('Invalid approval mode.')
      if (!Number.isSafeInteger(expectedRevision) || expectedRevision < 1) {
        throw new Error('Invalid approval policy revision.')
      }
      const current = await readApprovalPolicyUnlocked()
      if (current.revision !== expectedRevision) {
        throw new Error(
          'The Workspace approval policy changed. Review the latest setting and retry.'
        )
      }
      if (current.revision >= Number.MAX_SAFE_INTEGER) {
        throw new Error(
          'The Workspace approval policy revision limit was reached. Update Morpheus before changing this setting.'
        )
      }
      const next: CoworkApprovalPolicy = {
        schemaVersion: 1,
        id: APPROVAL_POLICY_ID,
        mode,
        revision: current.revision + 1,
        updatedAt: Date.now()
      }
      const tightening = approvalStrictness(mode) > approvalStrictness(current.mode)
      if (tightening) await syncLegacyProjectApprovalMode(mode)
      const replaced = await preferences().updateAsync(
        { id: APPROVAL_POLICY_ID, revision: current.revision },
        { $set: next },
        {}
      )
      if (replaced !== 1) {
        cachedApprovalPolicy = null
        throw new Error(
          'The Workspace approval policy changed. Review the latest setting and retry.'
        )
      }
      cachedApprovalPolicy = next
      if (!tightening) await syncLegacyProjectApprovalMode(mode)
      return clean(next)
    })
  )

async function withTaskMutationLock<T>(taskId: string, operation: () => Promise<T>): Promise<T> {
  const previous = taskMutationTails.get(taskId) ?? Promise.resolve()
  let release!: () => void
  const gate = new Promise<void>((resolve) => {
    release = resolve
  })
  const tail = previous.catch(() => undefined).then(() => gate)
  taskMutationTails.set(taskId, tail)
  await previous.catch(() => undefined)
  try {
    return await operation()
  } finally {
    release()
    if (taskMutationTails.get(taskId) === tail) taskMutationTails.delete(taskId)
  }
}

async function persistTaskDisplayMessages(task: CoworkTask): Promise<void> {
  const pending =
    task.messagesPersistedThrough === undefined
      ? task.messages
      : task.messages.filter(
          (message) =>
            message.sequence === undefined || message.sequence > task.messagesPersistedThrough!
        )
  if (pending.length) {
    const persisted = await persistCoworkMessages(task.id, pending)
    const sequenceById = new Map(persisted.map((message) => [message.id, message.sequence]))
    for (const message of task.messages) {
      const sequence = sequenceById.get(message.id)
      if (sequence !== undefined) message.sequence = sequence
    }
  }
  task.messagesPersistedThrough = task.messages.reduce(
    (highest, message) => Math.max(highest, message.sequence ?? 0),
    task.messagesPersistedThrough ?? 0
  )
}

function hasUnpersistedDisplayMessages(task: CoworkTask): boolean {
  if (task.messagesPersistedThrough === undefined) return task.messages.length > 0
  return task.messages.some(
    (message) => message.sequence === undefined || message.sequence > task.messagesPersistedThrough!
  )
}

function trimRecentDisplayMessages(task: CoworkTask): void {
  while (
    task.messages.length > 1 &&
    (task.messages.length > MAX_RECENT_DISPLAY_MESSAGES ||
      historyBytes(task.messages) > MAX_DISPLAY_HISTORY_BYTES)
  ) {
    removeHistoryPrefix(task.messages, 1)
  }
  task.hasEarlierMessages = Boolean((task.messages[0]?.sequence ?? 1) > 1)
}

function historyBytes(items: unknown[]): number {
  const cached = historyByteSizes.get(items)
  if (cached !== undefined) return cached
  const measured = Buffer.byteLength(JSON.stringify(items), 'utf8')
  historyByteSizes.set(items, measured)
  return measured
}

function appendHistoryItem<T>(items: T[], item: T): void {
  const previousLength = items.length
  const nextBytes =
    historyBytes(items) +
    Buffer.byteLength(JSON.stringify(item), 'utf8') +
    (previousLength === 0 ? 0 : 1)
  items.push(item)
  historyByteSizes.set(items, nextBytes)
}

function removeHistoryPrefix(items: unknown[], count: number): void {
  if (count <= 0) return
  if (count >= items.length) {
    items.splice(0, items.length)
    historyByteSizes.set(items, 2)
    return
  }
  const removedBytes = items
    .slice(0, count)
    .reduce<number>((total, item) => total + Buffer.byteLength(JSON.stringify(item), 'utf8'), 0)
  const nextBytes = historyBytes(items) - removedBytes - count
  items.splice(0, count)
  historyByteSizes.set(items, Math.max(2, nextBytes))
}

export const createProject = async (input: {
  name: string
  rootPath: string
  instructions?: string
  /** @deprecated File approval policy is Workspace-wide; this value is ignored. */
  approvalMode?: CoworkProject['approvalMode']
}): Promise<CoworkProject> => {
  const now = Date.now()
  const rootPath = await fs.realpath(input.rootPath)
  if (!(await fs.stat(rootPath)).isDirectory()) throw new Error('Choose a folder for this project.')
  const homePath = await fs.realpath(os.homedir()).catch(() => os.homedir())
  if (rootPath === path.parse(rootPath).root || rootPath === homePath) {
    throw new Error(
      'Choose a dedicated project folder, not an entire home drive or filesystem root.'
    )
  }
  if (!input.name.trim()) throw new Error('Project name is required.')
  return withCoworkApprovalPolicyReadLock(async () => {
    const approvalPolicy = await getCoworkApprovalPolicy()
    const project: CoworkProject = {
      schemaVersion: 1,
      id: randomUUID(),
      name: input.name.trim(),
      rootPath,
      instructions: input.instructions?.trim() ?? '',
      approvalMode: approvalPolicy.mode,
      createdAt: now,
      updatedAt: now
    }
    await projects().insertAsync(project)
    return clean(project)
  })
}

export const listProjects = async (): Promise<CoworkProject[]> => {
  const [result, approvalPolicy] = await Promise.all([
    projects().findAsync({}) as Promise<CoworkProject[]>,
    getCoworkApprovalPolicy()
  ])
  return clean(
    result
      .filter((project) => !project.archivedAt)
      .map((project) => ({ ...project, approvalMode: approvalPolicy.mode }))
      .sort((a, b) => b.updatedAt - a.updatedAt)
  )
}

export const getProject = async (id: string): Promise<CoworkProject | null> => {
  const [result, approvalPolicy] = await Promise.all([
    projects().findOneAsync({ id }) as Promise<CoworkProject | null>,
    getCoworkApprovalPolicy()
  ])
  return result ? clean({ ...result, approvalMode: approvalPolicy.mode }) : null
}

export const updateProject = async (
  id: string,
  patch: Partial<Pick<CoworkProject, 'name' | 'instructions' | 'extensionSettings'>>
): Promise<CoworkProject> => {
  const current = await getProject(id)
  if (!current) throw new Error('Workspace project not found.')
  if (patch.name !== undefined && !patch.name.trim()) throw new Error('Project name is required.')
  if (patch.extensionSettings !== undefined) {
    const settings = patch.extensionSettings
    if (typeof settings.folderInstructionsEnabled !== 'boolean') {
      throw new Error('Invalid folder-instruction setting.')
    }
    if (
      !Array.isArray(settings.enabledSkillIds) ||
      settings.enabledSkillIds.length > 8 ||
      new Set(settings.enabledSkillIds).size !== settings.enabledSkillIds.length ||
      settings.enabledSkillIds.some((id) => !/^[a-z0-9](?:[a-z0-9._-]{0,62}[a-z0-9])?$/.test(id))
    ) {
      throw new Error('Invalid enabled skill list.')
    }
    const hashPattern = /^[a-f0-9]{64}$/
    if (
      settings.folderInstructionsHash !== undefined &&
      !hashPattern.test(settings.folderInstructionsHash)
    ) {
      throw new Error('Invalid folder-instruction content hash.')
    }
    if (settings.skillInstructionHashes !== undefined) {
      if (
        !settings.skillInstructionHashes ||
        typeof settings.skillInstructionHashes !== 'object' ||
        Array.isArray(settings.skillInstructionHashes) ||
        Object.entries(settings.skillInstructionHashes).some(
          ([id, hash]) => !settings.enabledSkillIds.includes(id) || !hashPattern.test(hash)
        )
      ) {
        throw new Error('Invalid project-skill content hashes.')
      }
    }
  }
  const update = {
    ...(patch.name !== undefined ? { name: patch.name.trim() } : {}),
    ...(patch.instructions !== undefined ? { instructions: patch.instructions.trim() } : {}),
    ...(patch.extensionSettings !== undefined
      ? { extensionSettings: clean(patch.extensionSettings) }
      : {}),
    updatedAt: Date.now()
  }
  await projects().updateAsync({ id }, { $set: update }, {})
  return { ...current, ...update }
}

export const deleteProject = async (id: string): Promise<void> => {
  await projects().updateAsync(
    { id },
    { $set: { archivedAt: Date.now(), updatedAt: Date.now() } },
    {}
  )
}

export const createTask = async (input: {
  projectId: string
  title: string
  goal: string
  model: CoworkModelTarget
}): Promise<CoworkTask> => {
  if (!input.goal.trim()) throw new Error('Describe the outcome you want.')
  const project = await getProject(input.projectId)
  if (!project || project.archivedAt) throw new Error('Workspace project not found.')
  if ((await tasks().countAsync({ projectId: input.projectId })) >= MAX_TASKS_PER_PROJECT) {
    throw new Error(
      `This project has reached the ${MAX_TASKS_PER_PROJECT}-task history limit. Delete old tasks before creating another.`
    )
  }
  const now = Date.now()
  const userMessage: CoworkDisplayMessage = {
    id: randomUUID(),
    role: 'user',
    content: input.goal.trim(),
    createdAt: now,
    sequence: 1
  }
  const task: CoworkTask = {
    schemaVersion: 1,
    revision: 1,
    id: randomUUID(),
    projectId: input.projectId,
    title: input.title.trim() || input.goal.trim().slice(0, 80),
    goal: input.goal.trim(),
    status: 'queued',
    model: clean(input.model),
    modelBindings: [{ ...clean(input.model), boundAt: now }],
    modelContextStart: 0,
    plan: [],
    messages: [userMessage],
    agentMessages: [{ role: 'user', content: input.goal.trim() }],
    activities: [],
    artifacts: [],
    runSafety: {
      id: randomUUID(),
      startedAt: now,
      modelSteps: 0,
      mutations: 0,
      loopGuard: createCoworkLoopGuardState()
    },
    createdAt: now,
    updatedAt: now
  }
  await tasks().insertAsync(task)
  try {
    await persistTaskDisplayMessages(task)
    await tasks().updateAsync(
      { id: task.id },
      {
        $set: {
          messages: clean(task.messages),
          messagesPersistedThrough: task.messagesPersistedThrough,
          hasEarlierMessages: false
        }
      },
      {}
    )
  } catch (error) {
    await deleteTask(task.id)
    throw error
  }
  // Project archival and task creation can arrive on separate IPC calls. A
  // second check closes the window where creation read an active project just
  // before archival was persisted.
  const stillActive = await getProject(input.projectId)
  if (!stillActive || stillActive.archivedAt) {
    await deleteTask(task.id)
    throw new Error('This Workspace project is archived.')
  }
  await projects().updateAsync({ id: input.projectId }, { $set: { updatedAt: now } }, {})
  return clean(task)
}

export const listTasks = async (projectId?: string): Promise<CoworkTask[]> => {
  const query = projectId ? { projectId } : {}
  const result = (await tasks().findAsync(query)) as CoworkTask[]
  return clean(result.sort((a, b) => b.updatedAt - a.updatedAt))
}

export const listTaskMessages = (
  taskId: string,
  options: { beforeSequence?: number; limit?: number } = {}
) => listCoworkMessages(taskId, options)

/** Lightweight rail/list query; full model history is loaded only for an opened task. */
export const listTaskSummaries = async (projectId?: string): Promise<CoworkTaskSummary[]> => {
  const query = projectId ? { projectId } : {}
  const result = (await tasks().findAsync(query, {
    id: 1,
    projectId: 1,
    title: 1,
    status: 1,
    createdAt: 1,
    updatedAt: 1,
    startedAt: 1,
    completedAt: 1,
    _id: 0
  })) as CoworkTaskSummary[]
  return clean(result.sort((a, b) => b.updatedAt - a.updatedAt))
}

export const listRecentCompletedTaskMemories = async (
  projectId: string,
  excludeTaskId: string,
  limit = 5
): Promise<Array<Pick<CoworkTask, 'id' | 'title' | 'summary' | 'updatedAt'>>> => {
  const boundedLimit = Math.min(Math.max(Math.floor(limit), 1), 20)
  const taskCollection = tasks()
  await taskCollection.waitForIndexesAsync()
  const result = await new Promise<
    Array<Pick<CoworkTask, 'id' | 'title' | 'summary' | 'updatedAt'>>
  >((resolve, reject) => {
    taskCollection
      .find(
        {
          projectId,
          id: { $ne: excludeTaskId },
          status: 'completed',
          summary: { $exists: true }
        },
        { id: 1, title: 1, summary: 1, updatedAt: 1, _id: 0 }
      )
      .sort({ updatedAt: -1 })
      .limit(boundedLimit)
      .exec((error: Error | null, documents: any[]) => {
        if (error) reject(error)
        else resolve(documents)
      })
  })
  return clean(result)
}

async function getTaskUnlocked(id: string): Promise<CoworkTask | null> {
  const result = (await tasks().findOneAsync({ id })) as CoworkTask | null
  if (!result) return null
  const task = clean(result)
  if (hasUnpersistedDisplayMessages(task)) {
    await persistTaskDisplayMessages(task)
    trimRecentDisplayMessages(task)
    const updated = await tasks().updateAsync(
      { id, revision: task.revision },
      {
        $set: {
          messages: clean(task.messages),
          messagesPersistedThrough: task.messagesPersistedThrough,
          hasEarlierMessages: task.hasEarlierMessages
        }
      },
      {}
    )
    if (updated !== 1) return getTaskUnlocked(id)
  } else {
    trimRecentDisplayMessages(task)
  }
  return clean(task)
}

export const getTask = (id: string): Promise<CoworkTask | null> =>
  withTaskMutationLock(id, () => getTaskUnlocked(id))

export const replaceTask = (task: CoworkTask): Promise<CoworkTask> =>
  withTaskMutationLock(task.id, async () => {
    const expectedRevision = task.revision ?? 0
    const next = clean({ ...task, updatedAt: Date.now(), revision: expectedRevision + 1 })

    // Commit the authoritative task revision before copying its display messages
    // into the append-only transcript. A stale writer that loses this CAS can no
    // longer leave renderer-visible "ghost" messages behind.
    const replaced = await tasks().updateAsync(
      { id: task.id, revision: expectedRevision },
      next,
      {}
    )
    if (replaced !== 1) {
      throw new Error('This Workspace task changed in another operation. Refresh it and try again.')
    }

    // Advance the live object immediately. If transcript persistence is
    // interrupted, getTask sees messages beyond messagesPersistedThrough in the
    // accepted task revision and idempotently finishes the copy on the next read.
    task.updatedAt = next.updatedAt
    task.revision = next.revision
    await persistTaskDisplayMessages(task)
    trimRecentDisplayMessages(task)

    const transcriptState = {
      messages: clean(task.messages),
      messagesPersistedThrough: task.messagesPersistedThrough,
      hasEarlierMessages: task.hasEarlierMessages
    }
    // This projection-only finalization does not advance the revision. If
    // another process already advanced it, that accepted task remains the
    // authority and can finish its own idempotent transcript copy.
    await tasks().updateAsync(
      { id: task.id, revision: next.revision },
      { $set: transcriptState },
      {}
    )

    return clean({ ...next, ...transcriptState })
  })

export const setTaskStatus = async (
  id: string,
  status: CoworkTaskStatus,
  extra: Partial<CoworkTask> = {}
): Promise<CoworkTask> => {
  const task = await getTask(id)
  if (!task) throw new Error('Workspace task not found.')
  return replaceTask({ ...task, ...extra, status })
}

export const appendDisplayMessage = (
  task: CoworkTask,
  role: CoworkDisplayMessage['role'],
  content: string,
  author?: CoworkDisplayMessage['author']
): void => {
  const sequence =
    task.messages.reduce(
      (highest, message, index) => Math.max(highest, message.sequence ?? index + 1),
      0
    ) + 1
  appendHistoryItem(task.messages, {
    id: randomUUID(),
    role,
    content: content.slice(0, 200_000),
    createdAt: Date.now(),
    sequence,
    ...(author ? { author: clean(author) } : {})
  })
}

export const appendAgentMessage = (task: CoworkTask, message: CoworkAgentMessage): void => {
  appendHistoryItem(task.agentMessages, clean(message))
  while (
    task.agentMessages.length > 1 &&
    (task.agentMessages.length > 500 || historyBytes(task.agentMessages) > MAX_AGENT_HISTORY_BYTES)
  ) {
    // Trim only at a user-message boundary so an assistant tool call is never
    // separated from its tool results.
    const boundary = task.agentMessages.findIndex(
      (item, index) => index > 0 && item.role === 'user'
    )
    if (boundary <= 0) {
      throw new Error(
        'This task reached its 12 MB model-history limit. Start a new task to continue safely.'
      )
    }
    if (task.modelContextStart !== undefined) {
      // modelContextStart is an index into this exact array. Keep the current
      // binding boundary attached to the same message when an old prefix is
      // discarded, especially when the handoff user message itself triggers
      // the bounded-history trim.
      task.modelContextStart = Math.max(0, task.modelContextStart - boundary)
    }
    removeHistoryPrefix(task.agentMessages, boundary)
  }
}

/** Removes large generated content once a call no longer needs to execute. */
export function scrubCoworkToolArguments(task: CoworkTask, toolCall: CoworkToolCall): void {
  if (!LARGE_ARGUMENT_TOOLS.has(toolCall.function.name)) return
  let changed = false
  for (const message of task.agentMessages) {
    const stored = message.tool_calls?.find((call) => call.id === toolCall.id)
    if (!stored) continue
    try {
      const parsed = JSON.parse(stored.function.arguments || '{}')
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error()
      if (toolCall.function.name === 'write_file') {
        const length = typeof parsed.content === 'string' ? parsed.content.length : 0
        stored.function.arguments = JSON.stringify({
          ...parsed,
          content: `[omitted after execution: ${length} characters]`
        })
      } else {
        stored.function.arguments = JSON.stringify({
          path: parsed.path,
          title: parsed.title,
          content: `[omitted after execution: ${Buffer.byteLength(stored.function.arguments, 'utf8')} bytes]`
        })
      }
      changed = true
    } catch {
      stored.function.arguments = JSON.stringify({ content: '[omitted after execution]' })
      changed = true
    }
  }
  if (changed) historyByteSizes.delete(task.agentMessages)
}

/**
 * Converts mutations whose durable outcomes are unknown to ambiguous,
 * non-replayable results and closes every unresolved assistant tool call.
 */
export function reconcileInterruptedToolCalls(task: CoworkTask): {
  ambiguousMutations: number
  unresolvedCalls: number
} {
  let ambiguousMutations = 0
  for (const execution of task.toolExecutions ?? []) {
    if (execution.status !== 'prepared') continue
    ambiguousMutations++
    execution.status = 'ambiguous'
    execution.completedAt = Date.now()
    execution.resultMessage = JSON.stringify({
      ok: false,
      error:
        `A previous ${execution.toolName.replaceAll('_', ' ')} action was interrupted after it was prepared. ` +
        'Workspace did not repeat it because its outcome may be ambiguous. Inspect the connected project before issuing a new instruction.'
    })
  }

  const unresolvedCalls = task.agentMessages.reduce<CoworkToolCall[]>((pending, message) => {
    if (message.role === 'assistant') {
      pending.push(...(message.tool_calls ?? []))
    } else if (message.role === 'tool') {
      for (let index = pending.length - 1; index >= 0; index--) {
        if (pending[index].id !== message.tool_call_id) continue
        pending.splice(index, 1)
        break
      }
    }
    return pending
  }, [])

  for (const call of unresolvedCalls) {
    let execution: CoworkToolExecution | undefined
    for (let index = (task.toolExecutions?.length ?? 0) - 1; index >= 0; index--) {
      const candidate = task.toolExecutions![index]
      if (candidate.toolCallId !== call.id) continue
      execution = candidate
      break
    }
    const identityMatches = execution ? executionMatchesToolCall(execution, call) : false
    appendAgentMessage(task, {
      role: 'tool',
      tool_call_id: call.id,
      content:
        execution && !identityMatches
          ? JSON.stringify({
              ok: false,
              error:
                'Workspace blocked a reused tool call ID whose action name or arguments had changed. No file action ran.'
            })
          : (execution?.resultMessage ??
            JSON.stringify({
              ok: false,
              error:
                `The app stopped before the ${call.function.name.replaceAll('_', ' ')} action returned a durable result. ` +
                'Workspace did not replay it automatically.'
            }))
    })
    scrubCoworkToolArguments(task, call)
  }

  return { ambiguousMutations, unresolvedCalls: unresolvedCalls.length }
}

export const appendActivity = (
  task: CoworkTask,
  activity: Omit<CoworkActivity, 'id' | 'createdAt'>
): CoworkActivity => {
  const value: CoworkActivity = { ...activity, id: randomUUID(), createdAt: Date.now() }
  task.activities.push(value)
  if (task.activities.length > 500) task.activities.splice(0, task.activities.length - 500)
  return value
}

export const upsertArtifact = (task: CoworkTask, artifact: CoworkArtifact): void => {
  const index = task.artifacts.findIndex((item) => item.path === artifact.path)
  if (index >= 0) task.artifacts[index] = artifact
  else task.artifacts.push(artifact)
  if (task.artifacts.length > 500) task.artifacts.splice(0, task.artifacts.length - 500)
}

export const setPlan = (task: CoworkTask, plan: CoworkPlanStep[]): void => {
  task.plan = clean(plan)
}

export const deleteTask = (id: string): Promise<void> =>
  withTaskMutationLock(id, async () => {
    // Delete sensitive transcript rows first. If either datastore operation
    // fails, a retryable task record is preferable to unreachable orphaned data.
    await deleteCoworkMessages(id)
    await tasks().removeAsync({ id }, {})
  })

/**
 * A desktop process can exit while a model request or file operation is in
 * flight. There is no safe way to pretend that in-memory execution survived a
 * restart. Prepared mutations are therefore marked ambiguous and closed with
 * a tool result so a later resume cannot silently replay them. Any inconsistent
 * pending approval on an interrupted task is discarded.
 */
export const recoverInterruptedTasks = async (): Promise<number> => {
  // Prepared entries must be reconciled even if an error handler changed the
  // task from running to failed before it could persist the mutation result.
  const storedTasks = (await tasks().findAsync({})) as CoworkTask[]
  const interrupted = storedTasks.filter(
    (task) =>
      task.status === 'running' ||
      task.toolExecutions?.some((execution) => execution.status === 'prepared')
  )
  for (const stored of interrupted) {
    const task = clean(stored)
    const { ambiguousMutations } = reconcileInterruptedToolCalls(task)
    task.status = 'paused'
    delete task.pendingApproval
    task.error = ambiguousMutations
      ? `The app closed while ${ambiguousMutations} file action${ambiguousMutations === 1 ? ' was' : 's were'} in progress. ` +
        'Workspace did not repeat potentially ambiguous actions. Inspect the connected project before resuming.'
      : 'The app closed while this task was running. Review its activity, then resume when ready.'
    appendActivity(task, {
      type: 'system',
      label: 'Task paused after app restart',
      detail: task.error,
      status: 'waiting'
    })
    await replaceTask(task)
  }
  return interrupted.length
}
