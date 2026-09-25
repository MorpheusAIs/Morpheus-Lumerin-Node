'use strict'

import { ipcMain } from 'electron'
const stringify = require('json-stringify-safe')

import logger from '../../../logger'
import { isTrustedRendererEvent } from '../../../rendererTrust'
import WalletError from '../WalletError'

export function getLogData(data) {
  if (!data) {
    return ''
  }
  return stringify(data, (key, value) =>
    /password|private.?key|mnemonic|seed|authorization|auth.?header|token/i.test(key)
      ? '[redacted]'
      : value
  )
}

export const checkIfLoggableEvent = (eventName) => eventName !== 'persist-state'

export const isPromise = (p) => {
  if (p && (typeof p === 'object' || typeof p === 'function') && typeof p.then === 'function') {
    return true
  }

  return false
}

export const ignoreChain = (chain, data) =>
  chain !== 'multi' && chain !== 'none' && data?.chain && chain !== data.chain

export function onRendererEvent(eventName, handler, chain) {
  ipcMain.on(eventName, function (event, evProps) {
    if (!isTrustedRendererEvent(event)) {
      logger.warn(`Rejected ${eventName} from an untrusted renderer`)
      return
    }
    if (!evProps || typeof evProps !== 'object' || Array.isArray(evProps)) {
      logger.warn(`Rejected malformed ${eventName} IPC payload`)
      return
    }
    const { id, data } = evProps
    if ((typeof id !== 'string' && typeof id !== 'number') || String(id).length > 200) {
      logger.warn(`Rejected malformed ${eventName} IPC correlation id`)
      return
    }
    if (ignoreChain(chain, data)) {
      return
    }
    Promise.resolve()
      .then(() => handler(data))
      .then(function (res) {
        if (event.sender.isDestroyed()) {
          return
        }
        event.sender.send(eventName, { id, data: res })
      })
      .catch(function (err) {
        if (event.sender.isDestroyed()) {
          return
        }
        const message = String(err?.message ?? 'The desktop operation failed.').slice(0, 2_000)
        const error = new WalletError(message)
        event.sender.send(eventName, { id, data: { error } })
        logger.warn(`<-- ${eventName}:${id} ${message}`)
      })
      .catch(function (err) {
        logger.warn(`Could not send message to renderer: ${err.message}`)
      })
  })
}

export const subscribeTo = (types, chain) =>
  Object.keys(types).forEach((type) => {
    onRendererEvent(type, types[type], chain)
  })

export const unsubscribeTo = (types) =>
  Object.keys(types).forEach((type) => ipcMain.removeAllListeners(type, types[type]))

export default { subscribeTo, unsubscribeTo }
