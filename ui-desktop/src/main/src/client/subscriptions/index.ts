import { Core } from './core.types'
import * as handlers from './handlers'
import { ApplyFnWithFirstArg } from './parial-apply.types'
import utils from './utils'

const listeners = {
  'validate-password': handlers.validatePassword,
  'change-password': handlers.changePassword,
  'persist-state': handlers.persistState,
  'clear-cache': handlers.clearCache,
  'handle-client-error': handlers.handleClientSideError,
  'save-proxy-router-settings': handlers.saveProxyRouterSettings,
  'get-proxy-router-settings': handlers.getProxyRouterSettings,
  'get-default-currency-settings': handlers.getDefaultCurrency,
  'set-default-currency-settings': handlers.setDefaultCurrency,
  'get-custom-env-values': handlers.getCustomEnvs,
  'set-custom-env-values': handlers.setCustomEnvs,
  'get-profit-settings': handlers.getProfitSettings,
  'set-profit-settings': handlers.setProfitSettings,
  'get-contract-hashrate': handlers.getContractHashrate,
  'get-auto-adjust-price': handlers.getAutoAdjustPriceData,
  'set-auto-adjust-price': handlers.setAutoAdjustPriceData,
  'onboarding-completed': handlers.onboardingCompleted,
  'suggest-addresses': handlers.suggestAddresses,
  'login-submit': handlers.onLoginSubmit,
  // Api Gateway
  'get-all-models': handlers.getAllModels,
  'get-transactions': handlers.getTransactions,
  'get-balances': handlers.getBalances,
  'get-rates': handlers.getMorRate,
  'get-todays-budget': handlers.getTodaysBudget,
  'get-supply': handlers.getTokenSupply,
  'get-auth-headers': handlers.getAuthHeaders,
  // Chat history
  'get-chat-history-titles': handlers.getChatHistoryTitles,
  'get-chat-history': handlers.getChatHistory,
  'delete-chat-history': handlers.deleteChatHistory,
  'update-chat-history-title': handlers.updateChatHistoryTitle,
  // Failover
  'get-failover-setting': handlers.isFailoverEnabled,
  'set-failover-setting': handlers.setFailoverSetting,
  'check-provider-connectivity': handlers.checkProviderConnectivity,

  // IPFS
  'get-ipfs-version': handlers.getIpfsVersion,
  'get-ipfs-file': handlers.getIpfsFile,
  'pin-ipfs-file': handlers.pinIpfsFile,
  'unpin-ipfs-file': handlers.unpinIpfsFile,
  'add-file-to-ipfs': handlers.addFileToIpfs,
  'get-ipfs-pinned-files': handlers.getIpfsPinnedFiles,

  'open-select-folder-dialog': handlers.openSelectFolderDialog,

  // Agent
  'get-agent-users': handlers.getAgentUsers,
  'confirm-decline-agent-user': handlers.confirmDeclineAgentUser,
  'remove-agent-user': handlers.removeAgentUser,
  'get-agent-txs': handlers.getAgentTxs,
  'revoke-agent-allowance': handlers.revokeAgentAllowance,
  'get-agent-allowance-requests': handlers.getAgentAllowanceRequests,
  'confirm-decline-agent-allowance-request': handlers.confirmDeclineAgentAllowanceRequest,
  'start-services': handlers.startServices,
  'restart-service': handlers.restartService,
  'ping-service': handlers.pingService
}

// Keeps references to wrapped listeners (with core injected) to be able to unsubscribe from them
const wrappedListeners = {}

// Subscribe to messages where no core has to react
export const subscribe = (core: Core) => {
  // Inject core into listeners
  Object.entries(listeners).forEach(function ([event, handler]) {
    wrappedListeners[event] = (eventData: unknown) => handler(eventData, core)
  })
  utils.subscribeTo(wrappedListeners, 'none')
}

export const unsubscribe = () => utils.unsubscribeTo(wrappedListeners)

export default { subscribe, unsubscribe }
export type Client = ApplyFnWithFirstArg<typeof handlers>
