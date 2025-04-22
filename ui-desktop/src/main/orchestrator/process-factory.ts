import { GenericApiResponseDetector } from './detectors/generic-api-response'
import { ExternalProcess, ExternalProcessParams } from './external-process'
import { ProbeConfig } from './orchestrator.types'
import { ManagedProcess, ManagedProcessParams } from './managed-process'

const PING_TIMEOUT = 500

export const ProcessFactory = async (
  params: (ManagedProcessParams | ExternalProcessParams) & { probe: ProbeConfig }
) => {
  const { log } = params
  const pinger = new GenericApiResponseDetector({
    url: params.probe.url,
    method: params.probe.method,
    timeout: params.probe.timeout,
    pollInterval: params.probe.interval,
    log: log
    // responseRegexp: probeConfig.responseRegexp,
  })

  const isDetected = await pinger
    .ping(PING_TIMEOUT)
    .then(() => true)
    .catch(() => false)

  log?.info(`already running process was ${isDetected ? 'detected' : 'not detected'}`)

  if (!isDetected && (params as ManagedProcessParams).command) {
    log?.info('creating managed process')
    return new ManagedProcess({
      ...(params as ManagedProcessParams),
      pinger
    })
  }

  log?.info('creating external process')
  return new ExternalProcess({ pinger, log: params.log, onStateChange: params.onStateChange })
}
