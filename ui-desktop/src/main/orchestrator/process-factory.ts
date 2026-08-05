import { GenericApiResponseDetector } from './detectors/generic-api-response'
import { ExternalProcess, ExternalProcessParams } from './external-process'
import { ProbeConfig } from './orchestrator.types'
import { ManagedProcess, ManagedProcessParams } from './managed-process'

const PING_TIMEOUT = 500

export const ProcessFactory = async (
  params: (ManagedProcessParams | ExternalProcessParams) & {
    probe: ProbeConfig
    /**
     * When true, an already-listening service is still wrapped in a
     * ManagedProcess so the app retains the ability to stop/restart it.
     * Set for services we own (proxy-router) but not for genuinely external
     * dependencies like Docker.
     */
    reclaimIfDetected?: boolean
  }
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

  const hasCommand = !!(params as ManagedProcessParams).command

  // Previously ANY detected listener forced an ExternalProcess, permanently.
  // A zombie proxy-router left over from a previous crash would answer the
  // probe, get classified as "external", and from then on the app refused to
  // stop or restart it ("Cannot restart external service") — with no recovery
  // path short of the user killing the process by hand. For services we own we
  // now keep a ManagedProcess so restart still works; ManagedProcess.start()
  // short-circuits when the health check already passes, so a genuinely healthy
  // instance is left alone.
  if (hasCommand && (!isDetected || params.reclaimIfDetected)) {
    log?.info(
      isDetected
        ? 'detected an already-running instance; managing it so restart stays available'
        : 'creating managed process'
    )
    return new ManagedProcess({
      ...(params as ManagedProcessParams),
      pinger
    })
  }

  log?.info('creating external process')
  return new ExternalProcess({ pinger, log: params.log, onStateChange: params.onStateChange })
}
