'use strict';

const debug = require('debug')('lmr-wallet:core:contracts-stream');

/**
 * Create a "classic" stream that connects to Lumerin Contracts.
 *
 * @param {Web3} web3 The function to poll.
 * @returns {object} The stream instance.
 */
function createStream (web3) {
  const subscription = web3.eth.subscribe('newBlockHeaders');

  web3.eth.getBlock('latest')
    .then(function (block) {
      subscription.emit('data', block);
    })
    .catch(function (err) {
      subscription.emit('error', err);
    })
  // const emitTickerValue = () =>
  //   Promise.resolve()
  //     .then(fn)
  //     .then(function (data) {
  //       stream.emit('data', data);
  //     })
  //     .catch(function (err) {
  //       stream.emit('error', err);
  //     })
  //     .then(function () {
  //       if (!stop) {
  //         setTimeout(emitTickerValue, minInterval);
  //       }
  //     });

  // emitTickerValue();

  subscription.unsubscribe(function(error, success) {
    success || debug('Could not successfully unsubscribe from web3 block-stream');
  });


  return subscription;
}

module.exports = createStream;
