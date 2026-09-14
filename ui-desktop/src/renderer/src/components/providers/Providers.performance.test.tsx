import { act, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, expect, it, vi } from 'vitest';
import { Providers } from './Providers';

vi.mock('../common/LayoutHeader', () => ({
  LayoutHeader: ({ children, title }) => (
    <header>
      {title}
      {children}
    </header>
  ),
}));
vi.mock('../common/View', () => ({
  View: ({ children }) => <main>{children}</main>,
}));
vi.mock('../common/QueryError', () => ({ default: () => null }));
vi.mock('../dashboard/BalanceBlock.styles', () => ({
  BtnAccent: ({ children }) => <button>{children}</button>,
}));
vi.mock('./ProvidersList', () => ({
  default: ({ sessions, sessionsLoading, modelNames, balancesLoading }) => (
    <div>
      <span data-testid="session-count">{sessions.length}</span>
      <span data-testid="sessions-loading">{String(sessionsLoading)}</span>
      <span data-testid="model-name">{modelNames['model-1'] ?? 'pending'}</span>
      <span data-testid="balances-loading">{String(balancesLoading)}</span>
    </div>
  ),
}));

const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};

describe('Provider Hub progressive loading', () => {
  it('shows sessions before optional model names and balances resolve', async () => {
    const sessions = deferred<any[]>();
    const models = deferred<any[]>();
    const balance = deferred<string>();
    const getSessionsByProvider = vi.fn(() => sessions.promise);
    const getAllModels = vi.fn(() => models.promise);
    const getBalanceBySession = vi.fn(() => balance.promise);
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <Providers
          providerId="0x1111111111111111111111111111111111111111"
          getSessionsByProvider={getSessionsByProvider}
          getAllModels={getAllModels}
          getBalanceBySession={getBalanceBySession}
          claimFunds={vi.fn()}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByTestId('sessions-loading').textContent).toBe('true');
    expect(getAllModels).not.toHaveBeenCalled();

    await act(async () => {
      sessions.resolve([
        {
          Id: 'session-1',
          BidID: 'bid-1',
          ModelAgentId: 'model-1',
          ClosedAt: 0,
        },
      ]);
    });

    await waitFor(() => {
      expect(screen.getByTestId('session-count').textContent).toBe('1');
    });
    expect(screen.getByTestId('model-name').textContent).toBe('pending');
    expect(screen.getByTestId('balances-loading').textContent).toBe('true');
    expect(getAllModels).toHaveBeenCalledOnce();
    expect(getBalanceBySession).toHaveBeenCalledWith('session-1');

    await act(async () => {
      models.resolve([{ Id: 'model-1', Name: 'Fast model' }]);
      balance.resolve('1000000000000000000');
    });
    await waitFor(() => {
      expect(screen.getByTestId('model-name').textContent).toBe('Fast model');
      expect(screen.getByTestId('balances-loading').textContent).toBe('false');
    });
  });

  it('does not scan models or balances when the provider has no sessions', async () => {
    const getAllModels = vi.fn().mockResolvedValue([]);
    const getBalanceBySession = vi.fn().mockResolvedValue('0');
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <Providers
          providerId="0x2222222222222222222222222222222222222222"
          getSessionsByProvider={vi.fn().mockResolvedValue([])}
          getAllModels={getAllModels}
          getBalanceBySession={getBalanceBySession}
          claimFunds={vi.fn()}
        />
      </QueryClientProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId('sessions-loading').textContent).toBe('false');
    });
    expect(getAllModels).not.toHaveBeenCalled();
    expect(getBalanceBySession).not.toHaveBeenCalled();
  });
});
