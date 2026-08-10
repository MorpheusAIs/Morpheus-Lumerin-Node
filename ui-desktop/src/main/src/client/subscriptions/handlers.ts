import { app, dialog } from 'electron'
import restart from '../electron-restart'
import dbManager from '../database'
import storage from '../storage'
import auth from '../auth'
import wallet from '../wallet'
import {
  setProxyRouterConfig,
  getProxyRouterConfig,
  getDefaultCurrencySetting,
  setDefaultCurrencySetting,
  getKey,
  setKey,
  getFailoverSetting,
  setFailoverSetting as setFailoverSettingMain,
  setPasswordHash
} from '../settings'
import config from '../../../config'
import os from 'node:os'
import fs from 'node:fs'
import {
  AgentAllowanceRequestsRes,
  AgentTxRes,
  AgentUserRes,
  ChatHistory,
  ChatTitle,
  ResultResponse
} from './api.types'
import { Orchestrator } from '../../../orchestrator/orchestrator'
import log from '../../../logger'
import { Core } from './core.types'
import WalletError from '../../client/WalletError'
import keys from '../keys'
import * as wallets from '../wallets'
import { cfg } from '../../../../../orchestrator.config'
import { OrchestratorConfig } from '../../../orchestrator/orchestrator.types'

let authentication: Record<string, string> | null = null
let orchestrator: Orchestrator | null = null

/**
 * Error raised when the local proxy-router cannot be reached or returns a
 * non-2xx. Distinguished from a generic Error so the renderer can tell
 * "the node is down" apart from "the node said no".
 */
export class ProxyRouterError extends Error {
  readonly status?: number
  readonly unreachable: boolean

  constructor(message: string, opts: { status?: number; unreachable?: boolean } = {}) {
    super(message)
    this.name = 'ProxyRouterError'
    this.status = opts.status
    this.unreachable = opts.unreachable ?? false
  }
}

/**
 * Single entry point for proxy-router reads.
 *
 * Every one of these calls used to be wrapped in `catch { return [] }` or
 * `catch { return null }`. When the proxy-router was down or auth was broken,
 * the UI therefore rendered a zero balance, an empty model list and no
 * transactions — indistinguishable from "you genuinely have nothing". That is
 * the single most damaging behaviour in the app: it is why users conclude
 * their MOR has disappeared.
 *
 * Failing loudly here lets react-query put the affected view into an error
 * state that says what is actually wrong.
 */
async function proxyFetch<T>(
  pathname: string,
  init: RequestInit = {},
  label = pathname
): Promise<T> {
  const url = `${config.chain.localProxyRouterUrl}${pathname}`

  let response: Response
  try {
    response = await fetch(url, {
      ...init,
      headers: { ...(await getAuthHeaders()), ...(init.headers ?? {}) }
    })
  } catch (e: any) {
    log.error(`proxy-router unreachable for ${label}:`, e?.message ?? e)
    throw new ProxyRouterError(
      'Cannot reach the local proxy-router. Check that it is running in Settings.',
      { unreachable: true }
    )
  }

  const body = await response.json().catch(() => null)

  if (!response.ok) {
    const detail = (body && (body.error || body.message)) || `HTTP ${response.status}`
    log.error(`proxy-router error for ${label}: ${detail}`)
    throw new ProxyRouterError(detail, { status: response.status })
  }

  return body as T
}

export const validatePassword = (data) => auth.isValidPassword(data)

export const clearCache = () => {
  log.verbose('Clearing database cache')
  return dbManager.getDb().dropDatabase().then(restart)
}

export const clearCacheV2 = () => {
  log.verbose('Clearing database cache')
  return dbManager.getDb().dropDatabase()
}

export const persistState = (data) => storage.persistState(data).then(() => true)

export const changePassword = ({ oldPassword, newPassword }) => {
  return validatePassword(oldPassword).then(function (isValid) {
    if (!isValid) {
      return isValid
    }
    return auth.setPassword(newPassword).then(function () {
      const seed = wallet.getSeed(oldPassword)
      wallet.setSeed(seed, newPassword)

      return true
    })
  })
}

