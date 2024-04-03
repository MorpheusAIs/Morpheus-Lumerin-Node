import { useEffect, useRef } from 'react';

/**
 * SetInterval for react components
 * @param {Function} callback
 * @param {number} delay
 * @param {boolean} immediate whether to call callback immediately
 * @param {React.DependencyList | undefined} deps dependency list, like in useEffect
 * @returns {React.MutableRefObject<number>}
 */
export function useInterval(callback, delay, immediate = false, deps = []) {
  const intervalRef = useRef(null);
  const savedCallback = useRef(callback);
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback, ...deps]);
  useEffect(() => {
    const tick = () => savedCallback.current();
    if (immediate) {
      tick();
    }
    intervalRef.current = window.setInterval(tick, delay);
    return () => window.clearInterval(intervalRef.current);
  }, [delay, ...deps]);
  return intervalRef;
}
