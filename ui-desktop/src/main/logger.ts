import chalk from 'chalk'
import logger from 'electron-log'
import stringify from 'json-stringify-safe'
import config from './config'

export function getColorLevel(level = '') {
  const colors = {
    error: 'red',
    verbose: 'cyan',
    warn: 'yellow',
    debug: 'magenta',
    silly: 'blue'
  } as const

  if (level in colors) {
    return colors[level as keyof typeof colors]
  }
  // infer typing in the return
  return 'green'
}

logger.transports.console = function ({ date, level, data, scope }) {
  const color = getColorLevel(level)

  let meta = ''
  if (data.length) {
    meta += ' => '
    meta += data.map((d) => (typeof d === 'object' ? stringify(d) : d)).join(', ')
  }

  // eslint-disable-next-line no-console
  console.log(`${date.toISOString()} ${chalk[color](level)} ${scope ? `[${scope}]` : ''}: ${meta}`)
}

if (config.debug) {
  logger.transports.console.level = 'debug'
  logger.transports.file.level = 'debug'
} else {
  logger.transports.console.level = 'warn'
  logger.transports.file.level = 'warn'
}

export default logger
