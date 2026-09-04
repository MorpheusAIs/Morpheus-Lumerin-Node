import { app, BrowserWindow, dialog, ipcMain, session, shell, systemPreferences } from 'electron'
import { electronApp, is, optimizer } from '@electron-toolkit/utils'
import {
  default as install,
  REDUX_DEVTOOLS
  // REACT_DEVELOPER_TOOLS,
} from 'electron-devtools-installer'
import { createClient } from './src/client'
import config from './config'
import initContextMenu from './contextMenu'
import initMenu from './menu'
import errorHandler from './errorHandler'
import logger from './logger'
import { join } from 'path'
import { registerCoworkIpc } from './src/client/cowork-ipc'
import { isTrustedRendererEvent, isTrustedRendererUrl } from './rendererTrust'
import { registerChatStreamIpc } from './src/client/chat-stream-ipc'
import { registerIpfsDownloadIpc } from './src/client/ipfs-download-ipc'

const installExtension = (install as any).default as typeof install
const openExternalChannel = 'open-external-url'
const maxExternalUrlLength = 8192
const trustedExternalHosts = ['mor.org', 'lumerin.io', 'github.com', 'etherscan.io']
let externalConfirmationOpen = false
const ownsSingleInstanceLock = app.requestSingleInstanceLock()

if (!ownsSingleInstanceLock) {
  // Wallet state, the managed proxy-router, and Workspace's NeDB stores are all
  // process-owned. A second main process would contend for the same ports and
  // could retain a stale approval policy cache.
  app.quit()
} else {
  app.on('second-instance', () => {
    const mainWindow = BrowserWindow.getAllWindows()[0]
    if (!mainWindow) return
    if (mainWindow.isMinimized()) mainWindow.restore()
    mainWindow.show()
    mainWindow.focus()
  })
}

function normalizeExternalUrl(value: unknown): string | null {
  if (typeof value !== 'string' || value.length === 0 || value.length > maxExternalUrlLength) {
    return null
  }

  try {
    const url = new URL(value)
    if (
      url.protocol !== 'https:' ||
      !url.hostname ||
      url.username.length > 0 ||
      url.password.length > 0
    ) {
      return null
    }
    return url.href
  } catch {
    return null
  }
}

const trustedExternalHost = (hostname: string): boolean =>
  trustedExternalHosts.some((host) => hostname === host || hostname.endsWith(`.${host}`))

async function openExternalUrl(value: unknown): Promise<void> {
  const url = normalizeExternalUrl(value)
  if (!url) {
    logger.warn('Rejected unsafe external URL')
    return
  }

  try {
    const parsed = new URL(url)
    if (!trustedExternalHost(parsed.hostname)) {
      if (externalConfirmationOpen) {
        logger.warn('Ignored external URL while another link confirmation was open')
        return
      }
      externalConfirmationOpen = true
      try {
        const owner = BrowserWindow.getFocusedWindow()
        const options = {
          type: 'question' as const,
          title: 'Open external link',
          message: `Open ${parsed.hostname} in your browser?`,
          detail: url.slice(0, 2_000),
          buttons: ['Cancel', 'Open link'],
          defaultId: 0,
          cancelId: 0,
          noLink: true
        }
        const response = owner
          ? await dialog.showMessageBox(owner, options)
          : await dialog.showMessageBox(options)
        if (response.response !== 1) return
      } finally {
        externalConfirmationOpen = false
      }
    }
    await shell.openExternal(url)
  } catch (error) {
    logger.error('Failed to open external URL', error)
  }
}

function createWindow(): void {
  // Create the browser window.
  const mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    show: false,
    autoHideMenuBar: true,
    // ...(process.platform === 'linux' ? { icon } : {}),
    webPreferences: {
      preload: join(__dirname, '../preload/index.js'),
      devTools: is.dev,
      contextIsolation: true,
      nodeIntegration: false,
      webviewTag: false,
      // The preload is a bundled CJS bridge using only Electron's sandbox-safe
      // APIs. Renderer code therefore has neither Node integration nor an
      // unsandboxed preload escape hatch.
      sandbox: true
    }
  })

  mainWindow.on('ready-to-show', () => {
    mainWindow.show()
  })

  mainWindow.webContents.setWindowOpenHandler((details) => {
    void openExternalUrl(details.url)
    return { action: 'deny' }
  })

  const guardNavigation = (event: Electron.Event, navigationUrl: string): void => {
    if (isTrustedRendererUrl(navigationUrl)) {
      return
    }

    event.preventDefault()
    void openExternalUrl(navigationUrl)
  }

  mainWindow.webContents.on('will-navigate', guardNavigation)
  mainWindow.webContents.on('will-redirect', guardNavigation)

  // HMR for renderer base on electron-vite cli.
  // Load the remote URL for development or the local html file for production.
  if (is.dev && process.env['ELECTRON_RENDERER_URL']) {
    mainWindow.loadURL(process.env['ELECTRON_RENDERER_URL'])
  } else {
    mainWindow.loadFile(join(__dirname, '../renderer/index.html'))
  }
}

