import { createRequire } from 'node:module'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'

const runtime = vi.hoisted(() => ({
  ipcHandlers: new Map<string, (...args: any[]) => void>(),
  order: [] as string[],
  subscribed: false,
  subscriptionStateAtReady: [] as boolean[],
  stopped: false,
  unsubscribed: false
}))

vi.mock('../../core', () => ({
  default: vi.fn(() => ({
    start: vi.fn(() => {
      runtime.order.push('core-start')
      return {
        emitter: {
          setMaxListeners: vi.fn(),
          on: vi.fn()
        },
        events: [],
        api: {}
      }
    }),
    stop: vi.fn(() => {
      runtime.stopped = true
      runtime.order.push('core-stop')
    })
  }))
}))

vi.mock('../../logger', () => ({
  default: {
    verbose: vi.fn(),
    warn: vi.fn(),
    error: vi.fn()
  }
}))

vi.mock('./subscriptions', () => ({
  default: {
    subscribe: vi.fn(() => {
      runtime.subscribed = true
      runtime.order.push('subscribe')
    }),
    unsubscribe: vi.fn(() => {
      runtime.unsubscribed = true
      runtime.order.push('unsubscribe')
    })
  }
}))

vi.mock('./settings', () => ({
  presetDefaults: vi.fn(),
  getPasswordHash: vi.fn(() => 'configured-wallet')
}))

vi.mock('./storage', () => ({
  default: {
    getState: vi.fn(async () => ({ chain: { persisted: true } }))
  }
}))

const nodeRequire = createRequire(import.meta.url)
const electronModulePath = nodeRequire.resolve('electron')
let originalElectronModule: NodeJS.Module | undefined
let createClient: typeof import('./index').createClient

describe('renderer client bootstrap', () => {
  beforeAll(async () => {
    originalElectronModule = nodeRequire.cache[electronModulePath]
    nodeRequire.cache[electronModulePath] = {
      id: electronModulePath,
      path: electronModulePath,
      filename: electronModulePath,
      loaded: true,
      children: [],
      paths: [],
      exports: {
        ipcMain: {
          on: vi.fn((channel: string, handler: (...args: any[]) => void) => {
            runtime.ipcHandlers.set(channel, handler)
          }),
          removeAllListeners: vi.fn()
        }
      }
    } as unknown as NodeJS.Module
    ;({ createClient } = await import('./index'))
  })

  afterAll(() => {
    if (originalElectronModule) {
      nodeRequire.cache[electronModulePath] = originalElectronModule
    } else {
      delete nodeRequire.cache[electronModulePath]
    }
  })

  beforeEach(() => {
    runtime.ipcHandlers.clear()
    runtime.order.length = 0
    runtime.subscribed = false
    runtime.subscriptionStateAtReady.length = 0
    runtime.stopped = false
    runtime.unsubscribed = false
  })

  it('installs follow-up IPC subscriptions before acknowledging ui-ready', async () => {
    const config = {
      chain: {
        chainId: 'base',
        localProxyRouterUrl: 'http://127.0.0.1:8082'
      }
    }
    createClient(config)

    const send = vi.fn((channel: string) => {
      if (channel !== 'ui-ready') return
      runtime.subscriptionStateAtReady.push(runtime.subscribed)
      runtime.order.push('ui-ready')
    })
    const uiReady = runtime.ipcHandlers.get('ui-ready')

    expect(uiReady).toBeTypeOf('function')
    uiReady?.({ sender: { send } }, { id: 'ready-request' })

    await vi.waitFor(() => {
      expect(send).toHaveBeenCalledWith('ui-ready', {
        id: 'ready-request',
        data: {
          onboardingComplete: true,
          persistedState: { chain: { persisted: true } },
          config
        }
      })
    })

    expect(runtime.subscriptionStateAtReady).toEqual([true])
    expect(runtime.order).toEqual(['core-start', 'subscribe', 'ui-ready'])
  })

  it('rolls back the core and listeners when the ready response cannot be delivered', async () => {
    createClient({ chain: { chainId: 'base' } })
    const send = vi.fn(() => {
      runtime.order.push('ui-ready-failed')
      throw new Error('renderer was destroyed')
    })
    const uiReady = runtime.ipcHandlers.get('ui-ready')

    uiReady?.({ sender: { send } }, { id: 'ready-request' })

    await vi.waitFor(() => expect(runtime.stopped).toBe(true))
    expect(runtime.unsubscribed).toBe(true)
    expect(runtime.order).toEqual([
      'core-start',
      'subscribe',
      'ui-ready-failed',
      'unsubscribe',
      'core-stop'
    ])
  })
})
