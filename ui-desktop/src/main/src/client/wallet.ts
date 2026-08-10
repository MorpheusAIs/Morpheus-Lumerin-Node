import settings from 'electron-settings'
import { hdkey, Wallet as EthWallet } from '@ethereumjs/wallet'

import { aes256cbcIv } from './crypto'

export const getWallet = () =>
  Object.keys((settings.getSync('user.wallet') as object | undefined) ?? {})

export const getAddress = () => settings.getSync(`user.wallet.address`)

export const getToken = () => settings.getSync(`user.wallet.token`)

export function getSeed(password) {
  const encryptedSeed = settings.getSync(`user.wallet.encryptedSeed`)
  return aes256cbcIv.decrypt(password, encryptedSeed)
}
export const setAddress = (address) => settings.setSync(`user.wallet.address`, { address })

export const setSeed = (seed, password) =>
  settings.setSync(`user.wallet.encryptedSeed`, aes256cbcIv.encrypt(password, seed))

export const clearWallet = () => settings.setSync('user.wallet', {})

const getWalletFromSeed = (seed, index = 0) =>
  hdkey.EthereumHDKey.fromMasterSeed(Buffer.from(seed, 'hex'))
    .derivePath(`m/44'/60'/0'/0/${index}`)
    .getWallet()

const getAddress2 = (seed, index) => getWalletFromSeed(seed, index).getChecksumAddressString()

const getPrivateKey = (seed, index) => getWalletFromSeed(seed, index).getPrivateKey()

const getAddressAndPrivateKey = (seed, index) => ({
  address: getAddress2(seed, index),
  privateKey: Buffer.from(getPrivateKey(seed, index)).toString('hex')
})

/**
 * Derives the checksummed address for a raw private key.
 *
 * Used when importing a wallet, so the address can be shown and de-duplicated
 * without pushing the key into the proxy-router first — importing must never
 * disturb whichever wallet is currently active.
 */
export const privateKeyToAddress = (privateKeyHex: string): string => {
  const hex = privateKeyHex.startsWith('0x') ? privateKeyHex.slice(2) : privateKeyHex
  if (!/^[0-9a-fA-F]{64}$/.test(hex)) {
    throw new Error('Invalid private key (expected 64 hex characters).')
  }
  return EthWallet.fromPrivateKey(Buffer.from(hex, 'hex')).getChecksumAddressString()
}

export default {
  getAddress,
  setAddress,
  getActiveWallet: getWallet,
  setActiveWallet: setAddress,
  createAddress: getAddress2,
  getAddressAndPrivateKey,
  privateKeyToAddress,
  clearWallet,
  getWallet,
  getToken,
  getSeed,
  setSeed,
}
