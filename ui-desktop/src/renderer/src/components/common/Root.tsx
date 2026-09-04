import { withClient } from '../../store/hocs/clientContext';
import selectors from '../../store/selectors';
import { connect } from 'react-redux';
import React from 'react';
import { ToastsContext } from '../toasts';
import { LoadingState } from 'src/main/orchestrator.types';

type RootProps = {
  // From connect() mapStateToProps
  isSessionActive: boolean;
  isAuthBypassed: boolean;
  sellerDefaultCurrency: string;
  servicesState: LoadingState;
  config: any;
  // From connect()
  dispatch: (action: { type: string; payload?: any }) => void;
  // From withClient()
  client: any;
  // Render components passed from parent
  StartupComponent: React.ComponentType<{ onSkip: () => void }>;
  OnboardingComponent: React.ComponentType<{
    onOnboardingCompleted: (data: unknown) => Promise<void>;
  }>;
  RouterComponent: React.ComponentType;
  LoginComponent: React.ComponentType<{
    onLoginSubmit: (data: { password: string }) => Promise<void>;
  }>;
};

export class Root extends React.Component<RootProps> {
  static contextType = ToastsContext;
  declare context: React.ContextType<typeof ToastsContext>;

  state = {
    startupComplete: false,
    onboardingComplete: null,
  };

  startAuthenticatedSession = async (password: string): Promise<void> => {
    // Make the login acknowledgement carry the wallet identity. Relying only
    // on a separate `open-wallet` event allowed late bootstrap hydration to
    // erase it, while a second proxy request would make offline login slow.
    const walletState = await this.props.client.onLoginSubmit({ password });
    const address = walletState?.address;
    if (!address) {
      throw new Error(
        'Wallet login did not return an active address. Please try again.',
      );
    }

    this.props.dispatch({
      type: 'open-wallet',
      payload: { address, isActive: true },
    });
    this.props.dispatch({ type: 'session-started' });
  };

  componentDidMount() {
    this.props.client
      .onInit()
      .then(({ onboardingComplete, persistedState, config }) => {
        this.props.dispatch({
          type: 'initial-state-received',
          payload: { ...persistedState, config },
        });
        this.setState({ onboardingComplete });
      })
      .then(() => {
        if (this.props.isAuthBypassed) {
          // TODO: replace dummy password
          return this.startAuthenticatedSession('password').catch((_e) => {
            this.context.toast('error', 'Bypass auth failed');
          });
        }
        return undefined;
      })
      // The display currency is cosmetic. A missed/failed settings response
      // must not turn into "Failed to startup wallet" after the wallet and
      // proxy-router have already initialized successfully.
      .then(() =>
        this.props.client.getDefaultCurrencySetting().catch((error) => {
          // eslint-disable-next-line no-console
          console.warn(
            'Could not load the saved display currency; using the default.',
            error,
          );
          return null;
        }),
      )
      .then((defaultCurr) => {
        this.props.dispatch({
          type: 'set-seller-currency',
          payload: defaultCurr || this.props.sellerDefaultCurrency || 'BTC',
        });
      })
      // eslint-disable-next-line no-console
      .catch((e) => {
        console.error('root component error', e.message);
        this.context.toast(
          'error',
          'Failed to initialize Morpheus. Your wallet is unchanged; restart the app and try again.',
        );
      });
  }

  componentDidUpdate(): void {
    if (
      this.props.servicesState.orchestratorStatus === 'ready' &&
      !this.state.startupComplete
    ) {
      this.setState({ startupComplete: true });
    }
  }

  onOnboardingCompleted = (data) => {
    return (
      this.props.client
        .onOnboardingCompleted({
          proxyUrl: this.props.config.chain.localProxyRouterUrl,
          ...data,
        })
        .then((error) => {
          if (error) {
            this.context.toast('error', error);
            return;
          }
          this.setState({ onboardingComplete: true });
          this.props.dispatch({ type: 'session-started' });
        })
        // eslint-disable-next-line no-console
        .catch((_e) => {
          this.context.toast(
            'error',
            'Failed to finish onboarding. Please wait a few minutes and try again',
          );
        })
    );
  };

  onLoginSubmit = ({ password }) => this.startAuthenticatedSession(password);

  render() {
    const {
      StartupComponent,
      OnboardingComponent,
      RouterComponent,
      isSessionActive,
      LoginComponent,
    } = this.props;

    const { onboardingComplete, startupComplete } = this.state;

    // return <StartupComponent />;

    if (onboardingComplete === null) return null;

    if (!startupComplete) {
      return (
        <StartupComponent
          onSkip={() => this.setState({ startupComplete: true })}
        />
      );
    }

    if (!onboardingComplete) {
      return (
        <OnboardingComponent
          onOnboardingCompleted={this.onOnboardingCompleted}
        />
      );
    }

    if (!isSessionActive) {
      return <LoginComponent onLoginSubmit={this.onLoginSubmit} />;
    }

    // Service readiness and authentication are the only global gates. Wallet
    // balances, exchange rates, sessions, and model catalogs belong to their
    // individual tabs and load through their own cached queries. Holding the
    // whole application behind those optional network reads made navigation
    // appear frozen whenever one public API was slow or unavailable.
    return <RouterComponent />;
  }
}

const mapStateToProps = (state) => ({
  isSessionActive: selectors.isSessionActive(state),
  isAuthBypassed: selectors.getIsAuthBypassed(state),
  sellerDefaultCurrency: selectors.getSellerDefaultCurrency(state),
  servicesState: selectors.getServices(state),
  config: state.config,
});

export default connect(mapStateToProps)(withClient(Root));
