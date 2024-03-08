import React, { useState } from 'react';
import styled from 'styled-components';

import { LightLayout, LastUpdated, Text } from '../common';

const Container = styled.div`
  padding: 3.2rem 2.4rem;
  @media (min-width: 800px) {
    padding: 3.2rem 4.8rem;
  }
`;

const Body = styled.div`
  display: flex;
  margin-top: 3.2rem;
  align-items: center;
  flex-direction: column;

  @media (min-width: 1200px) {
    align-items: flex-start;
    margin-top: 4.8rem;
    flex-direction: row;
  }
`;

const LastUpdatedContainer = styled.div`
  padding: 0 2.4rem 3.2rem;
  @media (min-width: 800px) {
    padding: 0 4.8rem 3.2rem;
  }
`;

export function Totals(props) {
  const propTypes = {};

  const [state, setState] = useState({
    activeModal: false
  });
  const [socketData, setSocketData] = useState([]);

  const onOpenModal = e => {
    e.preventDefault();

    setState({ activeModal: e.target.dataset.modal });
  };

  const onCloseModal = e => {
    e.preventDefault();

    setState({ activeModal: false });
  };

  return (
    <LightLayout title="Lumerin Contracts" data-testid="contracts-container">
      <Container>
        <Text data-testid="title">{props.title}</Text>

        <Body></Body>
      </Container>
      <LastUpdatedContainer>
        <LastUpdated timestamp={props.lastUpdated} />
      </LastUpdatedContainer>
    </LightLayout>
  );
}
