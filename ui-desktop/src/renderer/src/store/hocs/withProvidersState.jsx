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

const withProvidersState = (WrappedComponent) => {
  class Container extends React.Component {
    static contextType = ToastsContext;

    static displayName = `withProvidersState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    getAllModels = async () => {
      const result = await this.props.client.getAllModels();
      return result;
    };

    getAllProviders = async () => {
      try {
        return (await this.props.client.getProviders()) || [];
      } catch (e) {
        console.log('Error', e);
        return [];
      }
    };

    getSessionsByProvider = async (provider) => {
      try {
        return (
          (await this.props.client.getSessionsByProvider({ provider })) || []
        );
      } catch (e) {
        console.log('Error', e);
        return [];
      }
    };

    getBalanceBySession = async (sessionId) => {
      try {
        return await this.props.client.getProviderClaimableBalance({
          sessionId,
        });
      } catch (e) {
        console.log('Error', e);
        return [];
      }
    };

    claimFunds = async (sessionId) => {
      // NOTE: this used to reference a bare `props` (instead of `this.props`),
      // so every claim threw a ReferenceError that the catch below swallowed —
      // the button silently did nothing. Errors now propagate to the caller so
      // the UI can actually report a failed claim.
      return this.props.client.claimProviderFunds({ sessionId });
    };

    fetchData = async (providerId) => {
      // Models and sessions are independent — fetch them together rather than
      // one after the other.
      const [models, providerSession] = await Promise.all([
        this.getAllModels(),
        this.getSessionsByProvider(providerId),
      ]);
      const modelsNames = (models ?? []).reduce(
        (a, b) => ({ ...a, [b.Id]: b.Name }),
        {},
      );

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
    };

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
    config: state.config,
  });

  const mapDispatchToProps = (dispatch) => ({
    setSelectedModel: (model) =>
      dispatch({ type: 'set-model', payload: model }),
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withProvidersState;
