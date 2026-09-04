const { ipcMain } = require('electron')
import createCore from '../../core'
import logger from '../../logger'
import subscriptions from './subscriptions'
import * as settings from './settings'
import storage from './storage'
import { isTrustedRendererEvent } from '../../rendererTrust'

export function startCore({ chain, core, config: coreConfig }, webContent) {
  logger.verbose(`Starting core ${chain}`)
  const { emitter, events, api } = core.start(coreConfig)

  emitter.setMaxListeners(50)

  events.push(
    'create-wallet',
    'open-wallet',
    'transactions-scan-started',
    'transactions-scan-finished',
    'contracts-scan-started',
    'contracts-scan-finished',
    'contract-updated',
    'services-state'
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

  emitter.on('wallet-error', function (err) {
    logger.warn(err.inner ? `${err.message} - ${err.inner.message}` : err.message)
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
  ipcMain.on('log.error', function (event, args) {
    if (!isTrustedRendererEvent(event)) return
    logger.error('ipcMain error ', String(args?.message ?? '').slice(0, 2000))
  })

  settings.presetDefaults()

  let core = {
    chain: config.chain.chainId,
    core: createCore(),
    config: Object.assign({}, config.chain, config)
  }
  let coreStarted = false
  let coreInitialized = false

  function cleanupCoreAfterFailedStartup() {
    try {
      subscriptions.unsubscribe(core)
    } catch (err) {
      logger.warn('Could not remove partially initialized renderer subscriptions', err.message)
    }

    if (coreInitialized) {
      try {
        stopCore(core)
      } catch (err) {
        logger.warn('Could not stop partially initialized wallet core', err.message)
      }
    }
    coreInitialized = false
  }

  ipcMain.on('ui-ready', function (webContent, args) {
    if (!isTrustedRendererEvent(webContent)) return
    if (coreStarted) return
    coreStarted = true
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

        // Install every follow-up listener before acknowledging ui-ready.
        // Root immediately requests settings after this response; replying
        // first made that request race subscriptions.subscribe() and time out.
        const { emitter, events, api } = startCore(core, webContent)
        coreInitialized = true
        core.emitter = emitter
        core.events = events
        core.api = api
        subscriptions.subscribe(core)

        webContent.sender.send('ui-ready', payload)
        // logger.verbose(`<-- ui-ready ${stringify(payload)}`);
      })
      .catch(function (err) {
        cleanupCoreAfterFailedStartup()
        coreStarted = false
        logger.error('Could not initialize renderer client', err.message)
      })
  })

  ipcMain.on('ui-unload', function (event) {
    if (!isTrustedRendererEvent(event)) return
    if (!coreStarted) return
    subscriptions.unsubscribe(core)
    stopCore(core)
    coreInitialized = false
    coreStarted = false
  })
}
