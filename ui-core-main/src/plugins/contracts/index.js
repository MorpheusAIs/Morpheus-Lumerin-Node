'use strict';

const debug = require('debug')('lmr-wallet:core:contracts');
const { Lumerin, CloneFactory } = require('contracts-js');
const Web3 = require('web3');

const { getActiveContracts, createContract, cancelContract, purchaseContract } = require('./api');

/**
 * Create a plugin instance.
 *
 * @returns {({ start: Function, stop: () => void})} The plugin instance.
 */
function createPlugin () {
  /**
   * Start the plugin instance.
   *
   * @param {object} options Start options.
   * @returns {{ events: string[] }} The instance details.
   */
  function start ({ config, eventBus, plugins }) {
    const { lmrTokenAddress, cloneFactoryAddress } = config;
    const { eth } = plugins;

    const web3 = new Web3(eth.web3Provider);
    const lumerin = Lumerin(web3, lmrTokenAddress);
    const cloneFactory = CloneFactory(web3, cloneFactoryAddress);

    const refreshContracts = (web3, lumerin, cloneFactory) => () => {
      eventBus.emit('contracts-scan-started', {});

      return getActiveContracts(web3, lumerin, cloneFactory)
        .then((contracts) => {
          console.log('----------------------------------------   ', { contracts })
          eventBus.emit('contracts-scan-finished', {
            actives: contracts
          });
        })
        .catch(function (error) {
          console.log('Could not sync contracts/events', error.stack);
          return {};
        });
    }

    return {
      api: {
        refreshContracts: refreshContracts(web3, lumerin, cloneFactory),
        createContract: createContract(web3, cloneFactory, plugins),
        cancelContract: cancelContract(web3),
        purchaseContract: purchaseContract(web3, cloneFactory, lumerin)
      },
      events: [
        'contracts-scan-started',
        'contracts-scan-finished',
      ],
      name: 'contracts'
    };
  }

  /**
   * Stop the plugin instance.
   */
  function stop () {
    debug('Plugin stopping');
  }

  return {
    start,
    stop
  };
}

module.exports = createPlugin;
