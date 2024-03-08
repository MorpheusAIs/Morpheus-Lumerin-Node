import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';

const withReceiveModalState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      address: PropTypes.string.isRequired,
      client: PropTypes.shape({
        copyToClipboard: PropTypes.func.isRequired
      }).isRequired
    };

    static displayName = `withReceiveModalState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = { copyBtnLabel: 'Copy to clipboard' };

    copyToClipboard = () => {
      this.props.client
        .copyToClipboard(this.props.address)
        .then(() => this.setState({ copyBtnLabel: 'Copied to clipboard!' }))
        .catch(err => this.setState({ copyBtnLabel: err.message }));
    };

    render() {
      return (
        <WrappedComponent
          copyToClipboard={this.copyToClipboard}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    address: selectors.getWalletAddress(state)
  });

  return connect(mapStateToProps)(withClient(Container));
};

export default withReceiveModalState;
