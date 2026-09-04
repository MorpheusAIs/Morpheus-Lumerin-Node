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
     * ManagedProcess. ManagedProcess then applies its own adoption policy when
     * start() verifies the occupied port.
     */
    reclaimIfDetected?: boolean
  }
) => {
  const { log } = params
  const managedParams = params as ManagedProcessParams
  const hasCommand = !!managedParams.command
  const pinger = new GenericApiResponseDetector({
    url: params.probe.url,
    method: params.probe.method,
    timeout: params.probe.timeout,
    pollInterval: params.probe.interval,
    log: log
    // responseRegexp: probeConfig.responseRegexp,
  })

  // A service that must be app-owned cannot be identified by its public
  // health response. Skip adoption detection and let ManagedProcess validate
  // the port plus a fresh per-spawn identity.
  const requiresOwnedProcess = hasCommand && managedParams.adoptIfDetected !== true
  const isDetected = requiresOwnedProcess
    ? false
    : await pinger
        .ping(PING_TIMEOUT)
        .then(() => true)
        .catch(() => false)

  log?.info(`already running process was ${isDetected ? 'detected' : 'not detected'}`)

  // Previously ANY detected listener forced an ExternalProcess, permanently.
  // A zombie service left over from a previous crash can answer the probe. For
  // services that should be app-owned, retain a ManagedProcess so start() can
  // apply the configured fail-open or fail-closed adoption policy rather than
  // permanently classifying the listener as an external dependency.
  if (hasCommand && (!isDetected || params.reclaimIfDetected)) {
    log?.info(
      isDetected
        ? 'detected an already-running instance; deferring ownership checks to startup'
        : 'creating managed process'
    )
    return new ManagedProcess({
      ...managedParams,
      pinger
    })
  }

  log?.info('creating external process')
  return new ExternalProcess({ pinger, log: params.log, onStateChange: params.onStateChange })
}
