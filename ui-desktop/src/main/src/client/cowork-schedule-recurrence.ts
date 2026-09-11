import type { CoworkScheduleCadence } from './cowork-schedule.types'

const MINUTE_MS = 60_000
const HOUR_MS = 60 * MINUTE_MS
const DAY_MS = 24 * HOUR_MS

interface LocalDate {
  year: number
  month: number
  day: number
}

interface ZonedParts extends LocalDate {
  hour: number
  minute: number
  second: number
}

const formatters = new Map<string, Intl.DateTimeFormat>()

function formatterFor(timeZone: string): Intl.DateTimeFormat {
  const existing = formatters.get(timeZone)
  if (existing) return existing

  const formatter = new Intl.DateTimeFormat('en-US', {
    timeZone,
    calendar: 'gregory',
    numberingSystem: 'latn',
    hourCycle: 'h23',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
  // Force Intl to validate before the formatter is cached.
  formatter.format(0)
  formatters.set(timeZone, formatter)
  return formatter
}

function integerPart(parts: Intl.DateTimeFormatPart[], type: Intl.DateTimeFormatPartTypes): number {
  const value = parts.find((part) => part.type === type)?.value
  if (!value) throw new Error(`Unable to read ${type} for schedule time zone`)
  return Number(value)
}

function zonedParts(instant: number, timeZone: string): ZonedParts {
  const parts = formatterFor(timeZone).formatToParts(new Date(instant))
  return {
    year: integerPart(parts, 'year'),
    month: integerPart(parts, 'month'),
    day: integerPart(parts, 'day'),
    hour: integerPart(parts, 'hour'),
    minute: integerPart(parts, 'minute'),
    second: integerPart(parts, 'second')
  }
}

function localDateAt(instant: number, timeZone: string): LocalDate {
  const { year, month, day } = zonedParts(instant, timeZone)
  return { year, month, day }
}

function addLocalDays(date: LocalDate, days: number): LocalDate {
  const result = new Date(Date.UTC(date.year, date.month - 1, date.day + days))
  return {
    year: result.getUTCFullYear(),
    month: result.getUTCMonth() + 1,
    day: result.getUTCDate()
  }
}

function localDayOfWeek(date: LocalDate): number {
  return new Date(Date.UTC(date.year, date.month - 1, date.day)).getUTCDay()
}

function sameLocalMinute(
  parts: ZonedParts,
  date: LocalDate,
  hour: number,
  minute: number
): boolean {
  return (
    parts.year === date.year &&
    parts.month === date.month &&
    parts.day === date.day &&
    parts.hour === hour &&
    parts.minute === minute
  )
}

function offsetAt(instant: number, timeZone: string): number {
  const parts = zonedParts(instant, timeZone)
  const instantAtWholeSecond = Math.floor(instant / 1000) * 1000
  return (
    Date.UTC(parts.year, parts.month - 1, parts.day, parts.hour, parts.minute, parts.second) -
    instantAtWholeSecond
  )
}

/**
 * Resolves one wall-clock minute into an instant. Ambiguous fall-back minutes
 * choose their first occurrence. A spring-forward minute that does not exist
 * runs at the first valid minute later on that same local date.
 */
function resolveLocalMinute(
  date: LocalDate,
  hour: number,
  minute: number,
  timeZone: string
): number | undefined {
  const nominal = Date.UTC(date.year, date.month - 1, date.day, hour, minute)
  const offsets = new Set<number>()

  // Sampling both sides of the date discovers each offset at a DST boundary.
  for (let delta = -36 * HOUR_MS; delta <= 36 * HOUR_MS; delta += 6 * HOUR_MS) {
    offsets.add(offsetAt(nominal + delta, timeZone))
  }

  const exactCandidates = [...offsets]
    .map((offset) => nominal - offset)
    .filter((candidate) => sameLocalMinute(zonedParts(candidate, timeZone), date, hour, minute))
    .sort((left, right) => left - right)
  if (exactCandidates.length > 0) return exactCandidates[0]

  // The local minute was skipped (normally a spring-forward gap). Search only
  // the transition window and select the first later wall-clock minute.
  let firstLater: number | undefined
  const requestedMinute = hour * 60 + minute
  for (
    let candidate = nominal - 18 * HOUR_MS;
    candidate <= nominal + 18 * HOUR_MS;
    candidate += MINUTE_MS
  ) {
    const parts = zonedParts(candidate, timeZone)
    if (parts.year !== date.year || parts.month !== date.month || parts.day !== date.day) continue
    if (parts.hour * 60 + parts.minute < requestedMinute) continue
    if (firstLater === undefined || candidate < firstLater) firstLater = candidate
  }
  return firstLater
}

function validateInteger(value: number, minimum: number, maximum: number, label: string): void {
  if (!Number.isInteger(value) || value < minimum || value > maximum) {
    throw new Error(`${label} must be an integer from ${minimum} through ${maximum}`)
  }
}

export function systemTimeZone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
}

export function validateTimeZone(timeZone: string): string {
  const normalized = timeZone.trim()
  if (!normalized) throw new Error('Schedule time zone is required')
  formatterFor(normalized)
  return normalized
}

export function validateCoworkCadence(cadence: CoworkScheduleCadence): CoworkScheduleCadence {
  if (!cadence || typeof cadence !== 'object') throw new Error('Schedule cadence is required')

  switch (cadence.kind) {
    case 'manual':
      return { kind: 'manual' }
    case 'hourly':
      validateInteger(cadence.minute, 0, 59, 'Minute')
      return { kind: 'hourly', minute: cadence.minute }
    case 'daily':
    case 'weekdays':
      validateInteger(cadence.hour, 0, 23, 'Hour')
      validateInteger(cadence.minute, 0, 59, 'Minute')
      return { kind: cadence.kind, hour: cadence.hour, minute: cadence.minute }
    case 'weekly':
      validateInteger(cadence.dayOfWeek, 0, 6, 'Day of week')
      validateInteger(cadence.hour, 0, 23, 'Hour')
      validateInteger(cadence.minute, 0, 59, 'Minute')
      return {
        kind: 'weekly',
        dayOfWeek: cadence.dayOfWeek,
        hour: cadence.hour,
        minute: cadence.minute
      }
    default: {
      const exhaustive: never = cadence
      throw new Error(`Unsupported schedule cadence: ${String(exhaustive)}`)
    }
  }
}

/** Returns the first scheduled instant strictly after `after`. */
export function computeNextCoworkRunAt(
  cadenceInput: CoworkScheduleCadence,
  timeZoneInput: string,
  after: number
): number | undefined {
  if (!Number.isFinite(after)) throw new Error('Schedule reference time must be finite')
  const cadence = validateCoworkCadence(cadenceInput)
  const timeZone = validateTimeZone(timeZoneInput)

  if (cadence.kind === 'manual') return undefined

  if (cadence.kind === 'hourly') {
    const firstWholeMinute = Math.floor(after / MINUTE_MS) * MINUTE_MS + MINUTE_MS
    // 27 hours covers clock changes and the largest practical UTC offsets.
    for (
      let candidate = firstWholeMinute;
      candidate <= firstWholeMinute + 27 * HOUR_MS;
      candidate += MINUTE_MS
    ) {
      if (zonedParts(candidate, timeZone).minute === cadence.minute) return candidate
    }
    throw new Error('Unable to determine the next hourly schedule occurrence')
  }

  const startDate = localDateAt(after, timeZone)
  for (let dayOffset = 0; dayOffset <= 14; dayOffset += 1) {
    const date = addLocalDays(startDate, dayOffset)
    const dayOfWeek = localDayOfWeek(date)
    if (cadence.kind === 'weekly' && dayOfWeek !== cadence.dayOfWeek) continue
    if (cadence.kind === 'weekdays' && (dayOfWeek === 0 || dayOfWeek === 6)) continue

    const candidate = resolveLocalMinute(date, cadence.hour, cadence.minute, timeZone)
    if (candidate !== undefined && candidate > after) return candidate
  }

  throw new Error('Unable to determine the next schedule occurrence')
}

export const COWORK_SCHEDULE_MINUTE_MS = MINUTE_MS
export const COWORK_SCHEDULE_DAY_MS = DAY_MS
