'use strict'

const { ipcMain, app } = require('electron')
const createCore = require('@lumerin/wallet-core')
const stringify = require('json-stringify-safe')

import logger from '../../logger'
import subscriptions from './subscriptions'
import settings from './settings'
import storage from './storage'
import { startMonitoringHashrate } from './contractsHashrateSyncer'

import { getAddressAndPrivateKey, refreshProxyRouterConnection } from './handlers/single-core'

import { runProxyRouter, isProxyRouterHealthy } from './proxyRouter'
let interval
export function startCore({ chain, core, config: coreConfig }, webContent) {
  logger.verbose(`Starting core ${chain}`)
  const { emitter, events, api } = core.start(coreConfig)
  const proxyRouterApi = api['proxy-router']
  if (interval) clearInterval(interval)

  interval = startMonitoringHashrate(coreConfig.localProxyRouterUrl, 5 * 60 * 1000)

  // emitter.setMaxListeners(30);
  emitter.setMaxListeners(50)

  events.push(
    'create-wallet',
    'transactions-scan-started',
    'transactions-scan-finished',
    'contracts-scan-started',
    'contracts-scan-finished',
    'contract-updated'
  )

  function send(eventName, data) {
    try {
      if (!webContent) {
        return
      }
      const payload = Object.assign({}, data, { chain })
      webContent.sender.send(eventName, payload)
    } catch (err) {
      logger.error('send error', err)
    }
  }

  events.forEach((event) =>
    emitter.on(event, function (data) {
      send(event, data)
    })
  )

  function syncTransactions({ address }, page = 1, pageSize = 15) {
    return storage
      .getSyncBlock(chain)
      .then(function (from) {
        send('transactions-scan-started', {})

        return api.explorer
          .syncTransactions(
            0,
            address,
            (number) => storage.setSyncBlock(number, chain),
            page,
            pageSize
          )
          .then(function () {
            send('transactions-scan-finished', { success: true })

            emitter.on('coin-block', function ({ number }) {
              storage.setSyncBlock(number, chain).catch(function (err) {
                logger.warn('Could not save new synced block', err)
              })
            })
          })
      })
      .catch(function (err) {
        logger.warn('Could not sync transactions/events', err.stack)
        send('transactions-scan-finished', {
          error: err.message,
          success: false
        })

        emitter.once('coin-block', () => syncTransactions({ address }, page, pageSize))
      })
  }

  emitter.on('open-wallet', syncTransactions)

  emitter.on('wallet-error', function (err) {
    logger.warn(err.inner ? `${err.message} - ${err.inner.message}` : err.message)
  })

  const shouldRestartProxyRouterAfterWalletUpdate = () => {
    const prevAppVersion = settings.getAppVersion()
    const isAppVersionChanged = prevAppVersion !== app.getVersion()

    if (isAppVersionChanged) {
      settings.setAppVersion(app.getVersion())
      return true
    }
    return false
  }

  emitter.on('open-proxy-router', async ({ password }) => {
    const proxyRouterUserConfig = settings.getProxyRouterConfig()
    if (!proxyRouterUserConfig.runWithoutProxyRouter) {
      const { address, privateKey } = await getAddressAndPrivateKey({ password }, { api })

      send('proxy-router-type-changed', {
        isLocal: true
      })

      const config = {
        privateKey,
        walletAddress: address,
        ...coreConfig.chain,
        ...proxyRouterUserConfig
      }

      const isProxyHealth = await isProxyRouterHealthy(api, config.localProxyRouterUrl)
      const shouldRestartProxy = shouldRestartProxyRouterAfterWalletUpdate()
      if (!isProxyHealth || shouldRestartProxy) {
        logger.debug('Seller is not healhy, restart...')
        await proxyRouterApi.kill(config.proxyPort).catch(logger.error)
        runProxyRouter(config)
      }

      refreshProxyRouterConnection(
        {
          proxyNodeUrl: config.localProxyRouterUrl
        },
        { api }
      )
    }
  })

  return {
    emitter,
    events,
    api
  }
}

export function stopCore({ core, chain }) {
  logger.verbose(`Stopping core ${chain}`)
  core.stop()
}

export function createClient(config) {
  ipcMain.on('log.error', function (_, args) {
    logger.error('ipcMain error ', args.message)
  })

  settings.presetDefaults()

  let core = {
    chain: config.chain.chainId,
    core: createCore(),
    config: Object.assign({}, config.chain, config)
  }

  ipcMain.on('ui-ready', function (webContent, args) {
    const onboardingComplete = !!settings.getPasswordHash()

    storage
      .getState()
      .catch(function (err) {
        logger.warn('Failed to get state', err.message)
        return {}
      })
      .then(function (persistedState) {
        const payload = Object.assign({}, args, {
          data: {
            onboardingComplete,
            persistedState: persistedState || {},
            config
          }
        })
        webContent.sender.send('ui-ready', payload)
        // logger.verbose(`<-- ui-ready ${stringify(payload)}`);
      })
      .catch(function (err) {
        logger.error('Could not send ui-ready message back', err.message)
      })
      .then(function () {
        const { emitter, events, api } = startCore(core, webContent)
        core.emitter = emitter
        core.events = events
        core.api = api
        console.log('Unknown chain =', core);
        subscriptions.subscribe(core)
      })
      .catch(function (err) {
        console.log('panic')
        console.log(err)
        console.log('Unknown chain =', err.message)
        logger.error('Could not start core', err.message)
      })
  })

  ipcMain.on('ui-unload', function () {
    stopCore(core)
    subscriptions.unsubscribe(core)
  })
}
