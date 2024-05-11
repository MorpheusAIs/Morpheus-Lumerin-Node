'use strict';

const logger = require('../../logger');
const { Lumerin } = require('contracts-js');
const Web3 = require('web3');

const { sendLmr, estimateGasTransfer } = require('./api');

/**
 * Creates an instance of the Lumerin plugin.
 *
 * @returns {{start:Function,stop:Function}} The plugin top-level API.
 */
function createPlugin () {
  /**
   * Start the plugin.
   *
   * @param {object} params The start parameters.
   * @param {object} params.config The configuration options.
   * @param {object} params.eventBus The cross-plugin event emitter.
   * @param {object} params.plugins All other plugins.
   * @returns {{api:object,events:string[],name:string}} The plugin API.
   */
  function start ({ config, eventBus, plugins }) {
    // debug.enabled = config.debug;

    const { lmrTokenAddress } = config;
    const { eth, explorer, token } = plugins;

    const web3 = new Web3(eth.web3Provider);
    const lumerin = Lumerin(web3, lmrTokenAddress)

    // Register LMR token
    token.registerToken(lumerin.address, {
      decimals: 18,
      name: 'Morpheus',
      symbol: 'MOR'
    });

    // eventBus.on('coin-block', emitLumerinStatus);

    // Collect meta parsers
    const metaParsers = Object.assign({},
      // {
      //   // auction: auctionEvents.auctionMetaParser,
      //   // converter: converterEvents.converterMetaParser,
      //   export: porterEvents.exportMetaParser,
      //   import: porterEvents.importMetaParser,
      //   importRequest: porterEvents.importRequestMetaParser
      // },
      token.metaParsers
    );

    // Build and return API
    return {
      api: {
        sendLmr: sendLmr(
          web3,
          lumerin,
          explorer.logTransaction,
          metaParsers
        ),
        estimateGasTransfer: estimateGasTransfer(lumerin),
      },
      events: [
        'wallet-error'
      ],
      name: 'lumerin'
    };
  }

  /**
   * Stop the plugin.
   */
  function stop () {}

  return {
    start,
    stop
  };
}

module.exports = createPlugin;
