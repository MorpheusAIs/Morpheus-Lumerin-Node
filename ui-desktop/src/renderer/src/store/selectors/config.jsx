import { createSelector } from 'reselect';

// Returns the "config" state branch
export const getConfig = state => state.config;

export const getCoinSymbol = state => state.config.chain.symbol;

export const getSymbolEth = state => state.config.chain.symbolEth;

export const getIsAuthBypassed = state => state.config.chain.bypassAuth;

export const getIp = state => state.config.ip;

export const getBuyerPool = state => state.config.buyerDefaultPool;

export const getSellerSelectedCurrency = state => state.config.sellerCurrency;

export const getSellerDefaultCurrency = state =>
  state.config.chain.defaultSellerCurrency;
