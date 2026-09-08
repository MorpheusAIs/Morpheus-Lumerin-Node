import { app, BrowserWindow, dialog } from 'electron'
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
import * as attachments from '../attachments'
import { cfg } from '../../../../../orchestrator.config'
import { OrchestratorConfig } from '../../../orchestrator/orchestrator.types'
import {
  validateAgentDecision,
  validateAgentToken,
  validateAgentUsername
} from './agentMutationSecurity'
import { InferenceTargetPayload, sessionInferenceHeaders } from './inference-session-target'
import { parseExistingSessionConflict } from './proxy-router-conflict'
import { showSessionConfirmation } from '../../../sessionConfirmation'
import type { SessionConfirmationDetails } from '../../../sessionConfirmationView'

let authentication: Record<string, string> | null = null
let orchestrator: Orchestrator | null = null
let sensitiveConfirmationOpen = false

async function confirmSessionAction(details: SessionConfirmationDetails): Promise<boolean> {
  if (sensitiveConfirmationOpen) throw new Error('Another security confirmation is already open.')
  sensitiveConfirmationOpen = true
  try {
    const owner =
      BrowserWindow.getFocusedWindow() ??
      BrowserWindow.getAllWindows().find((window) => !window.isDestroyed())
    return await showSessionConfirmation(owner, details)
  } finally {
    sensitiveConfirmationOpen = false
  }
}

async function confirmNativeAction(options: {
  title: string
  message: string
  detail: string
  confirmLabel: string
}): Promise<boolean> {
  if (sensitiveConfirmationOpen) throw new Error('Another security confirmation is already open.')
  sensitiveConfirmationOpen = true
  const dialogOptions = {
    type: 'warning' as const,
    title: options.title,
    message: options.message,
    detail: options.detail,
    buttons: ['Cancel', options.confirmLabel],
    defaultId: 0,
    cancelId: 0,
    noLink: true
  }
  const owner =
    BrowserWindow.getFocusedWindow() ??
    BrowserWindow.getAllWindows().find((window) => !window.isDestroyed())
  try {
    const result = owner
      ? await dialog.showMessageBox(owner, dialogOptions)
      : await dialog.showMessageBox(dialogOptions)
    return result.response === 1
  } finally {
    sensitiveConfirmationOpen = false
  }
}

function formatTokenAmount(wei: string): string {
  const padded = wei.padStart(19, '0')
  const whole = padded.slice(0, -18).replace(/^0+(?=\d)/, '') || '0'
  const fraction = padded.slice(-18).replace(/0+$/, '')
  return fraction ? `${whole}.${fraction}` : whole
}

/**
 * Error raised when the local proxy-router cannot be reached or returns a
 * non-2xx. Distinguished from a generic Error so the renderer can tell
 * "the node is down" apart from "the node said no".
 */
export class ProxyRouterError extends Error {
  readonly status?: number
  readonly unreachable: boolean
  readonly responseBody?: unknown

  constructor(
    message: string,
    opts: { status?: number; unreachable?: boolean; responseBody?: unknown } = {}
  ) {
    super(message)
    this.name = 'ProxyRouterError'
    this.status = opts.status
    this.unreachable = opts.unreachable ?? false
    this.responseBody = opts.responseBody
  }
}

