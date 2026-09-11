import { describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { loadAgentsPageData, useAgentTransactions } from './withAgentsState';

const deferred = <T>() => {
  let resolve!: (value: T) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<T>((done, fail) => {
    resolve = done;
    reject = fail;
  });
  return { promise, resolve, reject };
};

describe('loadAgentsPageData', () => {
  it('starts independent reads together and categorizes the cached result', async () => {
    const users = deferred<any>();
    const allowances = deferred<any>();
    const client = {
      getAgentUsers: vi.fn(() => users.promise),
      getAgentAllowanceRequests: vi.fn(() => allowances.promise),
    };

    const result = loadAgentsPageData(client);

    expect(client.getAgentUsers).toHaveBeenCalledOnce();
    expect(client.getAgentAllowanceRequests).toHaveBeenCalledOnce();
    users.resolve({
      agents: [
        { username: 'pending', isConfirmed: false },
        { username: 'active', isConfirmed: true },
      ],
    });
    allowances.resolve({ requests: [{ username: 'pending', token: 'MOR' }] });

    await expect(result).resolves.toMatchObject({
      pendingAgents: [{ username: 'pending' }],
      activeAgents: [{ username: 'active' }],
      allowanceRequests: [{ username: 'pending', token: 'MOR' }],
    });
  });
});

describe('useAgentTransactions', () => {
  it('fetches only while loading and does not refetch a successful result', async () => {
    const client = {
      getAgentTxs: vi.fn(async () => ({ txHashes: ['0x123'] })),
    };
    const { result } = renderHook(() => useAgentTransactions(client as any));
    expect(client.getAgentTxs).not.toHaveBeenCalled();
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'helper' }),
    );
    await waitFor(() =>
      expect(result.current.txModal).toEqual({
        state: 'success',
        agentName: 'helper',
        data: ['0x123'],
      }),
    );
    expect(client.getAgentTxs).toHaveBeenCalledOnce();
  });

  it('does not reopen a closed dialog when its request completes', async () => {
    const request = deferred<any>();
    const client = { getAgentTxs: vi.fn(() => request.promise) };
    const { result } = renderHook(() => useAgentTransactions(client as any));
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'helper' }),
    );
    act(() => result.current.setTxModal({ state: 'pending' }));
    await act(async () => request.resolve({ txHashes: ['late'] }));
    expect(result.current.txModal).toEqual({ state: 'pending' });
  });

  it('ignores a stale agent response after selecting another agent', async () => {
    const first = deferred<any>();
    const second = deferred<any>();
    const client = {
      getAgentTxs: vi
        .fn()
        .mockReturnValueOnce(first.promise)
        .mockReturnValueOnce(second.promise),
    };
    const { result } = renderHook(() => useAgentTransactions(client as any));
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'first' }),
    );
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'second' }),
    );
    await act(async () => second.resolve({ txHashes: ['second-result'] }));
    await act(async () => first.resolve({ txHashes: ['first-result'] }));
    expect(result.current.txModal).toEqual({
      state: 'success',
      agentName: 'second',
      data: ['second-result'],
    });
  });

  it('catches failures and allows retry for the same agent', async () => {
    const client = {
      getAgentTxs: vi
        .fn()
        .mockRejectedValueOnce(new Error('offline'))
        .mockResolvedValueOnce({ txHashes: [] }),
    };
    const { result } = renderHook(() => useAgentTransactions(client as any));
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'helper' }),
    );
    await waitFor(() => expect(result.current.txModal.state).toBe('error'));
    expect(client.getAgentTxs).toHaveBeenCalledOnce();
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'helper' }),
    );
    await waitFor(() =>
      expect(result.current.txModal).toEqual({
        state: 'success',
        agentName: 'helper',
        data: [],
      }),
    );
    expect(client.getAgentTxs).toHaveBeenCalledTimes(2);
  });

  it('handles missing transaction data as a recoverable error', async () => {
    const client = { getAgentTxs: vi.fn(async () => undefined) };
    const { result } = renderHook(() => useAgentTransactions(client as any));
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'helper' }),
    );
    await waitFor(() => expect(result.current.txModal.state).toBe('error'));
  });

  it('closes transaction history on a wallet switch and ignores its old response', async () => {
    const request = deferred<any>();
    const client = { getAgentTxs: vi.fn(() => request.promise) };
    const { result, rerender } = renderHook(
      ({ address }) => useAgentTransactions(client as any, address),
      { initialProps: { address: 'first-wallet' } },
    );
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'helper' }),
    );
    rerender({ address: 'second-wallet' });
    await act(async () => request.resolve({ txHashes: ['old-wallet'] }));
    expect(result.current.txModal).toEqual({ state: 'pending' });
    expect(client.getAgentTxs).toHaveBeenCalledOnce();
  });

  it('handles late rejection after unmount without surfacing an unhandled promise', async () => {
    const request = deferred<any>();
    const client = { getAgentTxs: vi.fn(() => request.promise) };
    const { result, unmount } = renderHook(() =>
      useAgentTransactions(client as any),
    );
    act(() =>
      result.current.setTxModal({ state: 'loading', agentName: 'helper' }),
    );
    unmount();
    await act(async () => request.reject(new Error('late rejection')));
  });
});
