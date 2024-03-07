import React from 'react';
import styled from 'styled-components';
import SentDetails from './SentDetails';
import ReceivedDetails from './ReceivedDetails';

const Container = styled.div`
  line-height: 1.6rem;
  font-size: 1rem;
  letter-spacing: 0px;
  color: ${p => p.theme.colors.primary};
  text-transform: uppercase;
  opacity: ${({ isPending }) => (isPending ? '0.5' : '1')};
  text-align: center;

  @media (min-width: 800px) {
    font-size: 1.1rem;
    letter-spacing: 0.4px;
  }
`;

const Failed = styled.span`
  line-height: 1.6rem;
  color: ${p => p.theme.colors.danger};
`;

export default function Details(props) {
  return (
    <Container isPending={props.isPending}>
      {props.isFailed ? (
        <Failed>FAILED TRANSACTION</Failed>
      ) : props.txType === 'sent' ? (
        <SentDetails {...props} />
      ) : props.txType === 'received' ? (
        <ReceivedDetails {...props} />
      ) : (
        <div>Waiting for metadata</div>
      )}
    </Container>
  );
}
