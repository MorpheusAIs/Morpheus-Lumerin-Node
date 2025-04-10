import React from 'react';
import { connect } from 'react-redux';
import { withClient } from './clientContext';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';

const withServicesState = (WrappedComponent: React.ComponentType<any>) => {
  class Container extends React.Component {
    static contextType = ToastsContext;

    render() {
      return <WrappedComponent {...this.state} {...this.props} />;
    }
  }

  const mapStateToProps = (state) => ({
    services: selectors.getServices(state),
  });

  return withClient(connect(mapStateToProps)(Container));
};

export default withServicesState;
