import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';

import selectors from '../selectors';

const withScanIndicatorState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      onWalletRefresh: PropTypes.func.isRequired,
      syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed'])
        .isRequired,
      syncBlock: PropTypes.number,
      isOnline: PropTypes.bool.isRequired
    };

    static displayName = `withScanIndicatorState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    onLabelClick = () => {
      if (this.props.isOnline && this.props.syncStatus !== 'syncing') {
        this.props.onWalletRefresh();
      }
    };

    render() {
      const label = this.props.isOnline
        ? this.props.syncStatus === 'syncing'
          ? 'Syncing…'
          : this.props.syncStatus === 'failed'
          ? 'Sync failed'
          : 'Up-to-date'
        : 'Offline';

      const tooltip = this.props.isOnline
        ? this.props.syncStatus === 'failed'
          ? 'Retry'
          : this.props.syncStatus === 'up-to-date'
          ? 'Refresh'
          : undefined
        : undefined;

      return (
        <WrappedComponent
          onLabelClick={this.onLabelClick}
          tooltip={tooltip}
          label={label}
          {...this.props}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    isOnline: selectors.getIsOnline(state)
  });

  return connect(mapStateToProps)(Container);
};

export default withScanIndicatorState;
