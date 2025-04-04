import chalk from 'chalk'
import logger from 'electron-log'
import stringify from 'json-stringify-safe'
import Logger, { LogMessage } from 'electron-log'

export function getColorLevel(level: Logger.LogLevel | string) {
  const colors = {
    error: 'red',
    warn: 'yellow',
    info: 'green',
    verbose: 'cyan',
    debug: 'magenta',
    silly: 'blue'
  } as const

  const key = colors[level] ? (level as Logger.LogLevel) : 'info'

  return colors[key]
}

const formatFn = (props: { message: LogMessage }) => {
  const { level, data, date, scope } = props.message
  const color = getColorLevel(level)

  let meta = ''
  if (data.length) {
    meta += ' => '
    meta += data.map((d) => (typeof d === 'object' ? stringify(d) : d)).join(', ')
  }

  return [`${date.toISOString()} ${chalk[color](level)} ${scope ? `[${scope}]` : ''}: ${meta}`]
}

logger.transports.console.format = formatFn
logger.transports.file.format = formatFn

logger.transports.console.level = process.env.LOG_LEVEL
logger.transports.file.level = process.env.LOG_LEVEL

export default logger
