import { app } from 'electron'
import isDev from 'electron-is-dev'
import * as os from 'os'
import { config as RavenConfig } from 'raven-js'
import { createWindow } from './src/main-window.js'
import { createClient } from './src/client'
import config from './config'
import initContextMenu from './contextMenu'
import initMenu from './menu'
import errorHandler from './errorHandler'
import logger from './logger'

errorHandler({ logger: logger.error })
console.log('electron bootstrap')
console.log('app config: ', config)
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
    RavenConfig(config.sentryDsn, {
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

console.log('app config: ', config)
createWindow()

app.on('ready', function () {
  logger.info('App ready, initializing...')

  initMenu()
  initContextMenu()

  createClient(config)
})
