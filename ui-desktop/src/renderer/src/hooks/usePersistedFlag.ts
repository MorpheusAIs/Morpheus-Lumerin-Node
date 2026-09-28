import { useCallback, useState } from 'react';

// Panel layout is a preference, not session state. Re-collapsing the sidebar on
// every launch is the kind of small friction that reads as the app forgetting
// what you told it, so the flag is written through to localStorage.
//
// Storage is deliberately best-effort: a renderer running without a DOM (tests,
// SSR-style rendering) or with storage disabled should fall back to the default
// rather than throw on the way up the tree.

function readFlag(key: string, fallback: boolean): boolean {
  try {
    const stored = window.localStorage.getItem(key);
    if (stored === 'true') return true;
    if (stored === 'false') return false;
  } catch {
    // Storage unavailable; the default is the honest answer.
  }
  return fallback;
}

function writeFlag(key: string, value: boolean): void {
  try {
    window.localStorage.setItem(key, String(value));
  } catch {
    // Losing the preference is survivable; crashing the panel is not.
  }
}

/**
 * A boolean that survives app restarts, with the same shape as useState.
 * The setter accepts a value or an updater, like useState's does.
 */
export function usePersistedFlag(
  key: string,
  fallback = false,
): [boolean, (next: boolean | ((prev: boolean) => boolean)) => void] {
  const [value, setValue] = useState(() => readFlag(key, fallback));

  const set = useCallback(
    (next: boolean | ((prev: boolean) => boolean)) => {
      setValue((prev) => {
        const resolved = typeof next === 'function' ? next(prev) : next;
        writeFlag(key, resolved);
        return resolved;
      });
    },
    [key],
  );

  return [value, set];
}

export const SIDEBAR_COLLAPSED_KEY = 'sidebar-collapsed';
export const CHAT_HISTORY_HIDDEN_KEY = 'chat-history-hidden';
