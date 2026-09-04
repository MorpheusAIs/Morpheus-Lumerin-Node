import { describe, expect, it, vi } from 'vitest'
import type { CoworkScheduleOccurrence } from './cowork-schedule.types'
import { runCoworkScheduledOccurrence } from './cowork-scheduled-occurrence'
import type { CoworkModelTarget, CoworkTask } from './cowork.types'

const NOW = 2_000_000

const model: CoworkModelTarget = {
  modelId: 'model-1',
  modelName: 'Marketplace LLM',
  isLocal: false,
  sessionId: 'session-1',
  sessionEndsAt: NOW + 60_000,
  dataBoundary: 'independent-provider'
}

const task = (): CoworkTask => ({
  schemaVersion: 1,
  revision: 1,
  id: 'task-1',
  projectId: 'project-1',
  title: 'Scheduled task',
  goal: 'Do the scheduled work',
  status: 'queued',
  model,
  plan: [],
  messages: [{ id: 'message-1', role: 'user', content: 'Do the work', createdAt: NOW }],
  agentMessages: [{ role: 'user', content: 'Do the work' }],
  activities: [],
  artifacts: [],
  createdAt: NOW,
  updatedAt: NOW
})

const occurrence = (): CoworkScheduleOccurrence => ({
  schedule: {
    schemaVersion: 1,
    revision: 1,
    id: 'schedule-1',
    projectId: 'project-1',
    name: 'Schedule',
    task: { title: 'Scheduled task', goal: 'Do the scheduled work', model },
    cadence: { kind: 'manual' },
    timeZone: 'UTC',
    status: 'active',
    createdAt: NOW,
    updatedAt: NOW
  },
  scheduledFor: NOW,
  triggeredAt: NOW,
  trigger: 'run-now',
  signal: new AbortController().signal
})

describe('runCoworkScheduledOccurrence', () => {
  it('deletes an untouched draft and pauses after the post-create session validation fails', async () => {
    const created = task()
    const refreshModelTarget = vi
      .fn()
      .mockResolvedValueOnce({ model, fingerprint: 'session-fingerprint' })
      .mockRejectedValueOnce(new Error('The selected model or session is no longer available.'))
    const createTask = vi.fn(async () => created)
    const getTask = vi.fn(async () => created)
    const deleteTask = vi.fn(async () => undefined)
    const pauseSchedule = vi.fn(async () => undefined)
    const startTask = vi.fn(async () => {
      // requestCoworkStart performs this same forced refresh before changing
      // the task, closing the create/start TOCTOU window.
      await refreshModelTarget(model)
    })

    await expect(
      runCoworkScheduledOccurrence(occurrence(), {
        now: () => NOW,
        pauseSchedule,
        refreshModelTarget,
        createTask,
        getTask,
        deleteTask,
        startTask,
        cancelTask: vi.fn(async () => undefined)
      })
    ).rejects.toThrow('selected model or session is no longer available')

    expect(refreshModelTarget).toHaveBeenCalledTimes(2)
    expect(createTask).toHaveBeenCalledTimes(1)
    expect(deleteTask).toHaveBeenCalledWith(created.id)
    expect(pauseSchedule).toHaveBeenCalledWith('schedule-1')
  })

  it('pauses but preserves a task that changed before start failed', async () => {
    const created = task()
    const changed = { ...created, revision: 2, status: 'waiting_approval' as const }
    const deleteTask = vi.fn(async () => undefined)
    const pauseSchedule = vi.fn(async () => undefined)

    await expect(
      runCoworkScheduledOccurrence(occurrence(), {
        now: () => NOW,
        pauseSchedule,
        refreshModelTarget: vi.fn(async () => ({ model, fingerprint: 'session-fingerprint' })),
        createTask: vi.fn(async () => created),
        getTask: vi.fn(async () => changed),
        deleteTask,
        startTask: vi.fn(async () => {
          throw new Error('Session closed while starting.')
        }),
        cancelTask: vi.fn(async () => undefined)
      })
    ).rejects.toThrow('Session closed while starting')

    expect(deleteTask).not.toHaveBeenCalled()
    expect(pauseSchedule).toHaveBeenCalledWith('schedule-1')
  })

  it('pauses without creating a task when the initial session preflight fails', async () => {
    const createTask = vi.fn(async () => task())
    const pauseSchedule = vi.fn(async () => undefined)

    await expect(
      runCoworkScheduledOccurrence(occurrence(), {
        now: () => NOW,
        pauseSchedule,
        refreshModelTarget: vi.fn(async () => {
          throw new Error('The marketplace session has expired.')
        }),
        createTask,
        getTask: vi.fn(async () => null),
        deleteTask: vi.fn(async () => undefined),
        startTask: vi.fn(async () => undefined),
        cancelTask: vi.fn(async () => undefined)
      })
    ).rejects.toThrow('marketplace session has expired')

    expect(createTask).not.toHaveBeenCalled()
    expect(pauseSchedule).toHaveBeenCalledWith('schedule-1')
  })
})
