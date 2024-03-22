'use strict';

const debug = require('debug')('lmr-wallet:core:block-stream');

function createStream (web3) {
  const subscription = web3.eth.subscribe('newBlockHeaders');

  web3.eth.getBlock('latest')
    .then(function (block) {
      subscription.emit('data', block);
    })
    .catch(function (err) {
      subscription.emit('error', err);
    })

  // subscription.destroy = subscription.unsubscribe;
  subscription.unsubscribe(function(error, success) {
    success || debug('Could not successfully unsubscribe from web3 block-stream');
  });

  return subscription;
}

module.exports = createStream;
