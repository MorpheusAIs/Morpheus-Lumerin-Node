import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  fetch: vi.fn(),
  confirm: vi.fn(),
  owner: { isDestroyed: () => false },
  readFileSync: vi.fn(() => 'test-user:test-password')
}))

vi.mock('electron', () => ({
  app: { quit: vi.fn(), relaunch: vi.fn() },
  BrowserWindow: {
    getFocusedWindow: vi.fn(() => mocks.owner),
    getAllWindows: vi.fn(() => [mocks.owner])
  },
  dialog: { showMessageBox: vi.fn() }
}))
vi.mock('../../../sessionConfirmation', () => ({ showSessionConfirmation: mocks.confirm }))
vi.mock('node:fs', () => ({ default: { readFileSync: mocks.readFileSync } }))
vi.mock('../electron-restart', () => ({ default: vi.fn() }))
vi.mock('../database', () => ({ default: { getDb: vi.fn() } }))
vi.mock('../storage', () => ({ default: { persistState: vi.fn() } }))
vi.mock('../auth', () => ({ default: {} }))
vi.mock('../wallet', () => ({ default: {} }))
vi.mock('../settings', () => ({
  setProxyRouterConfig: vi.fn(),
  getProxyRouterConfig: vi.fn(),
  getDefaultCurrencySetting: vi.fn(),
  setDefaultCurrencySetting: vi.fn(),
  getKey: vi.fn(),
  setKey: vi.fn(),
  getFailoverSetting: vi.fn(),
  setFailoverSetting: vi.fn(),
  setPasswordHash: vi.fn()
}))
vi.mock('../../../config', () => ({
  default: {
    chain: { bypassAuth: false, localProxyRouterUrl: 'http://127.0.0.1:8082' },
    isFailoverEnabled: false
  }
}))
vi.mock('../../../orchestrator/orchestrator', () => ({ Orchestrator: class {} }))
vi.mock('../../../logger', () => ({
  default: { error: vi.fn(), info: vi.fn(), verbose: vi.fn(), warn: vi.fn() }
}))
vi.mock('../keys', () => ({ default: {} }))
vi.mock('../wallets', () => ({}))
vi.mock('../attachments', () => ({}))
vi.mock('../../../../../orchestrator.config', () => ({ cfg: {} }))

import { dialog } from 'electron'
import { openSession, resetAuthHeaders, sendMor } from './handlers'

const modelId = `0x${'ab'.repeat(32)}`
const sessionId = `0x${'cd'.repeat(32)}`
const payload = { modelId, duration: 3600, directPayment: false, failover: true }
const response = (body: unknown, status = 200) => ({
  ok: status >= 200 && status < 300,
  status,
  json: vi.fn(async () => body)
})

function mockProxy(sessionResponse = response({ sessionID: sessionId })) {
  mocks.fetch.mockImplementation(async (url: string) => {
    if (url.endsWith('/auth/cookie/path')) {
      return response({ path: '/tmp/morpheus-session-opening-test-cookie' })
    }
    if (url.endsWith(`/blockchain/models/${modelId}/session`)) return sessionResponse
    throw new Error(`Unexpected proxy request in test: ${url}`)
  })
}

const sessionPosts = () => mocks.fetch.mock.calls.filter(([, init]) => init?.method === 'POST')

