import React from 'react';
import * as validators from '../validators';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import * as utils from '../utils';
import { toRfc2396, generatePoolUrl } from '../../utils';

const EMAIL_REGEX = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

const withOnboardingState = WrappedComponent => {
  class Container extends React.Component {
    // static propTypes = {
    //   onOnboardingCompleted: PropTypes.func.isRequired,
    //   client: PropTypes.shape({
    //     onTermsLinkClick: PropTypes.func.isRequired,
    //     getStringEntropy: PropTypes.func.isRequired,
    //     isValidMnemonic: PropTypes.func.isRequired,
    //     createMnemonic: PropTypes.func.isRequired
    //   }).isRequired,
    //   config: PropTypes.shape({
    //   }).isRequired
    // };

    static displayName = `withOnboardingState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      isPasswordDefined: false,
      areTermsAccepted: false,
      isMnemonicCopied: false,
      useUserMnemonic: false,
      isMnemonicVerified: false,
      licenseCheckbox: false,
      termsCheckbox: false,
      passwordAgain: null,
      mnemonicAgain: null,
      userMnemonic: null,
      password: null,
      mnemonic: null,
      proxyDefaultPool: null,
      lightningAddress: null,
      isTitanLightning: true,
      errors: {}
    };

    componentDidMount() {
      this.props.client
        .createMnemonic()
        .then(mnemonic => this.setState({ mnemonic }))
        // eslint-disable-next-line no-console
        .catch(() => console.warn("Couldn't create mnemonic"));
    }

    onTermsAccepted = () => {
      if (this.state.licenseCheckbox && this.state.termsCheckbox) {
        this.setState({ areTermsAccepted: true });
      }
    };

    onPasswordSubmit = ({ clearOnError = false }) => {
      const { password, passwordAgain } = this.state;

      const errors = validators.validatePasswordCreation(
        this.props.client,
        this.props.config,
        password
      );
      if (!errors.password && !passwordAgain) {
        errors.passwordAgain = `Repeat the ${
          clearOnError ? 'PIN' : 'password'
        }`;
      } else if (!errors.password && passwordAgain !== password) {
        errors.passwordAgain = `${
          clearOnError ? 'PINs' : 'Passwords'
        } don't match`;
      }
      if (Object.keys(errors).length > 0) {
        this.setState({
          passwordAgain: clearOnError ? '' : passwordAgain,
          errors
        });
        return;
      }
      this.setState({ isPasswordDefined: true });
    };

    onUseUserMnemonicToggled = () => {
      this.setState(state => ({
        ...state,
        useUserMnemonic: !state.useUserMnemonic,
        userMnemonic: null,
        errors: {
          ...state.errors,
          userMnemonic: null
        }
      }));
    };

    onMnemonicCopiedToggled = () => {
      this.setState(state => ({
        ...state,
        isMnemonicCopied: !state.isMnemonicCopied,
        mnemonicAgain: null,
        errors: {
          ...state.errors,
          mnemonicAgain: null
        }
      }));
    };

    onMnemonicAccepted = e => {
      if (e && e.preventDefault) e.preventDefault();

      const errors = this.state.useUserMnemonic
        ? validators.validateMnemonic(
            this.props.client,
            this.state.userMnemonic,
            'userMnemonic'
          )
        : validators.validateMnemonicAgain(
            this.props.client,
            this.state.mnemonic,
            this.state.mnemonicAgain
          );

      if (Object.keys(errors).length > 0) return this.setState({ errors });

      this.setState({ isMnemonicVerified: true });
      onFinishOnboarding();
    };

    validateDefaultPoolAddress() {
      const errors = validators.validatePoolAddress(
        this.state.proxyDefaultPool
      );
      validators.validatePoolUsername(this.state.proxyPoolUsername, errors);
      if (errors.proxyDefaultPool || errors.proxyPoolUsername) {
        this.setState({ errors });
        return false;
      }
      return true;
    }

    // HERE
    onFinishOnboarding = e => {
      if (e && e.preventDefault) e.preventDefault();

      return this.props.onOnboardingCompleted({
        password: this.state.password,
        mnemonic: this.state.useUserMnemonic
          ? utils.sanitizeMnemonic(this.state.userMnemonic)
          : this.state.mnemonic
      });
    };

    onRunWithoutProxyRouter = e => {
      return this.props.onOnboardingCompleted({
        proxyRouterConfig: {
          runWithoutProxyRouter: true
        },
        password: this.state.password,
        mnemonic: this.state.useUserMnemonic
          ? utils.sanitizeMnemonic(this.state.userMnemonic)
          : this.state.mnemonic
      });
    };

    onInputChange = ({ id, value }) => {
      this.setState(state => ({
        ...state,
        [id]: value,
        errors: {
          ...state.errors,
          [id]: null
        }
      }));
    };

    getCurrentStep() {
      if (!this.state.areTermsAccepted) return 'ask-for-terms';
      if (!this.state.isPasswordDefined) return 'define-password';
      if (this.state.isMnemonicVerified) return 'config-proxy-router';
      if (this.state.useUserMnemonic) return 'recover-from-mnemonic';
      if (this.state.isMnemonicCopied) return 'verify-mnemonic';

      return 'copy-mnemonic';
    }

    render() {
      const getWordsAmount = phrase =>
        utils.sanitizeMnemonic(phrase || '').split(' ').length;

      const shouldSubmit = phrase => getWordsAmount(phrase) === 12;

      const getTooltip = phrase =>
        shouldSubmit(phrase)
          ? null
          : 'A recovery phrase must have exactly 12 words';

      return (
        <WrappedComponent
          onUseUserMnemonicToggled={this.onUseUserMnemonicToggled}
          onMnemonicCopiedToggled={this.onMnemonicCopiedToggled}
          onMnemonicAccepted={this.onMnemonicAccepted}
          onTermsLinkClick={this.props.client.onTermsLinkClick}
          onPasswordSubmit={this.onPasswordSubmit}
          onTermsAccepted={this.onTermsAccepted}
          onInputChange={this.onInputChange}
          shouldSubmit={shouldSubmit}
          currentStep={this.getCurrentStep()}
          getTooltip={getTooltip}
          onFinishOnboarding={this.onFinishOnboarding}
          onRunWithoutProxyRouter={this.onRunWithoutProxyRouter}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    config: selectors.getConfig(state)
  });

  return connect(mapStateToProps)(withClient(Container));
};

export default withOnboardingState;
