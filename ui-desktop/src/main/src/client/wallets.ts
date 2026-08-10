import keytar from 'keytar'
import { randomUUID } from 'node:crypto'

import { getKey, setKey } from './settings'
import logger from '../../logger'

/**
 * Multi-wallet registry.
 *
 * The proxy-router holds exactly ONE key at a time — fixed keychain slots for
 * `private-key`, or `mnemonic` + `mnemonic-derivation-path`. Rather than
 * reshaping that (which would mean a new storage schema and a migration for
 * every existing install), this keeps a registry on the app side and swaps the
 * proxy-router's active key when the user switches. `proxyctl` already watches
 * `PrivateKeyUpdated()` and restarts cleanly, so the switch itself is handled.
 *
 * Two kinds of wallet:
 *
 *   hd        Derived from the mnemonic already stored in the proxy-router.
 *             NO SECRET IS STORED HERE — only the derivation index. Switching
 *             calls POST /wallet/derivationPath and the proxy-router re-derives
 *             from its own copy of the seed. The app never sees the phrase.
 *
 *   imported  An unrelated private key the user pasted in. This one genuinely
 *             is a secret, so it goes in the OS keychain via keytar (macOS
 *             Keychain / Windows Credential Vault / libsecret), never in
 *             electron-settings, which is a plaintext JSON file.
 *
 * Only non-secret metadata (id, label, kind, address, derivation index) lives
 * in settings.
 */

const REGISTRY_KEY = 'user.wallets'
const ACTIVE_KEY = 'user.activeWalletId'
const KEYCHAIN_SERVICE = 'morpheus-app-wallets'

export type WalletKind = 'hd' | 'imported'

export interface WalletRecord {
  id: string
  label: string
  kind: WalletKind
  /** Checksummed address, cached for display so the list renders without unlocking anything. */
  address: string
  /** HD only: the derivation path / account index passed to the proxy-router. */
  derivationPath?: string
}

// ---------------------------------------------------------------------------
// Registry (non-secret metadata)
// ---------------------------------------------------------------------------

export function listWallets(): WalletRecord[] {
  const raw = getKey(REGISTRY_KEY)
  if (!raw) {
    return []
  }
  try {
    const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
    return Array.isArray(parsed) ? parsed : []
  } catch (e) {
    logger.error('wallet registry is corrupt; treating as empty', e)
    return []
  }
}

function saveWallets(wallets: WalletRecord[]) {
  setKey(REGISTRY_KEY, JSON.stringify(wallets))
}

export function getActiveWalletId(): string | null {
  return getKey(ACTIVE_KEY) ?? null
}

export function setActiveWalletId(id: string | null) {
  setKey(ACTIVE_KEY, id)
}

export function getWallet(id: string): WalletRecord | undefined {
  return listWallets().find((w) => w.id === id)
}

const sameAddress = (a?: string, b?: string) =>
  !!a && !!b && a.toLowerCase() === b.toLowerCase()

/**
 * Ensures the currently-active proxy-router wallet appears in the registry.
 *
 * Existing installs onboarded before multi-wallet have a working wallet and an
 * empty registry. Without this they would open the switcher and see nothing,
 * or worse, "add" their own wallet a second time.
 */
export function adoptCurrentWallet(params: {
  address: string
  kind: string
  derivationPath?: string
}): WalletRecord {
  const wallets = listWallets()
  const existing = wallets.find((w) => sameAddress(w.address, params.address))
  if (existing) {
    setActiveWalletId(existing.id)
    return existing
  }

  const isHd = params.kind === 'mnemonic'
  const record: WalletRecord = {
    id: randomUUID(),
    label: isHd ? `Account ${params.derivationPath ?? '0'}` : 'Imported wallet',
    kind: isHd ? 'hd' : 'imported',
    address: params.address,
    derivationPath: isHd ? (params.derivationPath ?? '0') : undefined,
  }

  // Deliberately no secret written for an adopted wallet: if it is HD the
  // proxy-router already holds the mnemonic, and if it was imported we do not
  // have the key (the user pasted it during onboarding and it went straight to
  // the proxy-router). Such a wallet can be switched TO only while it is
  // already active — see assertSwitchable.
  saveWallets([...wallets, record])
  setActiveWalletId(record.id)
  return record
}