describe('session opening requires the main-owned in-app confirmation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.confirm.mockReset().mockResolvedValue(false)
    mocks.fetch.mockReset()
    resetAuthHeaders()
    vi.stubGlobal('fetch', mocks.fetch)
  })

  afterEach(() => vi.unstubAllGlobals())

  it('does not contact the router at all when confirmation is declined', async () => {
    await expect(openSession(payload)).rejects.toThrow('Session opening cancelled.')
    expect(mocks.confirm).toHaveBeenCalledWith(mocks.owner, payload)
    expect(mocks.fetch).not.toHaveBeenCalled()
    expect(dialog.showMessageBox).not.toHaveBeenCalled()
  })

  it('waits for approval and submits only the displayed immutable transaction values', async () => {
    let approve!: (value: boolean) => void
    mocks.confirm.mockImplementation(() => new Promise<boolean>((resolve) => (approve = resolve)))
    mockProxy()
    const mutablePayload = { ...payload }
    const pending = openSession(mutablePayload)
    expect(mocks.confirm).toHaveBeenCalledWith(mocks.owner, payload)
    expect(mocks.fetch).not.toHaveBeenCalled()

    // A renderer changing its source object after opening the prompt must not
    // change the transaction that the user is currently reviewing.
    mutablePayload.modelId = `0x${'ef'.repeat(32)}`
    mutablePayload.duration = 7200
    mutablePayload.directPayment = true
    mutablePayload.failover = false
    approve(true)

    await expect(pending).resolves.toEqual({ sessionID: sessionId })
    expect(sessionPosts()).toHaveLength(1)
    const [url, init] = sessionPosts()[0]
    expect(url).toBe(`http://127.0.0.1:8082/blockchain/models/${modelId}/session`)
    expect(JSON.parse(init.body)).toEqual({
      sessionDuration: payload.duration,
      directPayment: payload.directPayment,
      failover: payload.failover,
      rejectExisting: true
    })
    expect(dialog.showMessageBox).not.toHaveBeenCalled()
  })

  it('preserves direct payment explicitly when selected and approved', async () => {
    mocks.confirm.mockResolvedValue(true)
    mockProxy()
    await openSession({ ...payload, directPayment: true, failover: false })
    expect(mocks.confirm).toHaveBeenCalledWith(mocks.owner, {
      ...payload,
      directPayment: true,
      failover: false
    })
    expect(JSON.parse(sessionPosts()[0][1].body)).toMatchObject({
      directPayment: true,
      failover: false
    })
  })

  it.each([
    { ...payload, modelId: '' },
    { ...payload, modelId: 'bad\nmodel' },
    { ...payload, duration: 0 },
    { ...payload, duration: -1 },
    { ...payload, duration: 0.5 },
    { ...payload, duration: Number.NaN },
    { ...payload, duration: 315_360_001 }
  ])('rejects invalid transaction input before asking for approval (%j)', async (input) => {
    await expect(openSession(input)).rejects.toThrow(/invalid/iu)
    expect(mocks.confirm).not.toHaveBeenCalled()
    expect(mocks.fetch).not.toHaveBeenCalled()
  })

  it('fails closed if the confirmation host throws', async () => {
    mocks.confirm.mockRejectedValue(new Error('Another confirmation is already open.'))
    await expect(openSession(payload)).rejects.toThrow('Another confirmation')
    expect(mocks.fetch).not.toHaveBeenCalled()
  })

  it('shares the native financial-confirmation lock and releases it after dismissal', async () => {
    let decline!: (value: boolean) => void
    mocks.confirm.mockImplementation(() => new Promise<boolean>((resolve) => (decline = resolve)))
    const pending = openSession(payload)
    const pendingRejection = expect(pending).rejects.toThrow('Session opening cancelled.')
    const transfer = { to: `0x${'12'.repeat(20)}`, amount: '1000000000000000000' }

    await expect(sendMor(transfer)).rejects.toThrow(
      'Another security confirmation is already open.'
    )
    expect(dialog.showMessageBox).not.toHaveBeenCalled()
    expect(mocks.fetch).not.toHaveBeenCalled()

    decline(false)
    await pendingRejection
    vi.mocked(dialog.showMessageBox).mockResolvedValueOnce({ response: 0, checkboxChecked: false })
    await expect(sendMor(transfer)).rejects.toThrow('Transfer cancelled.')
    expect(dialog.showMessageBox).toHaveBeenCalledOnce()
    expect(mocks.fetch).not.toHaveBeenCalled()
  })

  it('keeps duplicate-session recovery after confirmation unchanged', async () => {
    mocks.confirm.mockResolvedValue(true)
    mockProxy(response({ existingSessionID: sessionId }, 409))
    await expect(openSession(payload)).resolves.toEqual({ existingSessionID: sessionId })
    expect(sessionPosts()).toHaveLength(1)
    expect(JSON.parse(sessionPosts()[0][1].body).rejectExisting).toBe(true)
  })

  it('preserves provider errors rather than reporting a successful opening', async () => {
    mocks.confirm.mockResolvedValue(true)
    mockProxy(response({ error: 'No provider accepted this session.' }, 400))
    await expect(openSession(payload)).rejects.toThrow('No provider accepted this session.')
    expect(sessionPosts()).toHaveLength(1)
  })
})
