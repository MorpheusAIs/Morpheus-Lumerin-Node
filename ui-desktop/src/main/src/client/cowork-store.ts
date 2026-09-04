import { randomUUID } from 'node:crypto'
import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { coworkCollection } from './cowork-database'
import { executionMatchesToolCall } from './cowork-mutation-journal'
import {
  CoworkActivity,
  CoworkAgentMessage,
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
const MAX_TASKS_PER_PROJECT = 500
const MAX_DISPLAY_HISTORY_BYTES = 4 * 1024 * 1024
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
  const project: CoworkProject = {
    schemaVersion: 1,
    id: randomUUID(),
    name: input.name.trim(),
    rootPath,
    instructions: input.instructions?.trim() ?? '',
    approvalMode: input.approvalMode ?? 'manual',
    createdAt: now,
    updatedAt: now
  }
  await projects().insertAsync(project)
  return clean(project)
}

export const listProjects = async (): Promise<CoworkProject[]> => {
  const result = (await projects().findAsync({})) as CoworkProject[]
  return clean(
    result.filter((project) => !project.archivedAt).sort((a, b) => b.updatedAt - a.updatedAt)
  )
}

export const getProject = async (id: string): Promise<CoworkProject | null> => {
  const result = (await projects().findOneAsync({ id })) as CoworkProject | null
  return result ? clean(result) : null
}

export const updateProject = async (
  id: string,
  patch: Partial<
    Pick<CoworkProject, 'name' | 'instructions' | 'approvalMode' | 'extensionSettings'>
  >
): Promise<CoworkProject> => {
  const current = await getProject(id)
  if (!current) throw new Error('Cowork project not found.')
  if (patch.name !== undefined && !patch.name.trim()) throw new Error('Project name is required.')
  if (
    patch.approvalMode !== undefined &&
    !['manual', 'auto', 'skip'].includes(patch.approvalMode)
  ) {
    throw new Error('Invalid approval mode.')
  }
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
    ...(patch.approvalMode !== undefined ? { approvalMode: patch.approvalMode } : {}),
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
  if (!project || project.archivedAt) throw new Error('Cowork project not found.')
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
    createdAt: now
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
    plan: [],
    messages: [userMessage],
    agentMessages: [{ role: 'user', content: input.goal.trim() }],
    activities: [],
    artifacts: [],
    createdAt: now,
    updatedAt: now
  }
  await tasks().insertAsync(task)
  // Project archival and task creation can arrive on separate IPC calls. A
  // second check closes the window where creation read an active project just
  // before archival was persisted.
  const stillActive = await getProject(input.projectId)
  if (!stillActive || stillActive.archivedAt) {
    await tasks().removeAsync({ id: task.id }, {})
    throw new Error('This Cowork project is archived.')
  }
  await projects().updateAsync({ id: input.projectId }, { $set: { updatedAt: now } }, {})
  return clean(task)
}

export const listTasks = async (projectId?: string): Promise<CoworkTask[]> => {
  const query = projectId ? { projectId } : {}
  const result = (await tasks().findAsync(query)) as CoworkTask[]
  return clean(result.sort((a, b) => b.updatedAt - a.updatedAt))
}

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

export const getTask = async (id: string): Promise<CoworkTask | null> => {
  const result = (await tasks().findOneAsync({ id })) as CoworkTask | null
  return result ? clean(result) : null
}

export const replaceTask = async (task: CoworkTask): Promise<CoworkTask> => {
  const expectedRevision = task.revision ?? 0
  const next = clean({ ...task, updatedAt: Date.now(), revision: expectedRevision + 1 })
  const replaced = await tasks().updateAsync({ id: task.id, revision: expectedRevision }, next, {})
  if (replaced !== 1) {
    throw new Error('This Cowork task changed in another operation. Refresh it and try again.')
  }
  // Keep the runner's existing history-array identities so their incremental
  // byte counters survive frequent saves. The persisted document is still a
  // detached JSON-safe clone.
  task.updatedAt = next.updatedAt
  task.revision = next.revision
  return clean(next)
}

export const setTaskStatus = async (
  id: string,
  status: CoworkTaskStatus,
  extra: Partial<CoworkTask> = {}
): Promise<CoworkTask> => {
  const task = await getTask(id)
  if (!task) throw new Error('Cowork task not found.')
  return replaceTask({ ...task, ...extra, status })
}

export const appendDisplayMessage = (
  task: CoworkTask,
  role: CoworkDisplayMessage['role'],
  content: string
): void => {
  appendHistoryItem(task.messages, {
    id: randomUUID(),
    role,
    content: content.slice(0, 200_000),
    createdAt: Date.now()
  })
  while (
    task.messages.length > 1 &&
    (task.messages.length > 500 || historyBytes(task.messages) > MAX_DISPLAY_HISTORY_BYTES)
  ) {
    removeHistoryPrefix(task.messages, 1)
  }
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
        'Cowork did not repeat it because its outcome may be ambiguous. Inspect the connected project before issuing a new instruction.'
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
                'Cowork blocked a reused tool call ID whose action name or arguments had changed. No file action ran.'
            })
          : (execution?.resultMessage ??
            JSON.stringify({
              ok: false,
              error:
                `The app stopped before the ${call.function.name.replaceAll('_', ' ')} action returned a durable result. ` +
                'Cowork did not replay it automatically.'
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

export const deleteTask = async (id: string): Promise<void> => {
  await tasks().removeAsync({ id }, {})
}

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
        'Cowork did not repeat potentially ambiguous actions. Inspect the connected project before resuming.'
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
