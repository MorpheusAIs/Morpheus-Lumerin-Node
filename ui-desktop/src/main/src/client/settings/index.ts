import settings from 'electron-settings'
import merge from 'lodash/merge'

import logger from '../../../logger'
import restart from '../electron-restart'
import { getDb } from '../database'
import defaultSettings from './defaultSettings.json'

const FAILOVER_KEY = 'user.failover'

//TODO: make sure default settings works as a static import.  it was getting imported every time
//      it was accessed.  if that's necessary, we have to use the async method
//      import() instead of require() with the new version of node
export const getKey = (key: string): any => settings.getSync(key)

export function setKey(key: string, value: any) {
  settings.setSync(key, value)
  logger.verbose('Settings changed', key)
}

export const getPasswordHash = () => {
  return getKey('user.passwordHash')
}

export function setPasswordHash(hash) {
  setKey('user.passwordHash', hash)
}

export const setProxyRouterConfig = (config) =>
  setKey('user.proxyRouterConfig', JSON.stringify(config))

export const getProxyRouterConfig = () => {
  try {
    const configJson = getKey('user.proxyRouterConfig')
    const data = JSON.parse(configJson)
    if (data.defaultPool) {
      if (!data.sellerDefaultPool) {
        data.sellerDefaultPool = data.defaultPool
      }
      if (!data.buyerDefaultPool) {
        data.buyerDefaultPool = data.defaultPool
      }
    }
    return data
  } catch (e) {
    console.error('error getting proxyrouter config', e)
    cleanupDb()
  }
}

export function upgradeSettings(defaultSettings, currentSettings) {
  let finalSettings = merge({}, currentSettings)
  // Remove no longer used settings as now are stored in config
  delete finalSettings.app
  delete finalSettings.coincap
  delete finalSettings.token

  // Convert previous addresses to checksum addresses
  // if (finalSettings.user && finalSettings.user.wallet) {

  //   Object.keys(finalSettings.user.wallets).forEach(function (key) {
  //     Object.keys(finalSettings.user.wallets[key].addresses).forEach(function (address) {
  //         if (!utils.checkAddressChecksum(address)) {
  //           finalSettings.user.wallets[key].addresses[utils.toChecksumAddress(address)] = finalSettings.user.wallets[key].addresses[address];
  //           // Remove previous lowercase address
  //           delete finalSettings.user.wallets[key].addresses[address];
  //         }
  //       }
  //     );
  //   });

  finalSettings.settingsVersion = defaultSettings.settingsVersion
  settings.setSync(finalSettings)
}

export function presetDefaults() {
  logger.verbose('Settings file', settings.file())
  const currentSettings = settings.getSync()
  settings.setSync(merge(defaultSettings, currentSettings))
  logger.verbose('Default settings applied')
  logger.debug('Current settings', settings.getSync())
}

export function cleanupDb() {
  const currentSettings = settings.getSync()

  logger.warn('Removing old user settings')
  delete currentSettings.user
  // Overwrite old settings and clear db if settings file version changed
  upgradeSettings(defaultSettings, currentSettings)
  const db = getDb()
  db.dropDatabase().catch(function (err) {
    logger.error('Possible database corruption', err.message)
  })
  restart(1)
}

export const resetSettings = () => {
  return settings.reset()
}

export const getDefaultCurrencySetting = () => getKey('selectedCurrency')

export const setDefaultCurrencySetting = (currency) => setKey('selectedCurrency', currency)

export const getAppVersion = () => getKey('app.version')

export const setAppVersion = (value) => setKey('app.version', value)

export const getFailoverSetting = async () => getKey(FAILOVER_KEY)

export const setFailoverSetting = async (isEnabled) => setKey(FAILOVER_KEY, { isEnabled })
