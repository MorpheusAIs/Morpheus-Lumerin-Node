import { electronAPI } from '@electron-toolkit/preload'
import { ipcRenderer, clipboard, shell, contextBridge } from 'electron'
import remote from '@electron/remote'

// Custom APIs for renderer
const api = {}
// Use `contextBridge` APIs to expose Electron APIs to
// renderer only if context isolation is enabled, otherwise
// just add to the DOM global.
if (process.contextIsolated) {
  try {
    // @see http://electronjs.org/docs/tutorial/security#2-disable-nodejs-integration-for-remote-content

    const copyToClipboard = function (text) {
      return clipboard.writeText(text)
    }

    const getAppVersion = function () {
      return remote.app.getVersion()
    }

    const openLink = function (url) {
      return shell.openExternal(url)
    }

    contextBridge.exposeInMainWorld('ipcRenderer', {
      send(eventName, payload) {
        return ipcRenderer.send(eventName, payload)
      },
      on(eventName, listener) {
        // For some reason the listener passed into this function doesn't work
        // if you want to use it to unsubscribe later (likely due to chrome/node connection).
        // So we wrap it in a function and provide an unsubscribe function both to event handler
        // and as a returned value
        function unsubscribe() {
          ipcRenderer.removeListener(eventName, subscription)
        }

        function subscription(event, payload) {
          listener(event, payload, unsubscribe)
        }

        ipcRenderer.on(eventName, subscription)

        return unsubscribe
      }
    })

    contextBridge.exposeInMainWorld('openLink', openLink)
    contextBridge.exposeInMainWorld('getAppVersion', getAppVersion)
    contextBridge.exposeInMainWorld('copyToClipboard', copyToClipboard)
    
    // contextBridge.exposeInMainWorld('isDev', !remote.app.isPackaged)
    contextBridge.exposeInMainWorld('isDev', true)

    contextBridge.exposeInMainWorld('electron', electronAPI)
    contextBridge.exposeInMainWorld('api', api)
  } catch (error) {
    console.error(error)
  }
} else {
  // @ts-ignore (define in dts)
  window.electron = electronAPI
  // @ts-ignore (define in dts)
  window.api = api
}
