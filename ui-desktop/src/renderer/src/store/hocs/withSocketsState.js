import React from 'react';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';

const withSocketsState = WrappedComponent => {
  class Container extends React.Component {
    // static propTypes = {
    //   sendLmrFeatureStatus: PropTypes.oneOf(['offline', 'no-funds', 'ok'])
    //     .isRequired,
    //   syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed'])
    //     .isRequired,
    //   address: PropTypes.string.isRequired,
    //   client: PropTypes.shape({
    //     refreshAllSockets: PropTypes.func.isRequired,
    //     copyToClipboard: PropTypes.func.isRequired
    //   }).isRequired
    // }

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
    incomingCount: selectors.getConnections(state).length,
    syncStatus: selectors.getProxyRouterSyncStatus(state),
    address: selectors.getWalletAddress(state)
  });

  return connect(mapStateToProps)(withClient(Container));
};

export default withSocketsState;
