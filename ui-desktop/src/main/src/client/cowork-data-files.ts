import { app } from 'electron'
import { chmodSync, closeSync, mkdirSync, openSync } from 'node:fs'
import path from 'node:path'

/**
 * Cowork transcripts can contain excerpts from connected project files. Keep
 * them outside the disposable cache and private to the current OS account.
 */
export function prepareCoworkDataFile(name: 'projects.db' | 'tasks.db' | 'schedules.db'): string {
  const directory = path.join(app.getPath('userData'), 'Cowork')
  mkdirSync(directory, { recursive: true, mode: 0o700 })
  chmodSync(directory, 0o700)

  const filename = path.join(directory, name)
  const descriptor = openSync(filename, 'a', 0o600)
  closeSync(descriptor)
  chmodSync(filename, 0o600)
  return filename
}