export const saveProxyRouterSettings = (data) => Promise.resolve(setProxyRouterConfig(data))

export const getProxyRouterSettings = async () => {
  return getProxyRouterConfig()
}

export const handleClientSideError = (data) => {
  log.error('client-side error', data.message, data.stack)
}

export const getDefaultCurrency = async () => getDefaultCurrencySetting()
export const setDefaultCurrency = async (curr) => setDefaultCurrencySetting(curr)

export const getCustomEnvs = async () => getKey('customEnvs')
export const setCustomEnvs = async (value) => setKey('customEnvs', value)

export const getProfitSettings = async () =>
  getKey('profitSettings') || {
    deviation: 2,
    target: 10,
    adaptExisting: false
  }
export const setProfitSettings = async (value) => setKey('profitSettings', value)

export const getAutoAdjustPriceData = async () => getKey('autoAdjustPriceData')
export const setAutoAdjustPriceData = async (value) => {
  const oldData = await getAutoAdjustPriceData()
  setKey('autoAdjustPriceData', {
    ...oldData,
    ...value
  })
}

export const getContractHashrate = async (params: { contractId: string; fromDate: Date }) => {
  const { contractId, fromDate } = params
  const collection = await dbManager.getDb().collection('hashrate').findAsync({ id: contractId })
  return collection
    .filter((x) => x.timestamp > fromDate.getTime())
    .sort((a, b) => a.timestamp - b.timestamp)
}

export const isFailoverEnabled = async () => {
  const settings = await getFailoverSetting()
  if (!settings) {
    return { isEnabled: config.isFailoverEnabled }
  }
  return settings
}

export const setFailoverSetting = (params) => setFailoverSettingMain(params)

export const restartWallet = () => restart(1)

export const openSelectFolderDialog = () => {
  return dialog.showOpenDialog({
    properties: ['openDirectory']
  })
}

