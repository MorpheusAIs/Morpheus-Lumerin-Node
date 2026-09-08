import { EventEmitter } from 'node:events'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  views: [] as any[],
  constructorError: null as Error | null,
  loadError: null as Error | null,
  setupError: null as Error | null,
  send: vi.fn(),
  expose: vi.fn()
}))

vi.mock('electron', async () => {
  const { EventEmitter } = await import('node:events')
  const makeSession = () => ({
    setPermissionRequestHandler: vi.fn(() => {
      if (mocks.setupError) throw mocks.setupError
    }),
    setPermissionCheckHandler: vi.fn(),
    webRequest: { onBeforeRequest: vi.fn() }
  })
  const isolatedSession = makeSession()
  return {
    ipcMain: new EventEmitter(),
    ipcRenderer: { send: mocks.send },
    contextBridge: { exposeInMainWorld: mocks.expose },
    session: { fromPartition: vi.fn(() => isolatedSession) },
    BrowserWindow: vi.fn(function (options) {
      if (mocks.constructorError) throw mocks.constructorError
      let destroyed = false
      let currentUrl = ''
      const webContents = Object.assign(new EventEmitter(), {
        mainFrame: {},
        session: isolatedSession,
        isDestroyed: vi.fn(() => destroyed),
        focus: vi.fn(),
        getURL: vi.fn(() => currentUrl),
        setWindowOpenHandler: vi.fn(),
        setIgnoreMenuShortcuts: vi.fn(),
        loadURL: vi.fn((url: string) => {
          currentUrl = url
          return mocks.loadError ? Promise.reject(mocks.loadError) : Promise.resolve()
        }),
        close: vi.fn(() => {
          destroyed = true
          webContents.emit('destroyed')
        })
      })
      const view = Object.assign(new EventEmitter(), {
        options,
        webContents,
        setBounds: vi.fn(),
        setMenu: vi.fn(),
        show: vi.fn(),
        isDestroyed: vi.fn(() => destroyed),
        destroy: vi.fn(() => {
          destroyed = true
          webContents.emit('destroyed')
          view.emit('closed')
        })
      })
      mocks.views.push(view)
      return view
    })
  }
})

import { BrowserWindow, ipcMain } from 'electron'
import {
  SESSION_CONFIRMATION_CHANNEL,
  SESSION_CONFIRMATION_TIMEOUT_MS,
  showSessionConfirmation
} from './sessionConfirmation'
import { renderSessionConfirmationHtml } from './sessionConfirmationView'

const details = Object.freeze({
  modelId: `0x${'ab'.repeat(32)}`,
  duration: 3600,
  directPayment: false,
  failover: true
})

const owners: any[] = []

function makeOwner() {
  const webContents = Object.assign(new EventEmitter(), {
    isDestroyed: vi.fn(() => false),
    focus: vi.fn(),
    getURL: vi.fn(() => 'file:///app/out/renderer/index.html')
  })
  const owner = Object.assign(new EventEmitter(), {
    webContents,
    isDestroyed: vi.fn(() => false),
    isMinimized: vi.fn(() => false),
    getContentSize: vi.fn(() => [1200, 800]),
    getContentBounds: vi.fn(() => ({ x: 0, y: 0, width: 1200, height: 800 })),
    focus: vi.fn(),
    restore: vi.fn()
  })
  owners.push(owner)
  return owner
}

function request(owner = makeOwner()) {
  const promise = showSessionConfirmation(owner as unknown as BrowserWindow, details)
  const view = mocks.views.at(-1)
  const argument = view.options.webPreferences.additionalArguments.find((value: string) =>
    value.startsWith('--session-confirmation-id=')
  )
  const requestId = argument.slice('--session-confirmation-id='.length)
  return { promise, owner, view, requestId }
}

function respond(pending: ReturnType<typeof request>, approved = false) {
  ipcMain.emit(
    SESSION_CONFIRMATION_CHANNEL,
    {
      sender: pending.view.webContents,
      senderFrame: pending.view.webContents.mainFrame
    },
    { requestId: pending.requestId, approved }
  )
}

