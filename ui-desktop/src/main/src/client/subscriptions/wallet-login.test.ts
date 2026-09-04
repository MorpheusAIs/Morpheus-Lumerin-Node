import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  emit: vi.fn(),
  fetch: vi.fn(),
  isValidPassword: vi.fn(),
  readFileSync: vi.fn(),
  getAddress: vi.fn(),
  setAddress: vi.fn(),
  setSeed: vi.fn()
}))

vi.mock('electron', () => ({
  app: { quit: vi.fn(), relaunch: vi.fn() },
  BrowserWindow: { getFocusedWindow: vi.fn(), getAllWindows: vi.fn(() => []) },
  dialog: { showMessageBox: vi.fn() }
}))

vi.mock('node:fs', () => ({
  default: { readFileSync: mocks.readFileSync }
}))

vi.mock('../electron-restart', () => ({ default: vi.fn() }))
vi.mock('../database', () => ({ default: { getDb: vi.fn() } }))
vi.mock('../storage', () => ({ default: { persistState: vi.fn() } }))
vi.mock('../auth', () => ({
  default: {
    isValidPassword: mocks.isValidPassword,
    setPassword: vi.fn()
  }
}))
vi.mock('../wallet', () => ({
  default: {
    getAddress: mocks.getAddress,
    setAddress: mocks.setAddress,
    setSeed: mocks.setSeed
  }
}))
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
    chain: {
      bypassAuth: false,
      localProxyRouterUrl: 'http://127.0.0.1:8082'
    },
    isFailoverEnabled: false
  }
}))
vi.mock('../../../orchestrator/orchestrator', () => ({ Orchestrator: class {} }))
vi.mock('../../../logger', () => ({
  default: {
    error: vi.fn(),
    info: vi.fn(),
    verbose: vi.fn(),
    warn: vi.fn()
  }
}))
vi.mock('../keys', () => ({ default: {} }))
vi.mock('../wallets', () => ({}))
vi.mock('../attachments', () => ({}))
vi.mock('../../../../../orchestrator.config', () => ({ cfg: {} }))

import { onboardingCompleted, onLoginSubmit, resetAuthHeaders } from './handlers'

const ADDRESS_A = `0x${'11'.repeat(20)}`
const ADDRESS_B = `0x${'22'.repeat(20)}`
const PASSWORD = 'correct-password'

const response = (body: unknown) =>
  ({
    ok: true,
    status: 200,
    json: vi.fn(async () => body)
  }) as unknown as Response

const core = {
  emitter: { emit: mocks.emit }
} as any

const mockWalletResponse = (body: unknown) => {
  mocks.fetch.mockImplementation(async (input: string | URL | Request) => {
    const url = String(input)
    if (url.endsWith('/auth/cookie/path')) {
      return response({ path: '/tmp/morpheus-test-auth-cookie' })
    }
    if (url.endsWith('/wallet')) {
      return response(body)
    }
    throw new Error(`Unexpected request: ${url}`)
  })
}

const walletRequests = () =>
  mocks.fetch.mock.calls.filter(([input]) => String(input).endsWith('/wallet'))

