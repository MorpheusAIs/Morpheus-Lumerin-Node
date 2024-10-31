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

    getConfig = async () => {
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/config`;
        const response = await fetch(path);
        const data = await response.json();
        return data;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    }

    updateEthNodeUrl = async (value) => {
      if(!value)
        return;

      if(!/\b(?:http|ws)s?:\/\/\S*[^\s."]/g.test(value)) {
        this.context.toast('error', "Invalid format");
        return;
      }

      const ethNodeResult = await fetch(`${this.props.config.chain.localProxyRouterUrl}/config/ethNode`, {
        method: 'POST',
        body: JSON.stringify({ urls: [value] })
      })

      const dataResponse = await ethNodeResult.json();
      if (dataResponse.error) {
        this.context.toast('error', dataResponse.error);
        return;
      }

      this.context.toast('success', "Changed");
    }


    render() {

      return (
        <WrappedComponent
          logout={this.logout}
          getConfig={this.getConfig}
          updateEthNodeUrl={this.updateEthNodeUrl}
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
