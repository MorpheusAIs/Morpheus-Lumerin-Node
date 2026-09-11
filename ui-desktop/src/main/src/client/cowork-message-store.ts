import { coworkCollection } from './cowork-database'
import type { CoworkDisplayMessage } from './cowork.types'

const DEFAULT_PAGE_SIZE = 50
const MAX_PAGE_SIZE = 100

export type CoworkMessageInput = CoworkDisplayMessage & { sequence?: number }

export interface CoworkStoredMessage extends CoworkDisplayMessage {
  taskId: string
  sequence: number
}

export interface CoworkMessagePage {
  messages: CoworkStoredMessage[]
  hasMore: boolean
  nextBeforeSequence?: number
}

let mutationTail: Promise<void> = Promise.resolve()

async function withMessageMutationLock<T>(operation: () => Promise<T>): Promise<T> {
  const previous = mutationTail
  let release!: () => void
  const gate = new Promise<void>((resolve) => {
    release = resolve
  })
  mutationTail = previous.catch(() => undefined).then(() => gate)
  await previous.catch(() => undefined)
  try {
    return await operation()
  } finally {
    release()
  }
}

function requiredId(value: unknown, label: string): string {
  if (typeof value !== 'string' || !value.trim()) throw new Error(`${label} is required.`)
  return value.trim()
}

function validSequence(value: unknown): value is number {
  return Number.isSafeInteger(value) && Number(value) > 0
}

function inputMessage(value: CoworkMessageInput): CoworkDisplayMessage {
  const id = requiredId(value?.id, 'Message ID')
  if (value.role !== 'user' && value.role !== 'assistant') {
    throw new Error(`Message ${id} has an unsupported role.`)
  }
  if (typeof value.content !== 'string') throw new Error(`Message ${id} content is required.`)
  if (!Number.isFinite(value.createdAt)) throw new Error(`Message ${id} timestamp is invalid.`)
  const author = value.author
  if (
    author !== undefined &&
    (typeof author !== 'object' ||
      (author.kind !== 'workspace' && author.kind !== 'model') ||
      ['modelId', 'modelName', 'sessionId'].some(
        (key) =>
          author[key as keyof typeof author] !== undefined &&
          typeof author[key as keyof typeof author] !== 'string'
      ))
  ) {
    throw new Error(`Message ${id} author is invalid.`)
  }
  return {
    id,
    role: value.role,
    content: value.content,
    createdAt: value.createdAt,
    ...(author ? { author: { ...author } } : {})
  }
}

function publicMessage(value: Record<string, unknown>): CoworkStoredMessage {
  const rawAuthor = value.author
  const author =
    rawAuthor &&
    typeof rawAuthor === 'object' &&
    !Array.isArray(rawAuthor) &&
    ((rawAuthor as Record<string, unknown>).kind === 'workspace' ||
      (rawAuthor as Record<string, unknown>).kind === 'model')
      ? (rawAuthor as CoworkDisplayMessage['author'])
      : undefined
  return {
    id: String(value.id),
    taskId: String(value.taskId),
    sequence: Number(value.sequence),
    role: value.role === 'assistant' ? 'assistant' : 'user',
    content: String(value.content ?? ''),
    createdAt: Number(value.createdAt),
    ...(author ? { author: { ...author } } : {})
  }
}

function cursorRows(db: any, query: Record<string, unknown>, limit: number): Promise<any[]> {
  return new Promise((resolve, reject) => {
    db.find(query)
      .sort({ sequence: -1 })
      .limit(limit)
      .exec((error: Error | null, rows: any[]) => {
        if (error) reject(error)
        else resolve(rows)
      })
  })
}

/**
 * Migrates or appends display messages without coupling them to a compute session.
 * Missing messages are never removed: callers may safely pass a legacy task.messages
 * array repeatedly, a suffix after migration, or retry the same save concurrently.
 */