errorHandler({ logger: logger.error })

const sleepBeforeStart = 3000

// This method will be called when Electron has finished
// initialization and is ready to create browser windows.
// Some APIs can only be used after this event occurs.
const appReady = ownsSingleInstanceLock ? app.whenReady() : new Promise<void>(() => undefined)

appReady
  .then(() => new Promise((r) => setTimeout(r, sleepBeforeStart)))
  .then(() => {
    // Set app user model id for windows
    electronApp.setAppUserModelId('com.electron')

    // Allow the renderer to use the microphone (STT recording) and other
    // media devices. Without an explicit handler Electron does not grant the
    // `media` permission, so getUserMedia() yields a silent track instead of
    // throwing. On macOS we also proactively request the OS-level mic grant.
    session.defaultSession.setPermissionRequestHandler(
      (webContents, permission, callback, details) => {
        const requestingUrl = details.requestingUrl || webContents.getURL()
        const mediaTypes = Array.isArray((details as any).mediaTypes)
          ? ((details as any).mediaTypes as string[])
          : []
        const audioOnly = mediaTypes.length === 0 || mediaTypes.every((type) => type === 'audio')
        callback(isTrustedRendererUrl(requestingUrl) && permission === 'media' && audioOnly)
      }
    )

    session.defaultSession.setPermissionCheckHandler(
      (webContents, permission, _requestingOrigin, details) => {
        const requestingUrl = details.requestingUrl || webContents?.getURL() || ''
        return isTrustedRendererUrl(requestingUrl) && permission === 'media'
      }
    )

    if (process.platform === 'darwin') {
      const micStatus = systemPreferences.getMediaAccessStatus('microphone')
      logger.info(`Microphone access status: ${micStatus}`)
      if (micStatus === 'denied' || micStatus === 'restricted') {
        logger.error(
          `Microphone access is "${micStatus}". macOS will return a SILENT audio track. ` +
            `Reset it with: tccutil reset Microphone com.github.Electron (dev) ` +
            `or enable it in System Settings > Privacy & Security > Microphone, then restart.`
        )
      }
      systemPreferences
        .askForMediaAccess('microphone')
        .then((granted) => logger.info(`Microphone access granted: ${granted}`))
        .catch((err) => logger.error('Failed requesting microphone access', err))
    }

    // Default open or close DevTools by F12 in development
    // and ignore CommandOrControl + R in production.
    // see https://github.com/alex8088/electron-toolkit/tree/master/packages/utils
    app.on('browser-window-created', (_, window) => {
      optimizer.watchWindowShortcuts(window)
    })

    // install devtools
    if (is.dev) {
      installExtension([REDUX_DEVTOOLS])
        .then((name) => console.log(`Added Extension:  ${name[0].name}`))
        .catch((err) => console.log('An error occurred: ', err))
    }

    // IPC test
    ipcMain.on('ping', (event) => {
      if (isTrustedRendererEvent(event)) console.log('pong')
    })

    // Synchronous so the preload can expose getAppVersion() as a plain
    // function. Replaces the previous @electron/remote round-trip.
    ipcMain.on('get-app-version', (event) => {
      event.returnValue = isTrustedRendererEvent(event) ? app.getVersion() : ''
    })

    ipcMain.handle(openExternalChannel, async (event, url: unknown) => {
      if (!isTrustedRendererEvent(event)) {
        logger.warn('Rejected external URL request from an untrusted renderer')
        return
      }
      await openExternalUrl(url)
    })

    // Register the renderer bootstrap/listener map before loading the renderer.
    // A packaged file:// page can execute quickly enough to emit `ui-ready`
    // before createWindow() returns; IPC events have no queue for a listener
    // that does not exist yet, so that race left Root waiting until its timeout
    // and displayed a misleading wallet-startup failure.
    createClient(config)

    registerCoworkIpc()
    registerChatStreamIpc()
    registerIpfsDownloadIpc()
    createWindow()

    app.on('activate', function () {
      // On macOS it's common to re-create a window in the app when the
      // dock icon is clicked and there are no other windows open.
      if (BrowserWindow.getAllWindows().length === 0) createWindow()
    })

    logger.info('App ready, initializing...')

    initMenu()
    initContextMenu()
  })

// Quit when all windows are closed, except on macOS. There, it's common
// for applications and their menu bar to stay active until the user quits
// explicitly with Cmd + Q.
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})
