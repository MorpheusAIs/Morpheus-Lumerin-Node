import { ThemeProvider as StyledThemeProvider } from 'styled-components';

// Cast: styled-components v4 ships React 16/17-era class component typings that
// React 18's stricter `JSX.LibraryManagedAttributes` resolution rejects. Until
// styled-components is upgraded to v6 (or the project drops v4), narrow it to a
// FC so TSC can use it. Runtime behavior is unchanged.
const ThemeProvider = StyledThemeProvider as unknown as React.FC<
  React.PropsWithChildren<{ theme: object }>
>;

import theme from './ui/theme';
import Root from './components/common/Root';
import { Provider as ClientProvider } from './store/hocs/clientContext';
import { Provider, createStore } from './store/store';

import createClient from './client';
import { subscribeToMainProcessMessages } from './subscriptions';

import Web3ConnectionNotifier from './components/Web3ConnectionNotifier';
import { ToastsProvider } from './components/toasts';
import { GlobalTooltips } from './components/common';
import Onboarding from './components/onboarding/Onboarding';
import Loading from './components/Loading';
import Router from './components/Router';
import Login from './components/Login';
import Startup from '@renderer/components/Startup';

const client = createClient(createStore);

// Initialize all the Main Process subscriptions
subscribeToMainProcessMessages(client.store);

function App(): JSX.Element {
  return (
    <>
      <ClientProvider value={client}>
        <Provider store={client.store}>
          <ThemeProvider theme={theme}>
            <ToastsProvider>
              <Root
                StartupComponent={Startup}
                OnboardingComponent={Onboarding}
                LoadingComponent={Loading}
                RouterComponent={Router}
                LoginComponent={Login}
              />
              <GlobalTooltips />
              <Web3ConnectionNotifier />
            </ToastsProvider>
          </ThemeProvider>
        </Provider>
      </ClientProvider>
    </>
  );
}

export default App;
