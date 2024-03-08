'use strict';

const debug = require('debug')('lmr-wallet:core:wallet');
const { add } = require('lodash');
const Web3 = require('web3');

const api = require('./api');
const hdkey = require('./hdkey');

function createPlugin () {
  let walletAddress;

  function start ({ config, eventBus, plugins }) {
    debug.enabled = config.debug;

    const web3 = new Web3(plugins.eth.web3Provider);

    function emitEthBalance (address) {
      web3.eth.getBalance(address)
        .then(function (balance) {
          eventBus.emit('eth-balance-changed', {
            ethBalance: balance
          });
        })
        .catch(function (err) {
          eventBus.emit('wallet-error', {
            inner: err,
            message: `Could not get ${config.symbol} balance`,
            meta: { plugin: 'wallet' }
          })
        });
    }

    eventBus.on('open-wallet', function ({ address }) {
      walletAddress = address;
      emitEthBalance(walletAddress);
    });

    eventBus.on('eth-tx', function () {
      if(walletAddress) {
        emitEthBalance(walletAddress);
      }
    });

    return {
      api: {
        createAddress: hdkey.getAddress,
        createPrivateKey: hdkey.getPrivateKey,
        getAddressAndPrivateKey: hdkey.getAddressAndPrivateKey,
        getGasLimit: api.estimateGas(web3),
        getGasPrice: api.getGasPrice(web3),
        sendEth: api.sendSignedTransaction(web3, plugins.explorer.logTransaction),
        enseureAccount: api.ensureAccount(web3)
      },
      events: [
        'open-wallet',
        'eth-tx',
        'eth-balance-changed',
        'wallet-state-changed',
        'wallet-error'
      ],
      name: 'wallet'
    };
  }

  function stop () {
    walletAddress = '';
  }

  return { start, stop };
}

module.exports = createPlugin;
