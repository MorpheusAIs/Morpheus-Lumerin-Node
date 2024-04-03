import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';

import selectors from '../selectors';

const withOfflineState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      isOnline: PropTypes.bool.isRequired
    };

    static displayName = `withOfflineState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      isVisible: !this.props.isOnline
    };

    componentDidUpdate(prevProps) {
      if (this.state.isVisible && this.props.isOnline && !prevProps.isOnline) {
        this.setState({ isVisible: false });
      } else if (
        !this.state.isVisible &&
        !this.props.isOnline &&
        prevProps.isOnline
      ) {
        this.setState({ isVisible: true });
      }
    }

    handleDismissClick = () => this.setState({ isVisible: false });

    render() {
      return (
        <WrappedComponent
          handleDismissClick={this.handleDismissClick}
          isVisible={this.state.isVisible}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    isOnline: selectors.getIsOnline(state)
  });

  return connect(mapStateToProps)(Container);
};

export default withOfflineState;
