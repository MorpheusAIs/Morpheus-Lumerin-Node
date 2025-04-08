import chalk from 'chalk'
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

Logger.transports.console.format = formatFn
Logger.transports.file.format = formatFn

Logger.transports.console.level = process.env.LOG_LEVEL
Logger.transports.file.level = process.env.LOG_LEVEL

export default Logger
