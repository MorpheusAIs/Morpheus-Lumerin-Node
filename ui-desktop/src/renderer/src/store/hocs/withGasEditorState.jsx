import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';

import { withClient } from './clientContext';
import selectors from '../selectors';
import * as utils from '../utils';

const withGasEditorState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      onInputChange: PropTypes.func.isRequired,
      useCustomGas: PropTypes.bool.isRequired,
      gasPrice: PropTypes.string.isRequired,
      gasLimit: PropTypes.string.isRequired,
      dispatch: PropTypes.func.isRequired,
      errors: PropTypes.shape({
        gasPrice: PropTypes.string,
        gasLimit: PropTypes.string
      }).isRequired,
      client: PropTypes.shape({
        getGasPrice: PropTypes.func.isRequired,
        fromWei: PropTypes.func.isRequired
      }).isRequired
    };

    static displayName = `withGasEditorState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = { priceError: false };

    componentDidMount() {
      this._isMounted = true;

      // Avoid getting current price if using custom price
      if (this.props.useCustomGas) return;

      this.props.client
        .getGasPrice({
          chain: this.props.chain
        })
        .then(this.updateGasPrice)
        .catch(() => {
          if (this._isMounted) this.setState({ priceError: true });
        });
    }

    componentWillUnmount() {
      this._isMounted = false;
    }

    updateGasPrice = ({ gasPrice }) => {
      this.props.dispatch({
        type: 'gas-price-updated',
        payload: gasPrice,
        chain: this.props.chain
      });
      if (
        !this._isMounted ||
        // Parity may return 0 as gasPrice if latests blocks are empty
        !utils.isGreaterThanZero(this.props.client, gasPrice)
      ) {
        return;
      }
      this.setState({ priceError: false });
      this.props.onInputChange({
        id: 'gasPrice',
        value: this.props.client.fromWei(gasPrice, 'gwei')
      });
    };

    onGasToggle = () => {
      const { useCustomGas, onInputChange } = this.props;
      onInputChange({ id: 'useCustomGas', value: !useCustomGas });
    };

    render() {
      return (
        <WrappedComponent
          onGasToggle={this.onGasToggle}
          {...this.props}
          {...this.state}
        />
      );
    }
  }

  const mapStateToProps = state => ({});

  return connect(mapStateToProps)(withClient(Container));
};

export default withGasEditorState;
