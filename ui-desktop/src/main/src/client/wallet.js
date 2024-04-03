'use strict'

const settings = require('electron-settings')

import { aes256cbcIv, sha256 } from './crypto'

export const getWallet = () => Object.keys(settings.getSync('user.wallet'))

export const getAddress = () => settings.getSync(`user.wallet.address`)

export const getToken = () => settings.getSync(`user.wallet.token`)

export function getSeed(password) {
  const encryptedSeed = settings.getSync(`user.wallet.encryptedSeed`)
  return aes256cbcIv.decrypt(password, encryptedSeed)
}

export const hasEntropy = () => !!settings.getSync(`user.wallet.encryptedEntropy`)

export function getEntropy(password) {
  const encryptedEntropy = settings.getSync(`user.wallet.encryptedEntropy`)
  return aes256cbcIv.decrypt(password, encryptedEntropy)
}

export const setAddress = (address) => settings.setSync(`user.wallet.address`, { address })

export const setSeed = (seed, password) =>
  settings.setSync(`user.wallet.encryptedSeed`, aes256cbcIv.encrypt(password, seed))

export const setEntropy = (entropy, password) =>
  settings.setSync(`user.wallet.encryptedEntropy`, aes256cbcIv.encrypt(password, entropy))

export const clearWallet = () => settings.setSync('user.wallet', {})

export default {
  getAddress,
  setAddress,
  getActiveWallet: getWallet,
  setActiveWallet: setAddress,
  clearWallet,
  getWallet,
  getToken,
  getSeed,
  setSeed,
  getEntropy,
  setEntropy,
  hasEntropy
}
