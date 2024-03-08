import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';

import selectors from '../selectors';

const withSendDrawerState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      sendLmrFeatureStatus: PropTypes.oneOf(['no-funds', 'offline', 'ok'])
        .isRequired,
      coinSymbol: PropTypes.string.isRequired
    };

    static displayName = `withSendDrawerState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    render() {
      const { sendLmrFeatureStatus, coinSymbol } = this.props;

      const sendLmrDisabledReason =
        sendLmrFeatureStatus === 'no-funds'
          ? `You need some ${coinSymbol} to send`
          : sendLmrFeatureStatus === 'offline'
          ? "Can't send while offline"
          : null;

      return (
        <WrappedComponent
          sendLmrDisabledReason={sendLmrDisabledReason}
          sendLmrDisabled={sendLmrFeatureStatus !== 'ok'}
          {...this.props}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    sendLmrFeatureStatus: selectors.sendLmrFeatureStatus(state),
    coinSymbol: selectors.getCoinSymbol(state)
  });

  return connect(mapStateToProps)(Container);
};

export default withSendDrawerState;
