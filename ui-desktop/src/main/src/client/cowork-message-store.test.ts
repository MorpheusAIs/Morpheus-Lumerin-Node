import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const electron = vi.hoisted(() => ({ userData: '' }))

vi.mock('electron', () => ({
  app: {
    getPath: (name: string) => {
      if (name !== 'userData') throw new Error(`Unexpected Electron path request: ${name}`)
      return electron.userData
    }
  }
}))

let suiteDirectory: string
const openedDatastores: any[] = []

beforeEach(async () => {
  vi.resetModules()
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-cowork-messages-'))
  electron.userData = path.join(suiteDirectory, 'user-data')
})

afterEach(async () => {
  for (const db of openedDatastores.splice(0)) db.persistence.stopAutocompaction()
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

async function store() {
  const api = await import('./cowork-message-store')
  const { coworkCollection } = await import('./cowork-database')
  const messages = coworkCollection('messages')
  openedDatastores.push(messages)
  return { ...api, collection: messages }
}

function message(id: string, createdAt: number, content = `message ${id}`) {
  return {
    id,
    role: (id.startsWith('assistant') ? 'assistant' : 'user') as 'user' | 'assistant',
    content,
    createdAt
  }
}

describe.sequential('Cowork display transcript store', () => {
  it('uses a private indexed datastore independent from task and session records', async () => {
    const { collection, persistCoworkMessages } = await store()
    await persistCoworkMessages('task-1', [message('user-1', 1)])
    await collection.waitForIndexesAsync()

    expect(Object.keys(collection.indexes)).toEqual(
      expect.arrayContaining(['_id', 'id', 'taskId', 'sequence'])
    )
    if (process.platform !== 'win32') {
      const directory = path.join(electron.userData, 'Cowork')
      expect((await fs.stat(directory)).mode & 0o777).toBe(0o700)
      expect((await fs.stat(path.join(directory, 'messages.db'))).mode & 0o777).toBe(0o600)
    }
  })

  it('idempotently migrates an existing display array and preserves stable sequences on upsert', async () => {
    const { collection, listCoworkMessages, persistCoworkMessages } = await store()
    const legacy = [message('user-1', 10), message('assistant-1', 20)]

    const first = await persistCoworkMessages('task-1', legacy)
    const retried = await persistCoworkMessages('task-1', [
      legacy[0],
      { ...legacy[1], content: 'updated streamed response', sequence: 999 }
    ])
    const page = await listCoworkMessages('task-1')

    expect(first.map((item) => item.sequence)).toEqual([1, 2])
    expect(retried.map((item) => item.sequence)).toEqual([1, 2])
    expect(page.messages.map((item) => [item.id, item.content, item.sequence])).toEqual([
      ['user-1', 'message user-1', 1],
      ['assistant-1', 'updated streamed response', 2]
    ])
    expect(await collection.countAsync({ taskId: 'task-1' })).toBe(2)
  })

  it('serializes concurrent retries without duplicating IDs or allocating unstable sequences', async () => {
    const { collection, listCoworkMessages, persistCoworkMessages } = await store()
    const first = message('user-1', 10)
    const second = message('assistant-1', 20)
    const third = message('user-2', 30)

    await Promise.all([
      persistCoworkMessages('task-1', [first]),
      persistCoworkMessages('task-1', [first, second]),
      persistCoworkMessages('task-1', [first, second, third]),
      persistCoworkMessages('task-1', [second, third])
    ])

    const page = await listCoworkMessages('task-1', { limit: 100 })
    expect(page.messages.map((item) => item.sequence)).toEqual([1, 2, 3])
    expect(new Set(page.messages.map((item) => item.id)).size).toBe(3)
    expect(await collection.countAsync({ taskId: 'task-1' })).toBe(3)
  })

  it('honours unused migration sequences, resolves collisions, and never removes omitted rows', async () => {
    const { listCoworkMessages, persistCoworkMessages } = await store()
    await persistCoworkMessages('task-1', [
      { ...message('user-10', 10), sequence: 10 },
      { ...message('assistant-20', 20), sequence: 20 },
      { ...message('user-collision', 30), sequence: 20 }
    ])
    await persistCoworkMessages('task-1', [message('user-new', 40)])
    await persistCoworkMessages('task-1', [{ ...message('assistant-20', 20), sequence: 1 }])

    const page = await listCoworkMessages('task-1', { limit: 100 })
    expect(page.messages.map((item) => [item.id, item.sequence])).toEqual([
      ['user-10', 10],
      ['assistant-20', 20],
      ['user-collision', 21],
      ['user-new', 22]
    ])
  })

  it('returns latest pages chronologically, caps pages at 100, and exposes an exclusive cursor', async () => {
    const { listCoworkMessages, persistCoworkMessages } = await store()
    await persistCoworkMessages(
      'task-pages',
      Array.from({ length: 205 }, (_, index) => message(`user-${index + 1}`, index + 1))
    )

    const latest = await listCoworkMessages('task-pages', { limit: 1_000 })
    const middle = await listCoworkMessages('task-pages', {
      beforeSequence: latest.nextBeforeSequence,
      limit: 100
    })
    const oldest = await listCoworkMessages('task-pages', {
      beforeSequence: middle.nextBeforeSequence,
      limit: 100
    })

    expect(latest.messages).toHaveLength(100)
    expect([latest.messages[0].sequence, latest.messages.at(-1)?.sequence]).toEqual([106, 205])
    expect(latest).toMatchObject({ hasMore: true, nextBeforeSequence: 106 })
    expect([middle.messages[0].sequence, middle.messages.at(-1)?.sequence]).toEqual([6, 105])
    expect(middle).toMatchObject({ hasMore: true, nextBeforeSequence: 6 })
    expect(oldest.messages.map((item) => item.sequence)).toEqual([1, 2, 3, 4, 5])
    expect(oldest.hasMore).toBe(false)
    expect(oldest.nextBeforeSequence).toBeUndefined()
  })

  it('deletes only the explicitly selected task transcript', async () => {
    const { deleteCoworkMessages, listCoworkMessages, persistCoworkMessages } = await store()
    await persistCoworkMessages('task-a', [message('user-a', 1)])
    await persistCoworkMessages('task-b', [message('user-b', 1)])

    expect(await deleteCoworkMessages('task-a')).toBe(1)
    expect((await listCoworkMessages('task-a')).messages).toEqual([])
    expect((await listCoworkMessages('task-b')).messages.map((item) => item.id)).toEqual(['user-b'])
  })
})
