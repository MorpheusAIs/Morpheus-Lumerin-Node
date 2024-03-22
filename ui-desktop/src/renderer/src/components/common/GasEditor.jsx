import withGasEditorState from '../../store/hocs/withGasEditorState';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

import { FieldBtn } from './Btn';
import TextInput from './TextInput';
import Flex from './Flex';
import Sp from './Spacing';

const GasLabel = styled.span`
  opacity: 0.5;
  font-size: 1.3rem;
  white-space: nowrap;
  margin-right: 1em;
`;

const ErrorMsg = styled.div`
  margin-top: 1em;
  color: ${p => p.theme.colors.danger};
`;

class GasEditor extends React.Component {
  static propTypes = {
    gasEstimateError: PropTypes.bool,
    onInputChange: PropTypes.func.isRequired,
    useCustomGas: PropTypes.bool.isRequired,
    onGasToggle: PropTypes.func.isRequired,
    priceError: PropTypes.bool,
    gasPrice: PropTypes.string.isRequired,
    gasLimit: PropTypes.string.isRequired,
    errors: PropTypes.shape({
      gasPrice: PropTypes.string,
      gasLimit: PropTypes.string
    }).isRequired
  };

  render() {
    return (
      <React.Fragment>
        {this.props.useCustomGas ? (
          <Flex.Row justify="space-between">
            <Flex.Item grow="1" basis="0">
              <TextInput
                data-testid="gas-limit-field"
                onChange={this.props.onInputChange}
                label="Gas Limit (Units)"
                error={this.props.errors.gasLimit}
                value={this.props.gasLimit}
                id="gasLimit"
              />
            </Flex.Item>
            <Sp mt={6} mx={2.5} />
            <Flex.Item grow="1" basis="0">
              <TextInput
                data-testid="gas-price-field"
                onChange={this.props.onInputChange}
                label="Gas Price (Gwei)"
                error={this.props.errors.gasPrice}
                value={this.props.gasPrice}
                id="gasPrice"
              />
            </Flex.Item>
          </Flex.Row>
        ) : (
          <Flex.Row justify="space-between" align="flex-start">
            <Flex.Row rowwrap shrink="1">
              <GasLabel>Gas Limit: {this.props.gasLimit} (Units)</GasLabel>
              <GasLabel>Gas Price: {this.props.gasPrice} (Gwei)</GasLabel>
            </Flex.Row>
            <FieldBtn
              data-testid="edit-gas-btn"
              tabIndex="-1"
              onClick={this.props.onGasToggle}
            >
              EDIT GAS
            </FieldBtn>
          </Flex.Row>
        )}
        {this.props.priceError && (
          <ErrorMsg>
            Current gas price could not be retrieved. Using last known price.
          </ErrorMsg>
        )}
        {this.props.gasEstimateError && (
          <ErrorMsg>Gas limit could not be estimated. Using default.</ErrorMsg>
        )}
      </React.Fragment>
    );
  }
}

export default withGasEditorState(GasEditor);
