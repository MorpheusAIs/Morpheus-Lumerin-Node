import { describe, expect, it } from 'vitest';
import { toBaseUnits, isValidAddress } from './amount';

// These assertions guard real money. The proxy-router parses `amount` straight
// into a big.Int, so any precision loss here becomes a wrong transfer amount
// on-chain — silently, with no validation to catch it server-side.
describe('toBaseUnits', () => {
  it('converts whole numbers to 18-decimal wei', () => {
    expect(toBaseUnits('1')).toBe('1000000000000000000');
    expect(toBaseUnits('42')).toBe('42000000000000000000');
  });

  it('converts fractional amounts', () => {
    expect(toBaseUnits('0.5')).toBe('500000000000000000');
    expect(toBaseUnits('1.5')).toBe('1500000000000000000');
    expect(toBaseUnits('0.1')).toBe('100000000000000000');
  });

  it('represents the smallest possible unit exactly', () => {
    expect(toBaseUnits('0.000000000000000001')).toBe('1');
  });

  it('keeps full precision on values that would lose digits as a float', () => {
    // 123456789.123456789 * 1e18 is not exactly representable in float64;
    // a Number-based implementation returns 123456789123456790000000000 here.
    expect(toBaseUnits('123456789.123456789')).toBe(
      '123456789123456789000000000',
    );
  });

  it('handles large amounts without exponent notation', () => {
    const result = toBaseUnits('1000000');
    expect(result).toBe('1000000000000000000000000');
    expect(result).not.toMatch(/e/i);
  });

  it('normalises leading zeros and bare decimals', () => {
    expect(toBaseUnits('00.5')).toBe('500000000000000000');
    expect(toBaseUnits('.5')).toBe('500000000000000000');
  });

  it('always returns a base-10 integer string', () => {
    for (const input of ['1', '0.5', '0.000000000000000001', '999999']) {
      expect(toBaseUnits(input)).toMatch(/^\d+$/);
    }
  });

  it('honours a custom decimal count', () => {
    expect(toBaseUnits('1', 6)).toBe('1000000');
    expect(toBaseUnits('1.5', 6)).toBe('1500000');
  });

  describe('rejects input the backend would reject or misread', () => {
    it.each([
      ['zero', '0'],
      ['zero with decimals', '0.0'],
      ['empty', ''],
      ['whitespace', '   '],
      ['non-numeric', 'abc'],
      ['negative', '-1'],
      ['exponent notation', '1e18'],
      ['comma separator', '1,000'],
      ['bare dot', '.'],
      ['null', null],
      ['undefined', undefined],
    ])('%s', (_label, input) => {
      expect(() => toBaseUnits(input)).toThrow();
    });

    it('more decimals than the token supports', () => {
      expect(() => toBaseUnits('1.0000000000000000001')).toThrow(
        /18 decimal places/,
      );
    });
  });
});

describe('isValidAddress', () => {
  it('accepts a checksummed address', () => {
    expect(isValidAddress('0x15dd2028C976beaA6668E286b496A518F457b5Cf')).toBe(
      true,
    );
  });

  it('accepts an all-lowercase address', () => {
    expect(isValidAddress('0x15dd2028c976beaa6668e286b496a518f457b5cf')).toBe(
      true,
    );
  });

  it('trims surrounding whitespace from pasted input', () => {
    expect(
      isValidAddress('  0x15dd2028C976beaA6668E286b496A518F457b5Cf  '),
    ).toBe(true);
  });

  it.each([
    ['missing 0x prefix', '15dd2028C976beaA6668E286b496A518F457b5Cf'],
    ['too short', '0x15dd2028C976beaA6668E286b496A518F457b5C'],
    ['too long', '0x15dd2028C976beaA6668E286b496A518F457b5Cff'],
    ['non-hex characters', '0xZZdd2028C976beaA6668E286b496A518F457b5Cf'],
    ['empty', ''],
    ['null', null],
    ['undefined', undefined],
    ['a transaction hash, not an address', '0x' + 'a'.repeat(64)],
  ])('rejects %s', (_label, input) => {
    expect(isValidAddress(input)).toBe(false);
  });
});
