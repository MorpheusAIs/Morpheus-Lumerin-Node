import { describe, expect, it, vi } from 'vitest';
import { sendToMainProcess } from './utils';

/**
 * Minimal stand-in for the contextBridge-exposed ipcRenderer.
 *
 * The real one hands every listener on a channel every message on that
 * channel — which is precisely the condition the correlation-id bug depended
 * on, so the fake must reproduce it faithfully.
 */
function createFakeIpc() {
  const listeners = new Map<string, Set<Function>>();

  return {
    sent: [] as Array<{ channel: string; payload: any }>,

    send(channel: string, payload: any) {
      this.sent.push({ channel, payload });
    },

    on(channel: string, listener: Function) {
      if (!listeners.has(channel)) {
        listeners.set(channel, new Set());
      }
      const set = listeners.get(channel)!;
      const unsubscribe = () => set.delete(listener);
      set.add(listener);
      return unsubscribe;
    },

    /** Deliver a response to EVERY listener on the channel, as Electron does. */
    respond(channel: string, message: { id: string; data?: any; error?: any }) {
      for (const listener of [...(listeners.get(channel) ?? [])]) {
        listener(message, () => listeners.get(channel)?.delete(listener));
      }
    },

    listenerCount(channel: string) {
      return listeners.get(channel)?.size ?? 0;
    },
  };
}

const idOf = (ipc: ReturnType<typeof createFakeIpc>, index = 0) =>
  ipc.sent[index].payload.id;

describe('sendToMainProcess', () => {
  it('resolves with the payload for a matching id', async () => {
    const ipc = createFakeIpc();
    const promise = sendToMainProcess('get-thing', undefined, 1000, ipc as any);

    ipc.respond('get-thing', { id: idOf(ipc), data: { value: 42 } });

    await expect(promise).resolves.toEqual({ value: 42 });
  });

  it('rejects when the response carries an error', async () => {
    const ipc = createFakeIpc();
    const promise = sendToMainProcess('do-thing', undefined, 1000, ipc as any);

    ipc.respond('do-thing', { id: idOf(ipc), error: 'nope' });

    await expect(promise).rejects.toBe('nope');
  });

  it('rejects when the payload contains an error field', async () => {
    const ipc = createFakeIpc();
    const promise = sendToMainProcess('do-thing', undefined, 1000, ipc as any);

    ipc.respond('do-thing', { id: idOf(ipc), data: { error: 'inner' } });

    await expect(promise).rejects.toBe('inner');
  });

  it('times out when no response arrives', async () => {
    vi.useFakeTimers();
    try {
      const ipc = createFakeIpc();
      const promise = sendToMainProcess('silent', undefined, 500, ipc as any);
      const assertion = expect(promise).rejects.toThrow(/timed out/i);

      await vi.advanceTimersByTimeAsync(600);
      await assertion;
    } finally {
      vi.useRealTimers();
    }
  });

  // ---------------------------------------------------------------------
  // Regression: the "buttons stop working after fast navigation" bug.
  //
  // The listener used to clear its timeout BEFORE comparing the correlation
  // id. Two concurrent calls on one channel each register a listener and each
  // sees both responses — so the response for call A ran call B's listener,
  // cleared B's timeout, then returned on the id mismatch. B was left with no
  // timer and no resolution: hung forever, no error, button dead.
  // ---------------------------------------------------------------------
  it('does not let one response cancel a different pending request', async () => {
    vi.useFakeTimers();
    try {
      const ipc = createFakeIpc();

      const first = sendToMainProcess('shared', { n: 1 }, 500, ipc as any);
      const second = sendToMainProcess('shared', { n: 2 }, 500, ipc as any);

      // Only the first request is answered.
      ipc.respond('shared', { id: idOf(ipc, 0), data: 'first-result' });
      await expect(first).resolves.toBe('first-result');

      // The second must still time out. Before the fix its timer had been
      // cleared by the response above, so this promise never settled and the
      // assertion below would hang until the test timeout.
      const assertion = expect(second).rejects.toThrow(/timed out/i);
      await vi.advanceTimersByTimeAsync(600);
      await assertion;
    } finally {
      vi.useRealTimers();
    }
  });

  it('resolves concurrent requests independently and correctly', async () => {
    const ipc = createFakeIpc();

    const a = sendToMainProcess('shared', { n: 'a' }, 1000, ipc as any);
    const b = sendToMainProcess('shared', { n: 'b' }, 1000, ipc as any);
    const c = sendToMainProcess('shared', { n: 'c' }, 1000, ipc as any);

    // Answer out of order to prove responses are matched by id, not arrival.
    ipc.respond('shared', { id: idOf(ipc, 2), data: 'C' });
    ipc.respond('shared', { id: idOf(ipc, 0), data: 'A' });
    ipc.respond('shared', { id: idOf(ipc, 1), data: 'B' });

    await expect(Promise.all([a, b, c])).resolves.toEqual(['A', 'B', 'C']);
  });

  it('gives each request a distinct correlation id', () => {
    const ipc = createFakeIpc();
    sendToMainProcess('shared', undefined, 1000, ipc as any);
    sendToMainProcess('shared', undefined, 1000, ipc as any);

    expect(idOf(ipc, 0)).not.toBe(idOf(ipc, 1));
  });

  it('unsubscribes after resolving so listeners do not accumulate', async () => {
    const ipc = createFakeIpc();
    const promise = sendToMainProcess('once', undefined, 1000, ipc as any);
    expect(ipc.listenerCount('once')).toBe(1);

    ipc.respond('once', { id: idOf(ipc), data: 'done' });
    await promise;

    expect(ipc.listenerCount('once')).toBe(0);
  });
});