// ---------------------------------------------------------------------------
// HD accounts
// ---------------------------------------------------------------------------

/** Lowest unused HD index, so adding accounts doesn't collide after a removal. */
export function nextHdIndex(): number {
  const used = new Set(
    listWallets()
      .filter((w) => w.kind === 'hd')
      .map((w) => Number(w.derivationPath))
      .filter((n) => Number.isInteger(n) && n >= 0),
  )
  let i = 0
  while (used.has(i)) {
    i++
  }
  return i
}

export function addHdWallet(params: {
  address: string
  derivationPath: string
  label?: string
}): WalletRecord {
  const wallets = listWallets()

  const clash = wallets.find((w) => sameAddress(w.address, params.address))
  if (clash) {
    throw new Error(`That account is already in your list as "${clash.label}".`)
  }

  const record: WalletRecord = {
    id: randomUUID(),
    label: params.label?.trim() || `Account ${params.derivationPath}`,
    kind: 'hd',
    address: params.address,
    derivationPath: params.derivationPath,
  }
  saveWallets([...wallets, record])
  return record
}

// ---------------------------------------------------------------------------
// Imported wallets (secret material)
// ---------------------------------------------------------------------------

export async function addImportedWallet(params: {
  address: string
  privateKey: string
  label?: string
}): Promise<WalletRecord> {
  const wallets = listWallets()

  const clash = wallets.find((w) => sameAddress(w.address, params.address))
  if (clash) {
    throw new Error(`That wallet is already in your list as "${clash.label}".`)
  }

  const record: WalletRecord = {
    id: randomUUID(),
    label: params.label?.trim() || `Imported ${params.address.slice(0, 6)}`,
    kind: 'imported',
    address: params.address,
  }

  // Secret first: if the keychain write fails we must not end up with a
  // registry entry that can never be switched to.
  await keytar.setPassword(KEYCHAIN_SERVICE, record.id, params.privateKey)
  saveWallets([...wallets, record])
  return record
}

export function getImportedPrivateKey(id: string): Promise<string | null> {
  return keytar.getPassword(KEYCHAIN_SERVICE, id)
}

export async function removeWallet(id: string): Promise<void> {
  const wallets = listWallets()
  const target = wallets.find((w) => w.id === id)
  if (!target) {
    return
  }
  if (wallets.length === 1) {
    throw new Error('Cannot remove your only wallet.')
  }
  if (getActiveWalletId() === id) {
    throw new Error('Switch to another wallet before removing this one.')
  }

  if (target.kind === 'imported') {
    await keytar.deletePassword(KEYCHAIN_SERVICE, id).catch((e) => {
      // Leaving an orphaned secret is worse than a noisy log.
      logger.error(`failed to delete keychain entry for wallet ${id}`, e)
    })
  }
  saveWallets(wallets.filter((w) => w.id !== id))
}

export function renameWallet(id: string, label: string): WalletRecord {
  const trimmed = label.trim()
  if (!trimmed) {
    throw new Error('Name cannot be empty.')
  }
  const wallets = listWallets()
  const target = wallets.find((w) => w.id === id)
  if (!target) {
    throw new Error('Wallet not found.')
  }
  target.label = trimmed
  saveWallets(wallets)
  return target
}

/**
 * Explains why a wallet cannot be activated, or null if it can.
 *
 * The awkward case is a wallet adopted from a pre-multi-wallet install that was
 * imported by private key: the proxy-router has the key but we do not, so we
 * cannot re-supply it after switching away. Better to say so up front than to
 * strand the user on a different wallet.
 */
export async function assertSwitchable(id: string): Promise<string | null> {
  const target = getWallet(id)
  if (!target) {
    return 'Wallet not found.'
  }
  if (target.kind === 'hd') {
    return null
  }
  const secret = await getImportedPrivateKey(id)
  if (!secret) {
    return (
      `The private key for "${target.label}" is not stored in this app, ` +
      `so it cannot be re-activated after switching away. Re-import it to make it switchable.`
    )
  }
  return null
}

export const __testing = { sameAddress, KEYCHAIN_SERVICE, REGISTRY_KEY, ACTIVE_KEY }
