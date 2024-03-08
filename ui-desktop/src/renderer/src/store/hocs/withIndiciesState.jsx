import React from 'react';
import { connect } from 'react-redux';

import { withClient } from './clientContext';
import selectors from '../selectors';

const withIndiciesState = WrappedComponent => {
  class Container extends React.Component {
    static displayName = `withIndiciesState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    render() {
      return (
        <WrappedComponent
          copyToClipboard={this.props.client.copyToClipboard}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    address: selectors.getWalletAddress(state)
  });

  return withClient(connect(mapStateToProps)(Container));
};

export default withIndiciesState;