function expectCleaned(pending: ReturnType<typeof request>) {
  expect(ipcMain.listenerCount(SESSION_CONFIRMATION_CHANNEL)).toBe(0)
  expect(pending.view.destroy).toHaveBeenCalled()
  expect(pending.owner.listenerCount('resize')).toBe(0)
  expect(pending.owner.listenerCount('move')).toBe(0)
  expect(pending.owner.listenerCount('closed')).toBe(0)
  expect(pending.owner.webContents.listenerCount('did-start-navigation')).toBe(0)
  expect(pending.owner.webContents.listenerCount('render-process-gone')).toBe(0)
  expect(vi.getTimerCount()).toBe(0)
}

describe('main-owned in-app session confirmation', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    mocks.views.length = 0
    mocks.constructorError = null
    mocks.loadError = null
    mocks.setupError = null
    owners.length = 0
  })

  afterEach(async () => {
    for (const owner of owners) owner.emit('closed')
    await vi.runAllTimersAsync()
    ipcMain.removeAllListeners()
    vi.useRealTimers()
  })

  it.each([null, undefined])(
    'fails closed when there is no owning app window (%j)',
    async (owner) => {
      await expect(showSessionConfirmation(owner, details)).resolves.toBe(false)
      expect(BrowserWindow).not.toHaveBeenCalled()
    }
  )

  it('fails closed for a destroyed owner without creating an overlay', async () => {
    const owner = makeOwner()
    owner.isDestroyed.mockReturnValue(true)
    await expect(showSessionConfirmation(owner as unknown as BrowserWindow, details)).resolves.toBe(
      false
    )
    expect(BrowserWindow).not.toHaveBeenCalled()
  })

  it('fails closed if the owning renderer is already destroyed', async () => {
    const owner = makeOwner()
    owner.webContents.isDestroyed.mockReturnValue(true)
    await expect(showSessionConfirmation(owner as unknown as BrowserWindow, details)).resolves.toBe(
      false
    )
    expect(BrowserWindow).not.toHaveBeenCalled()
  })

  it('releases the active-request lock if confirmation window construction fails', async () => {
    mocks.constructorError = new Error('Could not construct view')
    await expect(
      showSessionConfirmation(makeOwner() as unknown as BrowserWindow, details)
    ).rejects.toThrow('Could not construct view')
    expect(ipcMain.listenerCount(SESSION_CONFIRMATION_CHANNEL)).toBe(0)
    mocks.constructorError = null
    const next = request()
    respond(next)
    await expect(next.promise).resolves.toBe(false)
  })

  it('creates a main-owned frameless modal overlay with only its dedicated preload', async () => {
    const pending = request()
    expect(pending.view.options).toMatchObject({
      parent: pending.owner,
      modal: true,
      frame: false,
      transparent: true,
      show: false,
      resizable: false,
      movable: false,
      minimizable: false,
      maximizable: false,
      fullscreenable: false,
      skipTaskbar: true,
      hasShadow: false
    })
    expect(pending.view.options.webPreferences).toMatchObject({
      sandbox: true,
      contextIsolation: true,
      nodeIntegration: false,
      devTools: false,
      preload: expect.stringMatching(/preload[/\\]session-confirmation\.js$/u)
    })
    expect(pending.requestId).toMatch(/^[0-9a-f-]{36}$/u)
    expect(pending.view.show).not.toHaveBeenCalled()
    expect(pending.view.setBounds).toHaveBeenCalledWith({
      x: 0,
      y: 0,
      width: 1200,
      height: 800
    })
    await Promise.resolve()
    expect(pending.view.show).toHaveBeenCalledOnce()
    respond(pending)
    await expect(pending.promise).resolves.toBe(false)
    expectCleaned(pending)
  })

  it('rejects concurrent confirmations without replacing the current request', async () => {
    const pending = request()
    await expect(
      showSessionConfirmation(pending.owner as unknown as BrowserWindow, details)
    ).rejects.toThrow(/already open|another|confirmation/iu)
    expect(mocks.views).toHaveLength(1)
    respond(pending, true)
    await expect(pending.promise).resolves.toBe(true)
    expectCleaned(pending)
  })

  it('accepts only the current request from the exact overlay main frame', async () => {
    const pending = request()
    let settled = false
    void pending.promise.then(() => (settled = true))
    const trustedEvent = {
      sender: pending.view.webContents,
      senderFrame: pending.view.webContents.mainFrame
    }
    const validPayload = { requestId: pending.requestId, approved: true }

    for (const [event, payload] of [
      [{ sender: pending.owner.webContents, senderFrame: {} }, validPayload],
      [
        { ...trustedEvent, senderFrame: { parent: pending.view.webContents.mainFrame } },
        validPayload
      ],
      [{ ...trustedEvent, senderFrame: undefined }, validPayload],
      [trustedEvent, { ...validPayload, requestId: 'stale-request' }],
      [trustedEvent, { ...validPayload, approved: 'true' }],
      [trustedEvent, { ...validPayload, approved: 1 }],
      [trustedEvent, null],
      [trustedEvent, undefined],
      [trustedEvent, []]
    ]) {
      ipcMain.emit(SESSION_CONFIRMATION_CHANNEL, event, payload)
    }
    await Promise.resolve()
    expect(settled).toBe(false)

    respond(pending, true)
    await expect(pending.promise).resolves.toBe(true)
    expectCleaned(pending)
  })

  it('does not let a completed or replayed approval approve the next prompt', async () => {
    const first = request()
    respond(first, true)
    await expect(first.promise).resolves.toBe(true)
    const second = request()
    let settled = false
    void second.promise.then(() => (settled = true))

    respond(first, true)
    ipcMain.emit(
      SESSION_CONFIRMATION_CHANNEL,
      { sender: second.view.webContents, senderFrame: second.view.webContents.mainFrame },
      { requestId: first.requestId, approved: true }
    )
    await Promise.resolve()
    expect(settled).toBe(false)
    expect(first.requestId).not.toBe(second.requestId)

    respond(second)
    await expect(second.promise).resolves.toBe(false)
    expectCleaned(second)
  })

  it('does not accept an approval from the overlay while its expected document is absent', async () => {
    const pending = request()
    let settled = false
    void pending.promise.then(() => (settled = true))
    pending.view.webContents.getURL.mockReturnValueOnce('about:blank')
    respond(pending, true)
    await Promise.resolve()
    expect(settled).toBe(false)
    respond(pending)
    await expect(pending.promise).resolves.toBe(false)
  })

  it('expires before the renderer request timeout and releases its confirmation lock', async () => {
    expect(SESSION_CONFIRMATION_TIMEOUT_MS).toBeLessThan(120_000)
    const pending = request()
    await vi.advanceTimersByTimeAsync(SESSION_CONFIRMATION_TIMEOUT_MS)
    await expect(pending.promise).resolves.toBe(false)
    expectCleaned(pending)
    const next = request()
    respond(next)
    await expect(next.promise).resolves.toBe(false)
  })

  it.each([
    ['owner', 'closed'],
    ['ownerContents', 'destroyed'],
    ['ownerContents', 'render-process-gone'],
    ['viewContents', 'destroyed'],
    ['viewContents', 'render-process-gone']
  ])('cancels on %s %s', async (target, event) => {
    const pending = request()
    const emitter =
      target === 'owner'
        ? pending.owner
        : target === 'ownerContents'
          ? pending.owner.webContents
          : pending.view.webContents
    emitter.emit(event, {}, { reason: 'crashed' })
    await expect(pending.promise).resolves.toBe(false)
    expectCleaned(pending)
  })

  it('cancels on owner main-frame reload/navigation but not a child-frame navigation', async () => {
    const pending = request()
    let settled = false
    void pending.promise.then(() => (settled = true))
    pending.owner.webContents.emit('did-start-navigation', {}, 'about:blank', false, false)
    await Promise.resolve()
    expect(settled).toBe(false)
    pending.owner.webContents.emit('did-start-navigation', {}, 'about:blank', false, true)
    await expect(pending.promise).resolves.toBe(false)
    expectCleaned(pending)
  })

  it('cancels on overlay document load failure', async () => {
    const pending = request()
    pending.view.webContents.emit('did-fail-load', {}, -2, 'failed', 'data:text/html,test', true)
    await expect(pending.promise).resolves.toBe(false)
    expectCleaned(pending)
  })

  it('cleans up if loading the confirmation rejects', async () => {
    mocks.loadError = new Error('Failed to load confirmation')
    const pending = request()
    await expect(pending.promise).resolves.toBe(false)
    expectCleaned(pending)
  })

  it('destroys the hidden confirmation window and releases its lock when security setup throws', async () => {
    mocks.setupError = new Error('The session was destroyed during setup')
    const pending = request()
    await expect(pending.promise).resolves.toBe(false)
    expect(pending.view.show).not.toHaveBeenCalled()
    expect(pending.view.destroy).toHaveBeenCalled()
    expect(ipcMain.listenerCount(SESSION_CONFIRMATION_CHANNEL)).toBe(0)
    expect(pending.owner.listenerCount('resize')).toBe(0)
    expect(vi.getTimerCount()).toBe(0)

    mocks.setupError = null
    const next = request()
    respond(next, true)
    await expect(next.promise).resolves.toBe(true)
  })

  it.each(['show', 'focus'])(
    'fails closed and cleans up when the post-load %s operation throws',
    async (step) => {
      const pending = request()
      const failingOperation = step === 'show' ? pending.view.show : pending.view.webContents.focus
      failingOperation.mockImplementationOnce(() => {
        throw new Error('Native confirmation window disappeared after loading')
      })

      await expect(pending.promise).resolves.toBe(false)
      expectCleaned(pending)

      const next = request()
      respond(next)
      await expect(next.promise).resolves.toBe(false)
    }
  )

  it.each(['destroy', 'focus'])(
    'fails closed but finishes all cleanup when native %s throws',
    async (step) => {
      const pending = request()
      const failingOperation =
        step === 'destroy' ? pending.view.destroy : pending.owner.webContents.focus
      failingOperation.mockImplementationOnce(() => {
        throw new Error('Native object disappeared during teardown')
      })
      respond(pending, true)
      await expect(pending.promise).resolves.toBe(false)
      expectCleaned(pending)
      expect(pending.owner.webContents.focus).toHaveBeenCalled()

      const next = request()
      respond(next)
      await expect(next.promise).resolves.toBe(false)
    }
  )

  it('rejects an otherwise valid approval if the owner died before its close event arrives', async () => {
    const pending = request()
    pending.owner.webContents.isDestroyed.mockReturnValue(true)
    respond(pending, true)
    await expect(pending.promise).resolves.toBe(false)
    expectCleaned(pending)
    expect(pending.owner.webContents.focus).not.toHaveBeenCalled()
  })

  it('resizes with the owner content area and restores focus after dismissal', async () => {
    const pending = request()
    pending.owner.getContentSize.mockReturnValue([390, 700])
    pending.owner.getContentBounds.mockReturnValue({ x: 0, y: 0, width: 390, height: 700 })
    pending.owner.emit('resize')
    expect(pending.view.setBounds).toHaveBeenLastCalledWith({
      x: 0,
      y: 0,
      width: 390,
      height: 700
    })
    respond(pending)
    await expect(pending.promise).resolves.toBe(false)
    expect(pending.owner.webContents.focus).toHaveBeenCalled()
    expectCleaned(pending)
  })

  it('moves with the owning window using absolute content-area bounds', async () => {
    const pending = request()
    pending.owner.getContentBounds.mockReturnValue({ x: 175, y: 120, width: 1200, height: 800 })
    pending.owner.emit('move')
    expect(pending.view.setBounds).toHaveBeenLastCalledWith({
      x: 175,
      y: 120,
      width: 1200,
      height: 800
    })
    respond(pending)
    await expect(pending.promise).resolves.toBe(false)
    expectCleaned(pending)
  })

  it('destroys its child without restoring focus to an owner that was destroyed', async () => {
    const pending = request()
    pending.owner.isDestroyed.mockReturnValue(true)
    pending.owner.emit('closed')
    await expect(pending.promise).resolves.toBe(false)
    expect(pending.owner.webContents.focus).not.toHaveBeenCalled()
    expect(pending.view.destroy).toHaveBeenCalled()
    expect(ipcMain.listenerCount(SESSION_CONFIRMATION_CHANNEL)).toBe(0)
  })

  it('blocks overlay navigation, new windows, network access and permissions', async () => {
    const pending = request()
    const contents = pending.view.webContents
    const preventDefault = vi.fn()
    contents.emit('will-navigate', { preventDefault }, 'https://attacker.example')
    expect(preventDefault).toHaveBeenCalled()
    expect(contents.setWindowOpenHandler).toHaveBeenCalled()
    const windowHandler = contents.setWindowOpenHandler.mock.calls[0][0]
    expect(windowHandler({ url: 'https://attacker.example' })).toEqual({ action: 'deny' })
    const isolated = contents.session
    expect(isolated.setPermissionRequestHandler).toHaveBeenCalled()
    const permissionCallback = vi.fn()
    isolated.setPermissionRequestHandler.mock.calls[0][0](contents, 'media', permissionCallback, {})
    expect(permissionCallback).toHaveBeenCalledWith(false)
    expect(isolated.setPermissionCheckHandler.mock.calls[0][0]()).toBe(false)
    expect(isolated.webRequest.onBeforeRequest).toHaveBeenCalled()
    const requestHandler = isolated.webRequest.onBeforeRequest.mock.calls[0].at(-1)
    const networkCallback = vi.fn()
    requestHandler({ url: 'https://attacker.example' }, networkCallback)
    expect(networkCallback).toHaveBeenCalledWith({ cancel: true })
    respond(pending)
    await expect(pending.promise).resolves.toBe(false)
  })
})

