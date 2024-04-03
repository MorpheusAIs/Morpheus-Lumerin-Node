import logger from '../../../logger'

const pTimeout = require('p-timeout')
import auth from '../auth'
import keys from '../keys'
import config from '../../../config'
import wallet from '../wallet'
import noCore from './no-core'
import WalletError from '../WalletError'
import { setProxyRouterConfig, cleanupDb, getProxyRouterConfig } from '../settings'

export const withAuth =
  (fn) =>
  (data, { api }) => {
    if (typeof data.walletId !== 'string') {
      throw new WalletError('walletId is not defined')
    }

    return auth
      .isValidPassword(data.password)
      .then(() => {
        return wallet.getSeed(data.password)
      })
      .then((seed, index) => {
        return api.wallet.createPrivateKey(seed, index)
      })
      .then((privateKey) => fn(privateKey, data))
  }

export const createContract = async function (data, { api }) {
  data.walletId = wallet.getAddress().address
  data.password = await auth.getSessionPassword()

  if (typeof data.walletId !== 'string') {
    throw new WalletError('WalletId is not defined')
  }
  return withAuth((privateKey) =>
    api.contracts.createContract({
      price: data.price,
      speed: data.speed,
      duration: data.duration,
      profit: data.profit,
      sellerAddress: data.sellerAddress,
      password: data.password,
      privateKey
    })
  )(data, { api })
}

export const purchaseContract = async function (data, { api }) {
  data.walletId = wallet.getAddress().address
  data.minerPassword = data.password
  data.password = await auth.getSessionPassword()

  if (typeof data.walletId !== 'string') {
    throw new WalletError('WalletId is not defined')
  }

  return withAuth((privateKey) =>
    api.contracts.purchaseContract({
      ...data,
      privateKey
    })
  )(data, { api })
}

export const editContract = async function (data, { api }) {
  data.walletId = wallet.getAddress().address
  data.password = await auth.getSessionPassword()

  if (typeof data.walletId !== 'string') {
    throw new WalletError('WalletId is not defined')
  }
  return withAuth((privateKey) =>
    api.contracts.editContract({
      contractId: data.id,
      price: data.price,
      speed: data.speed,
      duration: data.duration,
      profit: data.profit,
      password: data.password,
      walletId: data.walletId,
      privateKey
    })
  )(data, { api })
}

export const claimFaucet = async function (data, { api }) {
  data.walletId = wallet.getAddress().address
  data.password = await auth.getSessionPassword()

  if (typeof data.walletId !== 'string') {
    throw new WalletError('WalletId is not defined')
  }

  return withAuth((privateKey) =>
    api.token.claimFaucet({
      ...data,
      privateKey
    })
  )(data, { api })
}

export const cancelContract = async function (data, { api }) {
  data.walletId = wallet.getAddress().address
  data.password = await auth.getSessionPassword()

  if (typeof data.walletId !== 'string') {
    throw new WalletError('WalletId is not defined')
  }
  return withAuth((privateKey) =>
    api.contracts.cancelContract({
      walletAddress: data.walletAddress,
      contractId: data.contractId,
      privateKey,
      closeOutType: data.closeOutType
    })
  )(data, { api })
}

export const setContractDeleteStatus = async function (data, { api }) {
  data.walletId = wallet.getAddress().address
  data.password = await auth.getSessionPassword()

  if (typeof data.walletId !== 'string') {
    throw new WalletError('WalletId is not defined')
  }
  return withAuth((privateKey) =>
    api.contracts.setContractDeleteStatus({
      walletAddress: data.walletAddress,
      contractId: data.contractId,
      deleteContract: data.deleteContract,
      privateKey
    })
  )(data, { api })
}

export function createWallet(data, core, isOpen = true) {
  const seed = keys.mnemonicToSeedHex(data.mnemonic)
  const entropy = keys.mnemonicToEntropy(data.mnemonic)
  const walletAddress = core.api.wallet.createAddress(seed)

  return Promise.all([
    wallet.setSeed(seed, data.password),
    wallet.setEntropy(entropy, data.password),
    wallet.setAddress(walletAddress)
  ])
    .then(() => core.emitter.emit('create-wallet', { address: walletAddress }))
    .then(() => isOpen && openWallet(core, data.password))
}

export const restartProxyRouter = async (data, { emitter, api }) => {
  const password = await auth.getSessionPassword()

  await api['proxy-router']
    .kill(config.chain.proxyPort)
    .catch((err) => logger.error('proxy router err', err))

  emitter.emit('open-proxy-router', { password })
}

export const stopProxyRouter = async (data, { emitter, api }) => {
  await api['proxy-router'].kill(config.chain.proxyPort).catch(logger.error)
}

export async function openWallet({ emitter }, password) {
  const { address } = wallet.getAddress()

  emitter.emit('open-wallet', { address, isActive: true })
  emitter.emit('open-proxy-router', { password })
}

export const onboardingCompleted = (data, core) => {
  setProxyRouterConfig(data.proxyRouterConfig)
  return auth
    .setPassword(data.password)
    .then(() =>
      createWallet(
        {
          mnemonic: data.mnemonic,
          password: data.password
        },
        core,
        true
      )
    )
    .then(() => true)
    .catch((err) => ({ error: new WalletError('Onboarding unable to be completed: ', err) }))
}

export const recoverFromMnemonic = function (data, core) {
  if (!auth.isValidPassword(data.password)) {
    return null
  }

  wallet.clearWallet()

  return createWallet(
    {
      mnemonic: data.mnemonic,
      password: data.password
    },
    core,
    false
  )
    .then(noCore.clearCache)
    .then((_) => auth.setSessionPassword(data.password))
}

