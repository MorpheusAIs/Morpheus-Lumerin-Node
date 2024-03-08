'use strict';

const { isNumber } = require('lodash');

function getTransactionStatus (transaction, receipt) {
  if (!receipt) {
    throw new Error('No transaction receipt');
  }

  const isMined = isNumber(receipt.blockNumber);

  const failed = receipt.status === false || (
    receipt.status === null && // no Byzantinum fork
      transaction.input !== '0x' && // is contract call
      transaction.gas === receipt.gasUsed && // used all gas
      !receipt.logs.length // and no logs
  );

  return !isMined || (isMined && !failed);
}

module.exports = getTransactionStatus;
