import { dialog } from 'electron'
import logger from '../../../logger'
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
  setFailoverSetting as setFailoverSettingMain
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
import { cfg } from '../../../../../orchestrator.config'

let authentication: Record<string, string> | null = null

export const validatePassword = (data) => auth.isValidPassword(data)

export const clearCache = () => {
  logger.verbose('Clearing database cache')
  return dbManager.getDb().dropDatabase().then(restart)
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
  logger.error('client-side error', data.message, data.stack)
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
  } catch (e) {
    console.log('Error', e)
    throw e
  }
}

export const getAllModels = async (): Promise<unknown[]> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/blockchain/models`
    const response = await fetch(path, {
      headers: await getAuthHeaders(),
      method: 'GET'
    })
    const data = await response.json()
    return data.models
  } catch (e) {
    console.log('Error', e)
    return []
  }
}

export const getBalances = async (): Promise<unknown[]> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/blockchain/balance`
    const response = await fetch(path, {
      headers: await getAuthHeaders()
    })
    const data = await response.json()
    return data
  } catch (e) {
    console.log('Error', e)
    return []
  }
}

export const sendEth = async (payload: {
  to: string
  amount: string
}): Promise<string | undefined> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/blockchain/send/eth`
    const response = await fetch(path, {
      method: 'POST',
      body: JSON.stringify({
        to: payload.to,
        amount: payload.amount
      }),
      headers: await getAuthHeaders()
    })
    const data = await response.json()
    return data.txHash
  } catch (e) {
    console.log('Error', e)
    return undefined
  }
}

export const sendMor = async (payload: {
  to: string
  amount: string
}): Promise<string | undefined> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/blockchain/send/mor`
    const response = await fetch(path, {
      method: 'POST',
      body: JSON.stringify({
        to: payload.to,
        amount: payload.amount
      }),
      headers: await getAuthHeaders()
    })
    const data = await response.json()
    return data.txHash
  } catch (e) {
    console.log('Error', e)
    return undefined
  }
}

export const getTransactions = async (payload: {
  page: number
  pageSize: number
}): Promise<unknown[]> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/blockchain/transactions?page=${payload.page}&limit=${payload.pageSize}`
    const response = await fetch(path, {
      headers: await getAuthHeaders()
    })
    const data = await response.json()
    return data.transactions
  } catch (e) {
    console.log('Error', e)
    return []
  }
}

export const getMorRate = async (payload?: {
  tokenAddress: string
  network: string
}): Promise<number | null> => {
  const tokenAddress = payload?.tokenAddress || '0x092baadb7def4c3981454dd9c0a0d7ff07bcfc86'
  const network = payload?.network || 'arbitrum'
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
  try {
    const path = `${config.chain.localProxyRouterUrl}/blockchain/sessions/budget`
    const response = await fetch(path, {
      headers: await getAuthHeaders()
    })
    const body = await response.json()
    return body.budget
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const getTokenSupply = async () => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/blockchain/token/supply`
    const response = await fetch(path, {
      headers: await getAuthHeaders()
    })
    const body = await response.json()
    return body.supply
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const getChatHistoryTitles = async (): Promise<ChatTitle[] | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/v1/chats`
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
  cid,
  destinationPath
}: {
  cid: string
  destinationPath: string
}): Promise<ResultResponse | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/download/${cid}`
    const response = await fetch(path, {
      headers: await getAuthHeaders(),
      method: 'POST',
      body: JSON.stringify({ destinationPath })
    })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const pinIpfsFile = async ({ cid }: { cid: string }): Promise<ResultResponse | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/pin`
    const response = await fetch(path, {
      method: 'POST',
      headers: await getAuthHeaders(),
      body: JSON.stringify({ cid })
    })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const unpinIpfsFile = async ({ cid }: { cid: string }): Promise<ResultResponse | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/unpin    `
    const response = await fetch(path, {
      method: 'POST',
      headers: await getAuthHeaders(),
      body: JSON.stringify({ cid })
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
}): Promise<{ hash: string; cid: string } | null> => {
  try {
    const path = `${config.chain.localProxyRouterUrl}/ipfs/add`
    const response = await fetch(path, {
      method: 'POST',
      headers: await getAuthHeaders(),
      body: JSON.stringify({ filePath }),
      signal: AbortSignal.timeout(10 * 60 * 1000) // 10 minutes timeout, because ipfs add can take a long time
    })
    const body = await response.json()
    return body
  } catch (e) {
    console.log('Error', e)
    return null
  }
}

export const getIpfsPinnedFiles = async (): Promise<{
  files: { cid: string; hash: string }[]
} | null> => {
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

export const startServices = async (_, core: Core) => {
  await new Orchestrator(
    cfg,
    (state) => {
      console.log('services state', state.startup)
      core.emitter.emit('services-state', state)
    },
    log
  ).startAll()
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

      console.log('Set Mnemonic To Wallet', await mnemonicRes.json())
    } else {
      const pKeyResp = await fetch(`${proxyUrl}/wallet/privateKey`, {
        method: 'POST',
        body: JSON.stringify({ PrivateKey: String(data.privateKey) }),
        headers: await getAuthHeaders()
      })
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
    .catch((err) => logger.error('onLoginSubmit err', err))
}

export async function openWallet(password: string, { emitter }: Core) {
  const storedAddress = wallet.getAddress()
  if (!storedAddress) {
    return
  }

  const { address } = storedAddress

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
