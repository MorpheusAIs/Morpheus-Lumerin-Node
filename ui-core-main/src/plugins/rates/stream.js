'use strict';

const EventEmitter = require('events');

/**
 * Create a "classic" stream that periodically emits the result of a fn.
 *
 * @param {() => any} fn The function to poll.
 * @param {number} minInterval The polling period in ms.
 * @returns {object} The stream instance.
 */
function createStream (fn, minInterval) {
  const stream = new EventEmitter();

  let stop = false; // Could have been called "flag" but...

  const emitTickerValue = () =>
    Promise.resolve()
      .then(fn)
      .then(function (data) {
        stream.emit('data', data);
      })
      .catch(function (err) {
        stream.emit('error', err);
      })
      .then(function () {
        if (!stop) {
          setTimeout(emitTickerValue, minInterval);
        }
      });

  emitTickerValue();

  stream.destroy = function () {
    stream.removeAllListeners();
    stop = true;
  };

  return stream;
}

module.exports = createStream;
