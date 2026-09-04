import type { CoworkScheduleOccurrence, CoworkScheduleRunResult } from './cowork-schedule.types'
import type { CoworkModelTarget, CoworkTask } from './cowork.types'

export interface CoworkScheduledOccurrenceDependencies {
  now?: () => number
  pauseSchedule: (scheduleId: string) => Promise<unknown>
  refreshModelTarget: (
    model: CoworkModelTarget
  ) => Promise<{ model: CoworkModelTarget; fingerprint: string }>
  createTask: (input: {
    projectId: string
    title: string
    goal: string
    model: CoworkModelTarget
  }) => Promise<CoworkTask>
  getTask: (taskId: string) => Promise<CoworkTask | null>
  deleteTask: (taskId: string) => Promise<void>
  startTask: (taskId: string) => Promise<unknown>
  cancelTask: (taskId: string) => Promise<unknown>
}

function isUntouchedScheduledTask(current: CoworkTask, created: CoworkTask): boolean {
  return (
    current.id === created.id &&
    current.revision === created.revision &&
    current.status === 'queued' &&
    current.startedAt === undefined &&
    current.pendingApproval === undefined
  )
}

async function removeUntouchedScheduledTask(
  created: CoworkTask,
  dependencies: CoworkScheduledOccurrenceDependencies
): Promise<void> {
  const current = await dependencies.getTask(created.id)
  if (current && isUntouchedScheduledTask(current, created)) {
    await dependencies.deleteTask(created.id)
  }
}

function asError(value: unknown): Error {
  return value instanceof Error ? value : new Error(String(value))
}

/**
 * Runs one persisted schedule occurrence with session checks on both sides of
 * task creation. The second check happens inside startTask. If it fails before
 * the new task is modified, the orphaned queued task is removed and the
 * schedule is paused so it cannot repeatedly create drafts against a closed
 * or expired session.
 */
export async function runCoworkScheduledOccurrence(
  { schedule, signal }: CoworkScheduleOccurrence,
  dependencies: CoworkScheduledOccurrenceDependencies
): Promise<CoworkScheduleRunResult> {
  const now = dependencies.now ?? Date.now
  if (signal.aborted) throw new Error('This schedule was deleted before its task started.')
  if (schedule.task.model.isLocal || !schedule.task.model.sessionId) {
    await dependencies.pauseSchedule(schedule.id)
    throw new Error('Cowork schedules require an active Morpheus marketplace session.')
  }
  if (!schedule.task.model.sessionEndsAt || schedule.task.model.sessionEndsAt <= now()) {
    await dependencies.pauseSchedule(schedule.id)
    throw new Error('The marketplace session selected by this schedule has expired.')
  }

  let refreshedModel: Awaited<
    ReturnType<CoworkScheduledOccurrenceDependencies['refreshModelTarget']>
  >
  try {
    refreshedModel = await dependencies.refreshModelTarget(schedule.task.model)
  } catch (error) {
    await dependencies.pauseSchedule(schedule.id)
    throw error
  }

  if (signal.aborted) throw new Error('This schedule was deleted before its task started.')
  const task = await dependencies.createTask({
    projectId: schedule.projectId,
    title: schedule.task.title,
    goal: schedule.task.goal,
    model: refreshedModel.model
  })
  if (signal.aborted) {
    await dependencies.deleteTask(task.id)
    throw new Error('This schedule was deleted before its task started.')
  }

  try {
    await dependencies.startTask(task.id)
  } catch (startError) {
    const cleanupResults = await Promise.allSettled([
      removeUntouchedScheduledTask(task, dependencies),
      dependencies.pauseSchedule(schedule.id)
    ])
    const cleanupFailure = cleanupResults.find(
      (result): result is PromiseRejectedResult => result.status === 'rejected'
    )
    if (cleanupFailure) {
      const original = asError(startError)
      throw new Error(
        `${original.message} Scheduled-task cleanup also failed: ${asError(cleanupFailure.reason).message}`,
        { cause: original }
      )
    }
    throw startError
  }

  if (signal.aborted) {
    await dependencies.cancelTask(task.id)
    await dependencies.deleteTask(task.id)
    throw new Error('This schedule was deleted while its task was starting.')
  }
  return { taskId: task.id }
}
