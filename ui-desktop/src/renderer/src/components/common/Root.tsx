import { withClient } from '../../store/hocs/clientContext';
import selectors from '../../store/selectors';
import { connect } from 'react-redux';
import React from 'react';
import { ToastsContext } from '../toasts';
import { LoadingState } from 'src/main/orchestrator.types';

class Root extends React.Component<{ servicesState: LoadingState }> {
  static contextType = ToastsContext;
  declare context: React.ContextType<typeof ToastsContext>;

  state = {
    startupComplete: false,
    onboardingComplete: null,
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
          this.props.client
            .onLoginSubmit({ password: 'password' })
            .then(() => this.props.dispatch({ type: 'session-started' }))
            .catch((e) => {
              this.context.toast('error', 'Bypass auth failed');
            });
        }
      })
      .then(() => this.props.client.getDefaultCurrencySetting())
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
          'Failed to startup wallet. Please wait a few minutes and try again',
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
        .catch((e) => {
          this.context.toast(
            'error',
            'Failed to finish onboarding. Please wait a few minutes and try again',
          );
        })
    );
  };

  onLoginSubmit = ({ password }) =>
    this.props.client
      .onLoginSubmit({ password })
      .then(() => this.props.dispatch({ type: 'session-started' }));

  render() {
    const {
      StartupComponent,
      OnboardingComponent,
      LoadingComponent,
      RouterComponent,
      isSessionActive,
      LoginComponent,
      hasEnoughData,
    } = this.props;

    const { onboardingComplete, startupComplete } = this.state;

    // return <StartupComponent />;

    if (onboardingComplete === null) return null;

    if (!startupComplete) {
      return <StartupComponent />;
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

    if (hasEnoughData) {
      return <RouterComponent />;
    }

    return <LoadingComponent />;
  }
}

const mapStateToProps = (state) => ({
  isSessionActive: selectors.isSessionActive(state),
  hasEnoughData: selectors.hasEnoughData(state),
  isAuthBypassed: selectors.getIsAuthBypassed(state),
  sellerDefaultCurrency: selectors.getSellerDefaultCurrency(state),
  servicesState: selectors.getServices(state),
  config: state.config,
});

export default connect(mapStateToProps)(withClient(Root));
