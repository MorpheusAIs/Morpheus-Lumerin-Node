const { app } = require('electron')
const remote = require('@electron/remote/main')
remote.initialize()
const path = require('path')

require('dotenv').config({ path: path.resolve(__dirname, '../.env') })

const isDev = import('electron-is-dev')
const os = require('os')
const Raven = require('raven')

const { createWindow } = require('./main/main-window.js')
const { createClient } = require('./main/client')
const config = require('./config')
const initContextMenu = require('./contextMenu')
const initMenu = require('./menu')
const errorHandler = require('./errorHandler')
const logger = require('./logger')
errorHandler({ logger: logger.error })

if (isDev) {
  // Development
  app.on('ready', function () {
    require('electron-debug')({ isEnabled: true })

    const {
      default: installExtension,
      REACT_DEVELOPER_TOOLS,
      REDUX_DEVTOOLS
    } = require('electron-devtools-installer')

    installExtension([REACT_DEVELOPER_TOOLS, REDUX_DEVTOOLS])
      .then((extName) => logger.debug(`Added Extension:  ${extName}`))
      .catch((err) => logger.debug('An error occurred: ', err))
  })
} else {
  // Production
  if (config.sentryDsn) {
    Raven.config(config.sentryDsn, {
      captureUnhandledRejections: true,
      release: app.getVersion(),
      tags: {
        process: process.type,
        electron: process.versions.electron,
        chrome: process.versions.chrome,
        platform: os.platform(),
        platform_release: os.release()
      }
    }).install()
  }
}

app.on('window-all-closed', function () {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

createWindow(config)

app.on('ready', function () {
  logger.info('App ready, initializing...')

  initMenu()
  initContextMenu()

  createClient(config)
})
