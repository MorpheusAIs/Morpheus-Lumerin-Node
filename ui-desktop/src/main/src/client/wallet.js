const settings = require('electron-settings')
const { hdkey } = require('ethereumjs-wallet')

import { aes256cbcIv } from './crypto'

export const getWallet = () => Object.keys(settings.getSync('user.wallet'))

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
  hdkey.fromMasterSeed(Buffer.from(seed, 'hex')).derivePath(`m/44'/60'/0'/0/${index}`).getWallet()

const getAddress2 = (seed, index) => getWalletFromSeed(seed, index).getChecksumAddressString()

const getPrivateKey = (seed, index) => getWalletFromSeed(seed, index).getPrivateKey()

const getAddressAndPrivateKey = (seed, index) => ({
  address: getAddress2(seed, index),
  privateKey: getPrivateKey(seed, index).toString('hex')
})

export default {
  getAddress,
  setAddress,
  getActiveWallet: getWallet,
  setActiveWallet: setAddress,
  createAddress: getAddress2,
  getAddressAndPrivateKey,
  clearWallet,
  getWallet,
  getToken,
  getSeed,
  setSeed,
}
