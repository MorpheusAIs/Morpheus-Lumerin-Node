import React from 'react';
import * as validators from '../validators';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import * as utils from '../utils';
import { toRfc2396, generatePoolUrl } from '../../utils';

const EMAIL_REGEX =
  /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
const UrlRegex = /\b(?:http|ws)s?:\/\/\S*[^\s."]/;

const withOnboardingState = (WrappedComponent) => {
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

    static displayName = `withOnboardingState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    state = {
      walletMode: null,
      isPreparingWallet: false,
      isSubmitting: false,
      setupError: '',
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
      userPrivateKey: null,
      derivationPath: null,
      customEthNode: null,
      password: null,
      mnemonic: null,
      useImportFlow: false,
      useEthStep: false,
      errors: {},
    };

    preparationId = 0;
    submitting = false;

    componentWillUnmount() {
      this.preparationId++;
      this.unmounted = true;
    }

    onWalletModeSelected = (walletMode) => {
      if (
        this.submitting ||
        !this.state.areTermsAccepted ||
        !['create', 'import'].includes(walletMode)
      )
        return;
      this.preparationId++;
      this.setState({
        walletMode,
        useImportFlow: walletMode === 'import',
        isPasswordDefined: false,
        isPreparingWallet: false,
        isMnemonicCopied: false,
        isMnemonicVerified: false,
        useUserMnemonic: false,
        useEthStep: false,
        password: null,
        passwordAgain: null,
        mnemonic: null,
        mnemonicAgain: null,
        userMnemonic: null,
        userPrivateKey: null,
        derivationPath: null,
        errors: {},
        setupError: '',
      });
    };

    onChooseWallet = () => {
      if (this.submitting) return;
      this.preparationId++;
      this.setState({
        walletMode: null,
        isPreparingWallet: false,
        setupError: '',
      });
    };

    onTermsAccepted = () => {
      if (this.state.licenseCheckbox && this.state.termsCheckbox) {
        this.setState({ areTermsAccepted: true });
      }
    };

    onPasswordSubmit = async ({ clearOnError = false } = {}) => {
      if (
        !this.state.walletMode ||
        !this.state.areTermsAccepted ||
        this.state.isPreparingWallet
      )
        return;
      const { password, passwordAgain } = this.state;

      const errors = validators.validatePasswordCreation(
        this.props.client,
        this.props.config,
        password,
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
          errors,
        });
        return;
      }
      if (this.state.useImportFlow) {
        this.setState({ isPasswordDefined: true, errors: {}, setupError: '' });
        return;
      }
      const preparationId = ++this.preparationId;
      this.setState({ isPreparingWallet: true, errors: {}, setupError: '' });
      try {
        const mnemonic = await this.props.client.createMnemonic();
        if (this.unmounted || preparationId !== this.preparationId) return;
        if (!mnemonic || !this.props.client.isValidMnemonic(mnemonic)) {
          throw new Error('Invalid generated phrase');
        }
        this.setState({
          mnemonic,
          isPasswordDefined: true,
          isPreparingWallet: false,
        });
      } catch {
        if (this.unmounted || preparationId !== this.preparationId) return;
        this.setState({
          isPreparingWallet: false,
          setupError:
            'Could not generate a recovery phrase. Your password is still here; try again.',
        });
      }
    };

    onUseUserMnemonicToggled = () => {
      this.setState((state) => ({
        ...state,
        useUserMnemonic: !state.useUserMnemonic,
        userMnemonic: null,
        errors: {
          ...state.errors,
          userMnemonic: null,
        },
      }));
    };

    onMnemonicCopiedToggled = () => {
      this.setState((state) => ({
        ...state,
        isMnemonicCopied: !state.isMnemonicCopied,
        mnemonicAgain: null,
        errors: {
          ...state.errors,
          mnemonicAgain: null,
        },
      }));
    };

    onMnemonicAccepted = (e) => {
      if (e && e.preventDefault) e.preventDefault();

      const errors = this.state.useUserMnemonic
        ? validators.validateMnemonic(
            this.props.client,
            this.state.userMnemonic,
            'userMnemonic',
          )
        : validators.validateMnemonicAgain(
            this.props.client,
            this.state.mnemonic,
            this.state.mnemonicAgain,
          );

      if (Object.keys(errors).length > 0) return this.setState({ errors });

      return this.onFinishOnboarding();
    };

    onPrivateKeyAccepted = (e) => {
      if (e && e.preventDefault) e.preventDefault();

      if (
        !/^(0x)?[0-9a-fA-F]{64}$/.test((this.state.userPrivateKey || '').trim())
      ) {
        this.setState({
          errors: {
            userPrivateKey:
              'Enter a valid private key (64 hexadecimal characters).',
          },
        });
        return;
      }
      this.setState({
        userPrivateKey: this.state.userPrivateKey.trim(),
        useUserMnemonic: false,
        userMnemonic: null,
        useEthStep: true,
        errors: {},
      });
    };

    validateDefaultPoolAddress() {
      const errors = validators.validatePoolAddress(
        this.state.proxyDefaultPool,
      );
      validators.validatePoolUsername(this.state.proxyPoolUsername, errors);
      if (errors.proxyDefaultPool || errors.proxyPoolUsername) {
        this.setState({ errors });
        return false;
      }
      return true;
    }

    onFinishOnboarding = async (
      e,
      ethNode = this.state.customEthNode || '',
    ) => {
      if (e && e.preventDefault) e.preventDefault();
      if (
        this.submitting ||
        !this.state.areTermsAccepted ||
        !this.state.isPasswordDefined
      )
        return;
      this.submitting = true;
      this.setState({ isSubmitting: true, setupError: '' });

      const payload = {
        password: this.state.password,
        ethNode,
        derivationPath: this.state.derivationPath || '0',
        privateKey: '',
      };

      if (this.state.userPrivateKey) {
        payload.privateKey = this.state.userPrivateKey;
      } else {
        payload.mnemonic = this.state.useUserMnemonic
          ? utils.sanitizeMnemonic(this.state.userMnemonic)
          : this.state.mnemonic;
      }

      try {
        await this.props.onOnboardingCompleted(payload);
      } catch (error) {
        if (!this.unmounted)
          this.setState({
            setupError:
              error?.message ||
              'Could not finish wallet setup. Your details are still here; try again.',
          });
      } finally {
        this.submitting = false;
        if (!this.unmounted) this.setState({ isSubmitting: false });
      }
    };

    onRunWithoutProxyRouter = (e) => {
      return this.props.onOnboardingCompleted({
        proxyRouterConfig: {
          runWithoutProxyRouter: true,
        },
        password: this.state.password,
        mnemonic: this.state.useUserMnemonic
          ? utils.sanitizeMnemonic(this.state.userMnemonic)
          : this.state.mnemonic,
      });
    };

    onSuggestAddress = async () => {
      const errors = validators.validateMnemonic(
        this.props.client,
        this.state.userMnemonic,
        'userMnemonic',
      );
      if (Object.keys(errors).length) {
        this.setState({ errors });
        throw new Error(
          'Enter a valid recovery phrase before selecting an address.',
        );
      }
      const addrs = await this.props.client.suggestAddresses(
        this.state.userMnemonic,
      );
      return addrs;
    };

    onMnemonicSet = async (e, path) => {
      if (e && e.preventDefault) e.preventDefault();

      this.setState({
        useUserMnemonic: true,
        userPrivateKey: null,
        derivationPath: path,
        useEthStep: true,
        errors: {},
      });
    };

    onEthNodeSet = async (e, useDefault = false) => {
      if (e && e.preventDefault) e.preventDefault();
      if (useDefault) return this.onFinishOnboarding(e, '');
      if (
        this.state.customEthNode &&
        !UrlRegex.test(this.state.customEthNode)
      ) {
        const errors = this.state.errors;
        errors.customEthNode = 'Url format is not valid';
        this.setState({ errors });
      } else {
        await this.onFinishOnboarding(e);
      }
    };

    onInputChange = ({ id, value }) => {
      this.setState((state) => ({
        ...state,
        [id]: value,
        errors: {
          ...state.errors,
          [id]: null,
        },
      }));
    };

    getCurrentStep() {
      if (!this.state.areTermsAccepted) return 'ask-for-terms';
      if (!this.state.walletMode) return 'choose-wallet';
      if (!this.state.isPasswordDefined) return 'define-password';
      if (this.state.useEthStep) return 'set-custom-eth';
      if (this.state.useUserMnemonic) return 'recover-from-mnemonic';
      if (this.state.isMnemonicCopied) return 'verify-mnemonic';
      if (this.state.useImportFlow) return 'import-flow';

      return 'copy-mnemonic';
    }

    render() {
      const getWordsAmount = (phrase) =>
        utils.sanitizeMnemonic(phrase || '').split(' ').length;

      const shouldSubmit = (phrase) => getWordsAmount(phrase) === 12;

      const getTooltip = (phrase) =>
        shouldSubmit(phrase)
          ? null
          : 'A recovery phrase must have exactly 12 words';

      return (
        <WrappedComponent
          onWalletModeSelected={this.onWalletModeSelected}
          onChooseWallet={this.onChooseWallet}
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
          onSuggestAddress={this.onSuggestAddress}
          onRunWithoutProxyRouter={this.onRunWithoutProxyRouter}
          onPrivateKeyAccepted={this.onPrivateKeyAccepted}
          onMnemonicSet={this.onMnemonicSet}
          onEthNodeSet={this.onEthNodeSet}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = (state) => ({
    config: selectors.getConfig(state),
  });

  return connect(mapStateToProps)(withClient(Container));
};

export default withOnboardingState;
