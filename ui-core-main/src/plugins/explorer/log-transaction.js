'use strict';

const createLogTransaction = queue =>
  function (promiEvent, from, meta) {
    if (promiEvent.once) {
      return new Promise(function (resolve, reject) {
        promiEvent.once('transactionHash', function (hash) {
          queue.addTransaction(from, meta)(hash);
        })
        promiEvent.once('receipt', function (receipt) {
          queue.addTransaction(from)(receipt.transactionHash);
          resolve({ receipt });
        });
        promiEvent.once('error', function (err) {
          promiEvent.removeAllListeners();
          reject(err);
        });
      });
    }

    // This is not a wrapped PromiEvent object. It shall be a plain promise
    // instead.
    return promiEvent.then(function (receipt) {
      queue.addTransaction(from)(receipt.transactionHash);
      return { receipt };
    });
  }

module.exports = createLogTransaction;
