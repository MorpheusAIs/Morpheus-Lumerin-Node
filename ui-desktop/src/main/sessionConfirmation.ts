import { BrowserWindow, ipcMain, type IpcMainEvent } from 'electron'
import { randomUUID } from 'node:crypto'
import { join } from 'node:path'
import {
  renderSessionConfirmationHtml,
  type SessionConfirmationDetails
} from './sessionConfirmationView'

export const SESSION_CONFIRMATION_CHANNEL = 'session-confirmation:respond'
export const SESSION_CONFIRMATION_TIMEOUT_MS = 60_000
let confirmationActive = false

/**
 * Main-owned, isolated web contents above the app, not an approval boolean
 * supplied by the app renderer. The underlying renderer cannot approve itself.
 */
export async function showSessionConfirmation(
  owner: BrowserWindow | null | undefined,
  details: SessionConfirmationDetails
): Promise<boolean> {
  if (confirmationActive) throw new Error('Another security confirmation is already open.')
  if (!owner || owner.isDestroyed() || owner.webContents.isDestroyed()) return false
  const ownerContents = owner.webContents
  confirmationActive = true
  try {
    const requestId = randomUUID()
    const html = renderSessionConfirmationHtml({ ...details }, requestId.replace(/-/g, ''))
    const url = `data:text/html;charset=utf-8,${encodeURIComponent(html)}`
    const view = new BrowserWindow({
      parent: owner,
      modal: true,
      frame: false,
      transparent: true,
      show: false,
      resizable: false,
      movable: false,
      minimizable: false,
      maximizable: false,
      fullscreenable: false,
      skipTaskbar: true,
      hasShadow: false,
      webPreferences: {
        preload: join(__dirname, '../preload/session-confirmation.js'),
        additionalArguments: [`--session-confirmation-id=${requestId}`],
        partition: 'session-confirmation',
        contextIsolation: true,
        sandbox: true,
        nodeIntegration: false,
        webviewTag: false,
        webSecurity: true,
        devTools: false
      }
    })
    const contents = view.webContents

    return await new Promise<boolean>((resolve) => {
      let settled = false
      const finish = (approved = false) => {
        if (settled) return
        settled = true
        clearTimeout(timeout)
        ipcMain.removeListener(SESSION_CONFIRMATION_CHANNEL, onDecision)
        owner.removeListener('closed', cancel)
        owner.removeListener('resize', resize)
        owner.removeListener('move', resize)
        ownerContents.removeListener('destroyed', cancel)
        ownerContents.removeListener('did-start-navigation', onOwnerNavigation)
        ownerContents.removeListener('render-process-gone', cancel)
        contents.removeListener('destroyed', cancel)
        contents.removeListener('render-process-gone', cancel)
        contents.removeListener('did-fail-load', cancel)
        contents.removeListener('will-navigate', denyNavigation)
        contents.removeListener('will-redirect', denyNavigation)
        // Native objects may disappear between events. Teardown must never
        // leave the caller waiting or turn a failed cleanup into permission.
        for (const cleanup of [
          () => {
            if (!view.isDestroyed()) view.destroy()
          },
          () => {
            if (!owner.isDestroyed() && !ownerContents.isDestroyed()) ownerContents.focus()
          }
        ]) {
          try {
            cleanup()
          } catch {
            approved = false
          }
        }
        resolve(approved)
      }
      const cancel = () => finish(false)
      const onDecision = (event: IpcMainEvent, payload: unknown) => {
        if (settled) return
        if (owner.isDestroyed() || ownerContents.isDestroyed() || contents.isDestroyed())
          return cancel()
        if (
          event.sender !== contents ||
          event.senderFrame !== contents.mainFrame ||
          contents.getURL() !== url ||
          !payload ||
          typeof payload !== 'object' ||
          Array.isArray(payload)
        )
          return
        const decision = payload as { requestId?: unknown; approved?: unknown }
        if (decision.requestId !== requestId || typeof decision.approved !== 'boolean') return
        finish(decision.approved)
      }
      const resize = () => {
        if (owner.isDestroyed() || contents.isDestroyed()) return cancel()
        const { x, y, width, height } = owner.getContentBounds()
        view.setBounds({ x, y, width: Math.max(1, width), height: Math.max(1, height) })
      }
      const onOwnerNavigation = (
        _event: Electron.Event,
        _url: string,
        _inPlace: boolean,
        isMainFrame: boolean
      ) => {
        if (isMainFrame) cancel()
      }
      const denyNavigation = (event: Electron.Event) => {
        event.preventDefault()
        cancel()
      }
      const timeout = setTimeout(cancel, SESSION_CONFIRMATION_TIMEOUT_MS)
      ipcMain.on(SESSION_CONFIRMATION_CHANNEL, onDecision)
      owner.once('closed', cancel)
      owner.on('resize', resize)
      owner.on('move', resize)
      ownerContents.once('destroyed', cancel)
      ownerContents.on('did-start-navigation', onOwnerNavigation)
      ownerContents.once('render-process-gone', cancel)
      contents.once('destroyed', cancel)
      contents.once('render-process-gone', cancel)
      contents.once('did-fail-load', cancel)
      contents.on('will-navigate', denyNavigation)
      contents.on('will-redirect', denyNavigation)
      try {
        contents.session.setPermissionRequestHandler((_contents, _permission, callback) =>
          callback(false)
        )
        contents.session.setPermissionCheckHandler(() => false)
        contents.session.webRequest.onBeforeRequest(
          { urls: ['http://*/*', 'https://*/*', 'ws://*/*', 'wss://*/*', 'file://*/*'] },
          (_details, callback) => callback({ cancel: true })
        )
        contents.setWindowOpenHandler(() => ({ action: 'deny' }))
        view.setMenu(null)
        resize()
        void contents
          .loadURL(url)
          .then(() => {
            if (!settled && !contents.isDestroyed()) {
              resize()
              view.show()
              contents.focus()
            }
          })
          .catch(cancel)
      } catch (error) {
        console.error(
          'Unable to display session confirmation:',
          error instanceof Error ? error.message : 'Unknown display error'
        )
        cancel()
      }
    })
  } finally {
    confirmationActive = false
  }
}
