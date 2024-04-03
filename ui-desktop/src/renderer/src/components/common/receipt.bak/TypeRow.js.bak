import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

const Label = styled.div`
  line-height: 1.4rem;
  font-size: 1.1rem;
  letter-spacing: 0.4px;
  text-transform: uppercase;
  color: ${p => p.theme.colors.copy};
  margin-right: 1.6rem;
  white-space: nowrap;
`;

const Value = styled.div`
  color: ${p => p.theme.colors.copy};
  text-align: right;
  line-height: 1.6rem;
  font-size: 1.3rem;
`;

const Type = styled(Value)`
  text-transform: capitalize;
`;

export default class TypeRow extends React.Component {
  static propTypes = {
    isCancelApproval: PropTypes.bool,
    isApproval: PropTypes.bool,
    txType: PropTypes.string.isRequired
  };

  render() {
    return (
      <React.Fragment>
        <Label>Type</Label>
        <Type>
          {this.props.isCancelApproval
            ? 'Allowance canceled'
            : this.props.isApproval
            ? 'Allowance set'
            : this.props.txType === 'import-requested'
            ? 'Import Request'
            : this.props.txType}
        </Type>
      </React.Fragment>
    );
  }
}
