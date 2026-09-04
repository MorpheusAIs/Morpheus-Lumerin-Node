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
  'get-ipfs-file',
  'pin-ipfs-file',
  'unpin-ipfs-file',
  'add-file-to-ipfs',
  'get-ipfs-pinned-files',
  'open-select-folder-dialog',
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
  } catch (error) {
    console.error(error)
  }
}
