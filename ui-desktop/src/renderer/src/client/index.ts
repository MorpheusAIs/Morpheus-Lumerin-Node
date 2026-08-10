import debounce from 'lodash/debounce';
import get from 'lodash/get';
import pickBy from 'lodash/pickBy';

import * as utils from './utils';
import keys from './keys';
import './sentry';

const createClient = function (createStore) {
  const reduxDevtoolsOptions = {
    // actionsBlacklist: ['price-updated$'],
    features: { dispatch: true },
    // maxAge: 100 // default: 50
  };

  const store = createStore(reduxDevtoolsOptions);

  const onUIReady = (_ev, payload) => {
    const debounceTime = get(
      payload,
      'data.config.statePersistanceDebounce',
      0,
    );

    // keysToPersist keys that are passed from global redux state to main process.
    // For now only chain data is used.
    // TODO: subscribe for changes only within listed branch of redux state
    const keysToPersist = ['chain'];

    store.subscribe(
      debounce(
        function () {
          const passedState = pickBy(store.getState(), function (_value, key) {
            return keysToPersist.includes(key);
          });

          utils
            .forwardToMainProcess('persist-state')(passedState)
            .catch((err) =>
              // eslint-disable-next-line no-console
              console.warn(`Error persisting state: ${err.message}`),
            );
        },
        debounceTime,
        { maxWait: 2 * debounceTime },
      ),
    );
  };

  window.ipcRenderer.on('ui-ready', onUIReady);

  const onTransactionLinkClick = (txHash) =>
    window.openLink('https://etherscan.io/tx/' + txHash);

  const onTermsLinkClick = () =>
    window.openLink(
      'https://github.com/Lumerin-protocol/WalletDesktop/blob/main/LICENSE',
    );

  const onHelpLinkClick = () => window.openLink('https://mor.org/fair-launch');

  const onLinkClick = (url) => window.openLink(url);

  const copyToClipboard = (text) =>
    Promise.resolve(window.copyToClipboard(text));

  const lockSendTransaction = () => {
    store.dispatch({
      type: 'allow-send-transaction',
      payload: { allowSendTransaction: false },
    });
  };

  const unlockSendTransaction = () => {
    store.dispatch({
      type: 'allow-send-transaction',
      payload: { allowSendTransaction: true },
    });
  };

  const onInit = () => {
    window.addEventListener('beforeunload', function () {
      utils.sendToMainProcess('ui-unload');
    });
    window.addEventListener('online', () => {
      store.dispatch({
        type: 'connectivity-state-changed',
        payload: { ok: true },
      });
    });
    window.addEventListener('offline', () => {
      store.dispatch({
        type: 'connectivity-state-changed',
        payload: { ok: false },
      });
    });
    return utils.sendToMainProcess('ui-ready');
  };

  const forwardedMethods = {
    refreshAllTransactions: utils.forwardToMainProcess(
      'refresh-all-transactions',
      120000,
    ),
    refreshAllContracts: utils.forwardToMainProcess(
      'refresh-all-contracts',
      120000,
    ),
    onOnboardingCompleted: utils.forwardToMainProcess('onboarding-completed'),
    suggestAddresses: utils.forwardToMainProcess('suggest-addresses'),
    validatePassword: utils.forwardToMainProcess('validate-password'),
    changePassword: utils.forwardToMainProcess('change-password'),
    onLoginSubmit: utils.forwardToMainProcess('login-submit'),
    // Transfers.
    //
    // These must comfortably EXCEED the proxy-router's own mining timeout
    // (lib.DefaultTxMineTimeout, currently 1 minute) — SendMOR/SendETH block on
    // WaitMinedWithTimeout, so the backend can legitimately take that long
    // before answering. An equal timeout on this side is a race: we would
    // report "Operation timed out" for a transaction that was broadcast and is
    // still mining, and the obvious user reaction is to send again. 90s leaves
    // the backend room to return its own success or timeout first, so the UI
    // always reflects a decided outcome.
    sendMor: utils.forwardToMainProcess('send-mor', 90000),
    sendEth: utils.forwardToMainProcess('send-eth', 90000),
    // Multi-wallet. Switching restarts the proxy-router's session machinery,
    // so it gets a longer budget than a plain read.
    getWallets: utils.forwardToMainProcess('get-wallets', 20000),
    addHdWallet: utils.forwardToMainProcess('add-hd-wallet', 30000),
    importWallet: utils.forwardToMainProcess('import-wallet', 20000),
    switchWallet: utils.forwardToMainProcess('switch-wallet', 45000),
    removeWallet: utils.forwardToMainProcess('remove-wallet', 20000),
    renameWallet: utils.forwardToMainProcess('rename-wallet', 20000),
    clearCache: utils.forwardToMainProcess('clear-cache'),
    handleClientSideError: utils.forwardToMainProcess('handle-client-error'),
    logout: utils.forwardToMainProcess('logout'),
    getProxyRouterSettings: utils.forwardToMainProcess(
      'get-proxy-router-settings',
    ),
    getDefaultCurrencySetting: utils.forwardToMainProcess(
      'get-default-currency-settings',
    ),
    setDefaultCurrencySetting: utils.forwardToMainProcess(
      'set-default-currency-settings',
    ),
    saveProxyRouterSettings: utils.forwardToMainProcess(
      'save-proxy-router-settings',
    ),
    // NOTE: `get-marketplace-fee` and `claim-faucet` were removed here — neither
    // had a handler registered in the main process, so calling them hung until
    // the timeout and then failed with a generic "Operation timed out". Nothing
    // in the UI referenced them. If a faucet is reintroduced, register the IPC
    // channel in src/main/src/client/subscriptions/index.ts at the same time.
    getCustomEnvValues: utils.forwardToMainProcess('get-custom-env-values'),
    setCustomEnvValues: utils.forwardToMainProcess('set-custom-env-values'),
    getProfitSettings: utils.forwardToMainProcess('get-profit-settings'),
    setProfitSettings: utils.forwardToMainProcess('set-profit-settings'),
    getAutoAdjustPriceData: utils.forwardToMainProcess('get-auto-adjust-price'),
    setAutoAdjustPriceData: utils.forwardToMainProcess('set-auto-adjust-price'),
    getContractHashrate: utils.forwardToMainProcess('get-contract-hashrate'),
    // API Gateway
    getAuthHeaders: utils.forwardToMainProcess('get-auth-headers'),
    getAllModels: utils.forwardToMainProcess('get-all-models'),

    getTransactions: utils.forwardToMainProcess('get-transactions'),
    getBalances: utils.forwardToMainProcess('get-balances'),
    getRates: utils.forwardToMainProcess('get-rates'),
    getTodaysBudget: utils.forwardToMainProcess('get-todays-budget'),
    getTokenSupply: utils.forwardToMainProcess('get-supply'),
    // Chat History
    getChatHistoryTitles: utils.forwardToMainProcess('get-chat-history-titles'),
    getChatHistory: utils.forwardToMainProcess('get-chat-history', 750000),
    deleteChatHistory: utils.forwardToMainProcess(
      'delete-chat-history',
      750000,
    ),
    updateChatHistoryTitle: utils.forwardToMainProcess(
      'update-chat-history-title',
      750000,
    ),
    // Failover
    getFailoverSetting: utils.forwardToMainProcess(
      'get-failover-setting',
      750000,
    ),
    setFailoverSetting: utils.forwardToMainProcess(
      'set-failover-setting',
      750000,
    ),
    checkProviderConnectivity: utils.forwardToMainProcess(
      'check-provider-connectivity',
      750000,
    ),

    // IPFS
    getIpfsVersion: utils.forwardToMainProcess('get-ipfs-version', 750000),
    getIpfsFile: utils.forwardToMainProcess('get-ipfs-file', null),
    pinIpfsFile: utils.forwardToMainProcess('pin-ipfs-file', 750000),
    unpinIpfsFile: utils.forwardToMainProcess('unpin-ipfs-file', 750000),
    addFileToIpfs: utils.forwardToMainProcess('add-file-to-ipfs', 750000),
    getIpfsPinnedFiles: utils.forwardToMainProcess(
      'get-ipfs-pinned-files',
      750000,
    ),

    openSelectFolderDialog: utils.forwardToMainProcess(
      'open-select-folder-dialog',
      750000,
    ),

    // Agents
    getAgentUsers: utils.forwardToMainProcess('get-agent-users', 750000),
    confirmDeclineAgentUser: utils.forwardToMainProcess(
      'confirm-decline-agent-user',
      750000,
    ),
    removeAgentUser: utils.forwardToMainProcess('remove-agent-user', 750000),
    getAgentTxs: utils.forwardToMainProcess('get-agent-txs', 750000),
    revokeAgentAllowance: utils.forwardToMainProcess(
      'revoke-agent-allowance',
      750000,
    ),
    getAgentAllowanceRequests: utils.forwardToMainProcess(
      'get-agent-allowance-requests',
      750000,
    ),
    confirmDeclineAgentAllowanceRequest: utils.forwardToMainProcess(
      'confirm-decline-agent-allowance-request',
      750000,
    ),

    // Startup services
    startServices: utils.forwardToMainProcess('start-services', 750000),
    restartService: utils.forwardToMainProcess('restart-service', 750000),
    pingService: utils.forwardToMainProcess('ping-service', 750000),
    quitApp: utils.forwardToMainProcess('quit-app', 750000),
  };

  const api = {
    ...utils,
    ...forwardedMethods,
    isValidMnemonic: keys.isValidMnemonic,
    createMnemonic: keys.createMnemonic,
    onTermsLinkClick,
    onTransactionLinkClick,
    copyToClipboard,
    onHelpLinkClick,
    getAppVersion: window.getAppVersion,
    onLinkClick,
    onInit,
    store,
    lockSendTransaction,
    unlockSendTransaction,
  };

  return api;
};

export default createClient;
export type Client = ReturnType<typeof createClient>;