export const getAuthHeaders = async () => {
  if (authentication) {
    return authentication
  }

  try {
    const path = `${config.chain.localProxyRouterUrl}/auth/cookie/path`
    const response = await fetch(path)
    const body = await response.json()
    let cookieFilePath = body.path

    const isWindows = os.platform() === 'win32'
    cookieFilePath = isWindows ? cookieFilePath.replace(/\//g, '\\') : cookieFilePath

    const cookieFile = fs.readFileSync(cookieFilePath, 'utf8').trim()
    const [username, password] = cookieFile.split(':')
    authentication = {
      Authorization: `Basic ${Buffer.from(`${username}:${password}`, 'utf-8').toString('base64')}`
    }
    return authentication
  } catch (e: any) {
    // This is the first call every other request depends on, so it is also the
    // most common failure when the node isn't up yet. Give it a message a user
    // can act on rather than propagating a bare ECONNREFUSED.
    log.error('failed to read proxy-router auth cookie:', e?.message ?? e)
    throw new ProxyRouterError(
      'Cannot authenticate with the local proxy-router. It may still be starting up — check Settings.',
      { unreachable: true }
    )
  }
}

/**
 * Clears the cached Basic-auth header.
 *
 * The credentials are read once and memoised for the process lifetime, so a
 * proxy-router restart (which regenerates the cookie) would otherwise leave
 * every subsequent request failing with 401 until the app itself restarted.
 */
export const resetAuthHeaders = () => {
  authentication = null
}

// ---------------------------------------------------------------------------
// Multi-wallet
// ---------------------------------------------------------------------------

/** Current wallet as the proxy-router sees it: address, storage kind, HD path. */
export const getActiveWallet = async (): Promise<{
  address: string
  kind?: string
  derivationPath?: string
}> => {
  return proxyFetch('/wallet', {}, 'active wallet')
}

/**
 * Returns the wallet list, adopting the proxy-router's current wallet if the
 * registry is empty (i.e. this install predates multi-wallet support).
 */
export const getWallets = async () => {
  const active = await getActiveWallet()

  if (active?.address && !wallets.listWallets().length) {
    wallets.adoptCurrentWallet({
      address: active.address,
      kind: active.kind ?? 'privateKey',
      derivationPath: active.derivationPath,
    })
  }

  // Keep the active marker honest: the proxy-router is the source of truth for
  // which key is loaded, not our stored pointer.
  const list = wallets.listWallets()
  const match = list.find(
    (w) => w.address?.toLowerCase() === active?.address?.toLowerCase()
  )
  if (match && wallets.getActiveWalletId() !== match.id) {
    wallets.setActiveWalletId(match.id)
  }

  return {
    wallets: list,
    activeId: match?.id ?? wallets.getActiveWalletId(),
    activeAddress: active?.address,
    // HD accounts can only be added when the proxy-router holds a mnemonic.
    canAddHd: active?.kind === 'mnemonic',
    nextHdIndex: wallets.nextHdIndex()
  }
}

/**
 * Adds the next HD account.
 *
 * Implemented as switch-read-switch-back: the proxy-router is the only party
 * that can derive from the seed, and it exposes no "derive without activating"
 * call. So we point it at the candidate path, read the resulting address, and
 * restore the previous path. Deliberately NOT left on the new account —
 * discovering an address should not silently move the user's funds context.
 */
export const addHdWallet = async (params: { label?: string }) => {
  const before = await getActiveWallet()
  if (before.kind !== 'mnemonic') {
    throw new Error(
      'This wallet was imported from a private key, so it has no seed phrase to derive further accounts from. Import another wallet instead.'
    )
  }

  const index = String(wallets.nextHdIndex())
  const previousPath = before.derivationPath ?? '0'

  const derived = await proxyFetch<{ address: string }>(
    '/wallet/derivationPath',
    { method: 'POST', body: JSON.stringify({ derivationPath: index }) },
    'derive HD account'
  )

  try {
    return wallets.addHdWallet({
      address: derived.address,
      derivationPath: index,
      label: params?.label
    })
  } finally {
    await proxyFetch(
      '/wallet/derivationPath',
      { method: 'POST', body: JSON.stringify({ derivationPath: previousPath }) },
      'restore derivation path'
    ).catch((e) => log.error('failed to restore previous derivation path', e))
  }
}

/** Imports an unrelated private key without activating it. */
export const importWallet = async (params: { privateKey: string; label?: string }) => {
  const privateKey = String(params.privateKey ?? '').trim()
  if (!/^(0x)?[0-9a-fA-F]{64}$/.test(privateKey)) {
    throw new Error('That is not a valid private key (expected 64 hex characters).')
  }
  const normalised = privateKey.startsWith('0x') ? privateKey : `0x${privateKey}`

  // Derive the address locally — importing must not disturb the active wallet.
  const address = keys.privateKeyToAddress(normalised)

  return wallets.addImportedWallet({
    address,
    privateKey: normalised,
    label: params.label
  })
}

/**
 * Makes `walletId` the proxy-router's active key.
 *
 * The proxy-router restarts its session machinery on key change (proxyctl
 * watches PrivateKeyUpdated), so callers should treat every address-scoped
 * cache as invalid afterwards.
 */
export const switchWallet = async (params: { walletId: string }, core: Core) => {
  const target = wallets.getWallet(params.walletId)
  if (!target) {
    throw new Error('Wallet not found.')
  }

  const blocker = await wallets.assertSwitchable(params.walletId)
  if (blocker) {
    throw new Error(blocker)
  }

  if (target.kind === 'hd') {
    await proxyFetch(
      '/wallet/derivationPath',
      {
        method: 'POST',
        body: JSON.stringify({ derivationPath: target.derivationPath ?? '0' })
      },
      'switch HD account'
    )
  } else {
    const privateKey = await wallets.getImportedPrivateKey(params.walletId)
    if (!privateKey) {
      throw new Error('The private key for this wallet is no longer in the keychain.')
    }
    await proxyFetch(
      '/wallet/privateKey',
      { method: 'POST', body: JSON.stringify({ privateKey }) },
      'switch wallet'
    )
  }

  wallets.setActiveWalletId(params.walletId)

  // The proxy-router tears down and restarts its session machinery when the key
  // changes (proxyctl watches PrivateKeyUpdated), so /wallet can briefly report
  // the old address or refuse the connection while that happens. Poll until it
  // settles rather than reading once and reporting a spurious mismatch.
  const now = await waitForAddress(target.address)
  if (!now) {
    log.error(
      `wallet switch mismatch: expected ${target.address} but the proxy-router never reported it`
    )
    throw new Error(
      'The node did not switch to the expected address. Check Settings, then try again.'
    )
  }

  // The renderer keeps the active address in redux, populated by the
  // 'create-wallet' event (see store/reducers/wallet.jsx). Without this the
  // switch succeeds at the node but every selector — and therefore every
  // address-scoped query key — keeps pointing at the previous wallet.
  wallet.setAddress(now)
  core?.emitter?.emit('create-wallet', { address: now })

  return { address: now, walletId: params.walletId }
}

/** Polls GET /wallet until it reports `expected`, or the budget runs out. */
async function waitForAddress(expected: string, timeoutMs = 20000): Promise<string | null> {
  const deadline = Date.now() + timeoutMs
  let last: string | undefined

  while (Date.now() < deadline) {
    try {
      const res = await getActiveWallet()
      last = res.address
      if (res.address?.toLowerCase() === expected.toLowerCase()) {
        return res.address
      }
    } catch (e) {
      // Expected while the router is mid-restart; keep waiting.
    }
    await new Promise((r) => setTimeout(r, 400))
  }

  log.error(`waitForAddress timed out; last seen ${last ?? 'none'}, wanted ${expected}`)
  return null
}

export const removeWallet = async (params: { walletId: string }) => {
  await wallets.removeWallet(params.walletId)
  return true
}

export const renameWallet = async (params: { walletId: string; label: string }) => {
  return wallets.renameWallet(params.walletId, params.label)
}

export const getAllModels = async (): Promise<unknown[]> => {
  const data = await proxyFetch<{ models: unknown[] }>('/blockchain/models', {}, 'models')
  return data.models ?? []
}

export const getBalances = async (): Promise<unknown> => {
  return proxyFetch('/blockchain/balance', {}, 'balances')
}

/**
 * Shared transfer helper for the two send endpoints.
 *
 * `amount` must be a base-10 **wei** string — the proxy-router decodes it into
 * a big.Int (see lib.BigInt.UnmarshalJSON), so decimals or exponent notation
 * are rejected server-side.
 *
 * Unlike most handlers in this file, transfers deliberately throw on failure.
 * Silently returning `undefined` for a money movement is the worst possible
 * outcome: the user can't tell "rejected" from "broadcast but not yet mined".
 */
const sendToken = async (
  token: 'eth' | 'mor',
  payload: { to: string; amount: string }
): Promise<string> => {
  const path = `${config.chain.localProxyRouterUrl}/blockchain/send/${token}`
  const response = await fetch(path, {
    method: 'POST',
    body: JSON.stringify({
      to: payload.to,
      amount: payload.amount
    }),
    headers: await getAuthHeaders()
  })

  const data = await response.json().catch(() => ({}))

  if (!response.ok) {
    throw new Error(data?.error || `Transfer failed (HTTP ${response.status})`)
  }
  if (!data?.tx) {
    throw new Error('Transfer did not return a transaction hash')
  }
  return data.tx
}

export const sendEth = async (payload: { to: string; amount: string }): Promise<string> =>
  sendToken('eth', payload)

export const sendMor = async (payload: { to: string; amount: string }): Promise<string> =>
  sendToken('mor', payload)

export const getTransactions = async (payload: {
  page: number
  pageSize: number
}): Promise<unknown[]> => {
  const data = await proxyFetch<{ transactions: unknown[] }>(
    `/blockchain/transactions?page=${payload.page}&limit=${payload.pageSize}`,
    {},
    'transactions'
  )
  return data.transactions ?? []
}

export const getMorRate = async (payload?: {
  tokenAddress: string
  network: string
}): Promise<number | null> => {
  const tokenAddress = payload?.tokenAddress || '0x7431ada8a591c955a994a21710752ef9b882b8e3'
  const network = payload?.network || 'base'
  try {
    const path = `https://api.geckoterminal.com/api/v2/simple/networks/${network}/token_price/${tokenAddress}`
    const response = await fetch(path)
    const body = await response.json()
    return body.data.attributes.token_prices[tokenAddress]
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const getTodaysBudget = async () => {
  const body = await proxyFetch<{ budget: unknown }>('/blockchain/sessions/budget', {}, 'budget')
  return body.budget
}

export const getTokenSupply = async () => {
  const body = await proxyFetch<{ supply: unknown }>('/blockchain/token/supply', {}, 'supply')
  return body.supply
}

export const getChatHistoryTitles = async (): Promise<ChatTitle[]> => {
  return proxyFetch<ChatTitle[]>('/v1/chats', {}, 'chat titles')
}

export const getChatHistory = async (chatId: string): Promise<ChatHistory | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/v1/chats/${chatId}`
    const response = await fetch(path, {
      headers: await getAuthHeaders()
    })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const deleteChatHistory = async (chatId: string): Promise<boolean> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/v1/chats/${chatId}`
    const response = await fetch(path, {
      method: 'DELETE',
      headers: await getAuthHeaders()
    })
    const body = await response.json()
    return body.result
  } catch (e) {
    console.log('Error', e)
    return false
  }
}

export const updateChatHistoryTitle = async (params: {
  id: string
  title: string
}): Promise<boolean> => {
  const { id, title } = params
  try {
    const path = `${config.chain.localProxyRouterUrl}/v1/chats/${id}`
    const response = await fetch(path, {
      method: 'POST',
      body: JSON.stringify({ title }),
      headers: await getAuthHeaders()
    })
    const body = await response.json()
    return body.result
  } catch (e) {
    console.log('Error', e)
    return false
  }
}

export const checkProviderConnectivity = async (params: {
  address: string
  endpoint: string
}): Promise<boolean> => {
  const { address, endpoint } = params
  try {
    const path = `${config.chain.localProxyRouterUrl}/proxy/provider/ping`
    const response = await fetch(path, {
      method: 'POST',
      body: JSON.stringify({
        providerAddr: address,
        providerUrl: endpoint
      }),
      headers: await getAuthHeaders()
    })

    if (!response.ok) {
      return false
    }

    const body = await response.json()
    return !!body.ping
  } catch (e) {
    console.log('checkProviderConnectivity: Error', e)
    return false
  }
}

export const clearEthNodeEnv = async () => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/config/ethNode`
    const response = await fetch(path, { method: 'DELETE', headers: await getAuthHeaders() })
    const data = await response.json()
    return data.status
  } catch (e) {
    console.log('CLEAR ETH NODE ERROR', e)
    return false
  }
}

export const clearWallet = async () => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/wallet`
    const response = await fetch(path, { method: 'DELETE', headers: await getAuthHeaders() })
    const data = await response.json()
    return data.status
  } catch (e) {
    console.log('CLEAR WALLET ERROR', e)
    return false
  }
}

