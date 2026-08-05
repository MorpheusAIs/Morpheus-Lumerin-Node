import * as validators from '../validators';
import { withClient } from './clientContext';
import * as utils from '../utils';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';
import { pooledMapSettled, withTimeout } from '../utils/concurrency';

const BALANCE_CONCURRENCY = 6;
const BALANCE_TIMEOUT_MS = 8000;

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
        const authHeaders = await this.props.client.getAuthHeaders();
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/providers`
        const response = await fetch(path, {
          headers: authHeaders
        });
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
        const authHeaders = await this.props.client.getAuthHeaders();
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/sessions/provider?provider=${provider}`;
        const response = await fetch(path, {
          headers: authHeaders
        });
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
        const authHeaders = await this.props.client.getAuthHeaders();
        const path = `${this.props.config.chain.localProxyRouterUrl}/proxy/sessions/${sessionId}/providerClaimableBalance`
        const response = await fetch(path, {
          headers: authHeaders
        });
        const data = await response.json();
        return data.balance;
      }
      catch(e) {
        console.log("Error", e)
        return [];
      }
    }

    claimFunds = async (sessionId) => {
      // NOTE: this used to reference a bare `props` (instead of `this.props`),
      // so every claim threw a ReferenceError that the catch below swallowed —
      // the button silently did nothing. Errors now propagate to the caller so
      // the UI can actually report a failed claim.
      const authHeaders = await this.props.client.getAuthHeaders();
      const path = `${this.props.config.chain.localProxyRouterUrl}/proxy/sessions/${sessionId}/providerClaim`;
      const response = await fetch(path, {
        method: 'POST',
        headers: authHeaders,
      });
      const dataResponse = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(
          dataResponse.error || `Claim failed (HTTP ${response.status})`,
        );
      }
      return dataResponse;
    }

    fetchData = async (providerId) => {
      // Models and sessions are independent — fetch them together rather than
      // one after the other.
      const [models, providerSession] = await Promise.all([
        this.getAllModels(),
        this.getSessionsByProvider(providerId),
      ]);
      const modelsNames = (models ?? []).reduce((a,b) => ({ ...a, [b.Id]: b.Name}), {});

      // Per-session claimable balance used to run in a sequential await loop —
      // N round-trips end to end, which is why this tab crawled for providers
      // with any real session history. Now bounded-parallel.
      const sessions = providerSession ?? [];
      const results = await pooledMapSettled(
        sessions,
        async (session) => {
          if (session.ClosedAt) {
            return { ...session, Balance: 0 };
          }
          const balance = await withTimeout(
            this.getBalanceBySession(session.Id),
            BALANCE_TIMEOUT_MS,
          );
          return { ...session, Balance: balance };
        },
        (session) => ({ ...session, Balance: 0 }),
        BALANCE_CONCURRENCY,
      );

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
