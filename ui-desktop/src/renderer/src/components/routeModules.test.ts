import { describe, expect, it, vi } from 'vitest';
import { createRoutePreloader } from './routeModules';

describe('route module intent preloading', () => {
  it('deduplicates repeated hover, focus, and click intent', async () => {
    const load = vi.fn(async () => ({ default: () => null }));
    const preload = createRoutePreloader({ '/chat': load });

    const first = preload('/chat');
    const second = preload('/chat');

    expect(second).toBe(first);
    await first;
    expect(load).toHaveBeenCalledOnce();
  });

  it('allows a failed local chunk read to be retried', async () => {
    const load = vi
      .fn()
      .mockRejectedValueOnce(new Error('chunk unavailable'))
      .mockResolvedValueOnce({ default: () => null });
    const preload = createRoutePreloader({ '/models': load });

    await expect(preload('/models')).rejects.toThrow('chunk unavailable');
    await expect(preload('/models')).resolves.toBeTruthy();
    expect(load).toHaveBeenCalledTimes(2);
  });
});
