import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest'

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

const localModel = {
  modelId: 'local-test',
  modelName: 'Local test model',
  isLocal: true
}

beforeAll(async () => {
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-cowork-schedules-'))
  electron.userData = path.join(suiteDirectory, 'user-data')
})

afterAll(async () => {
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

describe.sequential('Cowork schedule persistence and execution', () => {
  it('persists cadence fields and supports pause, resume, update, and delete', async () => {
    const store = await import('./cowork-schedule-store')
    const createdAt = Date.UTC(2026, 0, 5, 8, 0)
    const schedule = await store.createCoworkSchedule(
      {
        projectId: 'project-one',
        name: '  Morning brief  ',
        task: { title: '  Brief  ', goal: '  Prepare the morning brief.  ', model: localModel },
        cadence: { kind: 'weekdays', hour: 9, minute: 15 },
        timeZone: 'Asia/Kuala_Lumpur'
      },
      createdAt
    )

    expect(schedule).toMatchObject({
      schemaVersion: 1,
      revision: 1,
      name: 'Morning brief',
      task: { title: 'Brief', goal: 'Prepare the morning brief.' },
      cadence: { kind: 'weekdays', hour: 9, minute: 15 },
      timeZone: 'Asia/Kuala_Lumpur',
      status: 'active'
    })
    expect(schedule.nextRunAt).toBe(Date.UTC(2026, 0, 6, 1, 15))

    const paused = await store.pauseCoworkSchedule(schedule.id, createdAt + 1)
    expect(paused.status).toBe('paused')
    expect(paused.nextRunAt).toBeUndefined()

    const resumed = await store.resumeCoworkSchedule(schedule.id, createdAt + 2)
    expect(resumed.status).toBe('active')
    expect(resumed.nextRunAt).toBe(Date.UTC(2026, 0, 6, 1, 15))

    const updated = await store.updateCoworkSchedule(
      schedule.id,
      { cadence: { kind: 'hourly', minute: 45 }, timeZone: 'UTC' },
      Date.UTC(2026, 0, 5, 8, 30)
    )
    expect(updated.nextRunAt).toBe(Date.UTC(2026, 0, 5, 8, 45))

    vi.resetModules()
    const reloadedStore = await import('./cowork-schedule-store')
    expect(await reloadedStore.getCoworkSchedule(schedule.id)).toMatchObject({
      id: schedule.id,
      cadence: { kind: 'hourly', minute: 45 },
      timeZone: 'UTC'
    })

    await reloadedStore.deleteCoworkSchedule(schedule.id)
    expect(await reloadedStore.getCoworkSchedule(schedule.id)).toBeNull()
  })

  it('collapses missed occurrences into one catch-up run and advances from now', async () => {
    const store = await import('./cowork-schedule-store')
    const schedulerModule = await import('./cowork-scheduler')
    const originalDueAt = Date.UTC(2026, 0, 1, 9, 0)
    const currentTime = Date.UTC(2026, 0, 3, 12, 0)
    const schedule = await store.createCoworkSchedule(
      {
        projectId: 'project-two',
        name: 'Daily report',
        task: { title: 'Report', goal: 'Prepare one fresh report.', model: localModel },
        cadence: { kind: 'daily', hour: 9, minute: 0 },
        timeZone: 'UTC'
      },
      Date.UTC(2026, 0, 1, 8, 0)
    )
    expect(schedule.nextRunAt).toBe(originalDueAt)

    const occurrences: Array<{ scheduledFor: number; trigger: string }> = []
    const scheduler = schedulerModule.createCoworkScheduler(
      async (occurrence) => {
        occurrences.push({ scheduledFor: occurrence.scheduledFor, trigger: occurrence.trigger })
        return { taskId: 'fresh-task-one' }
      },
      { now: () => currentTime, onError: (error) => expect.unreachable(error.message) }
    )

    expect(await scheduler.tick()).toBe(1)
    expect(await scheduler.tick()).toBe(0)
    expect(occurrences).toEqual([{ scheduledFor: originalDueAt, trigger: 'scheduled' }])
    expect(await store.getCoworkSchedule(schedule.id)).toMatchObject({
      lastScheduledFor: originalDueAt,
      lastRunAt: currentTime,
      lastTaskId: 'fresh-task-one',
      nextRunAt: Date.UTC(2026, 0, 4, 9, 0)
    })
  })

  it('runs manual schedules on demand without inventing a recurring occurrence', async () => {
    const store = await import('./cowork-schedule-store')
    const { createCoworkScheduler } = await import('./cowork-scheduler')
    const now = Date.UTC(2026, 0, 4, 14, 0)
    const schedule = await store.createCoworkSchedule(
      {
        projectId: 'project-three',
        name: 'Manual cleanup',
        task: { title: 'Cleanup', goal: 'Organize the selected project.', model: localModel },
        cadence: { kind: 'manual' },
        timeZone: 'UTC',
        status: 'paused'
      },
      now - 1
    )

    let trigger = ''
    const scheduler = createCoworkScheduler(
      async (occurrence) => {
        trigger = occurrence.trigger
        return { taskId: 'manual-task-one' }
      },
      { now: () => now }
    )
    const result = await scheduler.runNow(schedule.id)

    expect(trigger).toBe('run-now')
    expect(result).toMatchObject({
      status: 'paused',
      lastTaskId: 'manual-task-one',
      lastRunAt: now
    })
    expect(result.nextRunAt).toBeUndefined()
  })

  it('aborts a claimed run-now occurrence when the scheduler is cancelled', async () => {
    const store = await import('./cowork-schedule-store')
    const { createCoworkScheduler } = await import('./cowork-scheduler')
    const now = Date.UTC(2026, 0, 4, 15, 0)
    const schedule = await store.createCoworkSchedule(
      {
        projectId: 'project-four',
        name: 'Cancellable manual task',
        task: {
          title: 'Wait for cancellation',
          goal: 'Do not create work after this occurrence is cancelled.',
          model: localModel
        },
        cadence: { kind: 'manual' },
        timeZone: 'UTC',
        status: 'paused'
      },
      now - 1
    )

    let markClaimed!: () => void
    const claimed = new Promise<void>((resolve) => {
      markClaimed = resolve
    })
    let observedSignal: AbortSignal | undefined
    let workCreated = false
    const reportedErrors: Error[] = []
    const scheduler = createCoworkScheduler(
      async (occurrence) => {
        observedSignal = occurrence.signal
        markClaimed()
        await new Promise<void>((_resolve, reject) => {
          occurrence.signal.addEventListener(
            'abort',
            () => reject(new Error('Scheduled occurrence aborted.')),
            { once: true }
          )
        })
        workCreated = true
      },
      { now: () => now, onError: (error) => reportedErrors.push(error) }
    )

    const running = scheduler.runNow(schedule.id)
    await claimed
    expect((await store.getCoworkSchedule(schedule.id))?.runningSince).toBe(now)

    scheduler.cancel(schedule.id)

    await expect(running).rejects.toThrow('Scheduled occurrence aborted.')
    expect(observedSignal?.aborted).toBe(true)
    expect(workCreated).toBe(false)
    expect(reportedErrors.map((error) => error.message)).toEqual(['Scheduled occurrence aborted.'])
    expect(await store.getCoworkSchedule(schedule.id)).toMatchObject({
      lastRunAt: now,
      lastError: 'Scheduled occurrence aborted.'
    })
    expect((await store.getCoworkSchedule(schedule.id))?.runningSince).toBeUndefined()
  })

  it('aborts every claimed occurrence when the scheduler stops', async () => {
    const store = await import('./cowork-schedule-store')
    const { createCoworkScheduler } = await import('./cowork-scheduler')
    const now = Date.UTC(2026, 0, 4, 16, 0)
    const schedules = await Promise.all(
      ['one', 'two'].map((suffix) =>
        store.createCoworkSchedule(
          {
            projectId: 'project-five',
            name: `Stoppable task ${suffix}`,
            task: {
              title: `Wait ${suffix}`,
              goal: 'Stop this occurrence when the scheduler stops.',
              model: localModel
            },
            cadence: { kind: 'manual' },
            timeZone: 'UTC',
            status: 'paused'
          },
          now - 1
        )
      )
    )

    const claimedIds = new Set<string>()
    let markBothClaimed!: () => void
    const bothClaimed = new Promise<void>((resolve) => {
      markBothClaimed = resolve
    })
    const observedSignals = new Map<string, AbortSignal>()
    const reportedErrors: Error[] = []
    const scheduler = createCoworkScheduler(
      async (occurrence) => {
        claimedIds.add(occurrence.schedule.id)
        observedSignals.set(occurrence.schedule.id, occurrence.signal)
        if (claimedIds.size === schedules.length) markBothClaimed()
        await new Promise<void>((_resolve, reject) => {
          occurrence.signal.addEventListener(
            'abort',
            () => reject(new Error(`Stopped ${occurrence.schedule.id}`)),
            { once: true }
          )
        })
      },
      { now: () => now, onError: (error) => reportedErrors.push(error) }
    )

    const running = schedules.map((schedule) => scheduler.runNow(schedule.id))
    await bothClaimed
    scheduler.stop()

    const results = await Promise.allSettled(running)
    expect(results.every((result) => result.status === 'rejected')).toBe(true)
    expect([...observedSignals.values()].every((signal) => signal.aborted)).toBe(true)
    expect(reportedErrors).toHaveLength(2)
    for (const schedule of schedules) {
      expect((await store.getCoworkSchedule(schedule.id))?.runningSince).toBeUndefined()
    }
  })

  it('wires schedule pause and deletion to cancel a claimed occurrence first', async () => {
    const ipcSource = await fs.readFile(
      path.join(process.cwd(), 'src/main/src/client/cowork-ipc.ts'),
      'utf8'
    )
    const handlerBody = (channel: 'pauseSchedule' | 'deleteSchedule', nextChannel: string) => {
      const start = ipcSource.indexOf(`handle(CHANNEL.${channel}`)
      const end = ipcSource.indexOf(`handle(CHANNEL.${nextChannel}`, start + 1)
      expect(start, `${channel} handler is missing`).toBeGreaterThan(-1)
      return ipcSource.slice(start, end < 0 ? ipcSource.length : end)
    }

    const pauseHandler = handlerBody('pauseSchedule', 'resumeSchedule')
    expect(pauseHandler.indexOf('scheduler.cancel(id)')).toBeLessThan(
      pauseHandler.indexOf('pauseCoworkSchedule(id)')
    )

    const deleteHandler = handlerBody('deleteSchedule', 'runScheduleNow')
    expect(deleteHandler.indexOf('scheduler.cancel(id)')).toBeLessThan(
      deleteHandler.indexOf('deleteCoworkSchedule(id)')
    )
  })
})
