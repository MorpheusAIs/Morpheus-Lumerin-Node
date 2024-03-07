export const subscribeToMainProcessMessages = function(store) {
  const ipcMessages = [
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
    'allow-send-transaction'
  ];

  // Subscribe to every IPC message defined above and dispatch a
  // Redux action of type { type: MSG_NAME, payload: MSG_ARG }
  ipcMessages.forEach(msgName =>
    window.ipcRenderer.on(msgName, (_, payload) =>
      store.dispatch({ type: msgName, payload })
    )
  );
};
