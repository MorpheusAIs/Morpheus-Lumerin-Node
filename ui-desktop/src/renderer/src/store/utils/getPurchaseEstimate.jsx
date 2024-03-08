import { sanitize } from './sanitizers';
import BigNumber from 'bignumber.js';

const format = {
  decimalSeparator: '.',
  groupSeparator: ',',
  groupSize: 3
};

BigNumber.config({ FORMAT: format });

/**
 * Given a coin amount, the current auction price and remaining token, returns
 * the expected LMR amount, the coin spent, if the purchase depletes the auction
 * and a possible coin return for that purchase.
 *
 * @param {Object} params - Params required for the estimate
 * @param {string} params.client - The client object
 * @param {string} params.amount - The user-provided amount (in ETH)
 * @param {string} params.rate - The current auction price
 * @param {string} params.remaining - The remaining token in current auction
 */
export function getPurchaseEstimate({ client, amount, rate, remaining }) {
  let isValidAmount;
  let weiAmount;
  try {
    weiAmount = new BigNumber(client.toWei(sanitize(amount)));
    isValidAmount = weiAmount.gte(new BigNumber(0));
  } catch (e) {
    isValidAmount = false;
  }

  const expectedLMRamount = isValidAmount
    ? client.toWei(
        weiAmount
          .dividedBy(new BigNumber(rate))
          .decimalPlaces(18)
          .toString(10)
      )
    : null;

  const excedes = isValidAmount
    ? client.toBN(expectedLMRamount).gt(client.toBN(remaining))
    : null;

  const usedCoinAmount =
    isValidAmount && excedes
      ? new BigNumber(remaining)
          .multipliedBy(new BigNumber(rate))
          .dividedBy(new BigNumber(client.toWei('1')))
          .decimalPlaces(0)
          .toString(10)
      : null;

  const excessCoinAmount =
    isValidAmount && excedes
      ? weiAmount
          .minus(usedCoinAmount)
          .decimalPlaces(0)
          .toString(10)
      : null;

  return { expectedLMRamount, excedes, usedCoinAmount, excessCoinAmount };
}
