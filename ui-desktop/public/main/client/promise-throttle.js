'use strict';

module.exports = function promiseThrottle (fn) {
  const promise = Promise.resolve();
  return function (...args) {
    return promise.catch().then(fn(...args));
  };
}
