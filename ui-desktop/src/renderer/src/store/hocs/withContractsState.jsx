import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';
import { ToastsContext } from '../../components/toasts';

const withContractsState = WrappedComponent => {
  class Container extends React.Component {
    // static propTypes = {
    //   sendLmrFeatureStatus: PropTypes.oneOf(['offline', 'no-funds', 'ok'])
    //     .isRequired,
    //   syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed'])
    //     .isRequired,
    //   address: PropTypes.string.isRequired,
    //   client: PropTypes.shape({
    //     refreshAllContracts: PropTypes.func.isRequired,
    //     copyToClipboard: PropTypes.func.isRequired
    //   }).isRequired
    // }

    static contextType = ToastsContext;

    static displayName = `withContractsState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      refreshStatus: 'init',
      refreshError: null
    };

    contractsRefresh = (force = false) => {
      const now = parseInt(Date.now() / 1000, 10);
      const timeout = 15; // seconds
      if (
        this.props.contractsLastUpdatedAt &&
        now - this.props.contractsLastUpdatedAt < timeout &&
        !force
      ) {
        this.props.setPendingRefresh();
        setTimeout(() => {
          this.props.setFinishedRefresh();
        }, 200);
        return;
      }
      this.setState({ refreshStatus: 'pending', refreshError: null });
      this.props.client
        .refreshAllContracts({})
        .then(() => this.setState({ refreshStatus: 'success' }))
        .catch(e => {
          this.context.toast(
            'error',
            'We’re experiencing connection problems.  Please wait a few minutes and try again'
          );
          this.props.setFailedRefresh();
        });
    };

    render() {
      return (
        <WrappedComponent
          copyToClipboard={this.props.client.copyToClipboard}
          contractsRefresh={this.contractsRefresh}
          getLocalIp={this.props.client.getLocalIp}
          getPoolAddress={this.props.client.getPoolAddress}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    hasContracts: selectors.hasContracts(state),
    activeCount: selectors.getActiveContractsCount(state),
    draftCount: selectors.getDraftContractsCount(state),
    syncStatus: selectors.getContractsSyncStatus(state),
    address: selectors.getWalletAddress(state),
    contracts: selectors.getMergeAllContracts(state),
    lmrBalance: selectors.getWalletLmrBalance(state),
    allowSendTransaction: selectors.isAllowSendTransaction(state),
    contractsLastUpdatedAt: selectors.getContractsLastUpdated(state),
    networkDifficulty: selectors.getNetworkDifficulty(state),
    lmrCoinPrice: selectors.getRate(state),
    ethCoinPrice: selectors.getRateEth(state),
    btcCoinPrice: selectors.getRateBtc(state),
    selectedCurrency: selectors.getSellerSelectedCurrency(state),
    formUrl: selectors.getSellerWhitelistForm(state),
    autoAdjustPriceInterval: selectors.getAutoAdjustPriceInterval(state),
    autoAdjustContractPriceTimeout: selectors.getAutoAdjustContractPriceTimeout(
      state
    )
  });

  const mapDispatchToProps = dispatch => ({
    setIp: ip => dispatch({ type: 'ip-received', payload: ip }),
    setFailedRefresh: () => dispatch({ type: 'contracts-scan-failed' }),
    setPendingRefresh: () => dispatch({ type: 'contracts-scan-started' }),
    setFinishedRefresh: () =>
      dispatch({ type: 'contracts-scan-finished', payload: {} }),
    setDefaultBuyerPool: pool =>
      dispatch({ type: 'buyer-default-pool-received', payload: pool })
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withContractsState;
