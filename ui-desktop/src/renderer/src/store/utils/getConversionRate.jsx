import BigNumber from 'bignumber.js';

const format = {
  decimalSeparator: '.',
  groupSeparator: ',',
  groupSize: 3
};

BigNumber.config({ FORMAT: format });

/**
 * Returns the coin/LMR rate of a conversion
 * Useful for displaying the obtained rate after a conversion estimate
 *
 * @param {Object} client - The client object
 * @param {string} lmrAmount - The LMR amount provided or obtained (in wei)
 * @param {string} coinAmount - The coin amount provided or obtained (in wei)
 */
export function getConversionRate(client, lmrAmount, coinAmount) {
  const compareAgainst = client.fromWei(lmrAmount);
  return new BigNumber(coinAmount)
    .dividedBy(new BigNumber(compareAgainst))
    .integerValue()
    .toString(10);
}
