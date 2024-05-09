import * as validators from '../validators';
import { withClient } from './clientContext';
import * as utils from '../utils';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';

const withBidsState = WrappedComponent => {
  class Container extends React.Component {
   
    static contextType = ToastsContext;

    static displayName = `withBidsState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    getBitsByModels = async (modelId) => {
        try {
            const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/models/${modelId}/bids`
            const response = await fetch(path);
            const data = await response.json();
            return data.bids;
          }
          catch(e) {
            console.log("Error", e)
            return [];
          }
    }

    getProviders = async () => {
        try {
            const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/providers`
            const response = await fetch(path);
            const data = await response.json();
            return data.providers;
          }
          catch(e) {
            console.log("Error", e)
            return [];
          }
    }
 
    render() {

      return (
        <WrappedComponent
            getProviders={this.getProviders}
            getBitsByModels={this.getBitsByModels}
            {...this.state}
            {...this.props}
        />
      );
    }
  }

  const mapStateToProps = (state, props) => ({
    // selectedCurrency: selectors.getSellerSelectedCurrency(state),
    // isLocalProxyRouter: selectors.getIsLocalProxyRouter(state),
    // titanLightningPool: state.config.chain.titanLightningPool,
    // titanLightningDashboard: state.config.chain.titanLightningDashboard,
    config: state.config,
    selectedModel: state.models.selectedModel
  });

  const mapDispatchToProps = dispatch => ({
    setBid: model => dispatch({ type: 'set-bid', payload: model })
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withBidsState;
