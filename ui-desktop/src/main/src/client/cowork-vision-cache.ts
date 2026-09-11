/**
 * Remembered vision-probe verdicts.
 *
 * A probe costs a request, so it runs once per model and the answer is kept.
 * The file is a cache, not a record of anything the user owns: if it is missing
 * or corrupt, every model simply falls back to the guess made from its name,
 * and the next task re-probes.
 */

import { app } from 'electron'
import { promises as fs } from 'node:fs'
import path from 'node:path'
import type { CoworkVisionProbeResult } from './cowork-vision-probe'

/** Providers change what they serve behind a model id, so a verdict expires. */
export const VISION_PROBE_TTL_MS = 30 * 24 * 60 * 60 * 1000
/** Bounded so a wallet with many marketplace models cannot grow the file forever. */
const MAX_CACHED_PROBES = 500

let cache: Map<string, CoworkVisionProbeResult> | null = null
let writeTail: Promise<void> = Promise.resolve()

const cacheFile = (): string => path.join(app.getPath('userData'), 'cowork-vision-probes.json')

const isFresh = (result: CoworkVisionProbeResult, now: number): boolean =>
  now - result.probedAt < VISION_PROBE_TTL_MS

function parseCache(raw: string): Map<string, CoworkVisionProbeResult> {
  const parsed = JSON.parse(raw)
  const entries = Array.isArray(parsed?.probes) ? parsed.probes : []
  const loaded = new Map<string, CoworkVisionProbeResult>()
  for (const entry of entries) {
    if (typeof entry?.modelId !== 'string' || !entry.modelId) continue
    if (typeof entry.sees !== 'boolean' || typeof entry.probedAt !== 'number') continue
    loaded.set(entry.modelId, {
      modelId: entry.modelId,
      sees: entry.sees,
      probedAt: entry.probedAt,
      answer: typeof entry.answer === 'string' ? entry.answer.slice(0, 200) : ''
    })
  }
  return loaded
}

export async function loadCoworkVisionProbes(): Promise<void> {
  if (cache) return
  cache = await fs
    .readFile(cacheFile(), 'utf8')
    .then(parseCache)
    .catch(() => new Map<string, CoworkVisionProbeResult>())
}

/**
 * Synchronous by design: the tool list is assembled inside request building,
 * where an await would put the verdict a turn behind the decision it informs.
 */
export function coworkVisionVerdict(modelId: string): CoworkVisionProbeResult | null {
  const result = cache?.get(modelId)
  return result && isFresh(result, Date.now()) ? result : null
}

export function recordCoworkVisionProbe(result: CoworkVisionProbeResult): void {
  if (!cache) cache = new Map()
  cache.set(result.modelId, result)
  if (cache.size > MAX_CACHED_PROBES) {
    const oldest = [...cache.values()].sort((left, right) => left.probedAt - right.probedAt)
    for (const stale of oldest.slice(0, cache.size - MAX_CACHED_PROBES)) {
      cache.delete(stale.modelId)
    }
  }
  const probes = [...cache.values()]
  // Serialised through a tail so two probes finishing together cannot interleave
  // their writes and leave the file half-written.
  writeTail = writeTail
    .then(() => fs.mkdir(path.dirname(cacheFile()), { recursive: true }))
    .then(() => fs.writeFile(cacheFile(), JSON.stringify({ version: 1, probes }), 'utf8'))
    .catch(() => undefined)
}

/** Test seam: forgets everything loaded, so a suite starts from a known state. */
export function resetCoworkVisionProbes(): void {
  cache = null
  writeTail = Promise.resolve()
}

export const coworkVisionProbeWritesSettled = (): Promise<void> => writeTail
