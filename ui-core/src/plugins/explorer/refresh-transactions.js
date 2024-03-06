'use strict';

const { isMatch, isNumber } = require('lodash');
const { utils: { toChecksumAddress } } = require('web3');

const createTryParseEventLog = require('./parse-log');

const refreshTransaction = (web3, eventsRegistry, queue) => (hash, address) =>
  web3.eth.getTransactionReceipt(hash).then(function (receipt) {
    const pending = [];

    // Skip unconfirmed transactions
    if (!receipt || !isNumber(receipt.blockNumber)) {
      return pending;
    }

    // Refresh transaction
    if (
      toChecksumAddress(receipt.from) === address ||
      toChecksumAddress(receipt.to) === address
    ) {
      pending.push(queue.addTransaction(address)(hash));
    }

    // Refresh transaction events
    if (receipt.logs && receipt.logs.length) {
      const tryParseEventLog = createTryParseEventLog(web3, eventsRegistry);

      receipt.logs.forEach(function (log) {
        tryParseEventLog(log, address).forEach(function (parsedLog) {
          const {
            contractAddress,
            filter,
            metaParser,
            parsed: { event, returnValues }
          } = parsedLog;

          if (isMatch(returnValues, filter)) {
            pending.push(
              queue.addEvent(address, metaParser)({
                address: contractAddress,
                event,
                returnValues,
                transactionHash: hash
              })
            );
          }
        })
      });
    }

    return Promise.all(pending);
  })

module.exports = refreshTransaction;
