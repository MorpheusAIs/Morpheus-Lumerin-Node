import { act, render, screen, waitFor } from '@testing-library/react';
import { createRef } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { Root } from './Root';
import { ToastsContext } from '../toasts';

function fixture(onboardingComplete: boolean, bypass = false) {
  const client = {
    onInit: vi.fn(async () => ({
      onboardingComplete,
      persistedState: {},
      config: { chain: {} },
    })),
    getDefaultCurrencySetting: vi.fn(async () => 'MOR'),
    onLoginSubmit: vi.fn<(...args: any[]) => Promise<any>>(),
    onOnboardingCompleted: vi.fn<(...args: any[]) => Promise<any>>(),
  };
  const dispatch = vi.fn();
  const ref = createRef<Root>();
  render(
    <ToastsContext.Provider value={{ toast: vi.fn() }}>
      <Root
        ref={ref}
        client={client}
        dispatch={dispatch}
        config={{ chain: { localProxyRouterUrl: 'http://127.0.0.1:8082' } }}
        servicesState={{ orchestratorStatus: 'ready' } as any}
        sellerDefaultCurrency="MOR"
        isSessionActive={false}
        isAuthBypassed={bypass}
        StartupComponent={() => <p>Starting</p>}
        OnboardingComponent={() => <h1>Create or import wallet</h1>}
        LoginComponent={() => <h1>Unlock wallet</h1>}
        RouterComponent={() => <h1>Authenticated app</h1>}
      />
    </ToastsContext.Provider>,
  );
  return { client, dispatch, ref };
}

describe('Root first-run and returning-wallet routing', () => {
  it.each([false, true])(
    'routes a fresh install to setup, never login or bypass (bypass=%s)',
    async (bypass) => {
      const { client, dispatch } = fixture(false, bypass);
      expect(
        await screen.findByRole('heading', { name: 'Create or import wallet' }),
      ).toBeVisible();
      expect(
        screen.queryByRole('heading', { name: 'Unlock wallet' }),
      ).not.toBeInTheDocument();
      expect(client.onLoginSubmit).not.toHaveBeenCalled();
      expect(dispatch).not.toHaveBeenCalledWith({ type: 'session-started' });
    },
  );

  it('preserves unlock for an existing installation without exposing replacement setup', async () => {
    fixture(true);
    expect(
      await screen.findByRole('heading', { name: 'Unlock wallet' }),
    ).toBeVisible();
    expect(
      screen.queryByRole('heading', { name: 'Create or import wallet' }),
    ).not.toBeInTheDocument();
  });

  it('returns to setup on authenticated empty-wallet recovery without starting a session', async () => {
    const { client, dispatch, ref } = fixture(true);
    await screen.findByRole('heading', { name: 'Unlock wallet' });
    client.onLoginSubmit.mockResolvedValue({ requiresOnboarding: true });
    await act(async () => {
      await ref.current!.onLoginSubmit({ password: 'test-only' });
    });
    expect(
      await screen.findByRole('heading', { name: 'Create or import wallet' }),
    ).toBeVisible();
    expect(dispatch).not.toHaveBeenCalledWith({ type: 'session-started' });
    expect(
      dispatch.mock.calls.some(([action]) => action.type === 'open-wallet'),
    ).toBe(false);
  });

  it.each(['rejection', 'response'])(
    'propagates setup %s failures so the form can retain inputs and retry',
    async (kind) => {
      const { client, dispatch, ref } = fixture(false);
      await screen.findByRole('heading', { name: 'Create or import wallet' });
      if (kind === 'rejection')
        client.onOnboardingCompleted.mockRejectedValue(
          new Error('Test setup failed'),
        );
      else
        client.onOnboardingCompleted.mockResolvedValue({
          error: { message: 'Test setup failed' },
        });
      await expect(
        ref.current!.onOnboardingCompleted({ password: 'test-only' }),
      ).rejects.toThrow('Test setup failed');
      expect(dispatch).not.toHaveBeenCalledWith({ type: 'session-started' });
      client.onOnboardingCompleted.mockResolvedValue(undefined);
      await act(async () => {
        await ref.current!.onOnboardingCompleted({ password: 'test-only' });
      });
      await waitFor(() =>
        expect(dispatch).toHaveBeenCalledWith({ type: 'session-started' }),
      );
    },
  );
});
