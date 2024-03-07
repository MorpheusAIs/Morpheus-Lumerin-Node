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

export default class NoSocketPlaceholder extends React.Component {
  render() {
    return (
      <Container data-testid="no-socket-placeholder">
        <Label>No connections yet!</Label>
      </Container>
    );
  }
}
