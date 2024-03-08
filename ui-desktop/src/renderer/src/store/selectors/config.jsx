import { createSelector } from 'reselect';

// Returns the "config" state branch
export const getConfig = state => state.config;

export const getCoinSymbol = state => state.config.chain.symbol;

export const getSymbolEth = state => state.config.chain.symbolEth;

export const getBuyerProxyPort = state => state.config.chain.proxyPort;

export const getSellerProxyPort = state => state.config.chain.proxyPort;

export const getIsAuthBypassed = state => state.config.chain.bypassAuth;

export const getIp = state => state.config.ip;

export const getPortCheckErrorLink = state =>
  state.config.chain.portCheckErrorLink;

export const getBuyerPool = state => state.config.buyerDefaultPool;

export const getRecaptchaSiteKey = state => state.config.recaptchaSiteKey;

export const getAutoAdjustPriceInterval = state =>
  state.config.autoAdjustPriceInterval;

export const getAutoAdjustContractPriceTimeout = state =>
  state.config.autoAdjustContractPriceTimeout;

export const getFaucetUrl = state => state.config.chain.faucetUrl;
export const showFaucet = state => state.config.chain.showFaucet;

export const getSellerSelectedCurrency = state => state.config.sellerCurrency;

export const getSellerDefaultCurrency = state =>
  state.config.chain.defaultSellerCurrency;

export const getSellerWhitelistForm = state =>
  state.config.chain.sellerWhitelistUrl;
