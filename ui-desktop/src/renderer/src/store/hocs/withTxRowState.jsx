import selectors from '../selectors';
import { connect } from 'react-redux';
import * as utils from '../utils';
import PropTypes from 'prop-types';
import React from 'react';

const withTxRowState = (WrappedComponent) => {
  class Container extends React.Component {
    // static propTypes = {
    //   confirmations: PropTypes.number.isRequired,
    //   coinSymbol: PropTypes.string.isRequired,
    //   tx: PropTypes.shape({
    //     contractCallFailed: PropTypes.bool,
    //     txType: PropTypes.string.isRequired
    //   }).isRequired
    // }

    static displayName = `withTxRowState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    render() {
      return <WrappedComponent {...this.props} />;
    }
  }

  const mapStateToProps = (state, props) => ({
    explorerUrl: selectors.getTransactionExplorerUrl(state, {
      hash: props.tx.hash,
    }),

    morAddress: state.config.chain.diamondAddress,

    morTokenAddress: state.config.chain.mainTokenAddress,
    diamondAddress: state.config.chain.diamondAddress,
    walletAddress: state.chain.wallet.address,
  });

  return connect(mapStateToProps)(Container);
};

export default withTxRowState;
