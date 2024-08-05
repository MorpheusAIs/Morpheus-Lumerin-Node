import { useEffect, useRef } from "react";

export function useInterval(
  callback: () => void,
  delay: number,
  immediate = false,
  deps: React.DependencyList | undefined = []
): React.MutableRefObject<number> {
  const intervalRef = useRef<number | null>(null);
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
    return () => window.clearInterval(intervalRef.current!);
  }, [delay, ...deps]);
  return intervalRef;
}