export const resetWallet = async () => {
  await clearWallet()
  await clearEthNodeEnv()
  await clearCacheV2()
  setPasswordHash('')
  app.relaunch()
  app.quit()
}

export const getAgentUsers = async (): Promise<AgentUserRes | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/auth/users`
    const response = await fetch(path, { method: 'GET', headers: await getAuthHeaders() })
    return await response.json()
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const confirmDeclineAgentUser = async (params: {
  username: string
  confirm: boolean
}): Promise<boolean> => {
  const { username, confirm } = params
  try {
    const path = `${config.chain.localProxyRouterUrl}/auth/users/confirm`

    const res = await fetch(path, {
      method: 'POST',
      body: JSON.stringify({ username, confirm }),
      headers: await getAuthHeaders()
    })
    await res.json()
    return true
  } catch (e) {
    console.log('Error', e)
    return false
  }
}

export const removeAgentUser = async (params: { username: string }): Promise<boolean> => {
  const { username } = params
  try {
    const path = `${config.chain.localProxyRouterUrl}/auth/users`
    await fetch(path, {
      method: 'DELETE',
      body: JSON.stringify({ username }),
      headers: await getAuthHeaders()
    })
    return true
  } catch (e) {
    console.log('Error', e)
    return false
  }
}

export const getAgentTxs = async (params: {
  username: string
  cursor: string
  limit: number
}): Promise<AgentTxRes | null> => {
  try {
    const query = new URLSearchParams()
    query.set('cursor', params.cursor)
    query.set('limit', params.limit.toString())

    const path = `${config.chain.localProxyRouterUrl}/auth/users/${encodeURIComponent(params.username)}/txs?${query.toString()}`
    const response = await fetch(path, {
      headers: await getAuthHeaders()
    })
    if (response.ok) {
      return await response.json()
    }
    throw new Error(await response.text())
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const revokeAgentAllowance = async (params: {
  username: string
  token: string
}): Promise<boolean> => {
  const { username, token } = params
  try {
    const path = `${config.chain.localProxyRouterUrl}/auth/allowance/revoke`
    await fetch(path, {
      method: 'POST',
      body: JSON.stringify({ username, token }),
      headers: await getAuthHeaders()
    })
    return true
  } catch (e) {
    console.log('Error', e)
    return false
  }
}

export const getAgentAllowanceRequests = async (): Promise<AgentAllowanceRequestsRes | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/auth/allowance/requests`
    const response = await fetch(path, { headers: await getAuthHeaders() })
    const data = await response.json()
    return data
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const confirmDeclineAgentAllowanceRequest = async (params: {
  username: string
  token: string
  confirm: boolean
}): Promise<boolean> => {
  const { username, token, confirm } = params
  try {
    const path = `${config.chain.localProxyRouterUrl}/auth/allowance/confirm`
    await fetch(path, {
      method: 'POST',
      body: JSON.stringify({ username, token, confirm }),
      headers: await getAuthHeaders()
    })
    return true
  } catch (e) {
    console.log('Error', e)
    return false
  }
}

