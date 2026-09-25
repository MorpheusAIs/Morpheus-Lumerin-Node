import { describe, expect, it, beforeEach, vi } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import { usePersistedFlag } from './usePersistedFlag';

describe('usePersistedFlag', () => {
  beforeEach(() => {
    window.localStorage.clear();
    vi.restoreAllMocks();
  });

  it('starts from the fallback when nothing is stored', () => {
    const { result } = renderHook(() => usePersistedFlag('panel', true));

    expect(result.current[0]).toBe(true);
  });

  it('reads a previously written value', () => {
    window.localStorage.setItem('panel', 'true');

    const { result } = renderHook(() => usePersistedFlag('panel'));

    expect(result.current[0]).toBe(true);
  });

  it('writes through on every change, including via an updater', () => {
    const { result } = renderHook(() => usePersistedFlag('panel'));

    act(() => result.current[1](true));
    expect(window.localStorage.getItem('panel')).toBe('true');

    act(() => result.current[1]((prev) => !prev));
    expect(result.current[0]).toBe(false);
    expect(window.localStorage.getItem('panel')).toBe('false');
  });

  // Anything other than the two values we write is a foreign key or a corrupted
  // one, and guessing at it is worse than falling back.
  it('ignores a value it did not write', () => {
    window.localStorage.setItem('panel', 'yes please');

    const { result } = renderHook(() => usePersistedFlag('panel', true));

    expect(result.current[0]).toBe(true);
  });

  it('survives storage that throws', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('denied');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('denied');
    });

    const { result } = renderHook(() => usePersistedFlag('panel', true));
    expect(result.current[0]).toBe(true);

    act(() => result.current[1](false));
    expect(result.current[0]).toBe(false);
  });
});
