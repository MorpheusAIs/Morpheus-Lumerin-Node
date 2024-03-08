import selectors from '../selectors';
import { connect } from 'react-redux';
import * as utils from '../utils';
import PropTypes from 'prop-types';
import React from 'react';

const withContractsRowState = WrappedComponent => {
  class Container extends React.Component {
    static displayName = `withContractsRowState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    render() {
      return <WrappedComponent {...this.props} {...this.state} />;
    }
  }

  const mapStateToProps = (state, props) => ({
    explorerUrl: selectors.getContractExplorerUrl(state, {
      hash: props.contract.id
    }),
    selectedCurrency: selectors.getSellerSelectedCurrency(state),
    networkDifficulty: selectors.getNetworkDifficulty(state),
    lmrRate: selectors.getRate(state),
    btcRate: selectors.getRateBtc(state),
    symbol: selectors.getCoinSymbol(state)
  });

  return connect(mapStateToProps)(Container);
};

export default withContractsRowState;
