import { beforeEach, describe, expect, it, vi } from 'vitest'

const runtime = vi.hoisted(() => ({
  listeners: new Map<string, (...args: any[]) => void>(),
  trusted: true,
  warn: vi.fn()
}))

vi.mock('electron', () => ({
  ipcMain: {
    on: vi.fn((channel: string, listener: (...args: any[]) => void) => {
      runtime.listeners.set(channel, listener)
    }),
    removeAllListeners: vi.fn()
  }
}))

vi.mock('../../../logger', () => ({ default: { warn: runtime.warn } }))
vi.mock('../../../rendererTrust', () => ({
  isTrustedRendererEvent: vi.fn(() => runtime.trusted)
}))

import { getLogData, onRendererEvent } from './utils'

const rendererEvent = () => ({
  sender: {
    isDestroyed: vi.fn(() => false),
    send: vi.fn()
  }
})

describe('legacy main-process IPC validation', () => {
  beforeEach(() => {
    runtime.listeners.clear()
    runtime.trusted = true
    runtime.warn.mockReset()
  })

  it('redacts nested secrets without mutating the source payload', () => {
    const payload = {
      profile: { password: 'secret', mnemonic: 'words', displayName: 'agent' },
      authorization: 'Basic secret',
      tokenAddress: 'sensitive-by-policy'
    }

    expect(JSON.parse(getLogData(payload))).toEqual({
      profile: { password: '[redacted]', mnemonic: '[redacted]', displayName: 'agent' },
      authorization: '[redacted]',
      tokenAddress: '[redacted]'
    })
    expect(payload.profile.password).toBe('secret')
  })

  it('rejects untrusted and malformed renderer messages before invoking a handler', async () => {
    const handler = vi.fn(() => true)
    onRendererEvent('secure-operation', handler, 'none')
    const listener = runtime.listeners.get('secure-operation')!

    runtime.trusted = false
    listener(rendererEvent(), { id: 'trusted-shape', data: {} })
    runtime.trusted = true
    listener(rendererEvent(), null)
    listener(rendererEvent(), { id: 'x'.repeat(201), data: {} })
    await Promise.resolve()

    expect(handler).not.toHaveBeenCalled()
    expect(runtime.warn).toHaveBeenCalledTimes(3)
  })

  it('normalizes synchronous handlers into bounded IPC responses', async () => {
    const handler = vi.fn(() => ({ ok: true }))
    onRendererEvent('secure-operation', handler, 'none')
    const listener = runtime.listeners.get('secure-operation')!
    const event = rendererEvent()

    listener(event, { id: 'request-1', data: { value: 7 } })
    await vi.waitFor(() => expect(event.sender.send).toHaveBeenCalled())

    expect(handler).toHaveBeenCalledWith({ value: 7 })
    expect(event.sender.send).toHaveBeenCalledWith('secure-operation', {
      id: 'request-1',
      data: { ok: true }
    })
  })
})
