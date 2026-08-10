import { describe, expect, it } from 'vitest';
import {
  DEFAULT_SESSION_DURATION_SECONDS,
  MAX_SESSION_DURATION_SECONDS,
  MIN_SESSION_DURATION_SECONDS,
  clampSessionDuration,
  estimateSessionTokenAmount,
  toSessionRequestDuration,
} from './sessionDuration';

describe('session duration', () => {
  it('passes a staked session duration through unchanged', () => {
    expect(
      toSessionRequestDuration(3600, false, { supply: 100, budget: 10 }),
    ).toBe(3600);
  });

  it('compensates direct pay for the current contract conversion', () => {
    expect(
      toSessionRequestDuration(300, true, { supply: 100, budget: 10 }),
    ).toBe(3001);
  });

  it('clamps durations to the supported session range', () => {
    expect(clampSessionDuration(1)).toBe(MIN_SESSION_DURATION_SECONDS);
    expect(clampSessionDuration(Number.NaN)).toBe(
      DEFAULT_SESSION_DURATION_SECONDS,
    );
    expect(clampSessionDuration(999_999)).toBe(MAX_SESSION_DURATION_SECONDS);
  });

  it('estimates the amount for the selected duration and payment method', () => {
    const meta = { supply: 100, budget: 10 };
    expect(estimateSessionTokenAmount(2, 300, false, meta)).toBe(6000);
    expect(estimateSessionTokenAmount(2, 300, true, meta)).toBe(6002);
  });

  it('rejects direct-pay conversion until pricing is available', () => {
    expect(() =>
      toSessionRequestDuration(300, true, { supply: 0, budget: 0 }),
    ).toThrow('Pricing data is not ready');
  });
});
