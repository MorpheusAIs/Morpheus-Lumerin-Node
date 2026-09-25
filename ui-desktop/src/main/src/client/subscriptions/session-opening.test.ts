import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  fetch: vi.fn(),
  confirm: vi.fn(),
  dismissal: vi.fn(() => ''),
  owner: { isDestroyed: () => false, isVisible: () => true, isMinimized: () => false },
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
vi.mock('../../../sessionConfirmation', () => ({
  showSessionConfirmation: mocks.confirm,
  consumeSessionConfirmationDismissal: mocks.dismissal
}))
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

const stakeWei = '123456789012345678901'

function mockProxy(sessionResponse = response({ sessionID: sessionId })) {
  mocks.fetch.mockImplementation(async (url: string) => {
    if (url.endsWith('/auth/cookie/path')) {
      return response({ path: '/tmp/morpheus-session-opening-test-cookie' })
    }
    // The confirmation window quotes the open before showing it, so the
    // estimate is a GET the router answers on the way to the dialog.
    if (url.includes('/session/estimate')) return response({ stake_wei: stakeWei })
    if (url.endsWith(`/blockchain/models/${modelId}/session`)) return sessionResponse
    throw new Error(`Unexpected proxy request in test: ${url}`)
  })
}

const sessionPosts = () => mocks.fetch.mock.calls.filter(([, init]) => init?.method === 'POST')

/**
 * The assertion that matters before approval is that nothing has been
 * *submitted*, not that nothing has been sent. Opening the confirmation window
 * now involves reads — the auth cookie path and the stake quote — and a
 * blanket "fetch was never called" would fail on those while saying nothing
 * about whether a transaction escaped.
 */
const expectNothingSubmitted = () => {
  expect(sessionPosts()).toHaveLength(0)
}

describe('session opening requires the main-owned in-app confirmation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.confirm.mockReset().mockResolvedValue(false)
    mocks.dismissal.mockReset().mockReturnValue('')
    mocks.fetch.mockReset()
    resetAuthHeaders()
    vi.stubGlobal('fetch', mocks.fetch)
  })

  afterEach(() => vi.unstubAllGlobals())

  it('submits no transaction when confirmation is declined', async () => {
    mockProxy()
    await expect(openSession(payload)).rejects.toThrow('Session opening cancelled.')
    expect(mocks.confirm).toHaveBeenCalledWith(mocks.owner, expect.objectContaining(payload))
    expectNothingSubmitted()
    expect(dialog.showMessageBox).not.toHaveBeenCalled()
  })

  it('tells the user why a prompt they never answered gave up', async () => {
    // A dismissal and a deliberate Cancel used to read identically, so a report
    // of "it cancels every time" could not be told apart from someone cancelling
    // every time. The reason belongs in the message the user actually sees.
    mockProxy()
    mocks.dismissal.mockReturnValue('the prompt document failed to load (ERR_ABORTED -3)')
    await expect(openSession(payload)).rejects.toThrow(/ERR_ABORTED/u)
    expectNothingSubmitted()
  })

  it('names the amount in the window that authorises it', async () => {
    mocks.confirm.mockResolvedValue(true)
    mockProxy()
    await openSession(payload)

    // The renderer shows a figure, but this window is the one that authorises
    // the transaction, so it quotes the router itself for the same duration the
    // request carries. A renderer bug cannot get a different amount approved
    // than the one on screen here.
    const [, details] = mocks.confirm.mock.calls[0]
    expect(details.amountMor).toBe('123.4567')

    const estimateCall = mocks.fetch.mock.calls.find((call) =>
      String(call[0]).includes('/session/estimate')
    )
    expect(String(estimateCall?.[0])).toContain(`sessionDuration=${payload.duration}`)
  })

  it('still asks for approval when the amount cannot be quoted', async () => {
    mocks.confirm.mockResolvedValue(true)
    mocks.fetch.mockImplementation(async (url: string) => {
      if (url.endsWith('/auth/cookie/path')) {
        return response({ path: '/tmp/morpheus-session-opening-test-cookie' })
      }
      if (url.includes('/session/estimate')) return response({ error: 'nope' }, 500)
      if (url.endsWith(`/blockchain/models/${modelId}/session`)) {
        return response({ sessionID: sessionId })
      }
      throw new Error(`Unexpected proxy request in test: ${url}`)
    })

    // A quote that fails is a reason to say the amount is unknown, not to block
    // an open the user explicitly asked for. The decision stays with them.
    await expect(openSession(payload)).resolves.toEqual({ sessionID: sessionId })
    expect(mocks.confirm.mock.calls[0][1].amountMor).toBeUndefined()
  })

  it('waits for approval and submits only the displayed immutable transaction values', async () => {
    let approve!: (value: boolean) => void
    mocks.confirm.mockImplementation(() => new Promise<boolean>((resolve) => (approve = resolve)))
    mockProxy()
    const mutablePayload = { ...payload }
    const pending = openSession(mutablePayload)
    await vi.waitFor(() => expect(mocks.confirm).toHaveBeenCalled())
    expect(mocks.confirm).toHaveBeenCalledWith(mocks.owner, expect.objectContaining(payload))
    expectNothingSubmitted()

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
    expect(mocks.confirm).toHaveBeenCalledWith(
      mocks.owner,
      expect.objectContaining({ ...payload, directPayment: true, failover: false })
    )
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
    mockProxy()
    mocks.confirm.mockRejectedValue(new Error('Another confirmation is already open.'))
    await expect(openSession(payload)).rejects.toThrow('Another confirmation')
    expectNothingSubmitted()
  })

  it('shares the native financial-confirmation lock and releases it after dismissal', async () => {
    let decline!: (value: boolean) => void
    mockProxy()
    mocks.confirm.mockImplementation(() => new Promise<boolean>((resolve) => (decline = resolve)))
    const pending = openSession(payload)
    const pendingRejection = expect(pending).rejects.toThrow('Session opening cancelled.')
    const transfer = { to: `0x${'12'.repeat(20)}`, amount: '1000000000000000000' }
    await vi.waitFor(() => expect(mocks.confirm).toHaveBeenCalled())

    await expect(sendMor(transfer)).rejects.toThrow(
      'Another security confirmation is already open.'
    )
    expect(dialog.showMessageBox).not.toHaveBeenCalled()
    expectNothingSubmitted()

    decline(false)
    await pendingRejection
    vi.mocked(dialog.showMessageBox).mockResolvedValueOnce({ response: 0, checkboxChecked: false })
    await expect(sendMor(transfer)).rejects.toThrow('Transfer cancelled.')
    expect(dialog.showMessageBox).toHaveBeenCalledOnce()
    expectNothingSubmitted()
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
