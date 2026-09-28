import { prepareCoworkDataFile } from './cowork-data-files'
import { createIndexedCoworkDatastore } from './cowork-datastore'

let schedules: any

/**
 * Schedules are durable user data, so they intentionally live beside (but not
 * inside) the legacy disposable cache database.
 */
export function coworkScheduleCollection(): any {
  if (schedules) return schedules

  const filename = prepareCoworkDataFile('schedules.db')
  // Schedules are repeatedly selected by ID and listed/counted per project.
  schedules = createIndexedCoworkDatastore(filename, ['id', 'projectId'], 60_000)
  return schedules
}