describe('isolated session-confirmation document', () => {
  it('renders reviewed fields as text rather than executable HTML', () => {
    const modelId = '</code><script>window.bad=true</script><img src=x onerror="bad()">&\'"'
    const html = renderSessionConfirmationHtml({ ...details, modelId }, 'safe-test-nonce')
    const document = new DOMParser().parseFromString(html, 'text/html')
    expect(document.querySelector('code')?.textContent).toBe(modelId)
    expect(document.querySelector('script')).toBeNull()
    expect(document.querySelector('img')).toBeNull()
    expect(document.querySelector('code')?.children).toHaveLength(0)
    expect(document.querySelector('style')?.nonce).toBe('safe-test-nonce')
    expect(
      document.querySelector<HTMLMetaElement>('meta[http-equiv="Content-Security-Policy"]')?.content
    ).toContain("script-src 'none'")
  })

  it('includes accessible dialog semantics, a default cancel action and exact transaction fields', () => {
    const html = renderSessionConfirmationHtml(details, 'safe-test-nonce')
    const document = new DOMParser().parseFromString(html, 'text/html')
    const dialog = document.querySelector('[role="dialog"]')
    expect(dialog?.getAttribute('aria-modal')).toBe('true')
    expect(document.getElementById(dialog!.getAttribute('aria-labelledby')!)?.textContent).toBe(
      'Open session'
    )
    expect(document.getElementById('session-cancel')?.hasAttribute('autofocus')).toBe(true)
    expect(document.querySelector('button[data-decision="approve"]')?.textContent).toContain(
      'Open session'
    )
    expect(document.body.textContent).toContain(details.modelId)
    expect(document.body.textContent).toContain('3,600 seconds')
    expect(document.body.textContent).toContain('Stake MOR · escrow')
    expect(document.body.textContent).toContain('Unused stake returns when the session closes.')
    expect(document.body.textContent).toContain('Enabled')
  })

  it('clearly distinguishes direct payment from refundable stake', () => {
    const html = renderSessionConfirmationHtml(
      { ...details, directPayment: true, failover: false },
      'safe-test-nonce'
    )
    const document = new DOMParser().parseFromString(html, 'text/html')
    expect(document.body.textContent).toContain('Direct MOR payment')
    expect(document.body.textContent).toContain('Disabled')
    expect(document.body.textContent).not.toContain('Unused stake returns')
  })
})

