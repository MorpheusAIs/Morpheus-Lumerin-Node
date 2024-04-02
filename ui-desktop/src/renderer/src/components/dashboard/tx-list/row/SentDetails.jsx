import FilteredMessage from '../../../../components/common/FilteredMessage';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

const Address = styled.span`
  letter-spacing: normal;
  line-height: 1.6rem;
  font-size: 1.1rem;
  font-weight: 600;
  text-transform: initial;
  color: white;

  @media (min-width: 800px) {
    font-size: 1.3rem;
  }
`;

export default class SentDetails extends React.Component {
  static propTypes = {
    isCancelApproval: PropTypes.bool,
    isApproval: PropTypes.bool,
    isPending: PropTypes.bool.isRequired,
    to: PropTypes.string.isRequired
  };

  render() {
    return (
      <div>
        {this.props.isPending ? 'Pending to ' : 'Sent to '}
        <Address>{this.props.to}</Address>
      </div>
    );
  }
}
