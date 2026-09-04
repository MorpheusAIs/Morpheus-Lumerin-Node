import { createRequire } from 'node:module'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'

const runtime = vi.hoisted(() => ({
  ipcHandlers: new Map<string, (...args: any[]) => void>(),
  order: [] as string[],
  subscribed: false,
  subscriptionStateAtReady: [] as boolean[],
  stopped: false,
  unsubscribed: false,
  startError: null as Error | null,
  getState: vi.fn(async () => ({ chain: { persisted: true } }))
}))

vi.mock('../../core', () => ({
  default: vi.fn(() => ({
    start: vi.fn(() => {
      if (runtime.startError) throw runtime.startError
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
    getState: runtime.getState
  }
}))

vi.mock('../../rendererTrust', () => ({
  isTrustedRendererEvent: vi.fn(() => true)
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
    runtime.startError = null
    runtime.getState.mockReset()
    runtime.getState.mockResolvedValue({ chain: { persisted: true } })
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

    // Snapshot the state at send-time. Checking only the eventual state would
    // miss the original race because subscribe() did run shortly afterward.
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

  it('coalesces concurrent ui-ready requests and acknowledges each correlation id', async () => {
    let resolveState!: (state: { chain: { persisted: boolean } }) => void
    runtime.getState.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveState = resolve
        })
    )
    createClient({ chain: { chainId: 'base' } })
    const send = vi.fn()
    const event = { sender: { send } }
    const uiReady = runtime.ipcHandlers.get('ui-ready')

    uiReady?.(event, { id: 'ready-request-1' })
    uiReady?.(event, { id: 'ready-request-2' })

    expect(runtime.getState).toHaveBeenCalledTimes(1)
    expect(runtime.order).not.toContain('core-start')

    resolveState({ chain: { persisted: true } })

    await vi.waitFor(() => expect(send).toHaveBeenCalledTimes(2))
    expect(runtime.order.filter((entry) => entry === 'core-start')).toHaveLength(1)
    expect(runtime.order.filter((entry) => entry === 'subscribe')).toHaveLength(1)
    expect(send.mock.calls.map(([, payload]) => payload.id)).toEqual([
      'ready-request-1',
      'ready-request-2'
    ])

    uiReady?.(event, { id: 'ready-request-3' })
    await vi.waitFor(() => expect(send).toHaveBeenCalledTimes(3))

    expect(runtime.getState).toHaveBeenCalledTimes(1)
    expect(runtime.order.filter((entry) => entry === 'core-start')).toHaveLength(1)
    expect(runtime.order.filter((entry) => entry === 'subscribe')).toHaveLength(1)
    expect(send.mock.calls[2][1].id).toBe('ready-request-3')
  })

  it('invalidates in-flight bootstrap on ui-unload without stopping an uninitialized core', async () => {
    let resolveFirstState!: (state: { chain: { persisted: boolean } }) => void
    runtime.getState.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveFirstState = resolve
        })
    )
    createClient({ chain: { chainId: 'base' } })
    const staleSend = vi.fn()
    const uiReady = runtime.ipcHandlers.get('ui-ready')
    const uiUnload = runtime.ipcHandlers.get('ui-unload')

    uiReady?.({ sender: { send: staleSend } }, { id: 'stale-request' })
    uiUnload?.({ sender: {} })
    resolveFirstState({ chain: { persisted: false } })
    await Promise.resolve()
    await Promise.resolve()

    expect(runtime.stopped).toBe(false)
    expect(runtime.unsubscribed).toBe(false)
    expect(runtime.order).not.toContain('core-start')
    expect(staleSend).not.toHaveBeenCalled()

    const freshSend = vi.fn()
    uiReady?.({ sender: { send: freshSend } }, { id: 'fresh-request' })

    await vi.waitFor(() => expect(freshSend).toHaveBeenCalledTimes(1))
    expect(runtime.order.filter((entry) => entry === 'core-start')).toHaveLength(1)
    expect(runtime.order.filter((entry) => entry === 'subscribe')).toHaveLength(1)
  })

  it('reports a bootstrap failure immediately instead of waiting for the renderer timeout', async () => {
    runtime.startError = new Error('wallet core failed to start')
    createClient({ chain: { chainId: 'base' } })
    const send = vi.fn()
    const uiReady = runtime.ipcHandlers.get('ui-ready')

    uiReady?.({ sender: { send } }, { id: 'failed-ready-request' })

    await vi.waitFor(() =>
      expect(send).toHaveBeenCalledWith('ui-ready', {
        id: 'failed-ready-request',
        error: { message: 'wallet core failed to start' }
      })
    )
    expect(runtime.stopped).toBe(false)
    expect(runtime.unsubscribed).toBe(false)
  })
})