export async function persistCoworkMessages(
  rawTaskId: string,
  messages: readonly CoworkMessageInput[]
): Promise<CoworkStoredMessage[]> {
  const taskId = requiredId(rawTaskId, 'Task ID')
  if (!Array.isArray(messages)) throw new Error('Messages must be an array.')

  return withMessageMutationLock(async () => {
    const db = coworkCollection('messages')
    const normalized = messages.map(inputMessage)
    const ids = [...new Set(normalized.map((message) => message.id))]
    const requestedSequences = [
      ...new Set(
        messages
          .map((message) => message.sequence)
          .filter((sequence): sequence is number => validSequence(sequence))
      )
    ]
    const [rowsWithIds, sequenceCandidates] = (await Promise.all([
      ids.length ? db.findAsync({ id: { $in: ids } }) : Promise.resolve([]),
      requestedSequences.length
        ? db.findAsync({ sequence: { $in: requestedSequences } })
        : Promise.resolve([])
    ])) as [any[], any[]]
    // NeDB can use only one index per query and prefers an exact taskId match,
    // which would materialize the task's complete history before applying the
    // sequence predicate. Query the sequence index directly, then isolate this
    // task's small candidate set in memory.
    const rowsWithRequestedSequences = sequenceCandidates.filter((row) => row.taskId === taskId)

    const foreign = rowsWithIds.find((row) => row.taskId !== taskId)
    if (foreign) {
      throw new Error(`Message ${String(foreign.id)} already belongs to another task.`)
    }

    const existingById = new Map<string, any>()
    for (const row of rowsWithIds) {
      if (typeof row.id !== 'string') continue
      const current = existingById.get(row.id)
      if (!current || (!validSequence(current.sequence) && validSequence(row.sequence))) {
        existingById.set(row.id, row)
      }
    }
    const usedSequences = new Set<number>(
      rowsWithRequestedSequences
        .map((row) => row.sequence)
        .filter((sequence): sequence is number => validSequence(sequence))
    )
    let maximumSequence: number | undefined

    const allocateSequence = async (): Promise<number> => {
      if (maximumSequence === undefined) {
        // Modern callers supply a sequence, so this indexed max lookup is only
        // needed for legacy migration or a real collision. Avoid loading every
        // historical message on each active-run save.
        const [latest] = await cursorRows(db, { taskId, sequence: { $gte: 1 } }, 1)
        let discoveredMaximum = validSequence(latest?.sequence) ? latest.sequence : 0
        for (const used of usedSequences) discoveredMaximum = Math.max(discoveredMaximum, used)
        maximumSequence = discoveredMaximum
      }
      let nextSequence = maximumSequence ?? 0
      do nextSequence += 1
      while (usedSequences.has(nextSequence))
      maximumSequence = nextSequence
      usedSequences.add(nextSequence)
      return nextSequence
    }

    const persistedById = new Map<string, CoworkStoredMessage>()
    for (let index = 0; index < normalized.length; index++) {
      const message = normalized[index]
      const requestedSequence = messages[index].sequence
      const existing = existingById.get(message.id)
      let sequence: number
      if (validSequence(existing?.sequence)) {
        sequence = existing.sequence
      } else if (validSequence(requestedSequence) && !usedSequences.has(requestedSequence)) {
        sequence = requestedSequence
        usedSequences.add(sequence)
      } else {
        sequence = await allocateSequence()
      }

      const stored: CoworkStoredMessage = {
        ...message,
        taskId,
        sequence,
        createdAt: Number.isFinite(existing?.createdAt) ? existing.createdAt : message.createdAt
      }
      const query = existing?._id ? { _id: existing._id } : { taskId, id: message.id }
      await db.updateAsync(query, { $set: stored }, { upsert: true })
      existingById.set(message.id, { ...stored, _id: existing?._id })
      persistedById.set(message.id, stored)
    }

    return normalized.map((message) => ({ ...persistedById.get(message.id)! }))
  })
}

export async function listCoworkMessages(
  rawTaskId: string,
  options: { beforeSequence?: number; limit?: number } = {}
): Promise<CoworkMessagePage> {
  const taskId = requiredId(rawTaskId, 'Task ID')
  const requestedLimit = Number.isFinite(options.limit)
    ? Math.floor(Number(options.limit))
    : DEFAULT_PAGE_SIZE
  const limit = Math.min(MAX_PAGE_SIZE, Math.max(1, requestedLimit))
  const beforeSequence = validSequence(options.beforeSequence) ? options.beforeSequence : undefined
  const query: Record<string, unknown> = {
    taskId,
    ...(beforeSequence ? { sequence: { $lt: beforeSequence } } : {})
  }
  const db = coworkCollection('messages')
  await db.waitForIndexesAsync()
  const rows = await cursorRows(db, query, limit + 1)
  const hasMore = rows.length > limit
  const messages = rows
    .slice(0, limit)
    .map((row) => publicMessage(row))
    .reverse()

  return {
    messages,
    hasMore,
    ...(hasMore && messages.length ? { nextBeforeSequence: messages[0].sequence } : {})
  }
}

/** Explicit project/task deletion is the only operation that removes transcript rows. */
export async function deleteCoworkMessages(rawTaskId: string): Promise<number> {
  const taskId = requiredId(rawTaskId, 'Task ID')
  return withMessageMutationLock(async () => {
    const removed = await coworkCollection('messages').removeAsync({ taskId }, { multi: true })
    return Number(removed) || 0
  })
}
