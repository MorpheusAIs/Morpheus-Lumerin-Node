import { useLayoutEffect, useRef, useState } from 'react';

type Overflow = { x: boolean; y: boolean };

export const useIsOverflow = (
  ref: React.RefObject<HTMLElement>,
  callback?: (overflow: Overflow) => void,
) => {
  const [isOverflow, setIsOverflow] = useState({ x: false, y: false });
  const callbackRef = useRef(callback);
  callbackRef.current = callback;

  useLayoutEffect(() => {
    const { current } = ref;
    if (!current) return;
    let disposed = false;

    const trigger = () => {
      if (disposed) return;
      const hasOverflow = {
        x: current.scrollWidth > current.clientWidth,
        y: current.scrollHeight > current.clientHeight,
      };
      setIsOverflow((previous) =>
        previous.x === hasOverflow.x && previous.y === hasOverflow.y
          ? previous
          : hasOverflow,
      );
      callbackRef.current?.(hasOverflow);
    };

    trigger();
    // Fixed-height previews can gain content without resizing themselves.
    const resizeObserver =
      typeof ResizeObserver === 'undefined'
        ? null
        : new ResizeObserver(trigger);
    resizeObserver?.observe(current);
    const mutationObserver =
      typeof MutationObserver === 'undefined'
        ? null
        : new MutationObserver(trigger);
    mutationObserver?.observe(current, {
      childList: true,
      characterData: true,
      subtree: true,
    });
    window.addEventListener('resize', trigger);
    void document.fonts?.ready.then(trigger, () => undefined);

    return () => {
      disposed = true;
      resizeObserver?.disconnect();
      mutationObserver?.disconnect();
      window.removeEventListener('resize', trigger);
    };
  }, [ref]);

  return isOverflow;
};
