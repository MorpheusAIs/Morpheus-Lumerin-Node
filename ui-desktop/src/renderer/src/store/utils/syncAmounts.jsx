import { sanitizeInput, sanitize } from './sanitizers';
import BigNumber from 'bignumber.js';
import * as web3Utils from 'web3-utils';

const ERROR_VALUE_PLACEHOLDER = 'Invalid amount';
export const PENDING_RATE_PLACEHOLDER = '--';

/**
 * The exchange-rate feed polls on an interval and reports nothing until its
 * first successful response, so `rate` is null for the first seconds after
 * launch and stays null whenever every rate service is unreachable. That is a
 * missing rate, not a malformed amount: treating the two alike made a cold
 * start render the wallet balance as "Invalid amount".
 */
function usableRate(rate) {
  const numeric = typeof rate === 'string' ? Number(rate) : rate;
  return typeof numeric === 'number' && Number.isFinite(numeric) && numeric > 0
    ? numeric
    : null;
}

const usdFormatter = new Intl.NumberFormat(navigator.language, {
  style: 'currency',
  currency: 'USD',
  currencyDisplay: 'narrowSymbol'
});

const getSmallUsdValuePlaceholder = () => `< ${usdFormatter.format(0.01)}`;

function getWeiUSDvalue(client, amount, rate) {
  const amountBN = client.toBN(amount);
  const rateBN = client.toBN(
    client.toWei(typeof rate === 'string' ? rate : rate.toString())
  );
  return amountBN.mul(rateBN).div(client.toBN(client.toWei('1')));
}

export const calculateEthFee = (gasLimit, gasPrice, client = web3Utils) => {
  const gasLimitBN = client.toBN(gasLimit);
  const gasPriceBN = client.toBN(gasPrice);
  return Number(client.fromWei(gasLimitBN.mul(gasPriceBN))).toFixed(6);
};

export function toUSD(amount, rate, client = web3Utils) {
  if (+amount === 0) {
    return 0;
  }
  if (usableRate(rate) === null) {
    return PENDING_RATE_PLACEHOLDER;
  }
  if (typeof amount === 'number') {
    amount = amount.toString();
  }

  let isValidAmount;
  let weiUSDvalue;
  try {
    weiUSDvalue = getWeiUSDvalue(client, client.toWei(sanitize(amount)), rate);
    isValidAmount = weiUSDvalue.gte(client.toBN('0'));
  } catch (e) {
    isValidAmount = false;
  }

  const expectedUSDamount = isValidAmount
    ? weiUSDvalue.isZero()
      ? usdFormatter.format(0)
      : weiUSDvalue.lt(client.toBN(client.toWei('0.01')))
      ? getSmallUsdValuePlaceholder()
      : usdFormatter.format(
          new BigNumber(client.fromWei(weiUSDvalue.toString()))
            .dp(2)
            .toString(10)
        )
    : ERROR_VALUE_PLACEHOLDER;

  return expectedUSDamount;
}

export function toCoin(amount, rate, client = web3Utils) {
  if (+amount === 0) {
    return 0;
  }
  if (usableRate(rate) === null) {
    return PENDING_RATE_PLACEHOLDER;
  }
  if (typeof amount === 'number') {
    amount = amount.toString();
  }

  let isValidAmount;
  let coinAmount;
  try {
    const weiAmount = new BigNumber(client.toWei(sanitize(amount)));
    // Converted inside the guard: the rate conversion used to sit in the
    // render expression below, where a bad rate threw instead of degrading.
    const weiRate = new BigNumber(client.toWei(String(usableRate(rate))));
    isValidAmount =
      weiAmount.gte(new BigNumber(0)) && weiRate.gt(new BigNumber(0));
    if (isValidAmount) {
      coinAmount = weiAmount.dividedBy(weiRate).decimalPlaces(4).toString(10);
    }
  } catch (e) {
    isValidAmount = false;
  }

  return isValidAmount ? coinAmount : ERROR_VALUE_PLACEHOLDER;
}

/**
 * Returns an updated state with coin and USD values are synced
 * Useful for updating a pair of coin - USD inputs
 *
 * @param {Object} params - Params required for the conversion
 * @param {string} params.state - The initial component state
 * @param {string} params.coinPrice - The coin/USD rate
 * @param {string} params.id - The id of the field being updated
 * @param {string} params.value - The new value of the field being updated
 * @param {string} params.client - The client object
 */
export function syncAmounts({ state, coinPrice, id, value, client }) {
  const sanitizedValue = sanitizeInput(value);
  return {
    ...state,
    usdAmount:
      id === 'coinAmount'
        ? toUSD(sanitizedValue, coinPrice, client)
        : state.usdAmount,
    coinAmount:
      id === 'usdAmount'
        ? toCoin(sanitizedValue, coinPrice, client)
        : state.coinAmount
  };
}
