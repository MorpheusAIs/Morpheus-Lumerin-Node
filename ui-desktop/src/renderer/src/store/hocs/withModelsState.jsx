import * as validators from '../validators';
import { withClient } from './clientContext';
import * as utils from '../utils';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';

const withModelsState = WrappedComponent => {
  class Container extends React.Component {
   
    static contextType = ToastsContext;

    static displayName = `withModelsState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    getAllModels = async () => {
        const result = await this.props.client.getAllModels();
        console.log("🚀 ~ Container ~ getAllModels= ~ result:", result)
        return result;
    }

    getAllProviders = async () => {

    }

    getBitsByModels = async (modelId) => {
        
    }
 
    render() {

      return (
        <WrappedComponent
            getAllModels={this.getAllModels}
            getAllProviders={this.getAllProviders}
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
    config: state.config
  });

  const mapDispatchToProps = dispatch => ({
    setSelectedModel: model => dispatch({ type: 'set-model', payload: model })
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withModelsState;
