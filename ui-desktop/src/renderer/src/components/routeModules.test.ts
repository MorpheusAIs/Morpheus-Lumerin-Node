import { describe, expect, it, vi } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
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

describe('secondary route bundle boundaries', () => {
  it('does not pull password-validation dictionaries into Models or Providers', () => {
    for (const file of ['withModelsState.jsx', 'withProvidersState.jsx']) {
      const source = readFileSync(
        resolve(process.cwd(), 'src/renderer/src/store/hocs', file),
        'utf8',
      );
      expect(source).not.toMatch(/validators|zxcvbn/);
    }
  });
});
