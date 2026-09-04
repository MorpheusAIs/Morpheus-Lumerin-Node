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
  let bootstrapGeneration = 0
  let bootstrapPromise = null
  let persistedBootstrapState = {}

  function cleanupCoreAfterFailedStartup() {
    if (coreInitialized) {
      try {
        subscriptions.unsubscribe(core)
      } catch (err) {
        logger.warn('Could not remove partially initialized renderer subscriptions', err.message)
      }

      try {
        stopCore(core)
      } catch (err) {
        logger.warn('Could not stop partially initialized wallet core', err.message)
      }
    }
    coreInitialized = false
    coreStarted = false
  }

  function ensureCoreStarted(webContent) {
    if (coreInitialized) {
      return Promise.resolve({
        generation: bootstrapGeneration,
        persistedState: persistedBootstrapState
      })
    }

    // A renderer reload or duplicate mount can emit ui-ready again while the
    // first request is still reading persisted state. Share that work and
    // reply to every request instead of dropping the later request and making
    // it time out. The generation also lets ui-unload invalidate work that is
    // still pending without calling stop() on a core that has not started.
    if (bootstrapPromise) {
      return bootstrapPromise
    }

    coreStarted = true
    const generation = ++bootstrapGeneration
    const pendingBootstrap = storage
      .getState()
      .catch(function (err) {
        logger.warn('Failed to get state', err.message)
        return {}
      })
      .then(function (persistedState) {
        if (generation !== bootstrapGeneration) {
          return null
        }

        // Install every follow-up listener before acknowledging ui-ready.
        // Root immediately requests settings after this response; replying
        // first made that request race subscriptions.subscribe() and time out.
        const { emitter, events, api } = startCore(core, webContent)
        coreInitialized = true
        core.emitter = emitter
        core.events = events
        core.api = api
        subscriptions.subscribe(core)

        persistedBootstrapState = persistedState || {}
        return { generation, persistedState: persistedBootstrapState }
      })
      .catch(function (err) {
        // A stale attempt belongs to a renderer that already unloaded. It must
        // neither tear down a newer attempt nor surface a false startup error.
        if (generation === bootstrapGeneration) {
          cleanupCoreAfterFailedStartup()
          logger.error('Could not initialize renderer client', err.message)
        }
        throw err
      })
      .finally(function () {
        if (bootstrapPromise === pendingBootstrap) {
          bootstrapPromise = null
        }
      })

    bootstrapPromise = pendingBootstrap
    return pendingBootstrap
  }

  ipcMain.on('ui-ready', function (webContent, args) {
    if (!isTrustedRendererEvent(webContent)) return
    ensureCoreStarted(webContent)
      .then(function (bootstrap) {
        if (!bootstrap || bootstrap.generation !== bootstrapGeneration || !coreInitialized) {
          return
        }

        const payload = Object.assign({}, args, {
          data: {
            onboardingComplete: !!settings.getPasswordHash(),
            persistedState: bootstrap.persistedState,
            config
          }
        })

        try {
          webContent.sender.send('ui-ready', payload)
        } catch (err) {
          // Invalidate every waiter from this renderer before cleanup so none
          // of them can acknowledge a core that was just stopped.
          if (bootstrap.generation === bootstrapGeneration) {
            bootstrapGeneration++
            bootstrapPromise = null
            cleanupCoreAfterFailedStartup()
            logger.error('Could not initialize renderer client', err.message)
          }
        }
        // logger.verbose(`<-- ui-ready ${stringify(payload)}`);
      })
      // Initialization failures are logged once inside the shared bootstrap;
      // acknowledge every coalesced requester immediately as well. Otherwise
      // each renderer waits for its generic IPC timeout and reports a much less
      // useful "operation timed out" error.
      .catch(function (err) {
        try {
          webContent.sender.send(
            'ui-ready',
            Object.assign({}, args, {
              error: {
                message: err?.message || 'Could not initialize the wallet core'
              }
            })
          )
        } catch {
          // The renderer disappeared while startup failed; there is nobody left
          // to notify and ensureCoreStarted already rolled back the partial core.
        }
      })
  })

  ipcMain.on('ui-unload', function (event) {
    if (!isTrustedRendererEvent(event)) return
    if (!coreStarted) return

    // Cancel in-flight state reads and prevent their continuations from
    // starting the core after this renderer is gone.
    bootstrapGeneration++
    bootstrapPromise = null
    cleanupCoreAfterFailedStartup()
  })
}
