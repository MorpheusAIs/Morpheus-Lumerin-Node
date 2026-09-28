import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterAll, beforeEach, describe, expect, it, vi } from 'vitest'

const electron = vi.hoisted(() => ({ userData: '' }))

vi.mock('electron', () => ({
  app: {
    getPath: (name: string) => {
      if (name !== 'userData') throw new Error(`Unexpected Electron path request: ${name}`)
      return electron.userData
    }
  }
}))

const {
  VISION_PROBE_TTL_MS,
  coworkVisionProbeWritesSettled,
  coworkVisionVerdict,
  loadCoworkVisionProbes,
  recordCoworkVisionProbe,
  resetCoworkVisionProbes
} = await import('./cowork-vision-cache')

const suite = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-vision-cache-'))

beforeEach(async () => {
  electron.userData = await fs.mkdtemp(path.join(suite, 'user-'))
  resetCoworkVisionProbes()
})

afterAll(async () => {
  await fs.rm(suite, { recursive: true, force: true })
})

const probe = (modelId: string, sees: boolean, probedAt = Date.now()) => ({
  modelId,
  sees,
  probedAt,
  answer: sees ? 'red' : 'none'
})

describe('vision probe cache', () => {
  it('survives a restart so a model is probed once, not once per session', async () => {
    await loadCoworkVisionProbes()
    recordCoworkVisionProbe(probe('model-a', true))
    await coworkVisionProbeWritesSettled()

    resetCoworkVisionProbes()
    await loadCoworkVisionProbes()

    expect(coworkVisionVerdict('model-a')).toMatchObject({ sees: true })
    expect(coworkVisionVerdict('model-b')).toBeNull()
  })

  it('forgets a verdict old enough that the provider may have changed the model', async () => {
    await loadCoworkVisionProbes()
    recordCoworkVisionProbe(probe('stale', true, Date.now() - VISION_PROBE_TTL_MS - 1))

    expect(coworkVisionVerdict('stale')).toBeNull()
  })

  it('falls back to no verdict rather than throwing on a corrupt cache file', async () => {
    await fs.writeFile(path.join(electron.userData, 'cowork-vision-probes.json'), '{ not json')

    await expect(loadCoworkVisionProbes()).resolves.toBeUndefined()
    expect(coworkVisionVerdict('anything')).toBeNull()
  })

  it('ignores entries that are not well-formed verdicts', async () => {
    await fs.writeFile(
      path.join(electron.userData, 'cowork-vision-probes.json'),
      JSON.stringify({
        version: 1,
        probes: [
          { modelId: 'good', sees: false, probedAt: Date.now(), answer: 'none' },
          { modelId: '', sees: true, probedAt: Date.now() },
          { modelId: 'no-verdict', probedAt: Date.now() },
          { modelId: 'no-time', sees: true }
        ]
      })
    )

    await loadCoworkVisionProbes()

    expect(coworkVisionVerdict('good')).toMatchObject({ sees: false })
    expect(coworkVisionVerdict('no-verdict')).toBeNull()
    expect(coworkVisionVerdict('no-time')).toBeNull()
  })

  it('serialises concurrent writes into one intact file', async () => {
    await loadCoworkVisionProbes()
    for (let index = 0; index < 25; index++) recordCoworkVisionProbe(probe(`model-${index}`, true))
    await coworkVisionProbeWritesSettled()

    const raw = await fs.readFile(path.join(electron.userData, 'cowork-vision-probes.json'), 'utf8')
    expect(JSON.parse(raw).probes).toHaveLength(25)
  })
})
