'use strict'
const electron = require('electron')
const debounce = require('lodash.debounce')

const dialog = electron.dialog || electron.remote.dialog

let installed = false

let options = {
  logger: console.error
}

const handleError = (error) => {
  try {
    options.logger('handle error: ', error?.message, error?.stack, error)
  } catch (loggerError) {
    // eslint-disable-line unicorn/catch-error-name
    dialog.showErrorBox(
      'The `logger` option function in electron-unhandled threw an error',
      loggerError.stack
    )
    return
  }
}

export default (inputOptions) => {
  if (installed) {
    return
  }

  installed = true

  options = {
    ...options,
    ...inputOptions
  }

  if (process.type === 'renderer') {
    const errorHandler = debounce((error) => {
      handleError(error)
    }, 200)
    window.addEventListener('error', (event) => {
      event.preventDefault()
      errorHandler(event.error || event)
    })

    const rejectionHandler = debounce((reason) => {
      handleError(reason)
    }, 200)
    window.addEventListener('unhandledrejection', (event) => {
      event.preventDefault()
      rejectionHandler(event.reason)
    })
  } else {
    process.on('uncaughtException', (error) => {
      handleError(error)
    })

    process.on('unhandledRejection', (error) => {
      handleError(error)
    })
  }
}
