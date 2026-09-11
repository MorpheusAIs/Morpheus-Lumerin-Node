import { withClient } from './clientContext';
import { connect } from 'react-redux';
import React from 'react';
import selectors from '../selectors';

const withProvidersState = (WrappedComponent) => {
  class Container extends React.Component {
    static displayName = `withProvidersState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    getAllModels = async () => {
      const result = await this.props.client.getAllModels();
      return result;
    };

    getSessionsByProvider = async (provider) => {
      return (
        (await this.props.client.getSessionsByProvider({ provider })) || []
      );
    };

    getBalanceBySession = async (sessionId) => {
      return this.props.client.getProviderClaimableBalance({ sessionId });
    };

    claimFunds = async (sessionId) => {
      // NOTE: this used to reference a bare `props` (instead of `this.props`),
      // so every claim threw a ReferenceError that the catch below swallowed —
      // the button silently did nothing. Errors now propagate to the caller so
      // the UI can actually report a failed claim.
      return this.props.client.claimProviderFunds({ sessionId });
    };

    render() {
      return (
        <WrappedComponent
          getAllModels={this.getAllModels}
          getBalanceBySession={this.getBalanceBySession}
          claimFunds={this.claimFunds}
          getSessionsByProvider={this.getSessionsByProvider}
          {...this.state}
          {...this.props}
        />
      );
    }
  }

  const mapStateToProps = (state, props) => ({
    providerId: selectors.getWalletAddress(state),
  });

  return withClient(connect(mapStateToProps)(Container));
};

export default withProvidersState;
