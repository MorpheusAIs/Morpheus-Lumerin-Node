import { createSelector } from 'reselect';

import { getConfig } from './config';

// Returns the array of transactions of the current chain/wallet/address.
// The items are mapped to contain properties useful for rendering.
export const getProxyRouter = state => state.proxyRouter;

export const getIsProxyRouterConnect = createSelector(
  getProxyRouter,
  proxyRouterData => proxyRouterData.isConnected
);

export const getIsLocalProxyRouter = createSelector(
  getProxyRouter,
  proxyRouterData => proxyRouterData.isLocal
);

export const getLocalProxyRouterUrl = createSelector(
  getConfig,
  configData => configData.chain.localProxyRouterUrl
);

// Returns the array of transactions of the current chain/wallet/address.
// The items are mapped to contain properties useful for rendering.
export const getConnections = createSelector(
  getProxyRouter,
  proxyRouterData => proxyRouterData.connections
);

// Returns if the current wallet/address has transactions on the active chain
export const hasConnections = createSelector(
  getConnections,
  connections => connections.length !== 0
);

// Returns contracts sync status on the active chain
export const getProxyRouterSyncStatus = createSelector(
  getProxyRouter,
  proxyRouterData => proxyRouterData.syncStatus
);
