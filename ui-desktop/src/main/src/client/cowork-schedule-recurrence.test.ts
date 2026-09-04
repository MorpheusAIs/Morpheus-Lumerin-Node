import { describe, expect, it } from 'vitest'
import { computeNextCoworkRunAt, validateCoworkCadence } from './cowork-schedule-recurrence'

describe('Cowork schedule recurrence', () => {
  it('computes hourly, daily, weekly, weekday, and manual cadences in UTC', () => {
    expect(
      computeNextCoworkRunAt({ kind: 'hourly', minute: 15 }, 'UTC', Date.UTC(2026, 0, 5, 10, 15))
    ).toBe(Date.UTC(2026, 0, 5, 11, 15))

    expect(
      computeNextCoworkRunAt(
        { kind: 'daily', hour: 9, minute: 30 },
        'UTC',
        Date.UTC(2026, 0, 5, 9, 29)
      )
    ).toBe(Date.UTC(2026, 0, 5, 9, 30))

    expect(
      computeNextCoworkRunAt(
        { kind: 'weekly', dayOfWeek: 1, hour: 8, minute: 0 },
        'UTC',
        Date.UTC(2026, 0, 5, 8, 0)
      )
    ).toBe(Date.UTC(2026, 0, 12, 8, 0))

    expect(
      computeNextCoworkRunAt(
        { kind: 'weekdays', hour: 9, minute: 0 },
        'UTC',
        Date.UTC(2026, 0, 9, 10, 0)
      )
    ).toBe(Date.UTC(2026, 0, 12, 9, 0))

    expect(computeNextCoworkRunAt({ kind: 'manual' }, 'UTC', Date.now())).toBeUndefined()
  })

  it('interprets persisted wall-clock fields in their explicit time zone', () => {
    expect(
      computeNextCoworkRunAt(
        { kind: 'daily', hour: 9, minute: 30 },
        'Asia/Tokyo',
        Date.UTC(2026, 0, 5, 0, 29)
      )
    ).toBe(Date.UTC(2026, 0, 5, 0, 30))
  })

  it('handles daylight-saving gaps and overlaps deterministically', () => {
    // New York jumps from 01:59 to 03:00 on 2026-03-08. A skipped 02:30
    // occurrence runs once at the first valid later minute.
    expect(
      computeNextCoworkRunAt(
        { kind: 'daily', hour: 2, minute: 30 },
        'America/New_York',
        Date.UTC(2026, 2, 8, 5, 0)
      )
    ).toBe(Date.UTC(2026, 2, 8, 7, 0))

    // New York repeats 01:30 on 2026-11-01. The first occurrence wins, and
    // asking after it advances to the following local day instead of doubling.
    expect(
      computeNextCoworkRunAt(
        { kind: 'daily', hour: 1, minute: 30 },
        'America/New_York',
        Date.UTC(2026, 10, 1, 4, 0)
      )
    ).toBe(Date.UTC(2026, 10, 1, 5, 30))
    expect(
      computeNextCoworkRunAt(
        { kind: 'daily', hour: 1, minute: 30 },
        'America/New_York',
        Date.UTC(2026, 10, 1, 5, 30)
      )
    ).toBe(Date.UTC(2026, 10, 2, 6, 30))
  })

  it('rejects invalid wall-clock values', () => {
    expect(() => validateCoworkCadence({ kind: 'hourly', minute: 60 })).toThrow(/Minute/)
    expect(() =>
      validateCoworkCadence({ kind: 'weekly', dayOfWeek: 7, hour: 9, minute: 0 })
    ).toThrow(/Day of week/)
    expect(() =>
      computeNextCoworkRunAt({ kind: 'daily', hour: 9, minute: 0 }, 'Not/A_Time_Zone', Date.now())
    ).toThrow()
  })
})
