'use strict'

const { ipcMain } = require('electron')
const stringify = require('json-stringify-safe')

import logger from '../../../logger'
import WalletError from '../WalletError'

export function getLogData(data) {
  if (!data) {
    return ''
  }
  const logData = Object.assign({}, data)

  const blackList = ['password']
  blackList.forEach((w) => delete logData[w])

  return stringify(logData)
}

export const checkIfLoggableEvent = (eventName) => eventName !== 'persist-state'

export const isPromise = (p) => {
  if (typeof p === 'object' && typeof p.then === 'function') {
    return true
  }

  return false
}

export const ignoreChain = (chain, data) =>
  chain !== 'multi' && chain !== 'none' && data.chain && chain !== data.chain

export function onRendererEvent(eventName, handler, chain) {
  ipcMain.on(eventName, function (event, evProps) {
    const { id, data } = evProps
    if (ignoreChain(chain, data)) {
      return
    }
    const result = handler(data)

    if (!isPromise(result)) {
      logger.warn(`<-- ${eventName} result is not a promise!. ${result}`)
      return
    }

    result
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
        const error = new WalletError(err.message)
        event.sender.send(eventName, { id, data: { error } })
        logger.warn(`<-- ${eventName}:${id} ${err.message}`)
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
