import React, { useState } from 'react';
import { connect } from 'react-redux';

import selectors from '../selectors';
import { withClient } from './clientContext';

const withTxListState = Component => {
  const WrappedComponent = props => {
    const [nextPageLoading, setNextPageLoading] = useState(false);
    const getPastTransactions = () => {
      setNextPageLoading(true);
      props.client
        .getPastTransactions({
          address: props.address,
          page: props.page,
          pageSize: props.pageSize
        })
        .finally(() => {
          setNextPageLoading(false);
        });
    };

    return (
      <Component
        transactions={props.transactions}
        syncStatus={props.syncStatus}
        isNextPageLoading={nextPageLoading}
        hasNextPage={props.hasNextPage}
        getPastTransactions={getPastTransactions}
      />
    );
  };

  const mapStateToProps = state => ({
    transactions: selectors.getTransactions(state),
    page: selectors.getTransactionPage(state),
    pageSize: selectors.getTransactionPageSize(state),
    hasNextPage: selectors.getHasNextPage(state),
    address: selectors.getWalletAddress(state)
  });

  return withClient(connect(mapStateToProps)(WrappedComponent));
};

export default withTxListState;