describe('wallet login acknowledgement', () => {
  beforeEach(() => {
    mocks.emit.mockReset()
    mocks.fetch.mockReset()
    mocks.isValidPassword.mockReset().mockResolvedValue(true)
    mocks.readFileSync.mockReset().mockReturnValue('test-user:test-password')
    mocks.getAddress.mockReset()
    mocks.setAddress.mockReset()
    mocks.setSeed.mockReset()
    resetAuthHeaders()
    vi.stubGlobal('fetch', mocks.fetch)
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('returns matching authoritative identity without rewriting the stored pointer', async () => {
    mocks.getAddress.mockReturnValue({ address: ADDRESS_A })
    mockWalletResponse({ address: ADDRESS_A })

    await expect(onLoginSubmit({ password: PASSWORD }, core)).resolves.toEqual({
      address: ADDRESS_A,
      isActive: true
    })

    expect(walletRequests()).toHaveLength(1)
    const authRequest = mocks.fetch.mock.calls.find(([input]) =>
      String(input).endsWith('/auth/cookie/path')
    )
    const walletRequest = walletRequests()[0]
    expect(authRequest?.[1]?.signal).toBe(walletRequest?.[1]?.signal)
    expect(walletRequest?.[1]?.signal).toBeInstanceOf(AbortSignal)
    expect(mocks.setAddress).not.toHaveBeenCalled()
    expect(mocks.emit.mock.calls).toEqual([
      ['open-wallet', { address: ADDRESS_A, isActive: true }],
      ['open-proxy-router', { password: PASSWORD }]
    ])
  })

  it('uses proxy identity and repairs a mismatched stored pointer', async () => {
    mocks.getAddress.mockReturnValue({ address: ADDRESS_A })
    mockWalletResponse({ address: ADDRESS_B })

    await expect(onLoginSubmit({ password: PASSWORD }, core)).resolves.toEqual({
      address: ADDRESS_B,
      isActive: true
    })

    expect(walletRequests()).toHaveLength(1)
    expect(mocks.setAddress).toHaveBeenCalledOnce()
    expect(mocks.setAddress).toHaveBeenCalledWith(ADDRESS_B)
    expect(mocks.emit).toHaveBeenCalledWith('open-wallet', {
      address: ADDRESS_B,
      isActive: true
    })
  })

  it.each([null, {}, { address: '' }, { address: 'not-an-address' }, { address: '0x1234' }])(
    'rejects missing or invalid authoritative wallet response %j without emitting',
    async (body) => {
      mocks.getAddress.mockReturnValue({ address: ADDRESS_A })
      mockWalletResponse(body)

      const result = await onLoginSubmit({ password: PASSWORD }, core)

      expect(result).toMatchObject({
        error: {
          name: 'WalletError',
          message:
            'The local proxy-router did not return a valid wallet address. Please wait a moment and try again.'
        }
      })
      expect(walletRequests()).toHaveLength(1)
      expect(mocks.setAddress).not.toHaveBeenCalled()
      expect(mocks.emit).not.toHaveBeenCalled()
    }
  )

  it('preserves an actionable proxy-unreachable error without emitting', async () => {
    mocks.getAddress.mockReturnValue({ address: ADDRESS_A })
    mocks.fetch.mockImplementation(async (input: string | URL | Request) => {
      const url = String(input)
      if (url.endsWith('/auth/cookie/path')) {
        return response({ path: '/tmp/morpheus-test-auth-cookie' })
      }
      if (url.endsWith('/wallet')) {
        throw new Error('connect ECONNREFUSED 127.0.0.1:8082')
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const result = await onLoginSubmit({ password: PASSWORD }, core)

    expect(result).toMatchObject({
      error: {
        name: 'WalletError',
        message:
          'Cannot connect to the local proxy-router yet. Wait a moment and try again; if it persists, restart MorpheusUI.',
        inner: 'Cannot reach the local proxy-router. Check that it is running in Settings.'
      }
    })
    expect(walletRequests()).toHaveLength(1)
    expect(mocks.setAddress).not.toHaveBeenCalled()
    expect(mocks.emit).not.toHaveBeenCalled()
  })

  it('aborts a hanging authoritative wallet read after the login budget', async () => {
    const controller = new AbortController()
    const timeout = vi.spyOn(AbortSignal, 'timeout').mockReturnValue(controller.signal)
    mocks.getAddress.mockReturnValue({ address: ADDRESS_A })
    mocks.fetch.mockImplementation(async (input: string | URL | Request, init?: RequestInit) => {
      const url = String(input)
      if (url.endsWith('/auth/cookie/path')) {
        return response({ path: '/tmp/morpheus-test-auth-cookie' })
      }
      if (url.endsWith('/wallet')) {
        return new Promise<Response>((_resolve, reject) => {
          init?.signal?.addEventListener('abort', () => {
            reject(init.signal?.reason ?? new Error('aborted'))
          })
        })
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const pending = onLoginSubmit({ password: PASSWORD }, core)
    await vi.waitFor(() => expect(walletRequests()).toHaveLength(1))
    controller.abort(new Error('login wallet check timed out'))

    await expect(pending).resolves.toMatchObject({
      error: {
        name: 'WalletError',
        message:
          'Cannot connect to the local proxy-router yet. Wait a moment and try again; if it persists, restart MorpheusUI.'
      }
    })
    expect(timeout).toHaveBeenCalledWith(2_000)
    expect(mocks.emit).not.toHaveBeenCalled()
  })

  it('does not contact the proxy-router when the password is invalid', async () => {
    mocks.getAddress.mockReturnValue({ address: ADDRESS_A })
    mocks.isValidPassword.mockResolvedValue(false)

    const result = await onLoginSubmit({ password: 'wrong-password' }, core)

    expect(result).toMatchObject({
      error: { name: 'WalletError', message: 'Invalid password' }
    })
    expect(mocks.fetch).not.toHaveBeenCalled()
    expect(mocks.getAddress).not.toHaveBeenCalled()
    expect(mocks.setAddress).not.toHaveBeenCalled()
    expect(mocks.emit).not.toHaveBeenCalled()
  })

  it('reuses onboarding wallet verification instead of reading it again after commit', async () => {
    vi.spyOn(console, 'log').mockImplementation(() => undefined)
    const privateKey = `0x${'33'.repeat(32)}`
    mocks.getAddress.mockReturnValue({ address: ADDRESS_A })
    mocks.fetch.mockImplementation(async (input: string | URL | Request) => {
      const url = String(input)
      if (url.endsWith('/auth/cookie/path')) {
        return response({ path: '/tmp/morpheus-test-auth-cookie' })
      }
      if (url.endsWith('/wallet/privateKey')) {
        return response({})
      }
      if (url.endsWith('/wallet')) {
        return response({ address: ADDRESS_A })
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    await expect(
      onboardingCompleted({ password: PASSWORD, privateKey }, core)
    ).resolves.toBeUndefined()

    expect(walletRequests()).toHaveLength(1)
    expect(mocks.setSeed).toHaveBeenCalledWith(ADDRESS_A, PASSWORD)
    expect(mocks.setAddress).toHaveBeenCalledWith(ADDRESS_A)
    expect(mocks.emit.mock.calls).toEqual([
      ['create-wallet', { address: ADDRESS_A }],
      ['open-wallet', { address: ADDRESS_A, isActive: true }],
      ['open-proxy-router', { password: PASSWORD }]
    ])
  })
})
