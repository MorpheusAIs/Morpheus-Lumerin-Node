import * as validators from '../validators';
import { withClient } from './clientContext';
import * as utils from '../utils';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';

const withToolsState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      client: PropTypes.shape({
        recoverFromMnemonic: PropTypes.func.isRequired,
        isValidMnemonic: PropTypes.func.isRequired,
        clearCache: PropTypes.func.isRequired
      }).isRequired
    };

    static contextType = ToastsContext;

    static displayName = `withToolsState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      mnemonic: null,
      privateKey: null,
      errors: {},
      hasStoredSecretPhrase: false
    };

    componentDidMount() {
      this.props.client.hasStoredSecretPhrase().then(value => {
        this.setState({ ...this.state, hasStoredSecretPhrase: value });
      });
    }

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

    onShowMnemonic = password => {
      this.props.client
        .revealSecretPhrase(password)
        .then(value => {
          this.setState({ ...this.state, mnemonic: value });
        })
        .catch(e => {
          this.context.toast('error', e.message);
        });
    };

    onExportPrivateKey = password => {
      this.props.client
        .getPrivateKey({ password })
        .then(({ privateKey }) => {
          this.setState({ ...this.state, privateKey: privateKey });
        })
        .catch(e => {
          this.context.toast('error', e.message);
        });
    };

    discardMnemonic = () => this.setState({ ...this.state, mnemonic: null });
    discardPrivateKey = () =>
      this.setState({ ...this.state, privateKey: null });

    onSubmit = password =>
      this.props.client.recoverFromMnemonic({
        mnemonic: utils.sanitizeMnemonic(this.state.mnemonic),
        password
      });

    validate = () => {
      const errors = {
        ...validators.validateMnemonic(this.props.client, this.state.mnemonic)
      };
      const hasErrors = Object.keys(errors).length > 0;
      if (hasErrors) this.setState({ errors });
      return !hasErrors;
    };

    onRescanTransactions = e => {
      if (e && e.preventDefault) e.preventDefault();
      this.props.client.clearCache();
    };

    onRunTest = e => {
      if (e && e.preventDefault) e.preventDefault();
      this.props.client.clearCache();
      console.log('RUN TEST');
    };

    logout = () => {
      return this.props.client
        .stopProxyRouter({})
        .then(() => this.props.client.logout());
    };

    setDefaultCurrency = async value => {
      await this.props.client.setDefaultCurrencySetting(value);
      this.context.toast('success', 'Changed default currency to ' + value);
      this.props.setSellerDefaultCurrency(value);
    };

    render() {
      const isRecoverEnabled =
        utils.sanitizeMnemonic(this.state.mnemonic || '').split(' ').length ===
        12;

      return (
        <WrappedComponent
          onRescanTransactions={this.onRescanTransactions}
          onRunTest={this.onRunTest}
          isRecoverEnabled={isRecoverEnabled}
          onInputChange={this.onInputChange}
          onSubmit={this.onSubmit}
          onShowMnemonic={this.onShowMnemonic}
          onExportPrivateKey={this.onExportPrivateKey}
          discardMnemonic={this.discardMnemonic}
          discardPrivateKey={this.discardPrivateKey}
          validate={this.validate}
          getDefaultCurrency={this.getDefaultCurrency}
          setDefaultCurrency={this.setDefaultCurrency}
          copyToClipboard={this.props.client.copyToClipboard}
          logout={this.logout}
          restartWallet={this.props.client.restartWallet}
          onRevealPhrase={this.props.client.revealSecretPhrase}
          getProxyRouterSettings={this.props.client.getProxyRouterSettings}
          saveProxyRouterSettings={this.props.client.saveProxyRouterSettings}
          restartProxyRouter={this.props.client.restartProxyRouter}
          getCustomEnvs={this.props.client.getCustomEnvValues}
          setCustomEnvs={this.props.client.setCustomEnvValues}
          getProfitSettings={this.props.client.getProfitSettings}
          setProfitSettings={this.props.client.setProfitSettings}
          {...this.state}
          {...this.props}
        />
      );
    }
  }

  const mapStateToProps = (state, props) => ({
    selectedCurrency: selectors.getSellerSelectedCurrency(state),
    isLocalProxyRouter: selectors.getIsLocalProxyRouter(state),
    titanLightningPool: state.config.chain.titanLightningPool,
    titanLightningDashboard: state.config.chain.titanLightningDashboard,
    config: state.config
  });

  const mapDispatchToProps = dispatch => ({
    setSellerDefaultCurrency: currency =>
      dispatch({ type: 'set-seller-currency', payload: currency })
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withToolsState;
