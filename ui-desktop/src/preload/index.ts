import { ipcRenderer, clipboard, contextBridge } from 'electron'

const coworkChannels = {
  getApprovalPolicy: 'cowork:get-approval-policy',
  updateApprovalPolicy: 'cowork:update-approval-policy',
  listProjects: 'cowork:list-projects',
  createProject: 'cowork:create-project',
  updateProject: 'cowork:update-project',
  deleteProject: 'cowork:delete-project',
  listTasks: 'cowork:list-tasks',
  getTask: 'cowork:get-task',
  listTaskMessages: 'cowork:list-task-messages',
  createTask: 'cowork:create-task',
  startTask: 'cowork:start-task',
  steerTask: 'cowork:steer-task',
  cancelTask: 'cowork:cancel-task',
  pauseTask: 'cowork:pause-task',
  rebindTask: 'cowork:rebind-task',
  resolveApproval: 'cowork:resolve-approval',
  deleteTask: 'cowork:delete-task',
  listModelOptions: 'cowork:list-model-options',
  previewArtifact: 'cowork:preview-artifact',
  revealArtifact: 'cowork:reveal-artifact',
  listSchedules: 'cowork:list-schedules',
  createSchedule: 'cowork:create-schedule',
  updateSchedule: 'cowork:update-schedule',
  pauseSchedule: 'cowork:pause-schedule',
  resumeSchedule: 'cowork:resume-schedule',
  deleteSchedule: 'cowork:delete-schedule',
  runScheduleNow: 'cowork:run-schedule-now',
  listExtensions: 'cowork:list-extensions',
  configureExtensions: 'cowork:configure-extensions',
  event: 'cowork:event'
} as const

// Temporary compatibility boundary for the legacy renderer client. Keeping a
// fixed list prevents injected renderer code from selecting arbitrary main
// process channels while the older screens are migrated to narrow APIs.
const legacyIpcChannels = new Set([
  'ui-ready',
  'ui-unload',
  'persist-state',
  'handle-client-error',
  'onboarding-completed',
  'suggest-addresses',
  'validate-password',
  'change-password',
  'login-submit',
  'send-mor',
  'send-eth',
  'parse-attachment',
  'get-wallets',
  'add-hd-wallet',
  'import-wallet',
  'switch-wallet',
  'remove-wallet',
  'rename-wallet',
  'clear-cache',
  'logout',
  'get-proxy-router-settings',
  'get-default-currency-settings',
  'set-default-currency-settings',
  'save-proxy-router-settings',
  'get-custom-env-values',
  'set-custom-env-values',
  'get-profit-settings',
  'set-profit-settings',
  'get-auto-adjust-price',
  'set-auto-adjust-price',
  'get-contract-hashrate',
  'get-all-models',
  'get-providers',
  'get-local-models',
  'get-node-config',
  'update-eth-node',
  'get-sessions-by-user',
  'get-bids-by-model',
  'get-bid-info',
  'close-session',
  'open-session',
  'get-sessions-by-provider',
  'get-provider-claimable-balance',
  'claim-provider-funds',
  'synthesize-speech',
  'transcribe-audio',
  'get-transactions',
  'get-balances',
  'get-rates',
  'get-todays-budget',
  'get-supply',
  'get-chat-history-titles',
  'get-chat-history',
  'delete-chat-history',
  'update-chat-history-title',
  'get-failover-setting',
  'set-failover-setting',
  'check-provider-connectivity',
  'get-ipfs-version',
  'pin-ipfs-file',
  'unpin-ipfs-file',
  'add-file-to-ipfs',
  'get-ipfs-pinned-files',
  'get-agent-users',
  'confirm-decline-agent-user',
  'remove-agent-user',
  'get-agent-txs',
  'revoke-agent-allowance',
  'get-agent-allowance-requests',
  'confirm-decline-agent-allowance-request',
  'start-services',
  'restart-service',
  'ping-service',
  'quit-app',
  'indexer-connection-status-changed',
  'lumerin-token-status-changed',
  'web3-connection-status-changed',
  'connectivity-state-changed',
  'proxy-router-connections-changed',
  'proxy-router-status-changed',
  'proxy-router-error',
  'transactions-scan-finished',
  'transactions-scan-started',
  'contracts-scan-finished',
  'contract-updated',
  'contracts-scan-started',
  'wallet-state-changed',
  'coin-price-updated',
  'network-difficulty-updated',
  'create-wallet',
  'open-wallet',
  'open-proxy-router',
  'eth-balance-changed',
  'token-balance-changed',
  'token-contract-received',
  'token-transactions-changed',
  'eth-tx',
  'lmr-tx',
  'coin-block',
  'transactions-next-page',
  'devices-device-updated',
  'devices-state-updated',
  'proxy-router-type-changed',
  'allow-send-transaction',
  'services-state',
  'wallet-error'
])

