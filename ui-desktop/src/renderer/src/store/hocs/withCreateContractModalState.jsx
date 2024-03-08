import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';

const withCreateContractModalState = WrappedComponent => {
  class Container extends React.Component {
    // static propTypes = {
    //   address: PropTypes.string.isRequired,
    //   client: PropTypes.shape({
    //     copyToClipboard: PropTypes.func.isRequired
    //   }).isRequired
    // }

    static displayName = `withCreateContractModalState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      copyBtnLabel: ''
    };

    render() {
      return (
        <WrappedComponent
          copyToClipboard={this.props.client.copyToClipboard}
          isProxyPortPublic={this.props.client.isProxyPortPublic}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = (state, props) => ({
    address: selectors.getWalletAddress(state),
    buyerPort: selectors.getBuyerProxyPort(state),
    lmrRate: selectors.getRate(state),
    isLocalProxyRouter: selectors.getIsLocalProxyRouter(state),
    ip: selectors.getIp(state),
    pool: selectors.getBuyerPool(state),
    btcRate: selectors.getRateBtc(state),
    lmrRate: selectors.getRate(state),
    symbol: selectors.getCoinSymbol(state),
    marketplaceFee: selectors.getMarketplaceFee(state),
    explorerUrl: props.contract
      ? selectors.getContractExplorerUrl(state, {
          hash: props.contract.id
        })
      : null,
    portCheckErrorLink: selectors.getPortCheckErrorLink(state)
  });

  return connect(mapStateToProps)(withClient(Container));
};

export default withCreateContractModalState;
