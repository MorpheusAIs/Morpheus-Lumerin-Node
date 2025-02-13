import { createSelector } from 'reselect';
import sortBy from 'lodash/sortBy';
import get from 'lodash/get';

import * as utils from '../utils';
import { getChain, getRate, getRateEth } from './chain';
import { getConfig } from './config';
import { getIsOnline } from './connectivity';
import {
  fromTokenBaseUnitsToETH,
  fromTokenBaseUnitsToLMR,
} from '../../utils/coinValue';
import { toUSD } from '../utils/syncAmounts';

export const getWallet = createSelector(
  getChain,
  (chainData) => chainData.wallet,
);

// Returns the Wallet address
export const getWalletAddress = createSelector(
  getWallet,
  (walletData) => walletData.address,
);

// Returns the marketplace fee
export const getMarketplaceFee = createSelector(
  getWallet,
  (walletData) => walletData.marketplaceFee,
);

// Returns the LMR balance of the active address in wei
export const getWalletEthBalance = createSelector(
  getWallet,
  (walletData) => fromTokenBaseUnitsToETH(walletData.ethBalance),
  // (walletData) => walletData.ethBalance / lmrDecimals
);

// Returns the LMR balance of the active address in wei
export const getWalletLmrBalance = createSelector(getWallet, (walletData) =>
  fromTokenBaseUnitsToLMR(get(walletData, 'token.lmrBalance', 0)),
);

export const getWalletLmrBalanceUSD = createSelector(
  getWalletLmrBalance,
  getRate,
  (lmrBalance, rate) => toUSD(lmrBalance, rate),
);

export const getWalletEthBalanceUSD = createSelector(
  getWalletEthBalance,
  getRateEth,
  (ethBalance, rateEth) => toUSD(ethBalance, rateEth),
);

// Returns the LMR balance of the active address in wei
export const getLmrBalanceWei = getWalletLmrBalance;

// Returns the array of transactions of the current chain/wallet/address.
// The items are mapped to contain properties useful for rendering.
export const getTransactions = createSelector(getWallet, (walletData) => {
  const transactionParser = utils.createTransactionParser(walletData.address);

  const transactions = Object.values(walletData?.token?.transactions) || [];

  const sorted = sortBy(transactions, [
    'receipt.blockNumber',
    'receipt.transactionIndex',
    'transaction.nonce',
  ]).reverse();

  return sorted.map(transactionParser);
});

export const getTransactionPage = createSelector(
  getWallet,
  (walletData) => walletData.page,
);

export const getTransactionPageSize = createSelector(
  getWallet,
  (walletData) => walletData.pageSize,
);

export const getHasNextPage = createSelector(
  getWallet,
  (walletData) => walletData.hasNextPage,
);

// Returns if the current wallet/address has transactions on the active chain
export const hasTransactions = createSelector(
  getTransactions,
  (transactions) => transactions.length !== 0,
);

// Returns wallet transactions sync status on the active chain
export const getTxSyncStatus = createSelector(
  getWallet,
  (walletData) => walletData.syncStatus,
);

// Returns the status of the "Send Lumerin" feature on the chain
export const sendLmrFeatureStatus = createSelector(
  getWalletLmrBalance,
  getIsOnline,
  getWalletEthBalance,
  (lmrBalance, isOnline, ethBalance) =>
    isOnline
      ? utils.hasFunds(lmrBalance) || utils.hasFunds(ethBalance)
        ? 'ok'
        : 'no-funds'
      : 'offline',
);

// Returns the status of the "Receive Lumerin" feature on the chain
export const receiveLmrFeatureStatus = createSelector(
  getWalletLmrBalance,
  getIsOnline,
  (lmrBalance, isOnline) =>
    isOnline ? (utils.hasFunds(lmrBalance) ? 'ok' : 'no-funds') : 'offline',
);

// Returns the status of the "Retry Import" feature on the chain
export const retryImportLmrFeatureStatus = createSelector(
  getIsOnline,
  getConfig,
  (isOnline, config) =>
    config.chain.chainId ? (isOnline ? 'ok' : 'no-coin') : 'offline',
);

export const getMergedTransactions = createSelector(
  getChain,
  (chain) => chain.wallet.token.transactions,
);

// Returns a transaction object given a transaction hash
export const getTransactionFromHash = createSelector(
  getTransactions,
  getWalletAddress,
  (props) => props.hash,
  (transactions, activeAddress, hash) =>
    transactions
      .map(utils.createTransactionParser(activeAddress))
      .find((tx) => tx.hash === hash),
);

export const isAllowSendTransaction = createSelector(getWallet, (walletData) =>
  get(walletData, 'allowSendTransaction', true),
);
