import { describe, expect, it } from 'vitest';
import {
  DEFAULT_SESSION_DURATION_SECONDS,
  MAX_SESSION_DURATION_SECONDS,
  MIN_SESSION_DURATION_SECONDS,
  SESSION_DURATION_OPTIONS,
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
      toSessionRequestDuration(900, true, { supply: 100, budget: 10 }),
    ).toBe(9001);
  });

  it('offers nothing shorter than a session a provider will accept', () => {
    // Five minutes fell under the provider minimum spend, so picking it only
    // ever produced a refused session request.
    const shortest = Math.min(
      ...SESSION_DURATION_OPTIONS.map((option) => option.seconds),
    );
    expect(shortest).toBe(MIN_SESSION_DURATION_SECONDS);
    expect(
      SESSION_DURATION_OPTIONS.some((option) => option.seconds === 300),
    ).toBe(false);
  });

  it('clamps durations to the supported session range', () => {
    expect(clampSessionDuration(1)).toBe(MIN_SESSION_DURATION_SECONDS);
    // A five-minute length saved by an older build is raised, not honoured.
    expect(clampSessionDuration(300)).toBe(MIN_SESSION_DURATION_SECONDS);
    expect(clampSessionDuration(Number.NaN)).toBe(
      DEFAULT_SESSION_DURATION_SECONDS,
    );
    expect(clampSessionDuration(999_999)).toBe(MAX_SESSION_DURATION_SECONDS);
  });

  it('estimates the amount for the selected duration and payment method', () => {
    const meta = { supply: 100, budget: 10 };
    expect(estimateSessionTokenAmount(2, 900, false, meta)).toBe(18000);
    expect(estimateSessionTokenAmount(2, 900, true, meta)).toBe(18002);
  });

  it('rejects direct-pay conversion until pricing is available', () => {
    expect(() =>
      toSessionRequestDuration(900, true, { supply: 0, budget: 0 }),
    ).toThrow('Pricing data is not ready');
  });
});