const coworkTaskStatuses = new Set([
  'draft',
  'queued',
  'running',
  'waiting_approval',
  'paused',
  'completed',
  'failed',
  'cancelled'
])

const boundedCoworkEventId = (value) =>
  typeof value === 'string' &&
  value.length >= 1 &&
  value.length <= 128 &&
  /^[A-Za-z0-9_-]+$/u.test(value)

const validCoworkTimestamp = (value) => Number.isSafeInteger(value) && value >= 0

const sanitizedCoworkTaskEvent = (payload) => {
  if (
    !payload ||
    typeof payload !== 'object' ||
    Array.isArray(payload) ||
    !boundedCoworkEventId(payload.taskId) ||
    !boundedCoworkEventId(payload.projectId) ||
    typeof payload.title !== 'string' ||
    payload.title.length < 1 ||
    payload.title.length > 240 ||
    /[\u0000-\u001f\u007f]/u.test(payload.title) ||
    !coworkTaskStatuses.has(payload.status) ||
    !validCoworkTimestamp(payload.createdAt) ||
    !validCoworkTimestamp(payload.updatedAt) ||
    (payload.startedAt !== undefined && !validCoworkTimestamp(payload.startedAt)) ||
    (payload.completedAt !== undefined && !validCoworkTimestamp(payload.completedAt))
  ) {
    return null
  }
  return {
    taskId: payload.taskId,
    projectId: payload.projectId,
    title: payload.title,
    status: payload.status,
    createdAt: payload.createdAt,
    updatedAt: payload.updatedAt,
    ...(payload.startedAt !== undefined ? { startedAt: payload.startedAt } : {}),
    ...(payload.completedAt !== undefined ? { completedAt: payload.completedAt } : {})
  }
}

// Unlike the legacy generic IPC bridge, Cowork exposes only fixed operations.
// The renderer cannot choose a main-process channel or supply a filesystem root.
const cowork = {
  getApprovalPolicy: () => ipcRenderer.invoke(coworkChannels.getApprovalPolicy),
  updateApprovalPolicy: (mode, expectedRevision) =>
    ipcRenderer.invoke(coworkChannels.updateApprovalPolicy, { mode, expectedRevision }),
  listProjects: () => ipcRenderer.invoke(coworkChannels.listProjects),
  createProject: (input) => ipcRenderer.invoke(coworkChannels.createProject, input),
  updateProject: (input) => ipcRenderer.invoke(coworkChannels.updateProject, input),
  deleteProject: (id) => ipcRenderer.invoke(coworkChannels.deleteProject, { id }),
  listTasks: (projectId) => ipcRenderer.invoke(coworkChannels.listTasks, { projectId }),
  getTask: (id) => ipcRenderer.invoke(coworkChannels.getTask, { id }),
  listTaskMessages: (taskId, beforeSequence, limit) =>
    ipcRenderer.invoke(coworkChannels.listTaskMessages, { taskId, beforeSequence, limit }),
  createTask: (input) => ipcRenderer.invoke(coworkChannels.createTask, input),
  startTask: (id) => ipcRenderer.invoke(coworkChannels.startTask, { id }),
  steerTask: (id, content) => ipcRenderer.invoke(coworkChannels.steerTask, { id, content }),
  cancelTask: (id) => ipcRenderer.invoke(coworkChannels.cancelTask, { id }),
  pauseTask: (id) => ipcRenderer.invoke(coworkChannels.pauseTask, { id }),
  rebindTask: (id, model) => ipcRenderer.invoke(coworkChannels.rebindTask, { id, model }),
  resolveApproval: (taskId, approvalId, approved) =>
    ipcRenderer.invoke(coworkChannels.resolveApproval, { taskId, approvalId, approved }),
  deleteTask: (id) => ipcRenderer.invoke(coworkChannels.deleteTask, { id }),
  listModelOptions: (force = false) =>
    ipcRenderer.invoke(coworkChannels.listModelOptions, { force }),
  previewArtifact: (taskId, path) =>
    ipcRenderer.invoke(coworkChannels.previewArtifact, { taskId, path }),
  revealArtifact: (taskId, path) =>
    ipcRenderer.invoke(coworkChannels.revealArtifact, { taskId, path }),
  listSchedules: (projectId) => ipcRenderer.invoke(coworkChannels.listSchedules, { projectId }),
  createSchedule: (input) => ipcRenderer.invoke(coworkChannels.createSchedule, input),
  updateSchedule: (input) => ipcRenderer.invoke(coworkChannels.updateSchedule, input),
  pauseSchedule: (id) => ipcRenderer.invoke(coworkChannels.pauseSchedule, { id }),
  resumeSchedule: (id) => ipcRenderer.invoke(coworkChannels.resumeSchedule, { id }),
  deleteSchedule: (id) => ipcRenderer.invoke(coworkChannels.deleteSchedule, { id }),
  runScheduleNow: (id) => ipcRenderer.invoke(coworkChannels.runScheduleNow, { id }),
  listExtensions: (projectId) => ipcRenderer.invoke(coworkChannels.listExtensions, { projectId }),
  configureExtensions: (projectId, settings) =>
    ipcRenderer.invoke(coworkChannels.configureExtensions, { projectId, ...settings }),
  onTaskEvent: (listener) => {
    const subscription = (_event, payload) => {
      const safeEvent = sanitizedCoworkTaskEvent(payload)
      if (safeEvent) listener(safeEvent)
    }
    ipcRenderer.on(coworkChannels.event, subscription)
    return () => ipcRenderer.removeListener(coworkChannels.event, subscription)
  }
}

