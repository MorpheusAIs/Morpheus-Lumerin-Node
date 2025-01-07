import * as validators from '../validators';
import { withClient } from './clientContext';
import * as utils from '../utils';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';

const withProvidersState = WrappedComponent => {
  class Container extends React.Component {
   
    static contextType = ToastsContext;

    static displayName = `withProvidersState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    getAllModels = async () => {
        const result = await this.props.client.getAllModels();
        return result;
    }

    getAllProviders = async () => {
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

    getSessionsByProvider = async (provider) => {
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/sessions/provider?provider=${provider}`;
        const response = await fetch(path);
        const data = await response.json();
        return data.sessions;
      }
      catch(e) {
        console.log("Error", e)
        return [];
      }
    }

    getBalanceBySession = async (sessionId) => {
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/proxy/sessions/${sessionId}/providerClaimableBalance`
        const response = await fetch(path);
        const data = await response.json();
        return data.balance;
      }
      catch(e) {
        console.log("Error", e)
        return [];
      }
    }

    claimFunds = async (sessionId) => {
      try {
        const path = `${props.config.chain.localProxyRouterUrl}/proxy/sessions/${sessionId}/providerClaim`;
        const response = await fetch(path, {
            method: "POST",
        });
        const dataResponse = await response.json();
      }
      catch(e) {
        console.log("Error", e)
      }
    }

    fetchData = async (providerId) => {
      const models = await this.getAllModels();
      // const providers = await getAllProviders();
      const providerSession = await this.getSessionsByProvider(providerId);
      const modelsNames = models.reduce((a,b) => ({ ...a, [b.Id]: b.Name}), {});
      
      let results = [];
      for (const session of providerSession) {
        const id = session.Id;
        let balance = 0;
        try {
          if(!session.ClosedAt) {
            balance = (await this.getBalanceBySession(id));
          }
        }
        catch(e) {
          console.log(e);
        }
        results.push({ ...session, Balance: balance })
      }

      return { results, modelsNames };
    }
 
    render() {

      return (
        <WrappedComponent
            getAllModels={this.getAllModels}
            getAllProviders={this.getAllProviders}
            getBalanceBySession={this.getBalanceBySession}
            claimFunds={this.claimFunds}
            getSessionsByProvider={this.getSessionsByProvider}
            fetchData={this.fetchData}
            {...this.state}
            {...this.props}
        />
      );
    }
  }

  const mapStateToProps = (state, props) => ({
    // selectedCurrency: selectors.getSellerSelectedCurrency(state),
    providerId: selectors.getWalletAddress(state),
    config: state.config
  });

  const mapDispatchToProps = dispatch => ({
    setSelectedModel: model => dispatch({ type: 'set-model', payload: model })
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withProvidersState;
