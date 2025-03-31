import React from 'react';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import { ToastsContext } from '../../components/toasts';
import { getSessionsByUser } from '../utils/apiCallsHelper';

const withDashboardState = (WrappedComponent) => {
  class Container extends React.Component {
    static propTypes = {
      sendLmrFeatureStatus: PropTypes.oneOf(['offline', 'no-funds', 'ok'])
        .isRequired,
      syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed'])
        .isRequired,
      address: PropTypes.string.isRequired,
      client: PropTypes.shape({
        refreshAllTransactions: PropTypes.func.isRequired,
        copyToClipboard: PropTypes.func.isRequired,
      }).isRequired,
    };

    static contextType = ToastsContext;

    static displayName = `withDashboardState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    state = {
      refreshStatus: 'init',
      refreshError: null,
    };

    onWalletRefresh = () => {
      this.setState({ refreshStatus: 'pending', refreshError: null });
      this.loadTransactions();
    };

    loadTransactions = async (page = 1, pageSize = 15) => {
      this.setState({ refreshStatus: 'pending', refreshError: null });
      const transactions = await this.props.client.getTransactions({
        page,
        pageSize,
      });
      this.setState({ refreshStatus: 'success' });

      // if (page && pageSize) {
      //   const hasNextPage = transactions.length;
      //   this.props.nextPage({
      //     hasNextPage: Boolean(hasNextPage),
      //     page: page + 1,
      //   })
      // }
      return transactions;
    };

    getStakedFunds = async (user) => {
      const isClosed = (item) =>
        item.ClosedAt || new Date().getTime() > item.EndsAt * 1000;

      if (!user) {
        return;
      }

      const authHeaders = await this.props.client.getAuthHeaders();
      const sessions = await getSessionsByUser(
        this.props.config.chain.localProxyRouterUrl,
        user,
        authHeaders,
      );

      try {
        const openSessions = sessions.filter((s) => !isClosed(s));
        const sum = openSessions.reduce((curr, next) => curr + next.Stake, 0);
        return (sum / 10 ** 18).toFixed(2);
      } catch (e) {
        console.log('Error', e);
        return 0;
      }
    };

    getBalances = async () => {
      const balances = await this.props.client.getBalances();
      const rate = await this.props.client.getRates();
      return { balances, rate };
    };

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
          getBalances={this.getBalances}
          sendDisabled={sendLmrFeatureStatus !== 'ok'}
          loadTransactions={this.loadTransactions}
          getStakedFunds={this.getStakedFunds}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = (state) => ({
    config: state.config,
    syncStatus: selectors.getTxSyncStatus(state),
    sendLmrFeatureStatus: selectors.sendLmrFeatureStatus(state),
    hasTransactions: selectors.hasTransactions(state),
    address: selectors.getWalletAddress(state),
    ethCoinPrice: selectors.getRateEth(state),
    symbol: selectors.getCoinSymbol(state),
    symbolEth: selectors.getSymbolEth(state),
    page: selectors.getTransactionPage(state),
    pageSize: selectors.getTransactionPageSize(state),
    hasNextPage: selectors.getHasNextPage(state),
    explorerUrl: selectors.getContractExplorerUrl(state, {
      hash: selectors.getWalletAddress(state),
    }),
  });

  const mapDispatchToProps = (dispatch) => ({
    nextPage: (data) =>
      dispatch({ type: 'transactions-next-page', payload: data }),
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withDashboardState;
