import React, { useState } from 'react';
import { connect } from 'react-redux';

import selectors from '../selectors';
import { withClient } from './clientContext';

const withTxListState = Component => {
  const WrappedComponent = props => {

    return (
      <Component
        {...props}
      />
    );
  };

  const mapStateToProps = state => ({
    address: selectors.getWalletAddress(state)
  });

  return withClient(connect(mapStateToProps)(WrappedComponent));
};

export default withTxListState;
