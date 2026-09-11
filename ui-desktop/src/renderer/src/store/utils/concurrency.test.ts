import { describe, expect, it, vi } from 'vitest';
import {
  pooledMap,
  pooledMapSettled,
  withTimeout,
  TimeoutError,
} from './concurrency';

describe('pooledMap', () => {
  it('preserves input order regardless of completion order', async () => {
    const items = [50, 10, 30, 5, 40];
    const result = await pooledMap(
      items,
      async (ms) => {
        await new Promise((r) => setTimeout(r, ms));
        return ms;
      },
      3,
    );
    expect(result).toEqual(items);
  });

  it('never exceeds the concurrency limit', async () => {
    let inFlight = 0;
    let peak = 0;

    await pooledMap(
      Array.from({ length: 50 }, (_, i) => i),
      async () => {
        inFlight++;
        peak = Math.max(peak, inFlight);
        await new Promise((r) => setTimeout(r, 2));
        inFlight--;
      },
      4,
    );

    expect(peak).toBeLessThanOrEqual(4);
  });

  // Regression guard: this is the property that stopped the Providers/Models
  // tabs from firing hundreds of simultaneous requests. If someone reverts to
  // Promise.all, peak concurrency jumps to items.length and this fails.
  it('does not run everything at once for large inputs', async () => {
    let inFlight = 0;
    let peak = 0;

    await pooledMap(
      Array.from({ length: 200 }, (_, i) => i),
      async () => {
        inFlight++;
        peak = Math.max(peak, inFlight);
        await new Promise((r) => setTimeout(r, 1));
        inFlight--;
      },
      6,
    );

    expect(peak).toBe(6);
    expect(peak).toBeLessThan(200);
  });

  it('handles an empty list without invoking the callback', async () => {
    const fn = vi.fn();
    await expect(pooledMap([], fn, 4)).resolves.toEqual([]);
    expect(fn).not.toHaveBeenCalled();
  });

  it('clamps the limit to the item count', async () => {
    let peak = 0;
    let inFlight = 0;
    await pooledMap(
      [1, 2],
      async () => {
        inFlight++;
        peak = Math.max(peak, inFlight);
        await new Promise((r) => setTimeout(r, 5));
        inFlight--;
      },
      100,
    );
    expect(peak).toBe(2);
  });

  it('propagates a rejection', async () => {
    await expect(
      pooledMap([1, 2, 3], async (n) => {
        if (n === 2) throw new Error('boom');
        return n;
      }),
    ).rejects.toThrow('boom');
  });
});

describe('pooledMapSettled', () => {
  it('substitutes the fallback for a failing item and keeps the rest', async () => {
    const result = await pooledMapSettled(
      [1, 2, 3, 4],
      async (n) => {
        if (n === 2) throw new Error('boom');
        return `ok${n}`;
      },
      (n) => `fallback${n}`,
      2,
    );
    expect(result).toEqual(['ok1', 'fallback2', 'ok3', 'ok4']);
  });

  it('passes the error to the fallback', async () => {
    const fallback = vi.fn(() => 'x');
    await pooledMapSettled(
      [1],
      async () => {
        throw new Error('specific failure');
      },
      fallback,
    );
    expect(fallback).toHaveBeenCalledWith(
      1,
      0,
      expect.objectContaining({ message: 'specific failure' }),
    );
  });
});

describe('withTimeout', () => {
  it('rejects with TimeoutError once the budget elapses', async () => {
    const pending = new Promise((r) => setTimeout(r, 1000));
    await expect(withTimeout(pending, 20)).rejects.toBeInstanceOf(TimeoutError);
  });

  it('resolves normally when the promise settles in time', async () => {
    await expect(withTimeout(Promise.resolve('value'), 1000)).resolves.toBe(
      'value',
    );
  });

  it('preserves the original rejection rather than masking it', async () => {
    await expect(
      withTimeout(Promise.reject(new Error('original')), 1000),
    ).rejects.toThrow('original');
  });

  it('is a passthrough for a non-positive budget', async () => {
    await expect(withTimeout(Promise.resolve(1), 0)).resolves.toBe(1);
  });

  // A hung item must free its worker slot, otherwise one dead provider stalls
  // the whole pool — the original Providers-tab symptom.
  it('lets a pool make progress past a hanging item', async () => {
    const completed: number[] = [];
    await pooledMapSettled(
      [1, 2, 3],
      async (n) => {
        // Item 2 never settles on its own.
        const work = n === 2 ? new Promise(() => {}) : Promise.resolve(n);
        const value = await withTimeout(work, 30);
        completed.push(n);
        return value;
      },
      (n) => n,
      1, // serial: item 2 would block 3 entirely without the timeout
    );
    expect(completed).toEqual([1, 3]);
  });
});
