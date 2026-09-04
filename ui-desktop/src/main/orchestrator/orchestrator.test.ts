import { beforeEach, describe, expect, it, vi } from 'vitest'

const electronApp = vi.hoisted(() => ({
  getPath: vi.fn(() => '/tmp/morpheus-orchestrator-test'),
  isPackaged: false,
  on: vi.fn()
}))

vi.mock('electron', () => ({ app: electronApp }))

import { Orchestrator } from './orchestrator'

const logger = {
  info: vi.fn(),
  warn: vi.fn(),
  error: vi.fn(),
  scope: vi.fn()
}
logger.scope.mockReturnValue(logger)

function instrumentStartup(events: string[]): any {
  const orchestrator: any = new Orchestrator({} as any, vi.fn(), logger as any)
  orchestrator.resetState = vi.fn(async () => events.push('reset'))
  orchestrator.emitStateUpdate = vi.fn(() => events.push('emit'))
  orchestrator.downloadProxyRouter = vi.fn(async () => events.push('download-proxy'))
  orchestrator.startProxyRouter = vi.fn(async () => events.push('start-proxy'))
  orchestrator.downloadOptionalAiRuntime = vi.fn(async () => events.push('download-ai-runtime'))
  orchestrator.downloadOptionalAiModel = vi.fn(async () => events.push('download-ai-model'))
  orchestrator.downloadOptionalIpfs = vi.fn(async () => events.push('download-ipfs'))
  orchestrator.startOptionalService = vi.fn(async (name: string) => events.push(`start-${name}`))
  return orchestrator
}

beforeEach(() => {
  vi.clearAllMocks()
  logger.scope.mockReturnValue(logger)
})

describe('Orchestrator startup order', () => {
  it('starts the required proxy before downloading optional services', async () => {
    const events: string[] = []
    const orchestrator = instrumentStartup(events)

    await orchestrator.startAll()

    const proxyStart = events.indexOf('start-proxy')
    expect(events.indexOf('download-proxy')).toBeLessThan(proxyStart)
    expect(proxyStart).toBeLessThan(events.indexOf('download-ai-runtime'))
    expect(proxyStart).toBeLessThan(events.indexOf('download-ai-model'))
    expect(proxyStart).toBeLessThan(events.indexOf('download-ipfs'))
  })

  it('does not start optional work when the required proxy cannot start', async () => {
    const events: string[] = []
    const orchestrator = instrumentStartup(events)
    orchestrator.startProxyRouter = vi.fn(async () => {
      events.push('start-proxy')
      throw new Error('proxy failed')
    })

    await expect(orchestrator.startAll()).rejects.toThrow('proxy failed')

    expect(events).not.toContain('download-ai-runtime')
    expect(events).not.toContain('download-ai-model')
    expect(events).not.toContain('download-ipfs')
  })
})
