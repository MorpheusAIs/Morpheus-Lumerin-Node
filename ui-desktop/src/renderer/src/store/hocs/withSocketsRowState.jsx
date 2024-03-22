import selectors from '../selectors';
import { connect } from 'react-redux';
import * as utils from '../utils';
import PropTypes from 'prop-types';
import React from 'react';

const withSocketsRowState = WrappedComponent => {
  class Container extends React.Component {
    // static propTypes = {
    //   confirmations: PropTypes.number.isRequired,
    //   coinSymbol: PropTypes.string.isRequired,
    //   sockets: PropTypes.isRequired
    // }

    static displayName = `withSocketsRowState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    render() {
      const { sockets } = this.props;

      return <WrappedComponent {...this.props} {...sockets} />;
    }
  }

  const mapStateToProps = (state, props) => ({
    // avoid unnecessary re-renders once transaction is confirmed
  });

  return connect(mapStateToProps)(Container);
};

export default withSocketsRowState;
