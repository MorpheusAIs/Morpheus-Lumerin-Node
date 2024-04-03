'use strict'

const settings = require('electron-settings')
// const utils = require("web3-utils");
const { merge } = require('lodash')

const logger = require('../../../logger')
const restart = require('../electron-restart')
const { getDb } = require('../database')

const getKey = (key) => settings.getSync(key)

function setKey(key, value) {
  settings.setSync(key, value)
  logger.verbose('Settings changed', key)
}

const getPasswordHash = () => getKey('user.passwordHash')

function setPasswordHash(hash) {
  setKey('user.passwordHash', hash)
}

const setProxyRouterConfig = (config) => setKey('user.proxyRouterConfig', JSON.stringify(config))

const getProxyRouterConfig = () => {
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

function upgradeSettings(defaultSettings, currentSettings) {
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

function presetDefaults() {
  logger.verbose('Settings file', settings.file())
  const currentSettings = settings.getSync()
  const defaultSettings = require('./defaultSettings')
  settings.setSync(merge(defaultSettings, currentSettings))
  logger.verbose('Default settings applied')
  logger.debug('Current settings', settings.getSync())
}

function cleanupDb() {
  const currentSettings = settings.getSync()
  const defaultSettings = require('./defaultSettings')

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

const getDefaultCurrencySetting = () => getKey('selectedCurrency')

const setDefaultCurrencySetting = (currency) => setKey('selectedCurrency', currency)

const getAppVersion = () => getKey('app.version')

const setAppVersion = (value) => setKey('app.version', value)

module.exports = {
  getPasswordHash,
  setPasswordHash,
  presetDefaults,
  setProxyRouterConfig,
  getProxyRouterConfig,
  cleanupDb,
  getDefaultCurrencySetting,
  setDefaultCurrencySetting,
  getKey,
  setKey,
  getAppVersion,
  setAppVersion
}
