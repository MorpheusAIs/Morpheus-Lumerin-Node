'use strict'

const chalk = require('chalk')

const logger = require('electron-log')
const stringify = require('json-stringify-safe')
import config from './config'

logger.transports.file.appName = 'lumerin-wallet-desktop'

export function getColorLevel(level = '') {
  const colors = {
    error: 'red',
    verbose: 'cyan',
    warn: 'yellow',
    debug: 'magenta',
    silly: 'blue'
  }
  return colors[level.toString()] || 'green'
}

logger.transports.console = function ({ date, level, data }) {
  const color = getColorLevel(level)

  let meta = ''
  if (data.length) {
    meta += ' => '
    meta += data.map((d) => (typeof d === 'object' ? stringify(d) : d)).join(', ')
  }

  // eslint-disable-next-line no-console
  console.log(`${date.toISOString()} - ${chalk[color](level)}:\t${meta}`)
}

if (config.debug) {
  logger.transports.console.level = 'debug'
  logger.transports.file.level = 'debug'
} else {
  logger.transports.console.level = 'warn'
  logger.transports.file.level = 'warn'
}

export default logger
