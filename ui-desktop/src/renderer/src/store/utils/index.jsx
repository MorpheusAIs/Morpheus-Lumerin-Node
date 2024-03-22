import BigNumber from 'bignumber.js';
import PropTypes from 'prop-types';

import { sanitize } from './sanitizers';

export { sanitizeMnemonic, sanitizeInput, sanitize } from './sanitizers';
export { createTransactionParser } from './createTransactionParser';
export { getAmountFieldsProps } from './getAmountFieldsProps';
export { getPurchaseEstimate } from './getPurchaseEstimate';
export { getConversionRate } from './getConversionRate';
export { mnemonicWords } from './mnemonicWords';
export { syncAmounts, calculateEthFee } from './syncAmounts';

export function hasFunds(value) {
  return value && new BigNumber(value).gt(new BigNumber(0));
}

export function isWeiable(client, amount, unit = 'ether') {
  let isValid;
  try {
    client.toWei(sanitize(amount), unit);
    isValid = true;
  } catch (e) {
    isValid = false;
  }
  return isValid;
}

export function isHexable(client, amount) {
  let isValid;
  try {
    client.toHex(amount);
    isValid = true;
  } catch (e) {
    isValid = false;
  }
  return isValid;
}

export function isGreaterThanZero(client, amount) {
  const weiAmount = client.toBN(client.toWei(sanitize(amount)));
  return weiAmount.gt(client.toBN(0));
}

// export function isFailed(tx, confirmations) {
//   return confirmations > 0 || tx.contractCallFailed
// }

export function isPending(tx, confirmations) {
  // return !isFailed(tx, confirmations) && confirmations < 6
  return false;
}

export const errorPropTypes = (...fields) => {
  const shape = fields.reduce((acc, fieldName) => {
    acc[fieldName] = PropTypes.oneOfType([
      PropTypes.arrayOf(PropTypes.string),
      PropTypes.string
    ]);
    return acc;
  }, {});
  return PropTypes.shape(shape).isRequired;
};

export const statusPropTypes = PropTypes.oneOf([
  'init',
  'pending',
  'success',
  'failure'
]).isRequired;
