import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createIndexedCoworkDatastore } from './cowork-datastore'

let suiteDirectory: string
const openedDatastores: any[] = []

beforeEach(async () => {
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-cowork-datastore-'))
})

afterEach(async () => {
  for (const db of openedDatastores.splice(0)) db.stopAutocompaction()
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

function open(filename = path.join(suiteDirectory, 'rows.db')) {
  const db = createIndexedCoworkDatastore(filename, ['id'], 60_000)
  openedDatastores.push(db)
  return db
}

describe('Indexed Cowork datastore', () => {
  it('keeps cursor queries working next to the gated async methods', async () => {
    const db = open()
    await db.insertAsync([
      { id: 'a', sequence: 1 },
      { id: 'b', sequence: 2 },
      { id: 'c', sequence: 3 }
    ])
    await db.waitForIndexesAsync()

    const rows = await new Promise<any[]>((resolve, reject) =>
      db
        .find({})
        .sort({ sequence: -1 })
        .limit(2)
        .exec((error: Error | null, documents: any[]) =>
          error ? reject(error) : resolve(documents)
        )
    )
    expect(rows.map((row) => row.id)).toEqual(['c', 'b'])
    expect(await db.findAsync({ id: 'a' })).toHaveLength(1)
    expect(await db.findOneAsync({ id: 'b' })).toMatchObject({ sequence: 2 })
    expect(await db.countAsync({})).toBe(3)
  })

  it('runs a read before a write issued after it', async () => {
    const db = open()
    const read = db.findAsync({})
    const write = db.insertAsync({ id: 'a' })

    expect(await read).toEqual([])
    await write
    expect(await db.countAsync({})).toBe(1)
  })

  it('resolves updates to the affected-row count', async () => {
    const db = open()
    await db.insertAsync({ id: 'a', revision: 1 })

    expect(await db.updateAsync({ id: 'a', revision: 1 }, { $set: { revision: 2 } }, {})).toBe(1)
    expect(await db.updateAsync({ id: 'a', revision: 1 }, { $set: { revision: 3 } }, {})).toBe(0)
    expect(await db.removeAsync({ id: 'a' }, {})).toBe(1)
  })

  it('rejects operations when the data file cannot be loaded', async () => {
    const corrupt = path.join(suiteDirectory, 'corrupt.db')
    await fs.writeFile(corrupt, 'not json\nstill not json\n', { mode: 0o600 })
    const db = open(corrupt)

    await expect(db.waitForIndexesAsync()).rejects.toThrow(/corrupt/)
    await expect(db.findAsync({})).rejects.toThrow(/corrupt/)
  })
})
