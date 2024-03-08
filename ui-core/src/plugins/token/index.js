'use strict'

const debug = require('debug')('lmr-wallet:core:token');
const Web3 = require('web3');
const { Lumerin } = require('contracts-js');

const events = require('./events');
const { registerToken, getTokenBalance, getTokenGasLimit } = require('./api');

function createPlugin () {
  let walletAddress;

  function start ({ config, eventBus, plugins }) {
    debug.enabled = config.debug;
    const { lmrTokenAddress } = config;


    const web3 = new Web3(plugins.eth.web3Provider);
    const lumerin = Lumerin(web3, lmrTokenAddress)

    function emitLmrBalance (walletAddress) {
      getTokenBalance(lumerin, walletAddress)
        .then(function (balance) {
          eventBus.emit('token-balance-changed', {
            lmrBalance: balance,
          });

          eventBus.emit('token-contract-received', {
            contract: Lumerin.address
          });
        })
        .catch(function (err) {
          eventBus.emit('wallet-error', {
            inner: err,
            message: `Could not get LMR token balance`,
            meta: { plugin: 'token' }
          });
        });
    }

    eventBus.on('open-wallet', function ({ address }) {
      walletAddress = address;

      emitLmrBalance(address);
    });

    eventBus.on('lmr-tx', function () {
      if (walletAddress) {
        emitLmrBalance(walletAddress);
      }
    });

    return {
      api: {
        getTokenGasLimit: getTokenGasLimit(lumerin),
        registerToken: registerToken(plugins),
        metaParsers: {
          approval: events.approvalMetaParser,
          transfer: events.transferMetaParser
        }
      },
      events: [
        'token-contract-received',
        'open-wallet',
        'lmr-tx',
        'token-state-changed',
        'token-balance-changed',
        'wallet-error'
      ],
      name: 'token'
    };
  }

  function stop () {
    walletAddress = null;
  }

  return { start, stop };
}

module.exports = createPlugin;