export function configuredLoopbackProxyUrl(): string {
  let url: URL
  try {
    url = new URL(config.chain.localProxyRouterUrl)
  } catch {
    throw new ProxyRouterError('The configured local proxy-router URL is invalid.')
  }
  const hostname = url.hostname.replace(/^\[|\]$/g, '').toLowerCase()
  if (
    (url.protocol !== 'http:' && url.protocol !== 'https:') ||
    (hostname !== '127.0.0.1' && hostname !== 'localhost' && hostname !== '::1') ||
    url.username ||
    url.password
  ) {
    throw new ProxyRouterError('The proxy-router admin API must use a loopback URL.')
  }
  return url.origin
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
export async function proxyFetch<T>(
  pathname: string,
  init: RequestInit = {},
  label = pathname
): Promise<T> {
  const url = `${configuredLoopbackProxyUrl()}${pathname}`

  let response: Response
  try {
    response = await fetch(url, {
      ...init,
      headers: { ...(await getAuthHeaders(init.signal)), ...(init.headers ?? {}) }
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
    throw new ProxyRouterError(detail, { status: response.status, responseBody: body })
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

export const getAuthHeaders = async (signal?: AbortSignal | null) => {
  if (authentication) {
    return authentication
  }

  try {
    const path = `${configuredLoopbackProxyUrl()}/auth/cookie/path`
    const response = await fetch(path, { signal })
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

/**
 * Extracts text from a chat attachment (PDF, OOXML, SVG, plain text, code).
 *
 * Lives in main so the bounded document parsers stay out of the renderer bundle.
 */
export const parseAttachment = (params: {
  name: string
  mime: string
  data: ArrayBuffer | ArrayBufferView
}) => attachments.parseAttachment(params)

// ---------------------------------------------------------------------------
// Multi-wallet
// ---------------------------------------------------------------------------

/** Current wallet as the proxy-router sees it: address, storage kind, HD path. */
export const getActiveWallet = async (
  init: RequestInit = {}
): Promise<{
  address: string
  kind?: string
  derivationPath?: string
}> => {
  return proxyFetch('/wallet', init, 'active wallet')
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
  const entry = wallets.getWallet(params.walletId)
  if (entry) {
    const approved = await confirmNativeAction({
      title: 'Remove wallet',
      message: `Remove “${String(entry.label || 'Wallet').slice(0, 120)}” from MorpheusUI?`,
      detail:
        `Address: ${entry.address}\n\n` +
        'This removes the wallet record and, for imported wallets, its private key from the operating-system keychain. It does not move funds.',
      confirmLabel: 'Remove wallet'
    })
    if (!approved) throw new Error('Wallet removal cancelled.')
  }
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

const MODEL_PAGE_LIMIT_MAX = 100

export const getModelsPage = async (payload?: {
  offset?: number
  limit?: number
}): Promise<unknown[]> => {
  const offset = Number(payload?.offset ?? 0)
  const limit = Number(payload?.limit ?? MODEL_PAGE_LIMIT_MAX)
  if (!Number.isSafeInteger(offset) || offset < 0) {
    throw new Error('Model page offset is invalid.')
  }
  if (!Number.isSafeInteger(limit) || limit < 1 || limit > MODEL_PAGE_LIMIT_MAX) {
    throw new Error(`Model page limit must be between 1 and ${MODEL_PAGE_LIMIT_MAX}.`)
  }
  const data = await proxyFetch<{ models: unknown[] }>(
    `/blockchain/models?offset=${offset}&limit=${limit}&order=asc`,
    {},
    'models'
  )
  return data.models ?? []
}

const boundedProxyId = (value: unknown, label: string): string => {
  const result = String(value ?? '').trim()
  if (!result || result.length > 256 || /[\u0000-\u001f\u007f]/u.test(result)) {
    throw new Error(`${label} is invalid.`)
  }
  return result
}

const walletAddress = (value: unknown, label: string): string => {
  const result = String(value ?? '').trim()
  if (!/^0x[0-9a-fA-F]{40}$/u.test(result)) throw new Error(`${label} is invalid.`)
  return result
}

export const getProviders = async (): Promise<unknown[]> => {
  const data = await proxyFetch<{ providers: unknown[] }>('/blockchain/providers', {}, 'providers')
  return data.providers ?? []
}

export const getLocalModels = async (): Promise<unknown> => {
  return proxyFetch('/v1/models', {}, 'local models')
}

export const getNodeConfig = async (): Promise<unknown> => {
  return proxyFetch('/config', {}, 'proxy-router configuration')
}

export const updateEthNode = async (payload: { url: string }): Promise<unknown> => {
  const raw = String(payload?.url ?? '').trim()
  if (!raw || raw.length > 2_048) throw new Error('Enter a valid Ethereum node URL.')

  let endpoint: URL
  try {
    endpoint = new URL(raw)
  } catch {
    throw new Error('Enter a valid Ethereum node URL.')
  }
  if (!['http:', 'https:', 'ws:', 'wss:'].includes(endpoint.protocol)) {
    throw new Error('Ethereum node URLs must use HTTP(S) or WS(S).')
  }

  const approved = await confirmNativeAction({
    title: 'Change Ethereum node',
    message: 'Use this Ethereum RPC endpoint?',
    detail: `${endpoint.origin}\n\nThe proxy-router will trust this endpoint for blockchain reads and transaction submission.`,
    confirmLabel: 'Change endpoint'
  })
  if (!approved) throw new Error('Ethereum node change cancelled.')

  return proxyFetch(
    '/config/ethNode',
    { method: 'POST', body: JSON.stringify({ urls: [raw] }) },
    'Ethereum node update'
  )
}

export const getSessionsByUser = async (payload: { user: string }): Promise<unknown[]> => {
  const user = walletAddress(payload?.user, 'Wallet address')
  const sessions: unknown[] = []
  const limit = 50
  for (let offset = 0; offset < 1_000; offset += limit) {
    const data = await proxyFetch<{ sessions: unknown[] }>(
      `/blockchain/sessions/user?user=${encodeURIComponent(user)}&offset=${offset}&limit=${limit}&order=desc`,
      {},
      'user sessions'
    )
    const page = data.sessions ?? []
    sessions.push(...page)
    if (page.length !== limit) return sessions
  }
  return sessions
}

export const getBidsByModel = async (payload: { modelId: string }): Promise<unknown[]> => {
  const modelId = boundedProxyId(payload?.modelId, 'Model ID')
  const bids: unknown[] = []
  const limit = 50
  for (let offset = 0; offset < 1_000; offset += limit) {
    const data = await proxyFetch<{ bids: unknown[] }>(
      `/blockchain/models/${encodeURIComponent(modelId)}/bids/active?offset=${offset}&limit=${limit}&order=desc`,
      {},
      'model bids'
    )
    const page = data.bids ?? []
    bids.push(...page)
    if (page.length !== limit) return bids
  }
  return bids
}

export const getBidInfo = async (payload: { id: string }): Promise<unknown> => {
  const id = boundedProxyId(payload?.id, 'Bid ID')
  const data = await proxyFetch<{ bid: unknown }>(
    `/blockchain/bids/${encodeURIComponent(id)}`,
    {},
    'bid information'
  )
  return data.bid
}

export const closeSession = async (payload: { sessionId: string }): Promise<unknown> => {
  const sessionId = boundedProxyId(payload?.sessionId, 'Session ID')
  const approved = await confirmNativeAction({
    title: 'Close session',
    message: 'Close this Morpheus session?',
    detail:
      `Session: ${sessionId}\n\n` +
      'Closing ends inference for this session and returns unused escrow according to the contract.',
    confirmLabel: 'Close session'
  })
  if (!approved) throw new Error('Session close cancelled.')
  return proxyFetch(
    `/blockchain/sessions/${encodeURIComponent(sessionId)}/close`,
    { method: 'POST' },
    'session close'
  )
}

export const openSession = async (payload: {
  modelId: string
  duration: number
  directPayment?: boolean
  failover?: boolean
}): Promise<unknown> => {
  const modelId = boundedProxyId(payload?.modelId, 'Model ID')
  const duration = Number(payload?.duration)
  if (!Number.isSafeInteger(duration) || duration <= 0 || duration > 315_360_000) {
    throw new Error('Session duration is invalid.')
  }
  const directPayment = payload?.directPayment === true
  const failover = payload?.failover === true
  const approved = await confirmSessionAction({
    modelId,
    duration,
    directPayment,
    failover
  })
  if (!approved) throw new Error('Session opening cancelled.')
  try {
    return await proxyFetch(
      `/blockchain/models/${encodeURIComponent(modelId)}/session`,
      {
        method: 'POST',
        body: JSON.stringify({
          failover,
          sessionDuration: duration,
          directPayment,
          rejectExisting: true
        })
      },
      'session opening'
    )
  } catch (error) {
    const conflict = parseExistingSessionConflict(error)
    if (conflict) return conflict
    throw error
  }
}

export const getSessionsByProvider = async (payload: { provider: string }): Promise<unknown[]> => {
  const provider = walletAddress(payload?.provider, 'Provider address')
  const data = await proxyFetch<{ sessions: unknown[] }>(
    `/blockchain/sessions/provider?provider=${encodeURIComponent(provider)}`,
    {},
    'provider sessions'
  )
  return data.sessions ?? []
}

export const getProviderClaimableBalance = async (payload: {
  sessionId: string
}): Promise<unknown> => {
  const sessionId = boundedProxyId(payload?.sessionId, 'Session ID')
  const data = await proxyFetch<{ balance: unknown }>(
    `/proxy/sessions/${encodeURIComponent(sessionId)}/providerClaimableBalance`,
    {},
    'provider claimable balance'
  )
  return data.balance
}

export const claimProviderFunds = async (payload: { sessionId: string }): Promise<unknown> => {
  const sessionId = boundedProxyId(payload?.sessionId, 'Session ID')
  const approved = await confirmNativeAction({
    title: 'Claim provider funds',
    message: 'Submit the provider claim transaction?',
    detail: `Session: ${sessionId}\n\nThis submits an on-chain transaction and incurs gas.`,
    confirmLabel: 'Claim funds'
  })
  if (!approved) throw new Error('Provider claim cancelled.')
  return proxyFetch(
    `/proxy/sessions/${encodeURIComponent(sessionId)}/providerClaim`,
    { method: 'POST' },
    'provider claim'
  )
}

async function boundedResponseBody(response: Response, limit: number): Promise<Buffer> {
  const declared = Number(response.headers.get('content-length') ?? 0)
  if (Number.isFinite(declared) && declared > limit) throw new Error('Proxy response is too large.')
  if (!response.body) return Buffer.alloc(0)
  const reader = response.body.getReader()
  const chunks: Buffer[] = []
  let total = 0
  while (true) {
    const { value, done } = await reader.read()
    if (done) return Buffer.concat(chunks, total)
    if (!value) continue
    total += value.byteLength
    if (total > limit) {
      await reader.cancel().catch(() => undefined)
      throw new Error('Proxy response is too large.')
    }
    chunks.push(Buffer.from(value))
  }
}

async function inferenceFetch(
  pathname: '/v1/chat/completions' | '/v1/audio/speech' | '/v1/audio/transcriptions',
  init: RequestInit,
  timeoutMs: number
): Promise<Response> {
  const signal = AbortSignal.timeout(timeoutMs)
  try {
    return await fetch(`${configuredLoopbackProxyUrl()}${pathname}`, {
      ...init,
      signal,
      headers: { ...(await getAuthHeaders()), ...(init.headers ?? {}) }
    })
  } catch (error: any) {
    if (signal.aborted) throw new Error('The inference request timed out.')
    throw new ProxyRouterError(error?.message || 'Cannot reach the local proxy-router.', {
      unreachable: true
    })
  }
}

export const openChatCompletionStream = async (
  payload: {
    target: InferenceTargetPayload
    messages: unknown[]
  },
  signal: AbortSignal
): Promise<Response> => {
  if (
    !Array.isArray(payload?.messages) ||
    payload.messages.length < 1 ||
    payload.messages.length > 500
  ) {
    throw new Error('Chat messages are invalid.')
  }
  const requestBody = JSON.stringify({ stream: true, messages: payload.messages })
  if (Buffer.byteLength(requestBody, 'utf8') > 24 * 1024 * 1024) {
    throw new Error('Chat request exceeds the 24 MB limit.')
  }
  try {
    return await fetch(`${configuredLoopbackProxyUrl()}/v1/chat/completions`, {
      method: 'POST',
      headers: {
        ...(await getAuthHeaders()),
        ...sessionInferenceHeaders(payload.target),
        Accept: 'application/json',
        'Content-Type': 'application/json'
      },
      body: requestBody,
      signal
    })
  } catch (error: any) {
    if (signal.aborted) throw new Error('The chat request was cancelled or timed out.')
    throw new ProxyRouterError(error?.message || 'Cannot reach the local proxy-router.', {
      unreachable: true
    })
  }
}

export const synthesizeSpeech = async (payload: {
  target: InferenceTargetPayload
  text: string
  voice: string
  speed: number
}): Promise<{
  ok: boolean
  status: number
  mimeType: string
  data: ArrayBuffer
  error?: string
}> => {
  const text = String(payload?.text ?? '')
  const voice = String(payload?.voice ?? '').trim()
  const speed = Number(payload?.speed)
  if (!text.trim() || text.length > 20_000) throw new Error('Speech input is invalid.')
  if (!voice || voice.length > 100 || /[\u0000-\u001f\u007f]/u.test(voice)) {
    throw new Error('Speech voice is invalid.')
  }
  if (!Number.isFinite(speed) || speed < 0.25 || speed > 4) {
    throw new Error('Speech speed must be between 0.25 and 4.')
  }
  const response = await inferenceFetch(
    '/v1/audio/speech',
    {
      method: 'POST',
      headers: { ...sessionInferenceHeaders(payload.target), 'Content-Type': 'application/json' },
      body: JSON.stringify({ input: text, voice, response_format: 'mp3', speed })
    },
    5 * 60_000
  )
  const body = await boundedResponseBody(response, 20 * 1024 * 1024)
  return {
    ok: response.ok,
    status: response.status,
    mimeType: response.headers.get('content-type') ?? 'audio/mpeg',
    data: response.ok ? Uint8Array.from(body).buffer : new ArrayBuffer(0),
    ...(response.ok ? {} : { error: body.toString('utf8').slice(0, 4_000) })
  }
}

export const transcribeAudio = async (payload: {
  target: InferenceTargetPayload
  fileName: string
  mimeType: string
  data: ArrayBuffer | ArrayBufferView
}): Promise<{ ok: boolean; status: number; contentType: string; body: string }> => {
  const fileName =
    String(payload?.fileName ?? 'audio')
      .trim()
      .slice(0, 255) || 'audio'
  const mimeType = String(payload?.mimeType ?? 'application/octet-stream')
    .trim()
    .slice(0, 200)
  const data = payload?.data
  if (!(data instanceof ArrayBuffer) && !ArrayBuffer.isView(data)) {
    throw new Error('Audio data is invalid.')
  }
  const audio =
    data instanceof ArrayBuffer
      ? Buffer.from(data)
      : Buffer.from(data.buffer, data.byteOffset, data.byteLength)
  if (!audio.length || audio.length > 20 * 1024 * 1024) {
    throw new Error('Audio must be between 1 byte and 20 MB.')
  }
  const form = new FormData()
  form.append('file', new Blob([Uint8Array.from(audio)], { type: mimeType }), fileName)
  form.append('response_format', 'json')
  const response = await inferenceFetch(
    '/v1/audio/transcriptions',
    { method: 'POST', headers: sessionInferenceHeaders(payload.target), body: form },
    5 * 60_000
  )
  const body = (await boundedResponseBody(response, 4 * 1024 * 1024)).toString('utf8')
  return {
    ok: response.ok,
    status: response.status,
    contentType: response.headers.get('content-type') ?? 'text/plain',
    body
  }
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
  const to = String(payload?.to ?? '').trim()
  const amount = String(payload?.amount ?? '').trim()
  if (!/^0x[0-9a-fA-F]{40}$/.test(to)) throw new Error('Enter a valid 0x wallet address.')
  if (!/^[0-9]{1,78}$/.test(amount))
    throw new Error('Transfer amount must be a positive wei integer.')
  const amountWei = BigInt(amount)
  if (amountWei <= 0n || amountWei >= 2n ** 256n)
    throw new Error('Transfer amount is outside the supported range.')
  const symbol = token.toUpperCase()
  const approved = await confirmNativeAction({
    title: `Send ${symbol}`,
    message: `Send ${formatTokenAmount(amount)} ${symbol}?`,
    detail: `Recipient: ${to}\nRaw amount: ${amount} wei\n\nBlockchain transfers are irreversible. Verify the address and amount carefully.`,
    confirmLabel: `Send ${symbol}`
  })
  if (!approved) throw new Error('Transfer cancelled.')

  const path = `${configuredLoopbackProxyUrl()}/blockchain/send/${token}`
  const response = await fetch(path, {
    method: 'POST',
    body: JSON.stringify({
      to,
      amount
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
  const approved = await confirmNativeAction({
    title: 'Erase wallet data',
    message: 'Erase this MorpheusUI wallet and start over?',
    detail:
      'This clears the local wallet, node settings, and application cache, then restarts the app. Make sure you have the recovery material you need.',
    confirmLabel: 'Erase and restart'
  })
  if (!approved) throw new Error('Wallet reset cancelled.')
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

async function mutateAgentAccess(
  pathname:
    | '/auth/users/confirm'
    | '/auth/users'
    | '/auth/allowance/revoke'
    | '/auth/allowance/confirm',
  method: 'POST' | 'DELETE',
  payload: Record<string, string | boolean>,
  label: string
): Promise<boolean> {
  const result = await proxyFetch<ResultResponse>(
    pathname,
    {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    },
    label
  )
  if (result?.result !== true) {
    throw new ProxyRouterError(`The local proxy-router did not confirm ${label}.`)
  }
  return true
}

export const confirmDeclineAgentUser = async (params: {
  username: string
  confirm: boolean
}): Promise<boolean> => {
  const username = validateAgentUsername(params?.username)
  const confirm = validateAgentDecision(params?.confirm)
  const approved = await confirmNativeAction({
    title: confirm ? 'Approve agent access' : 'Decline agent access',
    message: `${confirm ? 'Approve' : 'Decline'} access for agent “${username}”?`,
    detail: confirm
      ? 'This grants the agent its requested proxy-router permissions and requested token allowances. Review the request in the app before approving.'
      : 'This rejects and removes the pending agent access request.',
    confirmLabel: confirm ? 'Approve agent' : 'Decline request'
  })
  if (!approved) throw new Error(`Agent ${confirm ? 'approval' : 'decline'} cancelled.`)
  return mutateAgentAccess(
    '/auth/users/confirm',
    'POST',
    { username, confirm },
    confirm ? 'agent access approval' : 'agent access denial'
  )
}

export const removeAgentUser = async (params: { username: string }): Promise<boolean> => {
  const username = validateAgentUsername(params?.username)
  const approved = await confirmNativeAction({
    title: 'Remove agent access',
    message: `Remove agent “${username}”?`,
    detail:
      'This revokes the agent credentials and its proxy-router access. Requests using those credentials will stop working.',
    confirmLabel: 'Remove agent'
  })
  if (!approved) throw new Error('Agent removal cancelled.')
  return mutateAgentAccess('/auth/users', 'DELETE', { username }, 'agent removal')
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
  const username = validateAgentUsername(params?.username)
  const token = validateAgentToken(params?.token)
  const approved = await confirmNativeAction({
    title: 'Revoke agent allowance',
    message: `Revoke ${token} allowance for agent “${username}”?`,
    detail: 'The agent will no longer be authorized to spend this token through the proxy-router.',
    confirmLabel: 'Revoke allowance'
  })
  if (!approved) throw new Error('Agent allowance revocation cancelled.')
  return mutateAgentAccess(
    '/auth/allowance/revoke',
    'POST',
    { username, token },
    'agent allowance revocation'
  )
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
  const username = validateAgentUsername(params?.username)
  const token = validateAgentToken(params?.token)
  const confirm = validateAgentDecision(params?.confirm)
  const approved = await confirmNativeAction({
    title: confirm ? 'Approve agent allowance' : 'Decline agent allowance',
    message: `${confirm ? 'Approve' : 'Decline'} ${token} allowance for agent “${username}”?`,
    detail: confirm
      ? 'This authorizes the pending token spending limit shown in the app. Verify the requested amount before approving.'
      : 'This rejects and removes the pending token allowance request.',
    confirmLabel: confirm ? 'Approve allowance' : 'Decline request'
  })
  if (!approved) throw new Error(`Agent allowance ${confirm ? 'approval' : 'decline'} cancelled.`)
  return mutateAgentAccess(
    '/auth/allowance/confirm',
    'POST',
    { username, token, confirm },
    confirm ? 'agent allowance approval' : 'agent allowance denial'
  )
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
    // Never trust a renderer-provided destination for wallet setup. This flow
    // sends the seed or imported key, so even a well-formed remote URL would be
    // credential exfiltration. The admin API is deliberately loopback-only.
    const proxyUrl = configuredLoopbackProxyUrl()

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

    const activeAddress = await fetch(`${proxyUrl}/wallet`, {
      method: 'GET',
      headers: await getAuthHeaders()
    })
      .then((res) => res.json())
      .then((res) => res.address)
    const verifiedAddress = walletAddress(activeAddress, 'Wallet address')

    console.log('Wallet Address Is', verifiedAddress)

    wallet.setSeed(verifiedAddress, data.password)
    wallet.setAddress(verifiedAddress)
    core.emitter.emit('create-wallet', { address: verifiedAddress })
    await openWallet(data.password, core, verifiedAddress)
  } catch (err) {
    return { error: new WalletError('Onboarding unable to be completed: ', err) }
  }
}

export const onLoginSubmit = async ({ password }, core: Core) => {
  const isValid = config.chain.bypassAuth ? true : await auth.isValidPassword(password)
  if (!isValid) {
    return { error: new WalletError('Invalid password') }
  }

  try {
    return await openWallet(password, core)
  } catch (err) {
    log.error('onLoginSubmit err', err)
    const message =
      err instanceof ProxyRouterError && err.unreachable
        ? 'Cannot connect to the local proxy-router yet. Wait a moment and try again; if it persists, restart MorpheusUI.'
        : err instanceof ProxyRouterError
          ? err.message
          : 'Unable to open wallet.'
    return { error: new WalletError(message, err) }
  }
}

export async function openWallet(
  password: string,
  { emitter }: Core,
  knownAddress?: string
) {
  const storedAddress = wallet.getAddress()
  const storedAddressValue = (storedAddress as { address?: unknown } | undefined)?.address
  const previousAddress = typeof storedAddressValue === 'string' ? storedAddressValue : undefined
  const activeAddress =
    knownAddress ??
    (await getActiveWallet({ signal: AbortSignal.timeout(2_000) }))?.address

  let address: string
  try {
    address = walletAddress(activeAddress, 'Wallet address')
  } catch {
    throw new ProxyRouterError(
      'The local proxy-router did not return a valid wallet address. Please wait a moment and try again.'
    )
  }

  // The managed proxy-router owns the key that can sign. Never let a stale
  // desktop pointer make the UI display/query wallet A while the node signs as
  // wallet B. Repair the non-secret pointer when the node reports a change.
  if (previousAddress?.toLowerCase() !== address.toLowerCase()) {
    wallet.setAddress(address)
  }

  const openedWallet = { address, isActive: true }
  emitter.emit('open-wallet', openedWallet)
  emitter.emit('open-proxy-router', { password })
  return openedWallet
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
