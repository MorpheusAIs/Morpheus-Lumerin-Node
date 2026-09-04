import { clipboard, contextBridge, ipcRenderer } from 'electron'

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
  'get-auth-headers',
  'get-all-models',
  'get-providers',
  'get-local-models',
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
  'refresh-all-transactions',
  'refresh-all-contracts',
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
if (process.contextIsolated) {
  try {
    const copyToClipboard = function (text) {
      return clipboard.writeText(text)
    }

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
    contextBridge.exposeInMainWorld('chatStream', chatStream)
    contextBridge.exposeInMainWorld('ipfsDownload', ipfsDownload)
  } catch (error) {
    console.error(error)
  }
}
