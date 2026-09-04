import { app, BrowserWindow, ipcMain, session, shell, systemPreferences } from 'electron'
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

const installExtension = (install as any).default as typeof install

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
      // NOTE: sandbox is still disabled. The original reason (@electron/remote
      // in the preload) is gone, and the preload now only uses ipcRenderer,
      // clipboard, shell and contextBridge — all of which are available to a
      // sandboxed preload. Flipping this to `true` is very likely correct and
      // is a real hardening win, but it changes the preload's module
      // resolution and cannot be validated without launching the app on each
      // platform. Do it as its own change, with a manual smoke test.
      sandbox: false
    }
  })

  mainWindow.on('ready-to-show', () => {
    mainWindow.show()
  })

  mainWindow.webContents.setWindowOpenHandler((details) => {
    shell.openExternal(details.url)
    return { action: 'deny' }
  })

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
app
  .whenReady()
  .then(() => new Promise((r) => setTimeout(r, sleepBeforeStart)))
  .then(() => {
    // Set app user model id for windows
    electronApp.setAppUserModelId('com.electron')

    // Allow the renderer to use the microphone (STT recording) and other
    // media devices. Without an explicit handler Electron does not grant the
    // `media` permission, so getUserMedia() yields a silent track instead of
    // throwing. On macOS we also proactively request the OS-level mic grant.
    const grantedPermissions = new Set(['media', 'mediaKeySystem', 'audioCapture', 'videoCapture'])

    session.defaultSession.setPermissionRequestHandler((_webContents, permission, callback) => {
      callback(grantedPermissions.has(permission))
    })

    session.defaultSession.setPermissionCheckHandler((_webContents, permission) => {
      return grantedPermissions.has(permission)
    })

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
    ipcMain.on('ping', () => console.log('pong'))

    // Synchronous so the preload can expose getAppVersion() as a plain
    // function. Replaces the previous @electron/remote round-trip.
    ipcMain.on('get-app-version', (event) => {
      event.returnValue = app.getVersion()
    })

    // Register the renderer bootstrap/listener map before loading the renderer.
    // A packaged file:// page can execute quickly enough to emit `ui-ready`
    // before createWindow() returns; IPC events have no queue for a listener
    // that does not exist yet, so that race left Root waiting until its timeout
    // and displayed a misleading wallet-startup failure.
    createClient(config)
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
