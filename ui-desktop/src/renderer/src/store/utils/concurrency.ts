// Bounded-concurrency helpers.
//
// Several screens used to fan out one request per item with an unbounded
// `Promise.all(items.map(...))`. With a few hundred providers/models that
// issues a few hundred simultaneous IPC round-trips, saturates the single
// proxy-router, and starves the renderer's event loop — which is why the
// Providers / Models / Agents tabs took forever to load and why the UI stopped
// responding to clicks while they did.
//
// `pooledMap` keeps at most `limit` requests in flight and preserves input
// order in the result array. `withTimeout` makes sure one hung provider can't
// pin a worker slot forever.

export const DEFAULT_CONCURRENCY = 6;

/**
 * Maps over `items` with at most `limit` concurrent invocations of `fn`.
 * Results keep the order of `items`. If `fn` rejects, the rejection propagates
 * (wrap the callback yourself if you want per-item error tolerance).
 */
export async function pooledMap<T, R>(
  items: readonly T[],
  fn: (item: T, index: number) => Promise<R>,
  limit: number = DEFAULT_CONCURRENCY,
): Promise<R[]> {
  if (!items.length) {
    return [];
  }

  const size = Math.max(1, Math.min(limit, items.length));
  const results: R[] = new Array(items.length);
  let cursor = 0;

  const worker = async () => {
    while (true) {
      const index = cursor++;
      if (index >= items.length) {
        return;
      }
      results[index] = await fn(items[index], index);
    }
  };

  await Promise.all(Array.from({ length: size }, worker));
  return results;
}

/**
 * Like `pooledMap`, but a rejected item yields `fallback` instead of failing
 * the whole batch. Use for best-effort fan-outs (availability pings, per-item
 * balance lookups) where one bad provider must not blank the entire table.
 */
export async function pooledMapSettled<T, R>(
  items: readonly T[],
  fn: (item: T, index: number) => Promise<R>,
  fallback: (item: T, index: number, error: unknown) => R,
  limit: number = DEFAULT_CONCURRENCY,
): Promise<R[]> {
  return pooledMap(
    items,
    async (item, index) => {
      try {
        return await fn(item, index);
      } catch (error) {
        return fallback(item, index, error);
      }
    },
    limit,
  );
}

export class TimeoutError extends Error {
  constructor(ms: number) {
    super(`Timed out after ${ms}ms`);
    this.name = 'TimeoutError';
  }
}

/**
 * Rejects with `TimeoutError` if `promise` hasn't settled within `ms`.
 *
 * Note this does not cancel the underlying work — it just stops the caller
 * waiting on it. That is enough to free a `pooledMap` worker slot, which is
 * the property we need here.
 */
export function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  if (!Number.isFinite(ms) || ms <= 0) {
    return promise;
  }

  return new Promise<T>((resolve, reject) => {
    const timer = setTimeout(() => reject(new TimeoutError(ms)), ms);
    promise.then(
      (value) => {
        clearTimeout(timer);
        resolve(value);
      },
      (error) => {
        clearTimeout(timer);
        reject(error);
      },
    );
  });
}
