import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ToastsContext } from '../toasts';
import { Root } from './Root';

const EmptyComponent = () => <div />;

describe('Root bootstrap', () => {
  afterEach(() => vi.restoreAllMocks());

  it('falls back from a display-currency read failure without reporting wallet startup failure', async () => {
    const dispatch = vi.fn();
    const toast = vi.fn();
    const currencyError = new Error('settings request timed out');
    const consoleWarning = vi
      .spyOn(console, 'warn')
      .mockImplementation(() => undefined);
    const client = {
      onInit: vi.fn(async () => ({
        onboardingComplete: true,
        persistedState: {},
        config: {},
      })),
      getDefaultCurrencySetting: vi.fn(async () => {
        throw currencyError;
      }),
    };

    render(
      <ToastsContext.Provider value={{ toast }}>
        <Root
          isSessionActive={false}
          isAuthBypassed={false}
          sellerDefaultCurrency="MOR"
          servicesState={{ orchestratorStatus: 'starting' } as any}
          config={{}}
          dispatch={dispatch}
          client={client}
          StartupComponent={EmptyComponent}
          OnboardingComponent={EmptyComponent as any}
          RouterComponent={EmptyComponent}
          LoginComponent={EmptyComponent as any}
        />
      </ToastsContext.Provider>,
    );

    await waitFor(() =>
      expect(dispatch).toHaveBeenCalledWith({
        type: 'set-seller-currency',
        payload: 'MOR',
      }),
    );

    expect(toast).not.toHaveBeenCalled();
    expect(consoleWarning).toHaveBeenCalledWith(
      'Could not load the saved display currency; using the default.',
      currencyError,
    );
  });

  it('still reports a genuine renderer bootstrap failure and skips follow-up settings reads', async () => {
    const dispatch = vi.fn();
    const toast = vi.fn();
    const consoleError = vi
      .spyOn(console, 'error')
      .mockImplementation(() => undefined);
    const client = {
      onInit: vi.fn(async () => {
        throw new Error('renderer bootstrap failed');
      }),
      getDefaultCurrencySetting: vi.fn(),
    };

    render(
      <ToastsContext.Provider value={{ toast }}>
        <Root
          isSessionActive={false}
          isAuthBypassed={false}
          sellerDefaultCurrency="MOR"
          servicesState={{ orchestratorStatus: 'starting' } as any}
          config={{}}
          dispatch={dispatch}
          client={client}
          StartupComponent={EmptyComponent}
          OnboardingComponent={EmptyComponent as any}
          RouterComponent={EmptyComponent}
          LoginComponent={EmptyComponent as any}
        />
      </ToastsContext.Provider>,
    );

    await waitFor(() =>
      expect(toast).toHaveBeenCalledWith(
        'error',
        'Failed to initialize Morpheus. Your wallet is unchanged; restart the app and try again.',
      ),
    );

    expect(client.getDefaultCurrencySetting).not.toHaveBeenCalled();
    expect(consoleError).toHaveBeenCalledWith(
      'root component error',
      'renderer bootstrap failed',
    );
  });

  it('renders the authenticated app immediately without waiting for optional chain data', async () => {
    const dispatch = vi.fn();
    const client = {
      onInit: vi.fn(async () => ({
        onboardingComplete: true,
        persistedState: {},
        config: {},
      })),
      getDefaultCurrencySetting: vi.fn(async () => 'MOR'),
    };

    const Startup = () => <div data-testid="startup" />;
    const Router = () => <div data-testid="router" />;

    render(
      <ToastsContext.Provider value={{ toast: vi.fn() }}>
        <Root
          isSessionActive
          isAuthBypassed={false}
          sellerDefaultCurrency="MOR"
          servicesState={{ orchestratorStatus: 'ready' } as any}
          config={{}}
          dispatch={dispatch}
          client={client}
          StartupComponent={Startup}
          OnboardingComponent={EmptyComponent as any}
          RouterComponent={Router}
          LoginComponent={EmptyComponent as any}
        />
      </ToastsContext.Provider>,
    );

    await waitFor(() => expect(screen.getByTestId('router')).toBeTruthy());
    expect(screen.queryByTestId('startup')).toBeNull();
    expect(dispatch).not.toHaveBeenCalledWith({
      type: 'required-data-gathered',
    });
  });
});
