import type { CoworkModelTarget } from './cowork.types'

export type CoworkScheduleCadence =
  | { kind: 'manual' }
  | { kind: 'hourly'; minute: number }
  | { kind: 'daily'; hour: number; minute: number }
  | { kind: 'weekly'; dayOfWeek: number; hour: number; minute: number }
  | { kind: 'weekdays'; hour: number; minute: number }

export type CoworkScheduleStatus = 'active' | 'paused'

export interface CoworkScheduleTaskTemplate {
  title: string
  goal: string
  model: CoworkModelTarget
}

export interface CoworkSchedule {
  schemaVersion: 1
  revision: number
  id: string
  projectId: string
  name: string
  task: CoworkScheduleTaskTemplate
  cadence: CoworkScheduleCadence
  /** IANA time-zone name used to interpret the cadence's wall-clock fields. */
  timeZone: string
  status: CoworkScheduleStatus
  createdAt: number
  updatedAt: number
  nextRunAt?: number
  lastScheduledFor?: number
  lastRunAt?: number
  lastTaskId?: string
  lastError?: string
  runningSince?: number
}

export interface CreateCoworkScheduleInput {
  projectId: string
  name: string
  task: CoworkScheduleTaskTemplate
  cadence: CoworkScheduleCadence
  timeZone?: string
  status?: CoworkScheduleStatus
}

export interface UpdateCoworkScheduleInput {
  name?: string
  task?: CoworkScheduleTaskTemplate
  cadence?: CoworkScheduleCadence
  timeZone?: string
}

export type CoworkScheduleTrigger = 'scheduled' | 'run-now'

export interface CoworkScheduleOccurrence {
  schedule: CoworkSchedule
  scheduledFor: number
  triggeredAt: number
  trigger: CoworkScheduleTrigger
  /** Aborted when the schedule is deleted while this occurrence is being claimed. */
  signal: AbortSignal
}

export interface CoworkScheduleRunResult {
  taskId?: string
}

export type CoworkScheduleRunCallback = (
  occurrence: CoworkScheduleOccurrence
) => Promise<CoworkScheduleRunResult | void>
