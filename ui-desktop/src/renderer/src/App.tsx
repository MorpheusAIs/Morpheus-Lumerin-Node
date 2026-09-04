import { lazy, Suspense } from 'react';
import styled, {
  ThemeProvider as StyledThemeProvider,
} from 'styled-components';

// Cast: styled-components v4 ships React 16/17-era class component typings that
// React 18's stricter `JSX.LibraryManagedAttributes` resolution rejects. Until
// styled-components is upgraded to v6 (or the project drops v4), narrow it to a
// FC so TSC can use it. Runtime behavior is unchanged.
const ThemeProvider = StyledThemeProvider as unknown as React.FC<
  React.PropsWithChildren<{ theme: object }>
>;

import { QueryClientProvider } from '@tanstack/react-query';

import theme from './ui/theme';
import Root from './components/common/Root';
import { Provider as ClientProvider } from './store/hocs/clientContext';
import { Provider, createStore } from './store/store';
import { queryClient } from './store/queryClient';

import createClient from './client';
import { subscribeToMainProcessMessages } from './subscriptions';

import Web3ConnectionNotifier from './components/Web3ConnectionNotifier';
import { ToastsProvider } from './components/toasts';
import { GlobalTooltips } from './components/common/Tooltips';
import Loading from './components/Loading';
import ErrorBoundary from './components/common/ErrorBoundary';

const Startup = lazy(() => import('@renderer/components/Startup'));
const Onboarding = lazy(() => import('./components/onboarding/Onboarding'));
const Router = lazy(() => import('./components/Router'));
const Login = lazy(() => import('./components/Login'));

const ShellLoading = styled.div`
  align-items: center;
  background: #04130d;
  color: rgba(255, 255, 255, 0.62);
  display: flex;
  font-family: 'Roboto Mono', monospace;
  font-size: 1.2rem;
  height: 100vh;
  justify-content: center;
`;

const client = createClient(createStore);

// Initialize all the Main Process subscriptions
subscribeToMainProcessMessages(client.store);

function App(): JSX.Element {
  return (
    <>
      <ClientProvider value={client}>
        <Provider store={client.store}>
          <QueryClientProvider client={queryClient}>
            <ThemeProvider theme={theme}>
              <ToastsProvider>
                <ErrorBoundary resetKey="app-shell" label="app-shell">
                  <Suspense
                    fallback={
                      <ShellLoading role="status" aria-live="polite">
                        Loading Morpheus…
                      </ShellLoading>
                    }
                  >
                    <Root
                      StartupComponent={Startup}
                      OnboardingComponent={Onboarding}
                      LoadingComponent={Loading}
                      RouterComponent={Router}
                      LoginComponent={Login}
                    />
                  </Suspense>
                </ErrorBoundary>
                <GlobalTooltips />
                <Web3ConnectionNotifier />
              </ToastsProvider>
            </ThemeProvider>
          </QueryClientProvider>
        </Provider>
      </ClientProvider>
    </>
  );
}

export default App;
