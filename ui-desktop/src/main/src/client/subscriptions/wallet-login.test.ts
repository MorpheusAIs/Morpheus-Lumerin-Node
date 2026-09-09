import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  emit: vi.fn(),
  fetch: vi.fn(),
  isValidPassword: vi.fn(),
  readFileSync: vi.fn(),
  getAddress: vi.fn(),
  setAddress: vi.fn(),
  setSeed: vi.fn(),
  getPasswordHash: vi.fn(),
  setPassword: vi.fn(),
  privateKeyToAddress: vi.fn(),
  getAddressForDerivationPath: vi.fn(),
  isValidMnemonic: vi.fn(),
  addImportedWallet: vi.fn()
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
    setPassword: mocks.setPassword
  }
}))
vi.mock('../wallet', () => ({
  default: {
    getAddress: mocks.getAddress,
    setAddress: mocks.setAddress,
    setSeed: mocks.setSeed,
    privateKeyToAddress: mocks.privateKeyToAddress,
    getAddressForDerivationPath: mocks.getAddressForDerivationPath
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
  setPasswordHash: vi.fn(),
  getPasswordHash: mocks.getPasswordHash
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
vi.mock('../keys', () => ({
  default: {
    isValidMnemonic: mocks.isValidMnemonic,
    mnemonicToSeedHex: vi.fn(() => 'synthetic-test-seed')
  }
}))
vi.mock('../wallets', () => ({ addImportedWallet: mocks.addImportedWallet }))
vi.mock('../attachments', () => ({}))
vi.mock('../../../../../orchestrator.config', () => ({ cfg: {} }))

import { importWallet, onboardingCompleted, onLoginSubmit, resetAuthHeaders } from './handlers'

const ADDRESS_A = `0x${'11'.repeat(20)}`
const ADDRESS_B = `0x${'22'.repeat(20)}`
const PASSWORD = 'correct-password'

const response = (body: unknown, status = 200) =>
  ({
    ok: status >= 200 && status < 300,
    status,
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
    mocks.getPasswordHash.mockReset().mockReturnValue(undefined)
    mocks.setPassword.mockReset().mockResolvedValue(undefined)
    mocks.privateKeyToAddress.mockReset().mockReturnValue(ADDRESS_A)
    mocks.getAddressForDerivationPath.mockReset().mockReturnValue(ADDRESS_A)
    mocks.isValidMnemonic.mockReset().mockReturnValue(true)
    mocks.addImportedWallet.mockReset()
    resetAuthHeaders()
    vi.stubGlobal('fetch', mocks.fetch)
  })

  afterEach(() => {
    vi.useRealTimers()
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

  it('reconnects a matching node wallet and reuses verification after password commit', async () => {
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

    expect(walletRequests()).toHaveLength(2)
    expect(mocks.fetch.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false)
    expect(mocks.setSeed).toHaveBeenCalledWith(ADDRESS_A, PASSWORD)
    expect(mocks.setAddress).toHaveBeenCalledWith(ADDRESS_A)
    expect(mocks.emit.mock.calls).toEqual([
      ['create-wallet', { address: ADDRESS_A }],
      ['open-wallet', { address: ADDRESS_A, isActive: true }],
      ['open-proxy-router', { password: PASSWORD }]
    ])
  })

  const setupData = { password: PASSWORD, privateKey: `0x${'33'.repeat(32)}` }

  it('imports another wallet by deriving its address locally without changing the active node', async () => {
    const privateKey = '33'.repeat(32)
    const saved = { id: 'imported-wallet', address: ADDRESS_A, kind: 'imported' }
    mocks.addImportedWallet.mockReturnValue(saved)
    await expect(importWallet({ privateKey, label: 'Second wallet' })).resolves.toEqual(saved)
    expect(mocks.privateKeyToAddress).toHaveBeenCalledWith(`0x${privateKey}`)
    expect(mocks.addImportedWallet).toHaveBeenCalledWith({
      address: ADDRESS_A,
      privateKey: `0x${privateKey}`,
      label: 'Second wallet'
    })
    expect(mocks.fetch).not.toHaveBeenCalled()
    expect(mocks.setPassword).not.toHaveBeenCalled()
    expect(mocks.setAddress).not.toHaveBeenCalled()
    expect(mocks.emit).not.toHaveBeenCalled()
  })

  function setupProxy(options: { writeFails?: boolean; verificationFails?: boolean } = {}) {
    let configured = false
    mocks.fetch.mockImplementation(async (input: string | URL | Request, init?: RequestInit) => {
      const url = String(input)
      if (url.endsWith('/auth/cookie/path')) return response({ path: '/tmp/test-auth-cookie' })
      if (url.endsWith('/wallet') && init?.method !== 'POST') {
        if (!configured) return response({ error: 'wallet not set' }, 500)
        if (options.verificationFails)
          return response({ error: 'Keychain temporarily unavailable' }, 500)
        return response({ address: ADDRESS_A })
      }
      if (url.endsWith('/wallet/privateKey') || url.endsWith('/wallet/mnemonic')) {
        if (options.writeFails) return response({ error: 'Keychain write failed' }, 500)
        configured = true
        return response({ address: ADDRESS_A })
      }
      throw new Error(`Unexpected test request: ${url}`)
    })
    return options
  }

  it.each(['/config/ethNode', '/wallet/privateKey', '/wallet/mnemonic'])(
    'bounds a stalled %s write and releases setup for a safe retry',
    async (stalledPath) => {
      vi.useFakeTimers()
      const timeout = vi.spyOn(AbortSignal, 'timeout').mockImplementation((milliseconds) => {
        const controller = new AbortController()
        setTimeout(() => {
          const error = new Error('The operation timed out')
          error.name = 'TimeoutError'
          controller.abort(error)
        }, milliseconds)
        return controller.signal
      })
      let stalled = true
      let configured = false
      let stalledSignal: AbortSignal | null | undefined
      mocks.fetch.mockImplementation(async (input: string | URL | Request, init?: RequestInit) => {
        const url = String(input)
        if (url.endsWith('/auth/cookie/path')) return response({ path: '/tmp/test-auth-cookie' })
        if (url.endsWith('/wallet'))
          return configured
            ? response({ address: ADDRESS_A })
            : response({ error: 'wallet not set' }, 500)
        if (url.endsWith(stalledPath) && stalled) {
          // A timed-out POST may already have saved the key. Its retry must
          // detect that identity rather than sending another wallet write.
          configured = stalledPath !== '/config/ethNode'
          stalledSignal = init?.signal
          return new Promise<Response>((_resolve, reject) => {
            stalledSignal?.addEventListener('abort', () => reject(stalledSignal?.reason), {
              once: true
            })
          })
        }
        if (url.endsWith('/config/ethNode')) return response({})
        if (url.endsWith('/wallet/privateKey') || url.endsWith('/wallet/mnemonic')) {
          configured = true
          return response({ address: ADDRESS_A })
        }
        throw new Error(`Unexpected test request: ${url}`)
      })
      const input =
        stalledPath === '/wallet/mnemonic'
          ? {
              password: PASSWORD,
              mnemonic:
                'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about'
            }
          : {
              ...setupData,
              ...(stalledPath === '/config/ethNode' ? { ethNode: 'http://localhost:8545' } : {})
            }
      const pending = onboardingCompleted(input, core)
      await vi.advanceTimersByTimeAsync(0)
      expect(stalledSignal).toBeInstanceOf(AbortSignal)
      expect(timeout).toHaveBeenCalledWith(15_000)
      await vi.advanceTimersByTimeAsync(14_999)
      expect(stalledSignal?.aborted).toBe(false)
      await expect(onboardingCompleted(input, core)).resolves.toMatchObject({
        error: { message: expect.stringContaining('already in progress') }
      })
      await vi.advanceTimersByTimeAsync(1)
      await expect(pending).resolves.toMatchObject({
        error: { message: expect.stringContaining('setup timed out') }
      })
      expect(stalledSignal?.aborted).toBe(true)
      expect(mocks.setPassword).not.toHaveBeenCalled()
      expect(mocks.setAddress).not.toHaveBeenCalled()
      expect(mocks.emit).not.toHaveBeenCalled()

      stalled = false
      await expect(onboardingCompleted(input, core)).resolves.toBeUndefined()
      expect(mocks.setPassword).toHaveBeenCalledOnce()
      expect(
        mocks.fetch.mock.calls.filter(([url]) =>
          /\/wallet\/(privateKey|mnemonic)$/.test(String(url))
        )
      ).toHaveLength(1)
    }
  )

  it('commits the app password only after fresh wallet setup, verified identity and metadata', async () => {
    setupProxy()
    mocks.setPassword.mockImplementation(async () => {
      expect(walletRequests()).toHaveLength(2)
      expect(mocks.setSeed).toHaveBeenCalledWith(ADDRESS_A, PASSWORD)
      expect(mocks.setAddress).toHaveBeenCalledWith(ADDRESS_A)
      expect(mocks.emit).not.toHaveBeenCalled()
    })
    await expect(onboardingCompleted(setupData, core)).resolves.toBeUndefined()
    expect(mocks.setPassword).toHaveBeenCalledOnce()
    expect(mocks.setPassword).toHaveBeenCalledWith(PASSWORD)
    expect(
      mocks.fetch.mock.calls.filter(([url]) => String(url).endsWith('/wallet/privateKey'))
    ).toHaveLength(1)
  })

  it.each(['writeFails', 'verificationFails'] as const)(
    'does not create a password hash when wallet setup fails at %s',
    async (failure) => {
      const options = setupProxy({ [failure]: true })
      await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
      expect(mocks.setPassword).not.toHaveBeenCalled()
      expect(mocks.setSeed).not.toHaveBeenCalled()
      expect(mocks.setAddress).not.toHaveBeenCalled()
      expect(mocks.emit).not.toHaveBeenCalled()

      options[failure] = false
      await expect(onboardingCompleted(setupData, core)).resolves.toBeUndefined()
      expect(mocks.setPassword).toHaveBeenCalledOnce()
      expect(
        mocks.fetch.mock.calls.filter(([url]) => String(url).endsWith('/wallet/privateKey'))
      ).toHaveLength(failure === 'writeFails' ? 2 : 1)
    }
  )

  it('retries safely after a successful node write but failed local metadata without overwriting the node', async () => {
    setupProxy()
    mocks.setSeed.mockImplementationOnce(() => {
      throw new Error('Local settings unavailable')
    })
    await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
    expect(mocks.setPassword).not.toHaveBeenCalled()

    await expect(onboardingCompleted(setupData, core)).resolves.toBeUndefined()
    expect(mocks.setPassword).toHaveBeenCalledOnce()
    expect(
      mocks.fetch.mock.calls.filter(([url]) => String(url).endsWith('/wallet/privateKey'))
    ).toHaveLength(1)
  })

  it('rejects a node wallet with a different identity without writes or a password change', async () => {
    mockWalletResponse({ address: ADDRESS_B })
    await expect(onboardingCompleted(setupData, core)).resolves.toMatchObject({
      error: { message: expect.stringContaining('different wallet') }
    })
    expect(mocks.fetch.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false)
    expect(mocks.setPassword).not.toHaveBeenCalled()
    expect(mocks.setAddress).not.toHaveBeenCalled()
  })

  it('blocks setup on an existing password-protected installation without authenticated recovery', async () => {
    mocks.getPasswordHash.mockReturnValue('existing-completed-hash')
    mocks.getAddress.mockReturnValue({ address: ADDRESS_A })
    await expect(onboardingCompleted(setupData, core)).resolves.toMatchObject({
      error: { message: expect.stringContaining('already protected') }
    })
    expect(mocks.fetch).not.toHaveBeenCalled()
    expect(mocks.setPassword).not.toHaveBeenCalled()
  })

  it.each([
    [500, { error: 'Keychain access was denied' }],
    [401, { error: 'wallet not set' }],
    [404, { error: 'wallet not set' }],
    [200, {}],
    [200, { address: 'invalid' }]
  ])(
    'never overwrites an unknown node wallet during fresh setup (HTTP %s %j)',
    async (status, body) => {
      mocks.fetch.mockImplementation(async (input: string | URL | Request) => {
        if (String(input).endsWith('/auth/cookie/path'))
          return response({ path: '/tmp/test-auth-cookie' })
        return response(body, status as number)
      })
      await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
      expect(mocks.fetch.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false)
      expect(mocks.setPassword).not.toHaveBeenCalled()
      expect(mocks.setAddress).not.toHaveBeenCalled()
      expect(mocks.emit).not.toHaveBeenCalled()
    }
  )

  it('keeps fresh setup retryable without changing settings while the node is unreachable', async () => {
    mocks.fetch.mockRejectedValue(new Error('ECONNREFUSED'))
    await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
    expect(mocks.setPassword).not.toHaveBeenCalled()
    expect(mocks.setAddress).not.toHaveBeenCalled()
    expect(mocks.emit).not.toHaveBeenCalled()
    setupProxy()
    await expect(onboardingCompleted(setupData, core)).resolves.toBeUndefined()
    expect(mocks.setPassword).toHaveBeenCalledOnce()
  })

  it('permits retryable setup recovery only after a valid password and authoritative empty-wallet response', async () => {
    mocks.getPasswordHash.mockReturnValue('orphaned-setup-hash')
    const options = setupProxy({ writeFails: true })
    await expect(onLoginSubmit({ password: PASSWORD }, core)).resolves.toEqual({
      requiresOnboarding: true
    })
    expect(mocks.emit).not.toHaveBeenCalled()
    expect(mocks.setPassword).not.toHaveBeenCalled()

    await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
    expect(mocks.setPassword).not.toHaveBeenCalled()
    options.writeFails = false
    await expect(onboardingCompleted(setupData, core)).resolves.toBeUndefined()
    expect(mocks.setPassword).toHaveBeenCalledOnce()

    // Completing setup consumes the grant, even if an untrusted caller resends
    // the onboarding IPC rather than using the authenticated wallet workflow.
    await expect(onboardingCompleted(setupData, core)).resolves.toMatchObject({
      error: { message: expect.stringContaining('already protected') }
    })
  })

  it('binds recovery authorization to the password hash that was actually verified', async () => {
    mocks.getPasswordHash.mockReturnValue('verified-orphan-hash')
    setupProxy()
    await expect(onLoginSubmit({ password: PASSWORD }, core)).resolves.toEqual({
      requiresOnboarding: true
    })
    mocks.getPasswordHash.mockReturnValue('changed-in-another-operation')
    await expect(onboardingCompleted(setupData, core)).resolves.toMatchObject({
      error: { message: expect.stringContaining('already protected') }
    })
    expect(mocks.fetch.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false)
    expect(mocks.setPassword).not.toHaveBeenCalled()
  })

  it('never offers recovery after an invalid password', async () => {
    mocks.getPasswordHash.mockReturnValue('protected-hash')
    mocks.isValidPassword.mockResolvedValue(false)
    setupProxy()
    await expect(onLoginSubmit({ password: 'incorrect' }, core)).resolves.toMatchObject({
      error: { message: 'Invalid password' }
    })
    await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
    expect(mocks.fetch).not.toHaveBeenCalled()
  })

  it('does not grant recovery if password protection changes while login is pending', async () => {
    mocks.getPasswordHash.mockReturnValue('initial-password-hash')
    mocks.isValidPassword.mockImplementation(async () => {
      mocks.getPasswordHash.mockReturnValue('changed-while-verifying')
      return true
    })
    setupProxy()
    const result = await onLoginSubmit({ password: PASSWORD }, core)
    expect(result).toHaveProperty('error')
    expect(result).not.toHaveProperty('requiresOnboarding')
    await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
    expect(mocks.fetch.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false)
    expect(mocks.setPassword).not.toHaveBeenCalled()
  })

  it.each([
    [500, { error: 'Keychain access was denied' }],
    [401, { error: 'wallet not set' }],
    [404, { error: 'wallet not set' }],
    [200, {}],
    [200, { address: 'invalid' }]
  ])(
    'does not interpret HTTP %s %j as an empty wallet or authorize recovery',
    async (status, body) => {
      mocks.getPasswordHash.mockReturnValue('still-protected-hash')
      mocks.fetch.mockImplementation(async (input: string | URL | Request) => {
        if (String(input).endsWith('/auth/cookie/path'))
          return response({ path: '/tmp/test-auth-cookie' })
        return response(body, status as number)
      })
      const result = await onLoginSubmit({ password: PASSWORD }, core)
      expect(result).toHaveProperty('error')
      expect(result).not.toHaveProperty('requiresOnboarding')
      await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
      expect(mocks.fetch.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false)
      expect(mocks.setPassword).not.toHaveBeenCalled()
    }
  )

  it('does not grant recovery when the router cannot be reached', async () => {
    mocks.getPasswordHash.mockReturnValue('offline-protected-hash')
    mocks.fetch.mockRejectedValue(new Error('ECONNREFUSED'))
    const result = await onLoginSubmit({ password: PASSWORD }, core)
    expect(result).toHaveProperty('error')
    expect(result).not.toHaveProperty('requiresOnboarding')
    await expect(onboardingCompleted(setupData, core)).resolves.toHaveProperty('error')
    expect(mocks.setPassword).not.toHaveBeenCalled()
  })

  it('rejects overlapping setup attempts rather than writing two wallets', async () => {
    let finishRead!: (value: Response) => void
    let reads = 0
    mocks.fetch.mockImplementation(async (input: string | URL | Request) => {
      const url = String(input)
      if (url.endsWith('/auth/cookie/path')) return response({ path: '/tmp/test-auth-cookie' })
      if (url.endsWith('/wallet')) {
        if (++reads === 1) return new Promise<Response>((resolve) => (finishRead = resolve))
        return response({ address: ADDRESS_A })
      }
      if (url.endsWith('/wallet/privateKey')) return response({ address: ADDRESS_A })
      throw new Error(`Unexpected test request: ${url}`)
    })
    const first = onboardingCompleted(setupData, core)
    await vi.waitFor(() => expect(reads).toBe(1))
    await expect(onboardingCompleted(setupData, core)).resolves.toMatchObject({
      error: { message: expect.stringContaining('already in progress') }
    })
    finishRead(response({ error: 'wallet not set' }, 500))
    await expect(first).resolves.toBeUndefined()
    expect(
      mocks.fetch.mock.calls.filter(([url]) => String(url).endsWith('/wallet/privateKey'))
    ).toHaveLength(1)
    expect(mocks.setPassword).toHaveBeenCalledOnce()
  })

  it('supports mnemonic setup with the exact selected derivation path', async () => {
    setupProxy()
    const mnemonic =
      'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about'
    const derivationPath = "m/44'/60'/0'/0/2"
    await expect(
      onboardingCompleted({ password: PASSWORD, mnemonic, derivationPath }, core)
    ).resolves.toBeUndefined()
    expect(mocks.getAddressForDerivationPath).toHaveBeenCalledWith(
      'synthetic-test-seed',
      derivationPath
    )
    const write = mocks.fetch.mock.calls.find(([url]) => String(url).endsWith('/wallet/mnemonic'))
    expect(JSON.parse(write![1].body)).toEqual({ mnemonic, derivationPath })
  })
})