describe('dedicated confirmation preload', () => {
  const requestId = 'e1730217-a7cb-4d3f-b883-56ee79bcb91c'
  let originalArgv: string[]
  let documentListeners: ReturnType<typeof vi.spyOn>
  let windowListeners: ReturnType<typeof vi.spyOn>

  async function loadPreload(argument = `--session-confirmation-id=${requestId}`) {
    vi.resetModules()
    process.argv = ['electron', argument]
    await import('../preload/session-confirmation')
  }

  function invokeTrustedListener(type: 'click' | 'keydown', properties: Record<string, unknown>) {
    const match = documentListeners.mock.calls.find(([eventName]) => eventName === type)
    expect(match).toBeDefined()
    const listener = match![1] as EventListener
    const preventDefault = vi.fn()
    listener({ isTrusted: true, preventDefault, ...properties } as unknown as Event)
    return preventDefault
  }

  beforeEach(() => {
    originalArgv = process.argv
    mocks.send.mockClear()
    mocks.expose.mockClear()
    const html = renderSessionConfirmationHtml(details, 'safe-test-nonce')
    document.body.innerHTML = new DOMParser().parseFromString(html, 'text/html').body.innerHTML
    documentListeners = vi.spyOn(document, 'addEventListener')
    windowListeners = vi.spyOn(window, 'addEventListener')
  })

  afterEach(() => {
    for (const [name, listener, options] of documentListeners.mock.calls) {
      document.removeEventListener(name as string, listener as EventListener, options as boolean)
    }
    for (const [name, listener, options] of windowListeners.mock.calls) {
      window.removeEventListener(name as string, listener as EventListener, options as boolean)
    }
    documentListeners.mockRestore()
    windowListeners.mockRestore()
    process.argv = originalArgv
    document.body.innerHTML = ''
  })

  it('exposes no API bridge and puts initial focus on Cancel', async () => {
    await loadPreload()
    window.dispatchEvent(new Event('DOMContentLoaded'))
    expect(mocks.expose).not.toHaveBeenCalled()
    expect(mocks.send).not.toHaveBeenCalled()
    expect(document.activeElement).toBe(document.getElementById('session-cancel'))
  })

  it('does not accept synthetic DOM clicks or synthetic escape keys', async () => {
    await loadPreload()
    const approve = document.querySelector<HTMLButtonElement>('button[data-decision="approve"]')!
    approve.click()
    approve.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    expect(mocks.send).not.toHaveBeenCalled()
    expect(approve.disabled).toBe(false)
  })

  it('sends one nonce-bound decision from a trusted control click and disables all actions', async () => {
    await loadPreload()
    const approve = document.querySelector<HTMLButtonElement>('button[data-decision="approve"]')!
    invokeTrustedListener('click', { target: approve.querySelector('span') })
    expect(mocks.send).toHaveBeenCalledOnce()
    expect(mocks.send).toHaveBeenCalledWith(SESSION_CONFIRMATION_CHANNEL, {
      requestId,
      approved: true
    })
    expect(
      [...document.querySelectorAll<HTMLButtonElement>('button')].every((button) => button.disabled)
    ).toBe(true)
    invokeTrustedListener('click', { target: approve })
    invokeTrustedListener('keydown', { key: 'Escape' })
    expect(mocks.send).toHaveBeenCalledOnce()
  })

  it('cancels on a trusted Escape key without approving', async () => {
    await loadPreload()
    const preventDefault = invokeTrustedListener('keydown', { key: 'Escape' })
    expect(preventDefault).toHaveBeenCalledOnce()
    expect(mocks.send).toHaveBeenCalledWith(SESSION_CONFIRMATION_CHANNEL, {
      requestId,
      approved: false
    })
  })

  it('keeps Tab and Shift+Tab inside the confirmation controls', async () => {
    await loadPreload()
    const buttons = [...document.querySelectorAll<HTMLButtonElement>('button')]
    const first = buttons[0]
    const last = buttons.at(-1)!
    last.focus()
    expect(invokeTrustedListener('keydown', { key: 'Tab', shiftKey: false })).toHaveBeenCalled()
    expect(document.activeElement).toBe(first)
    expect(invokeTrustedListener('keydown', { key: 'Tab', shiftKey: true })).toHaveBeenCalled()
    expect(document.activeElement).toBe(last)
    expect(mocks.send).not.toHaveBeenCalled()
  })

  it.each(['--unrelated-argument', '--session-confirmation-id=not-a-valid-request'])(
    'fails closed without a valid main-owned request token (%s)',
    async (argument) => {
      await loadPreload(argument)
      invokeTrustedListener('click', {
        target: document.querySelector('button[data-decision="approve"]')
      })
      invokeTrustedListener('keydown', { key: 'Escape' })
      expect(mocks.send).not.toHaveBeenCalled()
    }
  )
})
