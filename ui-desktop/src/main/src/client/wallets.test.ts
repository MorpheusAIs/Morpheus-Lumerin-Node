import { beforeEach, describe, expect, it, vi } from 'vitest'

// electron-settings and keytar both require a real Electron/OS environment, so
// the module is exercised against in-memory doubles. The logic under test is
// the registry bookkeeping — index allocation, duplicate rejection, removal
// guards — which is where a mistake silently loses access to a wallet.
const store = new Map<string, any>()
const keychain = new Map<string, string>()

vi.mock('./settings', () => ({
  getKey: (k: string) => store.get(k),
  setKey: (k: string, v: any) => void store.set(k, v),
}))

vi.mock('keytar', () => ({
  default: {
    setPassword: async (_s: string, a: string, p: string) => void keychain.set(a, p),
    getPassword: async (_s: string, a: string) => keychain.get(a) ?? null,
    deletePassword: async (_s: string, a: string) => keychain.delete(a),
  },
}))

vi.mock('../../logger', () => ({
  default: { error: vi.fn(), info: vi.fn(), warn: vi.fn(), verbose: vi.fn() },
}))

const W = await import('./wallets')

const ADDR_A = '0x15dd2028C976beaA6668E286b496A518F457b5Cf'
const ADDR_B = '0x1111111111111111111111111111111111111111'
const ADDR_C = '0x2222222222222222222222222222222222222222'
const KEY = '0x' + '11'.repeat(32)

beforeEach(() => {
  store.clear()
  keychain.clear()
})

describe('adoptCurrentWallet', () => {
  it('registers a pre-existing mnemonic wallet as an HD account', () => {
    const w = W.adoptCurrentWallet({ address: ADDR_A, kind: 'mnemonic', derivationPath: '0' })
    expect(w.kind).toBe('hd')
    expect(w.derivationPath).toBe('0')
    expect(W.listWallets()).toHaveLength(1)
    expect(W.getActiveWalletId()).toBe(w.id)
  })

  it('registers a private-key wallet as imported', () => {
    expect(W.adoptCurrentWallet({ address: ADDR_A, kind: 'privateKey' }).kind).toBe('imported')
  })

  // Called on every wallet-list read, so it must not append a duplicate each time.
  it('is idempotent', () => {
    const first = W.adoptCurrentWallet({ address: ADDR_A, kind: 'mnemonic', derivationPath: '0' })
    const second = W.adoptCurrentWallet({ address: ADDR_A, kind: 'mnemonic', derivationPath: '0' })
    expect(second.id).toBe(first.id)
    expect(W.listWallets()).toHaveLength(1)
  })

  it('matches on address regardless of case', () => {
    const first = W.adoptCurrentWallet({ address: ADDR_A, kind: 'mnemonic', derivationPath: '0' })
    const again = W.adoptCurrentWallet({ address: ADDR_A.toLowerCase(), kind: 'mnemonic' })
    expect(again.id).toBe(first.id)
  })
})

describe('nextHdIndex', () => {
  it('starts at 0', () => {
    expect(W.nextHdIndex()).toBe(0)
  })

  it('advances past used indices', () => {
    W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    W.addHdWallet({ address: ADDR_B, derivationPath: '1' })
    expect(W.nextHdIndex()).toBe(2)
  })

  // Reusing a freed index is correct — it maps to the same address, so the
  // alternative (always incrementing) would strand accounts behind a gap.
  it('reuses a gap left by a removal', () => {
    W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    const b = W.addHdWallet({ address: ADDR_B, derivationPath: '1' })
    W.addHdWallet({ address: ADDR_C, derivationPath: '2' })
    W.setActiveWalletId('someone-else')
    return W.removeWallet(b.id).then(() => {
      expect(W.nextHdIndex()).toBe(1)
    })
  })

  it('ignores imported wallets, which have no index', async () => {
    await W.addImportedWallet({ address: ADDR_A, privateKey: KEY })
    expect(W.nextHdIndex()).toBe(0)
  })
})

