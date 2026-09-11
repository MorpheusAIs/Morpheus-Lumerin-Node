import { describe, expect, it } from 'vitest';

import { toCoin, toUSD, PENDING_RATE_PLACEHOLDER } from './syncAmounts';
import { getAmountFieldsProps } from './getAmountFieldsProps';

const ERROR_VALUE_PLACEHOLDER = 'Invalid amount';

describe('Wallet amount conversion before the exchange rate arrives', () => {
  // The rates plugin polls on an interval and reports nothing until its first
  // successful response, so every cold start renders balances with a null rate.
  it.each([
    ['null', null],
    ['undefined', undefined],
    ['zero', 0],
    ['a negative rate', -12],
    ['a non-numeric string', 'unavailable'],
    ['NaN', Number.NaN],
  ])(
    'reports a pending rate rather than an invalid amount for %s',
    (_label, rate) => {
      expect(toUSD('1.5', rate)).toBe(PENDING_RATE_PLACEHOLDER);
      expect(toUSD('1.5', rate)).not.toBe(ERROR_VALUE_PLACEHOLDER);
    },
  );

  it('never throws while the rate is missing', () => {
    expect(() => toCoin('1.5', null)).not.toThrow();
    expect(() => toCoin('1.5', undefined)).not.toThrow();
    // The rate conversion used to sit outside toCoin's try block, so a zero
    // rate divided by zero instead of degrading.
    expect(() => toCoin('1.5', 0)).not.toThrow();
    expect(toCoin('1.5', 0)).toBe(PENDING_RATE_PLACEHOLDER);
  });

  it('still converts once a real rate is present', () => {
    expect(toUSD('2', 1500)).toBe('$3,000.00');
    expect(toCoin('3000', 1500)).toBe('2');
  });

  it('still reports a genuinely malformed amount as invalid', () => {
    expect(toUSD('not-a-number', 1500)).toBe(ERROR_VALUE_PLACEHOLDER);
    expect(toCoin('not-a-number', 1500)).toBe(ERROR_VALUE_PLACEHOLDER);
  });

  it('keeps a zero balance readable without a rate', () => {
    expect(toUSD('0', null)).toBe(0);
    expect(toCoin('0', null)).toBe(0);
  });

  it('leaves a pending value out of the transaction inputs', () => {
    const props = getAmountFieldsProps({
      lmrAmount: PENDING_RATE_PLACEHOLDER,
      coinAmount: PENDING_RATE_PLACEHOLDER,
      usdAmount: PENDING_RATE_PLACEHOLDER,
    });
    expect(props.coinAmount).toBe('');
    expect(props.lmrAmount).toBe('');
    expect(props.usdAmount).toBe('0');
  });
});
