'use strict';

const logger = require('../../logger');
const EventEmitter = require('events');

function createStream (web3, updateInterval = 10000) {
  const ee = new EventEmitter();

  web3.eth.getBlock('latest')
    .then(function (block) {
      ee.emit('data', block);
    })
    .catch(function (err) {
      ee.emit('error', err);
    })

  const interval = setInterval(async () => {
    try {
      const block = await web3.eth.getBlock('latest');
      ee.emit('data', block);
    } catch (err) {
      ee.emit('error', err);
    }
  }, updateInterval);

  return { interval, stream: ee };
}

module.exports = createStream;
