import { describe, expect, it } from 'vitest';
import { desktopQueryRetryDelay, shouldRetryDesktopQuery } from './queryClient';

describe('desktop query startup recovery', () => {
  it('retries the short proxy-router readiness race three times', () => {
    const error = new Error('Cannot reach the local proxy-router.');

    expect(shouldRetryDesktopQuery(0, error)).toBe(true);
    expect(shouldRetryDesktopQuery(1, error)).toBe(true);
    expect(shouldRetryDesktopQuery(2, error)).toBe(true);
    expect(shouldRetryDesktopQuery(3, error)).toBe(false);
  });

  it('does not repeatedly retry authentication or ordinary API failures', () => {
    expect(
      shouldRetryDesktopQuery(0, {
        message: 'Cannot authenticate to the proxy-router.',
      }),
    ).toBe(true);
    expect(
      shouldRetryDesktopQuery(1, {
        message: 'Cannot authenticate to the proxy-router.',
      }),
    ).toBe(false);
    expect(shouldRetryDesktopQuery(1, new Error('HTTP 500'))).toBe(false);
  });

  it('backs off quickly without exceeding two seconds', () => {
    expect([0, 1, 2, 3, 8].map(desktopQueryRetryDelay)).toEqual([
      250, 500, 1_000, 2_000, 2_000,
    ]);
  });
});
