/**
 * Returns appropriate values and placeholders for coin/LMR - USD fields
 *
 * @param {Object} amounts - Input amounts to calculate props from
 * @param {string} amounts.lmrAmount - LMR amount as typed by user
 * @param {string} amounts.coinAmount - coin amount as typed by user
 * @param {string} amounts.usdAmount - USD amount as typed by user
 */
export function getAmountFieldsProps({ lmrAmount, coinAmount, usdAmount }) {
  const ERROR_VALUE_PLACEHOLDER = 'Invalid amount';
  const SMALL_VALUE_PLACEHOLDER = '< 0.01';

  return {
    lmrPlaceholder:
      lmrAmount === ERROR_VALUE_PLACEHOLDER ? ERROR_VALUE_PLACEHOLDER : '0.00',
    coinPlaceholder:
      coinAmount === ERROR_VALUE_PLACEHOLDER ? ERROR_VALUE_PLACEHOLDER : '0.00',
    usdPlaceholder:
      usdAmount === ERROR_VALUE_PLACEHOLDER
        ? ERROR_VALUE_PLACEHOLDER
        : usdAmount === SMALL_VALUE_PLACEHOLDER
        ? SMALL_VALUE_PLACEHOLDER
        : '0.00',
    coinAmount: coinAmount === ERROR_VALUE_PLACEHOLDER ? '' : coinAmount,
    lmrAmount: lmrAmount === ERROR_VALUE_PLACEHOLDER ? '' : lmrAmount,
    usdAmount: usdAmount === ERROR_VALUE_PLACEHOLDER ? '0' : usdAmount
  };
}
