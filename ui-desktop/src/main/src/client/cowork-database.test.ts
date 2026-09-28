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
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-cowork-indexes-'))
  electron.userData = path.join(suiteDirectory, 'user-data')
})

afterEach(async () => {
  for (const db of openedDatastores.splice(0)) {
    db.persistence.stopAutocompaction()
  }
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

const indexNames = (db: any): string[] => Object.keys(db.indexes)

describe.sequential('Cowork datastore indexes', () => {
  it('adds migration-safe task indexes without rejecting duplicate legacy IDs', async () => {
    const dataDirectory = path.join(electron.userData, 'Cowork')
    const taskFile = path.join(dataDirectory, 'tasks.db')
    await fs.mkdir(dataDirectory, { recursive: true })
    await fs.writeFile(
      taskFile,
      [
        {
          _id: 'legacy-row-one',
          id: 'duplicate-legacy-id',
          projectId: 'legacy-project',
          status: 'running'
        },
        {
          _id: 'legacy-row-two',
          id: 'duplicate-legacy-id',
          projectId: 'legacy-project',
          status: 'completed'
        }
      ]
        .map((row) => JSON.stringify(row))
        .join('\n') + '\n',
      { mode: 0o600 }
    )

    const { coworkCollection } = await import('./cowork-database')
    const tasks = coworkCollection('tasks')
    openedDatastores.push(tasks)
    await tasks.waitForIndexesAsync()

    expect(indexNames(tasks)).toEqual(expect.arrayContaining(['_id', 'id', 'projectId', 'status']))
    expect(tasks.indexes.id.unique).toBe(false)
    expect(tasks.indexes.projectId.unique).toBe(false)
    expect(tasks.indexes.status.unique).toBe(false)
    expect(await tasks.findAsync({ id: 'duplicate-legacy-id' })).toHaveLength(2)
    expect(await tasks.findAsync({ projectId: 'legacy-project' })).toHaveLength(2)

    const persisted = (await fs.readFile(taskFile, 'utf8'))
      .trim()
      .split('\n')
      .map((line) => JSON.parse(line))
      .filter((entry) => entry.$$indexCreated)
      .map((entry) => entry.$$indexCreated.fieldName)
    expect(persisted).toEqual(expect.arrayContaining(['id', 'projectId', 'status']))
  })

  it('indexes the project and schedule lookup/list keys', async () => {
    const { coworkCollection } = await import('./cowork-database')
    const { coworkScheduleCollection } = await import('./cowork-schedule-database')
    const projects = coworkCollection('projects')
    const schedules = coworkScheduleCollection()
    openedDatastores.push(projects, schedules)

    await Promise.all([projects.waitForIndexesAsync(), schedules.waitForIndexesAsync()])

    expect(indexNames(projects)).toEqual(expect.arrayContaining(['_id', 'id']))
    expect(indexNames(schedules)).toEqual(expect.arrayContaining(['_id', 'id', 'projectId']))
    expect(projects.indexes.id.unique).toBe(false)
    expect(schedules.indexes.id.unique).toBe(false)
    expect(schedules.indexes.projectId.unique).toBe(false)
  })
})
