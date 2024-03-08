import { createSelector } from 'reselect';
import get from 'lodash/get';

import { getConfig } from './config';

// Returns the "chains" state branch
export const getChain = state => state.chain;

// Returns the chain-specific config of the active chain
export const getChainConfig = createSelector(
  getConfig,
  configData => configData.chain
);

// Returns the active chain display name
export const getChainDisplayName = createSelector(
  getChainConfig,
  chainConfigData => chainConfigData.displayName
);

// Returns the active chain "meta" state branch
export const getChainMeta = createSelector(
  getChain,
  chainData => chainData.meta
);

// Returns the active chain height
export const getBlockHeight = createSelector(
  getChainMeta,
  chainMetaData => chainMetaData.height
);

// Returns the active chain current gas price
export const getChainGasPrice = createSelector(
  getChainConfig,
  getChainMeta,
  (chainConfigData, chainMetaData) =>
    // Parity may return 0 as gasPrice if latests blocks are empty
    !chainMetaData.gasPrice || parseInt(chainMetaData.gasPrice, 10) <= 0
      ? chainConfigData.defaultGasPrice
      : chainMetaData.gasPrice
);

// Returns the active chain connection status
export const getChainConnectionStatus = createSelector(
  getChain,
  getChainMeta,
  (chainData, chainMetaData) =>
    chainData ? chainMetaData.isWeb3Connected : null
);

// Returns the explorer URL for a specific transaction
export const getTransactionExplorerUrl = createSelector(
  getChainConfig,
  (_, props) => props.hash,
  (chainConfigData, hash) =>
    chainConfigData.explorerUrl
      ? chainConfigData.explorerUrl.replace('{{hash}}', hash)
      : '#'
);

export const getContractExplorerUrl = createSelector(
  getChainConfig,
  (_, props) => props.hash,
  (chainConfigData, hash) =>
    chainConfigData.explorerUrl
      ? chainConfigData.explorerUrl
          .replace('tx', 'address') // TODO: Replace with url builder
          .replace('{{hash}}', hash)
      : '#'
);

// Returns the ProxyRouter URL
export const getProxyRouterUrl = createSelector(
  getChainConfig,
  chainConfigData => chainConfigData.proxyRouterUrl
);

// Returns the indexer connection status
export const getIndexerConnectionStatus = createSelector(
  getChain,
  getChainMeta,
  (chainData, chainMetaData) =>
    chainData ? chainMetaData.isIndexerConnected : null
);

// Returns the amount of confirmations for a given transaction
export const getTxConfirmations = createSelector(
  getBlockHeight,
  ({ tx }) => tx,
  (blockHeight, txBlockNumber) =>
    txBlockNumber === null || txBlockNumber > blockHeight
      ? 0
      : blockHeight - txBlockNumber + 1
);

export const getRate = createSelector(
  getChain,
  getChainMeta,
  (chainData, chainMetaData) => (chainData ? chainMetaData.rate : null)
);

export const getNetworkDifficulty = createSelector(
  getChain,
  getChainMeta,
  (chainData, chainMetaData) =>
    chainData ? chainMetaData.networkDifficulty : null
);

export const getRateEth = createSelector(
  getChain,
  getChainMeta,
  (chainData, chainMetaData) => (chainData ? chainMetaData.rateEth : null)
);

export const getRateBtc = createSelector(
  getChain,
  getChainMeta,
  (chainData, chainMetaData) => (chainData ? chainMetaData.rateBtc : null)
);

export const getChainReadyStatus = createSelector(
  getChain,
  getConfig,
  (chainData, configData) => {
    const chainConfig = configData.chain;
    const chainMeta = chainData.meta;
    const walletData = chainData.wallet;

    return {
      hasLmrBalance: get(walletData, 'token.lmrBalance', null) !== null,
      hasBlockHeight: chainMeta.height > -1,
      displayName: chainConfig.displayName,
      symbol: chainConfig.symbol
    };
  }
);
