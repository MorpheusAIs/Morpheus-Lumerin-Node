import {
  claimCoworkScheduleRunNow,
  claimDueCoworkSchedule,
  finishCoworkScheduleRun,
  getCoworkSchedule,
  listCoworkSchedules,
  recoverInterruptedCoworkScheduleRuns
} from './cowork-schedule-store'
import type {
  CoworkSchedule,
  CoworkScheduleRunCallback,
  CoworkScheduleTrigger
} from './cowork-schedule.types'

export interface CoworkSchedulerOptions {
  now?: () => number
  pollIntervalMs?: number
  /** Limits startup catch-up work when many schedules are overdue. */
  maxRunsPerTick?: number
  onError?: (error: Error) => void
}

const asError = (value: unknown): Error =>
  value instanceof Error ? value : new Error(typeof value === 'string' ? value : 'Schedule failed')

const safeErrorMessage = (value: unknown): string => asError(value).message.slice(0, 2_000)

export class CoworkScheduler {
  private readonly now: () => number
  private readonly pollIntervalMs: number
  private readonly maxRunsPerTick: number
  private readonly onError?: (error: Error) => void
  private interval: NodeJS.Timeout | undefined
  private tickPromise: Promise<number> | undefined
  private readonly activeSchedules = new Set<string>()
  private readonly activeControllers = new Map<string, AbortController>()
  private started = false
  private stopped = false

  constructor(
    private readonly runOccurrence: CoworkScheduleRunCallback,
    options: CoworkSchedulerOptions = {}
  ) {
    if (typeof runOccurrence !== 'function') throw new Error('A schedule run callback is required.')
    this.now = options.now ?? Date.now
    this.pollIntervalMs = options.pollIntervalMs ?? 30_000
    this.maxRunsPerTick = options.maxRunsPerTick ?? 3
    this.onError = options.onError
    if (!Number.isFinite(this.pollIntervalMs) || this.pollIntervalMs < 1_000) {
      throw new Error('Schedule poll interval must be at least one second.')
    }
    if (!Number.isInteger(this.maxRunsPerTick) || this.maxRunsPerTick < 1) {
      throw new Error('Schedule run limit must be a positive integer.')
    }
  }

  async start(): Promise<void> {
    if (this.started) return
    this.stopped = false
    this.started = true
    try {
      await recoverInterruptedCoworkScheduleRuns(this.now())
    } catch (error) {
      this.started = false
      throw error
    }
    // stop() may have been called while durable recovery was in flight.
    if (!this.started) return
    this.interval = setInterval(() => this.wake(), this.pollIntervalMs)
    this.interval.unref?.()
    this.wake()
  }

  stop(): void {
    this.stopped = true
    this.started = false
    if (this.interval) clearInterval(this.interval)
    this.interval = undefined
    for (const controller of this.activeControllers.values()) controller.abort()
  }

  /** Promptly rechecks persisted schedules after external CRUD changes. */
  wake(): void {
    void this.tick().catch((error) => this.report(error))
  }

  /** Public for deterministic main-process tests and explicit app wake events. */
  tick(): Promise<number> {
    if (this.tickPromise) return this.tickPromise
    const operation = this.processDueSchedules().finally(() => {
      if (this.tickPromise === operation) this.tickPromise = undefined
    })
    this.tickPromise = operation
    return operation
  }

  async runNow(scheduleId: string): Promise<CoworkSchedule> {
    if (this.stopped) throw new Error('The Cowork scheduler is stopped.')
    if (this.activeSchedules.has(scheduleId)) throw new Error('This schedule is already running.')
    const claimedAt = this.now()
    const claimed = await claimCoworkScheduleRunNow(scheduleId, claimedAt)
    await this.execute(claimed, claimedAt, claimedAt, 'run-now', true)
    return (await getCoworkSchedule(scheduleId)) ?? claimed
  }

  /** Prevents a claimed occurrence from creating work after its schedule is deleted. */
  cancel(scheduleId: string): void {
    this.activeControllers.get(scheduleId)?.abort()
  }

  private async processDueSchedules(): Promise<number> {
    if (this.stopped) return 0
    const now = this.now()
    const due = (await listCoworkSchedules())
      .filter(
        (schedule) =>
          schedule.status === 'active' &&
          schedule.nextRunAt !== undefined &&
          schedule.nextRunAt <= now &&
          schedule.runningSince === undefined &&
          !this.activeSchedules.has(schedule.id)
      )
      .sort((left, right) => (left.nextRunAt ?? 0) - (right.nextRunAt ?? 0))
      .slice(0, this.maxRunsPerTick)

    let executed = 0
    for (const schedule of due) {
      if (this.stopped) break
      const scheduledFor = schedule.nextRunAt!
      const claimedAt = this.now()
      const claimed = await claimDueCoworkSchedule(schedule.id, scheduledFor, claimedAt)
      if (!claimed) continue
      executed += 1
      await this.execute(claimed, scheduledFor, claimedAt, 'scheduled', false)
    }
    return executed
  }

  private async execute(
    schedule: CoworkSchedule,
    scheduledFor: number,
    triggeredAt: number,
    trigger: CoworkScheduleTrigger,
    rethrow: boolean
  ): Promise<void> {
    this.activeSchedules.add(schedule.id)
    const controller = new AbortController()
    this.activeControllers.set(schedule.id, controller)
    try {
      if (this.stopped)
        throw new Error('The Cowork scheduler stopped before this occurrence began.')
      const result = await this.runOccurrence({
        schedule,
        scheduledFor,
        triggeredAt,
        trigger,
        signal: controller.signal
      })
      await finishCoworkScheduleRun(schedule.id, triggeredAt, {
        finishedAt: this.now(),
        taskId: result?.taskId
      })
    } catch (value) {
      const error = asError(value)
      try {
        await finishCoworkScheduleRun(schedule.id, triggeredAt, {
          finishedAt: this.now(),
          error: safeErrorMessage(error)
        })
      } catch (finishError) {
        this.report(finishError)
      }
      this.report(error)
      if (rethrow) throw error
    } finally {
      this.activeSchedules.delete(schedule.id)
      if (this.activeControllers.get(schedule.id) === controller) {
        this.activeControllers.delete(schedule.id)
      }
    }
  }

  private report(value: unknown): void {
    const error = asError(value)
    if (this.onError) this.onError(error)
    else console.error('[CoworkScheduler]', error)
  }
}

export function createCoworkScheduler(
  runOccurrence: CoworkScheduleRunCallback,
  options?: CoworkSchedulerOptions
): CoworkScheduler {
  return new CoworkScheduler(runOccurrence, options)
}
