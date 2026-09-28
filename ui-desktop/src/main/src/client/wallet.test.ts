import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest'

vi.mock('electron-settings', () => ({
  default: { getSync: vi.fn(), setSync: vi.fn() }
}))

import { getAddressForDerivationPath } from './wallet'

describe('onboarding public wallet identity', () => {
  // Main-process crypto operates on Node buffers, not jsdom's separate typed-array realm.
  beforeAll(() => vi.stubGlobal('Uint8Array', Object.getPrototypeOf(Buffer.prototype).constructor))
  afterAll(() => vi.unstubAllGlobals())

  // Public test vector shared with proxy-router/internal/repositories/wallet/hdwallet_test.go.
  // No installed wallet, keychain or settings are accessed.
  const seed = '00000000000000500000000000000000'
  const expected = '0xe39Be4d7E9D91D14e837589F3027798f3911A83c'

  it('derives the same relative index-zero address as the proxy-router', () => {
    expect(getAddressForDerivationPath(seed, '0')).toBe(expected)
  })

  it('accepts a complete derivation path without prefixing it a second time', () => {
    expect(getAddressForDerivationPath(seed, "m/44'/60'/0'/0/0")).toBe(expected)
  })

  it('respects a different selected account index', () => {
    const relative = getAddressForDerivationPath(seed, '2')
    expect(relative).toBe(getAddressForDerivationPath(seed, "m/44'/60'/0'/0/2"))
    expect(relative).not.toBe(expected)
  })

  it('rejects an invalid path instead of returning the default account', () => {
    expect(() => getAddressForDerivationPath(seed, 'invalid')).toThrow()
  })
})
