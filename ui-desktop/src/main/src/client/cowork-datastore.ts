import { chmodSync } from 'node:fs'
import { promisify } from 'node:util'

const Datastore = require('nedb')

const ASYNC_METHODS = ['find', 'findOne', 'insert', 'remove', 'update', 'count'] as const

const ensureIndex = (db: any, fieldName: string): Promise<void> =>
  new Promise((resolve, reject) => {
    // These indexes intentionally remain non-unique. UUID-backed IDs should be
    // unique in normal operation, but enforcing that during a migration would
    // make an otherwise-readable legacy database fail to load if it contains a
    // duplicate or incomplete historical row.
    db.ensureIndex(
      { fieldName, unique: false, sparse: false },
      (error: Error | null | undefined) => {
        if (error) reject(error)
        else resolve()
      }
    )
  })

export function createIndexedCoworkDatastore(
  filename: string,
  indexFields: readonly string[],
  compactionInterval: number
): any {
  let resolveIndexes!: () => void
  let rejectIndexes!: (error: unknown) => void
  const indexesReady = new Promise<void>((resolve, reject) => {
    resolveIndexes = resolve
    rejectIndexes = reject
  })

  let db: any
  db = new Datastore({
    filename,
    autoload: true,
    onload: (loadError: Error | null | undefined) => {
      if (loadError) {
        rejectIndexes(loadError)
        return
      }
      indexFields
        .reduce(
          (previous, fieldName) => previous.then(() => ensureIndex(db, fieldName)),
          Promise.resolve()
        )
        .then(() => resolveIndexes(), rejectIndexes)
    }
  })

  // NeDB's crash-safe compaction replaces the data file and inherits the
  // process umask. Restore private permissions after every such replacement.
  db.on('compaction.done', () => chmodSync(filename, 0o600))

  for (const method of ASYNC_METHODS) {
    const invoke = promisify(db[method].bind(db)) as (...args: any[]) => Promise<unknown>
    db[`${method}Async`] = async (...args: any[]) => {
      await indexesReady
      return invoke(...args)
    }
  }

  // Cursor-based callers cannot use the promisified wrappers, so expose the
  // same readiness gate without exposing how the indexes are initialized.
  db.waitForIndexesAsync = () => indexesReady
  // Observe initialization failures even if no operation has reached the store
  // yet; callers still receive the original rejection through the gate above.
  void indexesReady.catch((error) =>
    console.error(`Could not initialize Cowork datastore indexes for ${filename}:`, error)
  )

  db.persistence.setAutocompactionInterval(compactionInterval)
  return db
}
