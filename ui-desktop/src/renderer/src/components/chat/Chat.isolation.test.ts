import { describe, expect, it, vi } from 'vitest';
import { isStaleChatWork, resolveChatId } from './Chat';

/**
 * These cover the two mechanisms that keep one chat's context out of another.
 *
 * The third mechanism is structural rather than testable in isolation: the abort
 * flag used to be a module-level `let`, so two mounted chats shared one boolean
 * and stopping or switching in either cancelled the other's stream. It is a ref
 * on the component now, which a unit test cannot observe without mounting the
 * whole screen, so it is verified by the type of the declaration instead.
 */

describe('stale chat work', () => {
  const ref = <T,>(current: T) => ({ current });

  it('accepts work whose generation still matches the live one', () => {
    expect(isStaleChatWork(ref(true), ref(4), 4)).toBe(false);
  });

  it('rejects a response that arrives after the user switched chats', () => {
    // selectChat bumps the generation, so the in-flight load for the previous
    // chat resolves against a number that no longer matches.
    const generationRef = ref(4);
    generationRef.current += 1;
    expect(isStaleChatWork(ref(true), generationRef, 4)).toBe(true);
  });

  it('rejects work that resolves after the screen unmounted', () => {
    expect(isStaleChatWork(ref(false), ref(4), 4)).toBe(true);
  });

  it('rejects on unmount even when the generation is unchanged', () => {
    // Both conditions have to be independent: a teardown that forgot to bump
    // the generation must still stop the setState.
    expect(isStaleChatWork(ref(false), ref(0), 0)).toBe(true);
  });

  it('reads the refs at call time rather than closing over their values', () => {
    const mountedRef = ref(true);
    const generationRef = ref(1);
    const isStale = () => isStaleChatWork(mountedRef, generationRef, 1);

    expect(isStale()).toBe(false);
    generationRef.current = 2;
    // A single load checks this after the fetch and again after mapping the
    // messages, so a switch between those two points has to be caught.
    expect(isStale()).toBe(true);
  });
});

describe('chat id resolution', () => {
  it('keeps the id the chat already has', () => {
    const mint = vi.fn(() => 'minted');
    expect(resolveChatId('existing', mint)).toBe('existing');
    expect(mint).not.toHaveBeenCalled();
  });

  it('mints an id when the chat has none', () => {
    expect(resolveChatId(undefined, () => 'minted')).toBe('minted');
  });

  it('mints rather than passing an empty id through', () => {
    // An empty string reaches the router as an absent header, which is the
    // orphan case this exists to prevent.
    expect(resolveChatId('', () => 'minted')).toBe('minted');
  });

  it('gives two chats different ids', () => {
    let n = 0;
    const mint = () => `chat-${(n += 1)}`;
    expect(resolveChatId(undefined, mint)).not.toBe(
      resolveChatId(undefined, mint),
    );
  });

  it('gives the same chat one id across turns once it has been minted', () => {
    const mint = vi.fn(() => 'chat-1');
    const first = resolveChatId(undefined, mint);
    const second = resolveChatId(first, mint);
    expect(second).toBe(first);
    expect(mint).toHaveBeenCalledTimes(1);
  });
});
