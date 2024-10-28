import isDev from 'electron-is-dev'
// import { app, BrowserWindow, dialog, shell } from 'electron'
import { autoUpdater } from 'electron-updater'
// import * as windowStateKeeper from 'electron-window-state'

import logger from '../logger'
import '../analytics'
// import restart from './client/electron-restart'

// Disable electron security warnings since local content is served via http

// export function showUpdateNotification(info = {}) {
//   if (!Notification.isSupported()) {
//     return
//   }

//   const versionLabel = info.label ? `Version ${info.version}` : 'The latest version'

//   const notification = new Notification({
//     title: `${versionLabel} was installed`,
//     body: 'Lumerin Wallet will be automatically updated after restart.'
//   })

//   notification.show()
// }

export function initAutoUpdate() {
  if (isDev) {
    return
  }
  autoUpdater.on('checking-for-update', () => logger.info('Checking for update...'))
  autoUpdater.on('update-available', () => logger.info('Update available.'))
  autoUpdater.on('download-progress', function (progressObj) {
    let msg = `Download speed: ${progressObj.bytesPerSecond}`
    msg += ` - Downloaded ${progressObj.percent}%`
    msg += ` (${progressObj.transferred}/${progressObj.total})`
    logger.info(msg)
  })

  autoUpdater.on('update-not-available', () => logger.info('Update not available.'))
  autoUpdater.on('error', (err) => logger.error(`Error in auto-updater. ${err}`))

  autoUpdater
    .checkForUpdatesAndNotify()
    .then((res) => {
      logger.info(`Checked for the updates: ${res}`)
    })
    .catch(function (err) {
      logger.warn('Could not find updates', err.message)
    })
}

// TODO: reintegrate what's required to index.ts createWindow function
//
// export function loadWindow(config) {
//   // Ensure the app is ready before creating the main window
//   let appQuitting = false

//   if (!app.isReady()) {
//     logger.warn('Tried to load main window while app not ready. Reloading...')
//     restart(1)
//     return
//   }

//   if (mainWindow) {
//     return
//   }

//   const mainWindowState = windowStateKeeper.default({
//     // defaultWidth: 660,
//     defaultWidth: 820,
//     defaultHeight: 800
//   })

//   // TODO this should be read from config
//   mainWindow = new BrowserWindow({
//     show: false,
//     width: mainWindowState.width,
//     height: mainWindowState.height,
//     // maxWidth: 660,
//     // maxHeight: 700,
//     minWidth: 660,
//     minHeight: 800,
//     backgroundColor: '#323232',
//     webPreferences: {
//       // enableRemoteModule: true,
//       nodeIntegration: false,
//       contextIsolation: true,
//       preload: path.join(__dirname, 'preload.mjs'),
//       devTools: config.devTools || !app.isPackaged
//     },
//     x: mainWindowState.x,
//     y: mainWindowState.y
//   })

//   require('@electron/remote/main').enable(mainWindow.webContents)

//   mainWindowState.manage(mainWindow)

//   analytics.init(mainWindow.webContents.getUserAgent())

//   const appUrl = isDev
//     ? process.env.ELECTRON_START_URL
//     : `file://${path.join(__dirname, '../index.html')}`

//   logger.info('Roading renderer from URL:', appUrl)

//   mainWindow.loadURL(appUrl)

//   mainWindow.webContents.on('crashed', function (ev, killed) {
//     logger.error('Crashed', ev.sender.id, killed)
//   })

//   mainWindow.on('unresponsive', function (ev) {
//     logger.error('Unresponsive', ev.sender.id)
//   })

//   mainWindow.on('closed', function () {
//     mainWindow = null
//   })

//   mainWindow.once('ready-to-show', function () {
//     initAutoUpdate()
//     mainWindow.show()
//   })

//   mainWindow.on('close', (event) => {
//     event.preventDefault()
//     if (appQuitting || process.platform !== 'darwin') {
//       const choice = dialog.showMessageBoxSync(mainWindow, {
//         type: 'question',
//         buttons: ['Yes', 'No'],
//         title: 'Confirm',
//         message: 'Are you sure you want to quit?'
//       })
//       if (choice === 1) {
//         return
//       } else {
//         mainWindow.destroy()
//         mainWindow = null
//         app.quit()
//       }
//     } else {
//       mainWindow.hide()
//     }
//   })

//   app.on('activate', () => {
//     mainWindow.show()
//   })

//   app.on('before-quit', () => {
//     appQuitting = true
//   })
// }