export const getIpfsVersion = async (): Promise<{ version: string } | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/version`
    const response = await fetch(path, { headers: await getAuthHeaders() })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const getIpfsFile = async ({
  cidHash,
  destinationPath
}: {
  cidHash: string
  destinationPath: string
}): Promise<ResultResponse | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/download/${cidHash}?dest=${encodeURIComponent(destinationPath)}`
    const response = await fetch(path, {
      headers: await getAuthHeaders(),
      method: 'GET'
    })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const pinIpfsFile = async ({ cidHash }: { cidHash: string }): Promise<ResultResponse | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/pin`
    const response = await fetch(path, {
      method: 'POST',
      headers: await getAuthHeaders(),
      body: JSON.stringify({ cidHash })
    })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const unpinIpfsFile = async ({ cidHash }: { cidHash: string }): Promise<ResultResponse | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/unpin`
    const response = await fetch(path, {
      method: 'POST',
      headers: await getAuthHeaders(),
      body: JSON.stringify({ cidHash })
    })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const addFileToIpfs = async ({
  filePath
}: {
  filePath: string
}): Promise<{
  fileCID: string
  metadataCID: string
  fileCIDHash: string
  metadataCIDHash: string
} | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/add`
    const response = await fetch(path, {
      method: 'POST',
      headers: await getAuthHeaders(),
      body: JSON.stringify({ filePath }),
      signal: AbortSignal.timeout(10 * 60 * 1000)
    })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const getIpfsPinnedFiles = async (): Promise<
  {
    fileName: string
    fileSize: number
    fileCID: string
    fileCIDHash: string
    tags: string[]
    id: string
    modelName: string
    metadataCID: string
    metadataCIDHash: string
  }[] | null
> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/pin`
    const response = await fetch(path, { headers: await getAuthHeaders() })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

