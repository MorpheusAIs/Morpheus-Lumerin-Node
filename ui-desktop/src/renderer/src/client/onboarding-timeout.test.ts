import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('./keys', () => ({ default: {} }));
vi.mock('./sentry', () => ({}));

import createClient from './index';

describe('onboarding IPC deadline', () => {
  let originalIpc: typeof window.ipcRenderer;
  let listeners: Map<string, Function>;
  let sent: Array<{ channel: string; payload: any }>;
  let client: ReturnType<typeof createClient>;

  beforeEach(() => {
    vi.useFakeTimers();
    originalIpc = window.ipcRenderer;
    listeners = new Map();
    sent = [];
    window.ipcRenderer = {
      send: (channel, payload) => sent.push({ channel, payload }),
      on: (channel, listener) => {
        listeners.set(channel, listener);
        return () => listeners.delete(channel);
      },
    } as any;
    client = createClient(() => ({
      dispatch: vi.fn(),
      getState: vi.fn(),
      subscribe: vi.fn(),
    }));
  });

  afterEach(() => {
    window.ipcRenderer = originalIpc;
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('waits past the old 10-second cutoff and accepts the completed setup result', async () => {
    const result = client.onOnboardingCompleted({
      password: 'synthetic-test-password',
    });
    const settled = vi.fn();
    result.then(settled, settled);
    await vi.advanceTimersByTimeAsync(12_000);
    expect(settled).not.toHaveBeenCalled();
    const request = sent.find(
      ({ channel }) => channel === 'onboarding-completed',
    )!;
    listeners.get(request.channel)?.(
      { id: request.payload.id, data: undefined },
      () => listeners.delete(request.channel),
    );
    await expect(result).resolves.toBeUndefined();
    expect(listeners.has(request.channel)).toBe(false);
    expect(vi.getTimerCount()).toBe(0);
  });

  it('still terminates an unanswered onboarding IPC after its dedicated 60-second budget', async () => {
    vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    const result = client.onOnboardingCompleted({});
    const assertion = expect(result).rejects.toThrow('Operation timed out');
    await vi.advanceTimersByTimeAsync(59_999);
    expect(listeners.has('onboarding-completed')).toBe(true);
    await vi.advanceTimersByTimeAsync(1);
    await assertion;
    expect(listeners.has('onboarding-completed')).toBe(false);
  });

  it('does not extend the default login IPC deadline', async () => {
    vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    const result = client.onLoginSubmit({
      password: 'synthetic-test-password',
    });
    const assertion = expect(result).rejects.toThrow('Operation timed out');
    await vi.advanceTimersByTimeAsync(10_000);
    await assertion;
    expect(listeners.has('login-submit')).toBe(false);
  });
});
