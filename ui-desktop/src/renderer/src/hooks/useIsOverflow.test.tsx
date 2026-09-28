import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useIsOverflow } from './useIsOverflow';

type Observer = {
  trigger: () => void;
  observe: ReturnType<typeof vi.fn>;
  disconnect: ReturnType<typeof vi.fn>;
};
let observers: Observer[] = [];

function measuredElement() {
  const element = document.createElement('div');
  const dimensions = {
    clientWidth: 100,
    scrollWidth: 100,
    clientHeight: 40,
    scrollHeight: 40,
  };
  Object.entries(dimensions).forEach(([property]) => {
    Object.defineProperty(element, property, {
      configurable: true,
      get: () => dimensions[property],
    });
  });
  return { element, dimensions, ref: { current: element } };
}

describe('useIsOverflow', () => {
  beforeEach(() => {
    observers = [];
    vi.stubGlobal(
      'ResizeObserver',
      class {
        observe = vi.fn();
        disconnect = vi.fn();
        trigger: () => void;
        constructor(trigger: () => void) {
          this.trigger = trigger;
          observers.push(this);
        }
      },
    );
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('measures on mount and updates both directions after the available size changes', () => {
    const { dimensions, ref } = measuredElement();
    dimensions.scrollWidth = 150;
    const { result } = renderHook(() => useIsOverflow(ref));
    expect(result.current).toEqual({ x: true, y: false });

    act(() => {
      dimensions.clientWidth = 200;
      dimensions.scrollHeight = 80;
      observers[0].trigger();
    });
    expect(result.current).toEqual({ x: false, y: true });

    act(() => {
      dimensions.clientHeight = 100;
      observers[0].trigger();
    });
    expect(result.current).toEqual({ x: false, y: false });
  });

  it('detects updated content even if the fixed-size container did not resize', async () => {
    const { element, dimensions, ref } = measuredElement();
    const { result } = renderHook(() => useIsOverflow(ref));
    act(() => {
      dimensions.scrollHeight = 90;
      element.appendChild(document.createTextNode('Another allowance'));
    });
    await waitFor(() => expect(result.current.y).toBe(true));
  });

  it('falls back to window resize when ResizeObserver is unavailable', () => {
    vi.stubGlobal('ResizeObserver', undefined);
    const { dimensions, ref } = measuredElement();
    const { result } = renderHook(() => useIsOverflow(ref));
    act(() => {
      dimensions.scrollWidth = 200;
      window.dispatchEvent(new Event('resize'));
    });
    expect(result.current.x).toBe(true);
  });

  it('uses the latest callback without recreating observers and cleans up on unmount', () => {
    const { dimensions, ref } = measuredElement();
    const first = vi.fn();
    const latest = vi.fn();
    const { rerender, unmount } = renderHook(
      ({ callback }) => useIsOverflow(ref, callback),
      { initialProps: { callback: first } },
    );
    rerender({ callback: latest });
    expect(observers).toHaveLength(1);
    act(() => {
      dimensions.scrollWidth = 200;
      observers[0].trigger();
    });
    expect(latest).toHaveBeenLastCalledWith({ x: true, y: false });
    unmount();
    expect(observers[0].disconnect).toHaveBeenCalledOnce();
    latest.mockClear();
    act(() => {
      observers[0].trigger();
      window.dispatchEvent(new Event('resize'));
    });
    expect(latest).not.toHaveBeenCalled();
  });
});
