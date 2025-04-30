import chalk from 'chalk'
const electronLog = require('electron-log')
const stringify = require('json-stringify-safe')

// const logger = electronLog.create('wallet-core')
const Logger = electronLog

export function getColorLevel(level) {
  const colors = {
    error: 'red',
    warn: 'yellow',
    info: 'green',
    verbose: 'cyan',
    debug: 'magenta',
    silly: 'blue'
  }

  const key = colors[level] ? level : 'info'

  return colors[key]
}

const formatFn = (props) => {
  const { level, data, date, scope } = props.message
  const color = getColorLevel(level)

  let meta = ''
  if (data.length) {
    meta += ' => '
    meta += data.map((d) => (typeof d === 'object' ? stringify(d) : d)).join(', ')
  }

  return [`${date.toISOString()} ${chalk[color](level)} ${scope ? `[${scope}]` : ''}: ${meta}`]
}

Logger.transports.console.format = formatFn
Logger.transports.file.format = formatFn

Logger.transports.console.level = process.env.LOG_LEVEL
Logger.transports.file.level = process.env.LOG_LEVEL

export default Logger
