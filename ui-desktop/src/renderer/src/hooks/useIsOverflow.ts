import { useLayoutEffect, useState } from 'react';

export const useIsOverflow = (
  ref: React.RefObject<HTMLElement>,
  callback?: (oveflow: { x: boolean; y: boolean }) => void,
) => {
  const [isOverflow, setIsOverflow] = useState({ x: false, y: false });

  useLayoutEffect(() => {
    const { current } = ref;

    const trigger = () => {
      if (!current) return;
      const hasOverflow = {
        x: current.scrollWidth > current.clientWidth,
        y: current.scrollHeight > current.clientHeight,
      };

      setIsOverflow(hasOverflow);

      if (callback) callback(hasOverflow);
    };

    if (current) {
      trigger();
    }
  }, [callback, ref]);

  return isOverflow;
};