const chatStreamChannels = {
  start: 'chat-stream:start',
  cancel: 'chat-stream:cancel',
  event: 'chat-stream:event'
} as const

const validChatStreamRequestId = (value) =>
  typeof value === 'string' &&
  /^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$/iu.test(value)

const chatStream = {
  start: (requestId, payload) => {
    if (!validChatStreamRequestId(requestId)) {
      return Promise.reject(new Error('Chat stream request ID is invalid.'))
    }
    return ipcRenderer.invoke(chatStreamChannels.start, { requestId, payload })
  },
  cancel: (requestId) => {
    if (!validChatStreamRequestId(requestId)) return
    ipcRenderer.send(chatStreamChannels.cancel, { requestId })
  },
  onEvent: (requestId, listener) => {
    if (!validChatStreamRequestId(requestId) || typeof listener !== 'function') {
      throw new Error('Chat stream subscription is invalid.')
    }
    const subscription = (_event, payload) => {
      if (payload?.requestId !== requestId) return
      const safePayload =
        payload.kind === 'chunk' &&
        typeof payload.dataBase64 === 'string' &&
        payload.dataBase64.length <= 87_384 &&
        /^[A-Za-z0-9+/]*={0,2}$/u.test(payload.dataBase64)
          ? { requestId, kind: 'chunk', dataBase64: payload.dataBase64 }
          : payload.kind === 'end'
            ? { requestId, kind: 'end' }
            : payload.kind === 'error' && typeof payload.message === 'string'
              ? { requestId, kind: 'error', message: payload.message.slice(0, 2_000) }
              : null
      if (safePayload) listener(safePayload)
    }
    ipcRenderer.on(chatStreamChannels.event, subscription)
    return () => ipcRenderer.removeListener(chatStreamChannels.event, subscription)
  }
}

const ipfsDownloadChannels = {
  selectFolder: 'ipfs-download:select-folder',
  start: 'ipfs-download:start',
  cancel: 'ipfs-download:cancel',
  event: 'ipfs-download:event'
} as const

