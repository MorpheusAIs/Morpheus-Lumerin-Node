import React from 'react';
import { connect } from 'react-redux';

import { withClient } from './clientContext';
import selectors from '../selectors';
import PropTypes from 'prop-types';

const withReportsState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      sendLmrFeatureStatus: PropTypes.oneOf(['offline', 'no-funds', 'ok'])
        .isRequired,
      syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed'])
        .isRequired,
      address: PropTypes.string.isRequired,
      client: PropTypes.shape({
        // refreshAllSockets: PropTypes.func.isRequired,
        copyToClipboard: PropTypes.func.isRequired
      }).isRequired
    };

    static displayName = `withSocketsState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      refreshStatus: 'init',
      refreshError: null
    };

    onSocketRefresh = () => {
      this.setState({ refreshStatus: 'pending', refreshError: null });
      // this.props.client
      //   .refreshAllSockets()
      //   .then(() => this.setState({ refreshStatus: 'success' }))
      //   .catch(() =>
      //     this.setState({
      //       refreshStatus: 'failure',
      //       refreshError: 'Could not refresh'
      //     })
      //   );
    };

    render() {
      const { sendLmrFeatureStatus } = this.props;

      const sendDisabledReason =
        sendLmrFeatureStatus === 'offline'
          ? "Can't send while offline"
          : sendLmrFeatureStatus === 'no-funds'
          ? 'You need some funds to send'
          : null;

      return (
        <WrappedComponent
          sendDisabledReason={sendDisabledReason}
          copyToClipboard={this.props.client.copyToClipboard}
          onSocketRefresh={this.onSocketRefresh}
          sendDisabled={sendLmrFeatureStatus !== 'ok'}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    sendLmrFeatureStatus: selectors.sendLmrFeatureStatus(state),
    syncStatus: selectors.getSocketsSyncStatus(state),
    address: selectors.getWalletAddress(state)
  });

  return withClient(connect(mapStateToProps)(Container));
};

export default withReportsState;
