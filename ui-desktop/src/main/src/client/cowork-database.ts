import { prepareCoworkDataFile } from './cowork-data-files'
import { createIndexedCoworkDatastore } from './cowork-datastore'

// Kept separate from the legacy cache database. Settings → Clear cache calls
// dbManager.dropDatabase(), and a Cowork project/task history is user data, not
// a disposable cache.
const collections = new Map<string, any>()

const INDEX_FIELDS = {
  // Project IDs are used for every task creation/update and project edit.
  projects: ['id'],
  // These are the task hot paths: direct lookup, project history, and startup
  // recovery of interrupted work.
  tasks: ['id', 'projectId', 'status'],
  // Display transcripts are independent from a marketplace session and page
  // backwards by their durable, per-task sequence.
  messages: ['id', 'taskId', 'sequence'],
  // One fail-closed policy applies to every current and future Workspace task.
  preferences: ['id']
} as const

export function coworkCollection(name: keyof typeof INDEX_FIELDS): any {
  const existing = collections.get(name)
  if (existing) return existing

  const filename = prepareCoworkDataFile(`${name}.db`)
  const db = createIndexedCoworkDatastore(
    filename,
    INDEX_FIELDS[name],
    name === 'tasks' || name === 'messages' ? 5 * 60_000 : 60_000
  )
  // Task records may contain bounded multi-megabyte transcripts. Compacting
  // every 30 seconds repeatedly rewrote the whole store during active runs and
  // competed with model/file work on the main process. Keep compaction durable
  // but infrequent enough to avoid a regular UI-stall cadence.
  collections.set(name, db)
  return db
}