const ipfsDownload = {
  selectFolder: () => ipcRenderer.invoke(ipfsDownloadChannels.selectFolder),
  start: (requestId, folderToken, cidHash) => {
    if (!validChatStreamRequestId(requestId) || !validChatStreamRequestId(folderToken)) {
      return Promise.reject(new Error('IPFS download identifiers are invalid.'))
    }
    if (typeof cidHash !== 'string' || !/^0x[0-9a-f]{64}$/iu.test(cidHash)) {
      return Promise.reject(new Error('IPFS metadata CID hash is invalid.'))
    }
    return ipcRenderer.invoke(ipfsDownloadChannels.start, { requestId, folderToken, cidHash })
  },
  cancel: (requestId) => {
    if (!validChatStreamRequestId(requestId)) return
    ipcRenderer.send(ipfsDownloadChannels.cancel, { requestId })
  },
  onEvent: (requestId, listener) => {
    if (!validChatStreamRequestId(requestId) || typeof listener !== 'function') {
      throw new Error('IPFS download subscription is invalid.')
    }
    const subscription = (_event, payload) => {
      if (payload?.requestId !== requestId) return
      let safePayload: any = null
      if (payload.kind === 'error' && typeof payload.message === 'string') {
        safePayload = { requestId, kind: 'error', message: payload.message.slice(0, 2_000) }
      } else if (payload.kind === 'progress') {
        const progress = payload.progress
        if (
          progress &&
          (progress.status === 'downloading' || progress.status === 'completed') &&
          Number.isSafeInteger(progress.downloaded) &&
          progress.downloaded >= 0 &&
          progress.downloaded <= 256 * 1024 * 1024 * 1024 &&
          Number.isSafeInteger(progress.total) &&
          progress.total >= 0 &&
          progress.total <= 256 * 1024 * 1024 * 1024 &&
          Number.isFinite(progress.percentage) &&
          progress.percentage >= 0 &&
          progress.percentage <= 100 &&
          Number.isSafeInteger(progress.timeUpdated) &&
          progress.timeUpdated >= 0
        ) {
          safePayload = {
            requestId,
            kind: 'progress',
            progress: {
              status: progress.status,
              downloaded: progress.downloaded,
              total: progress.total,
              percentage: progress.percentage,
              timeUpdated: progress.timeUpdated
            }
          }
        }
      }
      if (safePayload) listener(safePayload)
    }
    ipcRenderer.on(ipfsDownloadChannels.event, subscription)
    return () => ipcRenderer.removeListener(ipfsDownloadChannels.event, subscription)
  }
}
// Use `contextBridge` APIs to expose Electron APIs to
// renderer only if context isolation is enabled, otherwise
// just add to the DOM global.
if (process.contextIsolated) {
  try {
    // @see http://electronjs.org/docs/tutorial/security#2-disable-nodejs-integration-for-remote-content

    const copyToClipboard = function (text) {
      return clipboard.writeText(text)
    }

    // Was `remote.app.getVersion()` via @electron/remote. That module is
    // deprecated, widens the renderer's reach into main-process objects, and
    // was never initialised here anyway (no `@electron/remote/main`
    // initialize()/enable() call exists), so the call was failing at runtime.
    // A synchronous IPC call is both safer and actually works.
    const getAppVersion = function () {
      return ipcRenderer.sendSync('get-app-version')
    }

    const openLink = function (url) {
      return ipcRenderer.invoke('open-external-url', url)
    }

    contextBridge.exposeInMainWorld('ipcRenderer', {
      send(eventName, payload) {
        if (!legacyIpcChannels.has(eventName)) throw new Error('Unsupported desktop IPC channel.')
        return ipcRenderer.send(eventName, payload)
      },
      on(eventName, listener) {
        if (!legacyIpcChannels.has(eventName)) throw new Error('Unsupported desktop IPC channel.')
        // For some reason the listener passed into this function doesn't work
        // if you want to use it to unsubscribe later (likely due to chrome/node connection).
        // So we wrap it in a function and provide an unsubscribe function both to event handler
        // and as a returned value
        function unsubscribe() {
          ipcRenderer.removeListener(eventName, subscription)
        }

        function subscription(_event, payload) {
          // Electron's event object exposes sender methods and must never cross
          // the contextBridge. Only copy the application payload and a narrow
          // unsubscribe capability into the renderer.
          listener(payload, unsubscribe)
        }

        ipcRenderer.on(eventName, subscription)

        return unsubscribe
      }
    })

    contextBridge.exposeInMainWorld('openLink', openLink)
    contextBridge.exposeInMainWorld('getAppVersion', getAppVersion)
    contextBridge.exposeInMainWorld('copyToClipboard', copyToClipboard)
    contextBridge.exposeInMainWorld('cowork', cowork)
    contextBridge.exposeInMainWorld('chatStream', chatStream)
    contextBridge.exposeInMainWorld('ipfsDownload', ipfsDownload)

    // `isDev` used to be hardcoded to `true` (with the real check commented
    // out), so production builds reported themselves as development builds.
    // Nothing in the renderer consumes it, so rather than ship a value that is
    // both dead and wrong, it is gone. If it is needed again, derive it in the
    // main process and pass it over IPC — the renderer cannot determine it.
  } catch (error) {
    console.error(error)
  }
} else {
  // @ts-ignore (define in dts)
  window.cowork = cowork
  // @ts-ignore (define in dts)
  window.chatStream = chatStream
  // @ts-ignore (define in dts)
  window.ipfsDownload = ipfsDownload
}
