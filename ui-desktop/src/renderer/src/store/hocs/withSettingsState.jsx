import * as validators from '../validators';
import { withClient } from './clientContext';
import * as utils from '../utils';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';

const withSettingsState = WrappedComponent => {
  class Container extends React.Component {

    static contextType = ToastsContext;

    static displayName = `withSettingsState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    logout = () => {
      return this.props.client.logout();
    };

    render() {

      return (
        <WrappedComponent
          logout={this.logout}
          {...this.state}
          {...this.props}
        />
      );
    }
  }

  const mapStateToProps = (state, props) => ({
    config: state.config
  });

  const mapDispatchToProps = dispatch => ({
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withSettingsState;