const getOrchestrator = (core: Core): Orchestrator => {
  if (!orchestrator) {
    orchestrator = new Orchestrator(
      cfg,
      (state) => {
        core.emitter.emit('services-state', state)
      },
      log
    )
  }
  return orchestrator
}

export const startServices = async (_, core: Core) => {
  await getOrchestrator(core).startAll()
}

export const restartService = async (data: { service: keyof OrchestratorConfig }, core: Core) => {
  await getOrchestrator(core).restartService(data.service)
  // The proxy-router regenerates its auth cookie on start, so the memoised
  // header is stale the moment it comes back up. Without this, every request
  // after a restart failed with 401 until the whole app was relaunched.
  if (data.service === 'proxyRouter') {
    resetAuthHeaders()
  }
}

export const pingService = async (data: { service: keyof OrchestratorConfig }, core: Core) => {
  return await getOrchestrator(core).ping(data.service)
}

export const onboardingCompleted = async (data, core: Core) => {
  try {
    const { proxyUrl } = data

    if (data.ethNode) {
      const ethNodeResult = await fetch(`${proxyUrl}/config/ethNode`, {
        method: 'POST',
        body: JSON.stringify({ urls: [data.ethNode] }),
        headers: await getAuthHeaders()
      })

      const dataResponse = await ethNodeResult.json()
      if (dataResponse.error) {
        return dataResponse.error
      }
    }

    await auth.setPassword(data.password)

    if (data.mnemonic) {
      const mnemonicRes = await fetch(`${proxyUrl}/wallet/mnemonic`, {
        method: 'POST',
        body: JSON.stringify({
          mnemonic: data.mnemonic,
          derivationPath: String(data.derivationPath || 0)
        }),
        headers: await getAuthHeaders()
      })
      if (!mnemonicRes.ok) {
        throw new Error(await mnemonicRes.text())
      }

      console.log('Set Mnemonic To Wallet', await mnemonicRes.json())
    } else {
      const pKeyResp = await fetch(`${proxyUrl}/wallet/privateKey`, {
        method: 'POST',
        body: JSON.stringify({ privateKey: String(data.privateKey) }),
        headers: await getAuthHeaders()
      })
      if (!pKeyResp.ok) {
        throw new Error(await pKeyResp.text())
      }
      console.log('Set Private Key To Wallet', await pKeyResp.json())
    }

    const walletAddress = await fetch(`${proxyUrl}/wallet`, {
      method: 'GET',
      headers: await getAuthHeaders()
    })
      .then((res) => res.json())
      .then((res) => res.address)

    console.log('Wallet Address Is', walletAddress)

    wallet.setSeed(walletAddress, data.password)
    wallet.setAddress(walletAddress)
    core.emitter.emit('create-wallet', { address: walletAddress })
    openWallet(data.password, core)
  } catch (err) {
    return { error: new WalletError('Onboarding unable to be completed: ', err) }
  }
}

export const onLoginSubmit = ({ password }, core: Core) => {
  var checkPassword = config.chain.bypassAuth
    ? new Promise((r) => r(true))
    : auth.isValidPassword(password)

  return checkPassword
    .then(function (isValid) {
      if (!isValid) {
        return { error: new WalletError('Invalid password') }
      }
      openWallet(password, core)

      return isValid
    })
    .catch((err) => log.error('onLoginSubmit err', err))
}

export async function openWallet(password: string, { emitter }: Core) {
  const storedAddress = wallet.getAddress()
  if (!storedAddress) {
    return
  }

  const { address } = storedAddress as { address?: string }

  emitter.emit('open-wallet', { address, isActive: true })
  emitter.emit('open-proxy-router', { password })
}

export const suggestAddresses = async (mnemonic: string) => {
  const seed = keys.mnemonicToSeedHex(mnemonic)
  let results: any[] = []
  for (let i = 0; i < 10; i++) {
    const walletAddress = wallet.createAddress(seed, i)
    results.push(walletAddress)
  }
  return results
}

export const quitApp = async () => {
  app.quit()
}