function onLoginSubmit({ password }, core) {
  var checkPassword = config.chain.bypassAuth
    ? new Promise((r) => r(true))
    : auth.isValidPassword(password)

  return checkPassword
    .then(function (isValid) {
      if (!isValid) {
        return { error: new WalletError('Invalid password') }
      }
      openWallet(core, password)

      return isValid
    })
    .catch((err) => logger.error('onLoginSubmit err', err))
}
export function refreshAllSockets({ url }, { api, emitter }) {
  emitter.emit('sockets-scan-started', {})
  return api.sockets
    .getConnections()
    .then(function () {
      emitter.emit('sockets-scan-finished', { success: true })
      return {}
    })
    .catch(function (error) {
      logger.warn('Could not sync sockets/connections', error.stack)
      emitter.emit('sockets-scan-finished', {
        error: error.message,
        success: false
      })
      // emitter.once('coin-block', () =>
      //   refreshAllTransactions({ address }, { api, emitter })
      // );
      return {}
    })
}

export function refreshAllTransactions({ address }, { api, emitter }) {
  emitter.emit('transactions-scan-started', {})
  return api.explorer
    .refreshAllTransactions(address)
    .then(function () {
      emitter.emit('transactions-scan-finished', { success: true })
      return {}
    })
    .catch(function (error) {
      logger.warn('Could not sync transactions/events', error.stack)
      emitter.emit('transactions-scan-finished', {
        error: error.message,
        success: false
      })
      emitter.once('coin-block', () => refreshAllTransactions({ address }, { api, emitter }))
      return {}
    })
}

export const getMarketplaceFee = async function (data, { api }) {
  return api.contracts.getMarketplaceFee(data)
}

export function refreshAllContracts({}, { api }) {
  const walletId = wallet.getAddress().address
  return api.contracts.refreshContracts(null, walletId)
}

export function refreshTransaction({ hash, address }, { api }) {
  return pTimeout(api.explorer.refreshTransaction(hash, address), config.scanTransactionTimeout)
    .then(() => ({ success: true }))
    .catch((error) => ({ error, success: false }))
}

export const getGasLimit = (data, { api }) => api.wallet.getGasLimit(data)

export const getGasPrice = (data, { api }) => api.wallet.getGasPrice(data)

export const sendLmr = async (data, { api }) =>
  withAuth(api.lumerin.sendLmr)(
    {
      ...data,
      walletId: wallet.getAddress().address,
      password: await auth.getSessionPassword()
    },
    { api }
  )

export const sendEth = async (data, { api }) =>
  withAuth(api.wallet.sendEth)(
    {
      ...data,
      walletId: wallet.getAddress().address,
      password: await auth.getSessionPassword()
    },
    { api }
  )

export const startDiscovery = (data, { api }) => api.devices.startDiscovery(data)

export const stopDiscovery = (data, { api }) => api.devices.stopDiscovery()

export const setMinerPool = (data, { api }) => api.devices.setMinerPool(data)

export const getLmrTransferGasLimit = (data, { api }) => api.lumerin.estimateGasTransfer(data)

export const getAddressAndPrivateKey = async (data, { api }) => {
  const isValid = await auth.isValidPassword(data.password)
  if (!isValid) {
    return { error: new WalletError('Invalid password') }
  }

  const seed = wallet.getSeed(data.password)
  return api.wallet.getAddressAndPrivateKey(seed, undefined)
}

export const refreshProxyRouterConnection = async (data, { api }) =>
  api['proxy-router'].refreshConnectionsStream(data)

export const getLocalIp = async ({}, { api }) => api['proxy-router'].getLocalIp()

export const isProxyPortPublic = async (data, { api }) => api['proxy-router'].isProxyPortPublic(data)

export const logout = async (data) => {
  return cleanupDb()
}

export const getPoolAddress = async (data) => {
  const config = getProxyRouterConfig()
  return config.buyerDefaultPool || config.defaultPool
}

export const hasStoredSecretPhrase = async (data) => {
  return wallet.hasEntropy()
}

export const revealSecretPhrase = async (password) => {
  const isValid = await auth.isValidPassword(password)
  if (!isValid) {
    return { error: new WalletError('Invalid password') }
  }

  const entropy = wallet.getEntropy(password)
  const mnemonic = keys.entropyToMnemonic(entropy)
  return mnemonic
}

export function getPastTransactions({ address, page, pageSize }, { api }) {
  return api.explorer.getPastCoinTransactions(0, undefined, address, page, pageSize)
}

export default {
  // refreshAllSockets,
  refreshAllContracts,
  purchaseContract,
  createContract,
  cancelContract,
  onboardingCompleted,
  recoverFromMnemonic,
  onLoginSubmit,
  refreshAllTransactions,
  refreshTransaction,
  createWallet,
  getGasLimit,
  getGasPrice,
  openWallet,
  sendLmr,
  sendEth,
  startDiscovery,
  stopDiscovery,
  setMinerPool,
  getLmrTransferGasLimit,
  getAddressAndPrivateKey,
  refreshProxyRouterConnection,
  logout,
  getLocalIp,
  getPoolAddress,
  restartProxyRouter,
  claimFaucet,
  revealSecretPhrase,
  hasStoredSecretPhrase,
  getPastTransactions,
  setContractDeleteStatus,
  editContract,
  getMarketplaceFee,
  isProxyPortPublic,
  stopProxyRouter
}
