import utils from 'web3-utils'
import cuid from 'cuid'
import * as Deferred from '../lib/Deferred'

export const fromWei = (str, unit = 'ether') => utils.fromWei(str, unit)
export const toWei = (bn, unit = 'ether') => utils.toWei(bn, unit)

export const isAddress = (str) => utils.isAddress(str)

export const toBN = (str) => utils.toBN(str)
export const toHex = (bn) => utils.toHex(bn)

export function forwardToMainProcess(eventName, timeout = 10000) {
  return function (data) {
    return sendToMainProcess(eventName, data, timeout)
  }
}

/**
 * Sends a message to Main Process and returns a Promise.
 *
 * This makes it easier to handle IPC inside components
 * without the need of manual (un)subscriptions.
 * @param {string} eventName
 * @param {*} data
 * @param {number} timeout
 * @param {import('electron').IpcRenderer} ipcRenderer
 * @returns
 */
export function sendToMainProcess(
  eventName,
  data,
  timeout = 10000,
  ipcRenderer = window.ipcRenderer
) {
  const id = cuid()

  const deferred = new Deferred()
  let timeoutId

  function listener(ev, { id: _id, data: _data, error }, unsubscribe) {
    if (timeoutId) {
      window.clearTimeout(timeoutId)
    }
    if (_id !== id) {
      return
    }

    const responseError = error || (_data && _data.error)

    if (responseError) {
      deferred.reject(responseError)
      ipcRenderer.send('handle-client-error', {
        id: cuid(),
        data: responseError
      })
    } else {
      deferred.resolve(_data)
    }

    return unsubscribe()
  }

  const unsubscribe = ipcRenderer.on(eventName, listener)
  ipcRenderer.send(eventName, { id, data })

  if (timeout) {
    timeoutId = setTimeout(() => {
      console.warn(`Event "${eventName}" timed out after ${timeout}ms.`)
      deferred.reject(new Error('Operation timed out. Please try again later.'))
      unsubscribe()
    }, timeout)
  }

  return deferred.promise
}
