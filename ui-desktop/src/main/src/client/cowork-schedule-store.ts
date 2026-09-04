import { randomUUID } from 'node:crypto'
import { coworkScheduleCollection } from './cowork-schedule-database'
import {
  computeNextCoworkRunAt,
  systemTimeZone,
  validateCoworkCadence,
  validateTimeZone
} from './cowork-schedule-recurrence'
import type {
  CoworkSchedule,
  CoworkScheduleTaskTemplate,
  CreateCoworkScheduleInput,
  UpdateCoworkScheduleInput
} from './cowork-schedule.types'

const schedules = () => coworkScheduleCollection()
const clean = <T>(value: T): T => JSON.parse(JSON.stringify(value))
const MAX_SCHEDULES_PER_PROJECT = 100

function validateTaskTemplate(task: CoworkScheduleTaskTemplate): CoworkScheduleTaskTemplate {
  if (!task || typeof task !== 'object') throw new Error('Scheduled task details are required.')
  const goal = task.goal?.trim()
  if (!goal) throw new Error('Describe the outcome for each scheduled task.')
  const title = task.title?.trim() || goal.slice(0, 80)
  if (!task.model?.modelId?.trim() || !task.model?.modelName?.trim()) {
    throw new Error('Choose a model for this schedule.')
  }
  if (typeof task.model.isLocal !== 'boolean') throw new Error('Scheduled task model is invalid.')
  return clean({
    title,
    goal,
    model: {
      ...task.model,
      modelId: task.model.modelId.trim(),
      modelName: task.model.modelName.trim()
    }
  })
}

function nextRunAtFor(
  schedule: Pick<CoworkSchedule, 'status' | 'cadence' | 'timeZone'>,
  now: number
) {
  if (schedule.status !== 'active') return undefined
  return computeNextCoworkRunAt(schedule.cadence, schedule.timeZone, now)
}

async function replaceSchedule(
  current: CoworkSchedule,
  patch: Partial<CoworkSchedule>,
  now = Date.now()
): Promise<CoworkSchedule> {
  const next = clean({
    ...current,
    ...patch,
    schemaVersion: 1 as const,
    revision: current.revision + 1,
    updatedAt: now
  })
  const replaced = await schedules().updateAsync(
    { id: current.id, revision: current.revision },
    next,
    {}
  )
  if (replaced !== 1) {
    throw new Error('This Workspace schedule changed in another operation. Refresh it and try again.')
  }
  return clean(next)
}

export async function createCoworkSchedule(
  input: CreateCoworkScheduleInput,
  now = Date.now()
): Promise<CoworkSchedule> {
  if (!Number.isFinite(now)) throw new Error('Schedule creation time must be finite.')
  const projectId = input.projectId?.trim()
  if (!projectId) throw new Error('Choose a project for this schedule.')
  const name = input.name?.trim()
  if (!name) throw new Error('Schedule name is required.')
  const cadence = validateCoworkCadence(input.cadence)
  const timeZone = validateTimeZone(input.timeZone ?? systemTimeZone())
  const status = input.status ?? 'active'
  if (status !== 'active' && status !== 'paused') throw new Error('Invalid schedule status.')
  if ((await schedules().countAsync({ projectId })) >= MAX_SCHEDULES_PER_PROJECT) {
    throw new Error(`This project has reached the ${MAX_SCHEDULES_PER_PROJECT}-schedule limit.`)
  }

  const schedule: CoworkSchedule = {
    schemaVersion: 1,
    revision: 1,
    id: randomUUID(),
    projectId,
    name,
    task: validateTaskTemplate(input.task),
    cadence,
    timeZone,
    status,
    createdAt: now,
    updatedAt: now
  }
  schedule.nextRunAt = nextRunAtFor(schedule, now)
  const stored = clean(schedule)
  await schedules().insertAsync(stored)
  return clean(stored)
}

export async function listCoworkSchedules(projectId?: string): Promise<CoworkSchedule[]> {
  const query = projectId ? { projectId } : {}
  const result = (await schedules().findAsync(query)) as CoworkSchedule[]
  return clean(result.sort((left, right) => right.updatedAt - left.updatedAt))
}

export async function getCoworkSchedule(id: string): Promise<CoworkSchedule | null> {
  const result = (await schedules().findOneAsync({ id })) as CoworkSchedule | null
  return result ? clean(result) : null
}

