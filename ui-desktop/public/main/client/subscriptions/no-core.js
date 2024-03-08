'use strict';

const handlers = require('../handlers');
const utils = require('./utils');

const listeners = {
  'validate-password': handlers.validatePassword,
  'change-password': handlers.changePassword,
  'persist-state': handlers.persistState,
  'clear-cache': handlers.clearCache,
  'handle-client-error': handlers.handleClientSideError,
  'get-pool-address': handlers.getPoolAddress,
  'reveal-secret-phrase': handlers.revealSecretPhrase,
  'restart-wallet': handlers.restartWallet,
  'has-stored-secret-phrase': handlers.hasStoredSecretPhrase,
  "logout": handlers.logout,
  "save-proxy-router-settings": handlers.saveProxyRouterSettings,
  "get-proxy-router-settings": handlers.getProxyRouterSettings,
  "get-default-currency-settings": handlers.getDefaultCurrency,
  "set-default-currency-settings": handlers.setDefaultCurrency,
  "get-custom-env-values": handlers.getCustomEnvs,
  "set-custom-env-values": handlers.setCustomEnvs,
  "get-profit-settings": handlers.getProfitSettings,
  "set-profit-settings": handlers.setProfitSettings,
  "get-contract-hashrate": handlers.getContractHashrate,
  "get-auto-adjust-price": handlers.getAutoAdjustPriceData,
  "set-auto-adjust-price": handlers.setAutoAdjustPriceData,
};

// Subscribe to messages where no core has to react
const subscribeWithoutCore = () =>
  utils.subscribeTo(listeners, 'none');

const unsubscribeWithoutCore = () =>
  utils.unsubscribeTo(listeners);

module.exports = { subscribeWithoutCore, unsubscribeWithoutCore };
