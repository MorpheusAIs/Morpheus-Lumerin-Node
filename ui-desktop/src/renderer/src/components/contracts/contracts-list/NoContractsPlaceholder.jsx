import styled from 'styled-components';
import React from 'react';

const Container = styled.div`
  padding: 6.4rem 3.2rem;
`;

const Label = styled.div`
  line-height: 3rem;
  font-size: 2.4rem;
  font-weight: 600;
  text-align: center;
  color: #c2c2c2;
  margin-top: 0.8rem;
`;

const Emoji = styled.svg`
  display: block;
  margin: 0 auto;
`;

export default function NoContractsPlaceholder({ message }) {
  return (
    <Container data-testid="no-contract-placeholder">
      <Label>{`${message || 'No contracts yet!'}`}</Label>
    </Container>
  );
}
