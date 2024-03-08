import BigNumber from 'bignumber.js';
import PropTypes from 'prop-types';
import { fromWei, toBN, toWei, toHex } from 'web3-utils';

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

export function isWeiable(amount, unit = 'ether') {
  let isValid;
  try {
    toWei(amount.replace(',', '.'), unit);
    isValid = true;
  } catch (e) {
    isValid = false;
  }
  return isValid;
}

export function isHexable(amount) {
  try {
    toHex(amount);
    return true;
  } catch (e) {
    return false;
  }
}

export function isGreaterThanZero(amount) {
  const weiAmount = new BigNumber(toWei(amount.replace(',', '.')));
  return weiAmount.gt(new BigNumber(0));
}

export function toLMR(amount, rate, errorValue = 'Invalid amount', remaining) {
  let isValidAmount;
  let weiAmount;
  try {
    weiAmount = new BigNumber(toWei(amount.replace(',', '.')));
    isValidAmount = weiAmount.gte(new BigNumber(0));
  } catch (e) {
    isValidAmount = false;
  }

  const expectedLMRamount = isValidAmount
    ? toWei(
        weiAmount
          .dividedBy(new BigNumber(rate))
          .decimalPlaces(18)
          .toString(10)
      )
    : errorValue;

  const excedes = isValidAmount
    ? toBN(expectedLMRamount).gte(toBN(remaining))
    : null;

  const usedCoinAmount =
    isValidAmount && excedes
      ? new BigNumber(remaining)
          .multipliedBy(new BigNumber(rate))
          .dividedBy(new BigNumber(toWei('1')))
          .integerValue()
          .toString(10)
      : null;

  const excessCoinAmount =
    isValidAmount && excedes
      ? weiAmount
          .minus(usedCoinAmount)
          .integerValue()
          .toString(10)
      : null;

  return { expectedLMRamount, excedes, usedCoinAmount, excessCoinAmount };
}

export function weiToGwei(amount) {
  return fromWei(amount, 'gwei');
}

export function gweiToWei(amount) {
  return toWei(amount, 'gwei');
}

export function smartRound(weiAmount) {
  const n = Number.parseFloat(fromWei(weiAmount), 10);
  let decimals = -Math.log10(n) + 10;
  if (decimals < 2) {
    decimals = 2;
  } else if (decimals >= 18) {
    decimals = 18;
  }
  // round extra decimals and remove trailing zeroes
  return new BigNumber(n.toFixed(Math.ceil(decimals))).toString(10);
}

/**
 * Removes extra spaces and converts to lowercase
 * Useful for sanitizing user input before recovering a wallet.
 *
 * @param {string} str The string to sanitize
 */
export function sanitizeMnemonic(str) {
  return str
    .replace(/\s+/gi, ' ')
    .trim()
    .toLowerCase();
}

export function getConversionRate(lmrAmount, coinAmount) {
  const compareAgainst = fromWei(lmrAmount);
  return new BigNumber(coinAmount)
    .dividedBy(new BigNumber(compareAgainst))
    .integerValue()
    .toString(10);
}
export function abbreviateAddress(addr, length = 6) {
  return `${addr.slice(0, length)}...${addr.slice(
    addr.length - length,
    addr.length
  )}`;
}

export const toRfc2396 = (pool, username) => {
  const addressParts = pool.replace('stratum+tcp://', '').split(':');
  const address = addressParts[0];
  const port = addressParts[1];
  const password = '';
  // This worker name and password won't be forwarded from the seller to the buyer.
  // Set url with {username}:{password} to preserve backward compatibility
  // This also should maintain consistency of data between UI/Blockchain/Proxy Router
  return `stratum+tcp://${username}:${password}@${address}:${port}`;
};

export const generatePoolUrl = (account, poolAddress) => {
  const password = '';
  const username = encodeURIComponent(account);
  return `stratum+tcp://${username}:${password}@${poolAddress}`;
};
