// Amount parsing for on-chain transfers.
//
// Kept in its own module (rather than inside withTransactionModalState) so the
// money-critical conversion can be unit-tested without dragging in React,
// react-redux and the IPC client.

export const DEFAULT_DECIMALS = 18;

/**
 * Converts a human-entered decimal amount into a base-10 wei string.
 *
 * Uses string maths rather than Number arithmetic on purpose: the proxy-router
 * parses this value straight into a big.Int, and float64 cannot represent most
 * 18-decimal values exactly. `123456789.123456789 * 1e18`, for example, comes
 * out as ...790000000 — a silently wrong transfer amount with nothing
 * server-side to catch it.
 *
 * @throws if the input is empty, non-numeric, zero, negative, in exponent
 *         notation, or has more precision than the token supports.
 */
export const toBaseUnits = (
  amount: unknown,
  decimals: number = DEFAULT_DECIMALS,
): string => {
  const raw = String(amount ?? '').trim();
  if (!raw) {
    throw new Error('Enter an amount');
  }
  if (!/^\d*\.?\d*$/.test(raw) || raw === '.') {
    throw new Error('Amount is not a valid number');
  }

  const [whole = '', fraction = ''] = raw.split('.');
  if (fraction.length > decimals) {
    throw new Error(`At most ${decimals} decimal places are supported`);
  }

  const padded = (whole + fraction.padEnd(decimals, '0')).replace(/^0+/, '');
  const value = padded === '' ? '0' : padded;

  if (value === '0') {
    throw new Error('Amount must be greater than zero');
  }
  return value;
};

/** Strict 0x-prefixed 20-byte address check. */
export const isValidAddress = (address: unknown): boolean =>
  /^0x[a-fA-F0-9]{40}$/.test(String(address ?? '').trim());
