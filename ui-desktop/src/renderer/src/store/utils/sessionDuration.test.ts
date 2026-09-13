import { describe, expect, it } from 'vitest';
import {
  DEFAULT_SESSION_DURATION_SECONDS,
  MAX_SESSION_DURATION_SECONDS,
  MIN_SESSION_DURATION_SECONDS,
  SESSION_DURATION_OPTIONS,
  clampSessionDuration,
  estimateSessionTokenAmount,
  sessionComputeCost,
  sessionDurationOptions,
} from './sessionDuration';

describe('session duration', () => {
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

  it('honours a ceiling lower than the hardcoded one', () => {
    // getSessionEnd clamps to the contract's own maximum, so a length above it
    // is paid for and not received.
    expect(clampSessionDuration(24 * 60 * 60, 3600)).toBe(3600);
    expect(sessionDurationOptions(3600).map((option) => option.seconds)).toEqual(
      [900, 1800, 3600],
    );
    expect(sessionDurationOptions()).toEqual(SESSION_DURATION_OPTIONS);
  });

  it('never leaves the picker empty when the ceiling is misreported', () => {
    expect(sessionDurationOptions(1)).toEqual([SESSION_DURATION_OPTIONS[0]]);
    expect(sessionDurationOptions(Number.NaN)).toEqual(
      SESSION_DURATION_OPTIONS,
    );
  });

  it('charges the same amount whichever way the provider is paid', () => {
    // The payment method decides who pays the provider at close, not what the
    // session costs to open. Direct pay used to be quoted at price x duration,
    // which buys a few hundredth of the length asked for.
    const meta = { supply: 100, budget: 10 };
    // 2 wei/s x (900 + 1 headroom) x 100 / 10.
    expect(estimateSessionTokenAmount(2, 900, meta)).toBe(18020);
  });

  it('rounds the amount up so the session is never short', () => {
    expect(estimateSessionTokenAmount(1, 900, { supply: 10, budget: 3 })).toBe(
      Math.ceil((901 * 10) / 3),
    );
  });

  it('reports the compute cost separately from the amount locked up', () => {
    const meta = { supply: 100, budget: 10 };
    expect(sessionComputeCost(2, 900)).toBe(1800);
    expect(estimateSessionTokenAmount(2, 900, meta)).toBeGreaterThan(
      sessionComputeCost(2, 900),
    );
  });

  it('refuses to quote until pricing data has loaded', () => {
    expect(() =>
      estimateSessionTokenAmount(2, 900, { supply: 0, budget: 0 }),
    ).toThrow('Pricing data is not ready');
  });
});