export async function updateCoworkSchedule(
  id: string,
  patch: UpdateCoworkScheduleInput,
  now = Date.now()
): Promise<CoworkSchedule> {
  const current = await getCoworkSchedule(id)
  if (!current) throw new Error('Workspace schedule not found.')
  if (patch.name !== undefined && !patch.name.trim()) throw new Error('Schedule name is required.')

  const cadence =
    patch.cadence !== undefined ? validateCoworkCadence(patch.cadence) : current.cadence
  const timeZone =
    patch.timeZone !== undefined ? validateTimeZone(patch.timeZone) : current.timeZone
  const recurrenceChanged = patch.cadence !== undefined || patch.timeZone !== undefined
  const nextPatch: Partial<CoworkSchedule> = {
    ...(patch.name !== undefined ? { name: patch.name.trim() } : {}),
    ...(patch.task !== undefined ? { task: validateTaskTemplate(patch.task) } : {}),
    ...(patch.cadence !== undefined ? { cadence } : {}),
    ...(patch.timeZone !== undefined ? { timeZone } : {})
  }
  if (recurrenceChanged) {
    nextPatch.nextRunAt = nextRunAtFor({ status: current.status, cadence, timeZone }, now)
  }
  return replaceSchedule(current, nextPatch, now)
}

export async function pauseCoworkSchedule(id: string, now = Date.now()): Promise<CoworkSchedule> {
  const current = await getCoworkSchedule(id)
  if (!current) throw new Error('Workspace schedule not found.')
  return replaceSchedule(current, { status: 'paused', nextRunAt: undefined }, now)
}

export async function resumeCoworkSchedule(id: string, now = Date.now()): Promise<CoworkSchedule> {
  const current = await getCoworkSchedule(id)
  if (!current) throw new Error('Workspace schedule not found.')
  const active = { ...current, status: 'active' as const }
  return replaceSchedule(
    current,
    { status: 'active', nextRunAt: nextRunAtFor(active, now), lastError: undefined },
    now
  )
}

export async function deleteCoworkSchedule(id: string): Promise<void> {
  await schedules().removeAsync({ id }, {})
}

/**
 * Atomically advances a due schedule before its callback starts. Advancing from
 * `claimedAt`, rather than the old due time, collapses any number of missed
 * occurrences into one safe catch-up run.
 */
export async function claimDueCoworkSchedule(
  id: string,
  expectedNextRunAt: number,
  claimedAt: number
): Promise<CoworkSchedule | null> {
  const current = await getCoworkSchedule(id)
  if (
    !current ||
    current.status !== 'active' ||
    current.nextRunAt !== expectedNextRunAt ||
    current.runningSince !== undefined
  ) {
    return null
  }

  const next = clean({
    ...current,
    revision: current.revision + 1,
    updatedAt: claimedAt,
    lastScheduledFor: expectedNextRunAt,
    runningSince: claimedAt,
    nextRunAt: computeNextCoworkRunAt(current.cadence, current.timeZone, claimedAt)
  })
  const replaced = await schedules().updateAsync(
    {
      id,
      revision: current.revision,
      status: 'active',
      nextRunAt: expectedNextRunAt,
      runningSince: { $exists: false }
    },
    next,
    {}
  )
  return replaced === 1 ? clean(next) : null
}

export async function claimCoworkScheduleRunNow(
  id: string,
  claimedAt: number
): Promise<CoworkSchedule> {
  const current = await getCoworkSchedule(id)
  if (!current) throw new Error('Workspace schedule not found.')
  if (current.runningSince !== undefined) throw new Error('This schedule is already running.')

  const next = clean({
    ...current,
    revision: current.revision + 1,
    updatedAt: claimedAt,
    runningSince: claimedAt
  })
  const replaced = await schedules().updateAsync(
    { id, revision: current.revision, runningSince: { $exists: false } },
    next,
    {}
  )
  if (replaced !== 1) throw new Error('This schedule is already running.')
  return clean(next)
}

export async function finishCoworkScheduleRun(
  id: string,
  runningSince: number,
  result: { finishedAt: number; taskId?: string; error?: string }
): Promise<void> {
  const set: Record<string, unknown> = {
    updatedAt: result.finishedAt,
    lastRunAt: result.finishedAt
  }
  const unset: Record<string, true> = { runningSince: true }
  if (result.taskId) set.lastTaskId = result.taskId
  else unset.lastTaskId = true
  if (result.error) set.lastError = result.error
  else unset.lastError = true

  await schedules().updateAsync(
    { id, runningSince },
    { $set: set, $unset: unset, $inc: { revision: 1 } },
    {}
  )
}

/** Clears stale execution leases without replaying a possibly-created task. */
export async function recoverInterruptedCoworkScheduleRuns(now = Date.now()): Promise<number> {
  const interrupted = (await schedules().findAsync({
    runningSince: { $exists: true }
  })) as CoworkSchedule[]
  let recovered = 0
  for (const schedule of interrupted) {
    const replaced = await schedules().updateAsync(
      { id: schedule.id, runningSince: schedule.runningSince },
      {
        $set: {
          updatedAt: now,
          lastError:
            'The app closed while this schedule was starting. The occurrence was not replayed.'
        },
        $unset: { runningSince: true },
        $inc: { revision: 1 }
      },
      {}
    )
    recovered += replaced
  }
  return recovered
}
