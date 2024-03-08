import React from 'react';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';

import { withClient } from './clientContext';
import selectors from '../selectors';
import * as utils from '../utils';

const withReceiptState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      confirmations: PropTypes.number.isRequired,
      explorerUrl: PropTypes.string.isRequired,
      coinSymbol: PropTypes.string.isRequired,
      address: PropTypes.string.isRequired,
      client: PropTypes.shape({
        refreshTransaction: PropTypes.func.isRequired,
        copyToClipboard: PropTypes.func.isRequired,
        onLinkClick: PropTypes.func.isRequired
      }).isRequired,
      hash: PropTypes.string,
      tx: PropTypes.object.isRequired
    };

    static displayName = `withReceiptState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      refreshStatus: 'init',
      refreshError: null
    };

    copyToClipboard = address => {
      this.props.client
        .copyToClipboard(address)
        .then(() => {})
        .catch(() => {});
    };

    onRefreshRequest = () => {
      this.setState({ refreshStatus: 'pending', refreshError: null });
      this.props.client
        .refreshTransaction({
          address: this.props.address,
          chain: this.props.chain,
          hash: this.props.hash
        })
        .then(() => this.setState({ refreshStatus: 'success' }))
        .catch(() =>
          this.setState({
            refreshStatus: 'failure',
            refreshError: 'Could not refresh'
          })
        );
    };

    onExplorerLinkClick = () =>
      this.props.client.onLinkClick(this.props.explorerUrl);

    render() {
      return (
        <WrappedComponent
          onExplorerLinkClick={this.onExplorerLinkClick}
          onRefreshRequest={this.onRefreshRequest}
          copyToClipboard={this.copyToClipboard}
          isPending={utils.isPending(this.props.tx, this.props.confirmations)}
          // isFailed={utils.isFailed(this.props.tx, this.props.confirmations)}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = (state, { hash }) => {
    const tx = selectors.getTransactionFromHash(state, { hash }) || {};

    return {
      confirmations: selectors.getTxConfirmations(state, { tx }),
      explorerUrl: selectors.getTransactionExplorerUrl(state, { hash }),
      coinSymbol: selectors.getCoinSymbol(state),
      address: selectors.getWalletAddress(state),
      tx
    };
  };

  return withClient(connect(mapStateToProps)(Container));
};

export default withReceiptState;