describe('duplicate protection', () => {
  it('rejects an HD account that is already listed', () => {
    W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    expect(() => W.addHdWallet({ address: ADDR_A, derivationPath: '5' })).toThrow(/already in your list/)
  })

  it('rejects importing a wallet that is already listed', async () => {
    W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    await expect(W.addImportedWallet({ address: ADDR_A, privateKey: KEY })).rejects.toThrow(
      /already in your list/,
    )
  })
})

describe('imported wallet secrets', () => {
  it('stores the key in the keychain, never in settings', async () => {
    const w = await W.addImportedWallet({ address: ADDR_A, privateKey: KEY })

    expect(await W.getImportedPrivateKey(w.id)).toBe(KEY)
    // The settings blob is a plaintext JSON file on disk — the key must not be in it.
    expect(JSON.stringify([...store.values()])).not.toContain(KEY)
  })

  it('does not leave a registry entry behind if the keychain write fails', async () => {
    const kt = (await import('keytar')).default
    const spy = vi.spyOn(kt, 'setPassword').mockRejectedValueOnce(new Error('keychain locked'))

    await expect(W.addImportedWallet({ address: ADDR_A, privateKey: KEY })).rejects.toThrow(
      /keychain locked/,
    )
    expect(W.listWallets()).toHaveLength(0)
    spy.mockRestore()
  })

  it('deletes the secret when the wallet is removed', async () => {
    const a = W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    const b = await W.addImportedWallet({ address: ADDR_B, privateKey: KEY })
    W.setActiveWalletId(a.id)

    await W.removeWallet(b.id)
    expect(await W.getImportedPrivateKey(b.id)).toBeNull()
  })
})

describe('removal guards', () => {
  it('refuses to remove the only wallet', async () => {
    const a = W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    W.setActiveWalletId(a.id)
    await expect(W.removeWallet(a.id)).rejects.toThrow(/only wallet/)
  })

  it('refuses to remove the active wallet', async () => {
    const a = W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    W.addHdWallet({ address: ADDR_B, derivationPath: '1' })
    W.setActiveWalletId(a.id)
    await expect(W.removeWallet(a.id)).rejects.toThrow(/Switch to another wallet/)
  })

  it('is a no-op for an unknown id', async () => {
    W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    await expect(W.removeWallet('nope')).resolves.toBeUndefined()
    expect(W.listWallets()).toHaveLength(1)
  })
})

describe('assertSwitchable', () => {
  it('allows HD accounts, which need no stored secret', async () => {
    const a = W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    expect(await W.assertSwitchable(a.id)).toBeNull()
  })

  it('allows an imported wallet whose key is in the keychain', async () => {
    const a = await W.addImportedWallet({ address: ADDR_A, privateKey: KEY })
    expect(await W.assertSwitchable(a.id)).toBeNull()
  })

  // A wallet adopted from a pre-multi-wallet install was imported by private
  // key, but the app never kept that key. Switching away would strand the user.
  it('blocks an adopted imported wallet with no stored key', async () => {
    const a = W.adoptCurrentWallet({ address: ADDR_A, kind: 'privateKey' })
    expect(await W.assertSwitchable(a.id)).toMatch(/not stored in this app/)
  })

  it('reports an unknown wallet', async () => {
    expect(await W.assertSwitchable('nope')).toMatch(/not found/)
  })
})

describe('renameWallet', () => {
  it('renames and persists', () => {
    const a = W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    W.renameWallet(a.id, '  Trading  ')
    expect(W.getWallet(a.id)?.label).toBe('Trading')
  })

  it('rejects an empty name', () => {
    const a = W.addHdWallet({ address: ADDR_A, derivationPath: '0' })
    expect(() => W.renameWallet(a.id, '   ')).toThrow(/cannot be empty/)
  })
})

describe('listWallets', () => {
  it('returns an empty list rather than throwing on a corrupt registry', () => {
    store.set('user.wallets', '{not json')
    expect(W.listWallets()).toEqual([])
  })
})
