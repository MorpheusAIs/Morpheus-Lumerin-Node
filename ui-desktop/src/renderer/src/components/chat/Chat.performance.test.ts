import { describe, expect, it, vi } from 'vitest';
import {
  createAnimationFrameBatch,
  disposeActiveChatStream,
  isNearChatBottom,
  observeChatAutoScroll,
  revokeInactiveObjectUrls,
} from './Chat';

describe('Chat streaming performance helpers', () => {
  it('treats only the configured bottom threshold as eligible for auto-scroll', () => {
    expect(
      isNearChatBottom({
        clientHeight: 100,
        scrollHeight: 1_000,
        scrollTop: 804,
      }),
    ).toBe(true);
    expect(
      isNearChatBottom({
        clientHeight: 100,
        scrollHeight: 1_000,
        scrollTop: 803,
      }),
    ).toBe(false);
  });

  it('removes the exact scroll and wheel listeners that it registered', () => {
    const addEventListener = vi.fn();
    const removeEventListener = vi.fn();
    const element = {
      addEventListener,
      clientHeight: 100,
      removeEventListener,
      scrollHeight: 1_000,
      scrollTop: 850,
    } as unknown as HTMLElement;
    const onAutoScrollChange = vi.fn();

    const cleanup = observeChatAutoScroll(element, onAutoScrollChange);
    const scrollListener = addEventListener.mock.calls.find(
      ([eventName]) => eventName === 'scroll',
    )?.[1] as EventListener;
    const wheelListener = addEventListener.mock.calls.find(
      ([eventName]) => eventName === 'wheel',
    )?.[1] as EventListener;

    scrollListener(new Event('scroll'));
    wheelListener(new WheelEvent('wheel', { deltaY: -1 }));
    cleanup();

    expect(onAutoScrollChange).toHaveBeenNthCalledWith(1, true);
    expect(onAutoScrollChange).toHaveBeenNthCalledWith(2, false);
    expect(removeEventListener).toHaveBeenCalledWith('scroll', scrollListener);
    expect(removeEventListener).toHaveBeenCalledWith('wheel', wheelListener);
  });

  it('coalesces stream updates into one animation-frame commit', () => {
    let nextHandle = 0;
    const frames = new Map<number, FrameRequestCallback>();
    const requestFrame = vi.fn((callback: FrameRequestCallback) => {
      const handle = ++nextHandle;
      frames.set(handle, callback);
      return handle;
    });
    const cancelFrame = vi.fn((handle: number) => frames.delete(handle));
    const commits: string[][] = [];
    const batch = createAnimationFrameBatch<string[]>(
      (value) => commits.push(value),
      requestFrame,
      cancelFrame,
    );

    batch.schedule(['first']);
    batch.schedule(['latest']);

    expect(requestFrame).toHaveBeenCalledTimes(1);
    expect(commits).toEqual([]);
    frames.get(1)?.(16);
    expect(commits).toEqual([['latest']]);

    batch.schedule(['final']);
    batch.flush();
    expect(cancelFrame).toHaveBeenCalledWith(2);
    expect(commits).toEqual([['latest'], ['final']]);
  });

  it('can cancel a pending stream render without committing stale state', () => {
    const commits: string[] = [];
    const cancelFrame = vi.fn();
    const batch = createAnimationFrameBatch<string>(
      (value) => commits.push(value),
      () => 42,
      cancelFrame,
    );

    batch.schedule('stale');
    batch.cancel();
    batch.flush();

    expect(cancelFrame).toHaveBeenCalledWith(42);
    expect(commits).toEqual([]);
  });

  it('cancels the reader on unmount and blocks its queued stale commit', async () => {
    const mounted = { current: true };
    const generation = { current: 7 };
    const cancel = vi.fn().mockResolvedValue(undefined);
    const reader = {
      cancel,
    } as unknown as ReadableStreamDefaultReader<Uint8Array>;
    const activeReader = { current: reader };
    const commits: string[] = [];
    const requestGeneration = generation.current;
    const batch = createAnimationFrameBatch<string>(
      (value) => {
        if (mounted.current && generation.current === requestGeneration) {
          commits.push(value);
        }
      },
      () => 42,
      vi.fn(),
    );

    batch.schedule('stale');
    await disposeActiveChatStream(mounted, generation, activeReader);
    batch.flush();

    expect(cancel).toHaveBeenCalledOnce();
    expect(activeReader.current).toBeNull();
    expect(mounted.current).toBe(false);
    expect(generation.current).toBe(8);
    expect(commits).toEqual([]);
  });

  it('revokes only object URLs no longer referenced by audio messages', () => {
    const owned = new Set(['blob:old', 'blob:active']);
    const revoke = vi.fn();

    revokeInactiveObjectUrls(owned, new Set(['blob:active']), revoke);

    expect(revoke).toHaveBeenCalledOnce();
    expect(revoke).toHaveBeenCalledWith('blob:old');
    expect([...owned]).toEqual(['blob:active']);
  });
});
