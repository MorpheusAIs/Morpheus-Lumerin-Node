import { chmodSync } from 'node:fs'

// The original `nedb` calls util.isDate, which Node 23 removed (Electron 43
// embeds Node 24). This fork reads the same data files and has native *Async
// methods.
const Datastore = require('@seald-io/nedb')

const GATED_METHODS = [
  'findAsync',
  'findOneAsync',
  'insertAsync',
  'removeAsync',
  'countAsync'
] as const

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

  const db = new Datastore({
    filename,
    autoload: true,
    onload: (loadError: Error | null | undefined) => {
      if (loadError) {
        rejectIndexes(loadError)
        return
      }
      // These indexes intentionally remain non-unique. UUID-backed IDs should
      // be unique in normal operation, but enforcing that during a migration
      // would make an otherwise-readable legacy database fail to load if it
      // contains a duplicate or incomplete historical row.
      indexFields
        .reduce(
          (previous, fieldName) =>
            previous.then(() => db.ensureIndexAsync({ fieldName, unique: false, sparse: false })),
          Promise.resolve()
        )
        .then(() => resolveIndexes(), rejectIndexes)
    }
  })

  // NeDB's crash-safe compaction replaces the data file and inherits the
  // process umask. Restore private permissions after every such replacement.
  db.on('compaction.done', () => chmodSync(filename, 0o600))

  const gated: Record<string, (...args: any[]) => unknown> = {}
  for (const method of GATED_METHODS) {
    gated[method] = async (...args: any[]) => {
      await indexesReady
      // find/findOne/count return a lazy Cursor that is queued only once it is
      // awaited. Queue it now, as nedb did, so it keeps its place before any
      // write issued after it.
      const result = db[method](...args)
      return typeof result?.execAsync === 'function' ? result.execAsync() : result
    }
  }
  // Callers compare the result with an affected-row count, which is what the
  // original nedb callback reported. The fork resolves { numAffected, ... }.
  gated.updateAsync = async (...args: any[]) => {
    await indexesReady
    return (await db.updateAsync(...args)).numAffected
  }
  // Cursor-based callers cannot use the gated methods, so expose the same
  // readiness gate without exposing how the indexes are initialized.
  gated.waitForIndexesAsync = () => indexesReady
  // Observe initialization failures even if no operation has reached the store
  // yet; callers still receive the original rejection through the gate above.
  void indexesReady.catch((error) =>
    console.error(`Could not initialize Workspace datastore indexes for ${filename}:`, error)
  )

  db.setAutocompactionInterval(compactionInterval)

  // The fork implements find(), update() and the rest on top of its own
  // *Async methods, so the gated versions cannot replace them on the datastore
  // itself: db.find() would then return a Promise instead of a Cursor. Serve
  // them from a proxy and bind everything else to the real datastore.
  return new Proxy(db, {
    get(target, property) {
      if (typeof property === 'string' && Object.hasOwn(gated, property)) return gated[property]
      const value = Reflect.get(target, property, target)
      return typeof value === 'function' ? value.bind(target) : value
    }
  })
}
