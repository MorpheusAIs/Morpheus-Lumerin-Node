import React from 'react';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import { ToastsContext } from '../../components/toasts';

const withDashboardState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      sendLmrFeatureStatus: PropTypes.oneOf(['offline', 'no-funds', 'ok'])
        .isRequired,
      syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed'])
        .isRequired,
      address: PropTypes.string.isRequired,
      client: PropTypes.shape({
        refreshAllTransactions: PropTypes.func.isRequired,
        copyToClipboard: PropTypes.func.isRequired
      }).isRequired
    };

    static contextType = ToastsContext;

    static displayName = `withDashboardState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      refreshStatus: 'init',
      refreshError: null
    };

    onWalletRefresh = () => {
      this.setState({ refreshStatus: 'pending', refreshError: null });
      this.props.client
        .refreshAllTransactions({
          address: this.props.address,
          chain: this.props.chain
        })
        .then(() => this.setState({ refreshStatus: 'success' }))
        .catch(() => {
          this.context.toast(
            'error',
            'We’re experiencing connection problems.  Please wait a few minutes and try again'
          );
          this.setState({
            refreshStatus: 'failure',
            refreshError: 'Could not refresh'
          });
        });
    };

    onInit = async () => {
      var balances = await this.props.client.getBalances();
      console.log("🚀 ~ Container ~ getBalances= ~ balances:", balances)
      var transactions = await this.props.client.getTransactions();
      console.log("🚀 ~ Container ~ getBalances= ~ transactions:", transactions)
      var rate = await this.props.client.getMorRate();
      console.log("🚀 ~ Container ~ onInit= ~ rate:", rate)
      return { balances, transactions, rate };
    }

    render() {
      const { sendLmrFeatureStatus } = this.props;

      const sendDisabledReason =
        sendLmrFeatureStatus === 'offline'
          ? "Can't send while offline"
          : sendLmrFeatureStatus === 'no-funds'
          ? 'You need some funds to send'
          : null;

      return (
        <WrappedComponent
          sendDisabledReason={sendDisabledReason}
          copyToClipboard={this.props.client.copyToClipboard}
          onWalletRefresh={this.onWalletRefresh}
          onInit={this.onInit}
          sendDisabled={sendLmrFeatureStatus !== 'ok'}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    syncStatus: selectors.getTxSyncStatus(state),
    sendLmrFeatureStatus: selectors.sendLmrFeatureStatus(state),
    hasTransactions: selectors.hasTransactions(state),
    address: selectors.getWalletAddress(state)
  });

  return withClient(connect(mapStateToProps)(Container));
};

export default withDashboardState;
