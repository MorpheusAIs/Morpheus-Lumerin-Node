import { QueryClient } from '@tanstack/react-query';

const queryErrorMessage = (error: unknown): string =>
  String(
    (error as { message?: unknown } | null)?.message ?? error ?? '',
  ).toLowerCase();

/**
 * The orchestrator can report ready a fraction before the proxy-router's HTTP
 * listener accepts its first request. Retry that narrow connection race for a
 * few seconds, while preserving the existing single retry for real API errors.
 */
export const shouldRetryDesktopQuery = (
  failureCount: number,
  error: unknown,
): boolean => {
  const message = queryErrorMessage(error);
  const nodeStillStarting =
    /cannot reach|econnrefused|connection refused|failed to fetch|socket hang up/.test(
      message,
    );
  return nodeStillStarting ? failureCount < 3 : failureCount < 1;
};

export const desktopQueryRetryDelay = (attempt: number): number =>
  Math.min(250 * 2 ** attempt, 2_000);

// A single app-level QueryClient. It lives above the router (see App.tsx) so the
// cache survives route unmounts. Revisiting a tab serves cached data instantly
// and revalidates in the background (stale-while-revalidate), instead of showing
// a blocking full-screen loader every time.
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Data is considered fresh for this long; within the window a remount
      // reuses the cache with no network call at all.
      staleTime: 30_000,
      // Keep unused data around so navigating back is instant.
      gcTime: 5 * 60_000,
      // The desktop window focus/blur churn would otherwise trigger constant
      // refetches; we rely on staleTime + explicit invalidation instead.
      refetchOnWindowFocus: false,
      retry: shouldRetryDesktopQuery,
      retryDelay: desktopQueryRetryDelay,
    },
  },
});
