import React, { useState, useEffect } from 'react';
import withSocketsState from '../../store/hocs/withSocketsState';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import axios from 'axios';

import { LightLayout, LastUpdated, LoadingBar, Text, Btn, Sp } from '../common';

const Container = styled.div`
  padding: 3.2rem 2.4rem;
  @media (min-width: 800px) {
    padding: 3.2rem 4.8rem;
  }
`;

const LoadingContainer = styled.div`
  text-align: center;
  max-width: 400px;
  margin: 0 auto;
`;

const Body = styled.div`
  display: flex;
  margin-top: 3.2rem;
  align-items: center;
  flex-direction: column;

  @media (min-width: 1200px) {
    align-items: flex-start;
    flex-direction: row;
  }
`;

const BuyBtn = styled(Btn)`
  order: 0;
  white-space: nowrap;
  margin-bottom: 3.2rem;
  min-width: 300px;

  @media (min-width: 1200px) {
    margin-bottom: 0;
    order: 1;
    min-width: auto;
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

  useEffect(() => {
    setTimeout(pingMessengerAPI, 5000);
  }, [socketData]);

  const pingMessengerAPI = async () => {
    const { data } = await axios('http://localhost:8080/connection');
    console.log('connection data: ', data);

    setSocketData(data);
  };

  const onOpenModal = e => {
    e.preventDefault();

    setState({ activeModal: e.target.dataset.modal });
  };

  const onCloseModal = e => {
    e.preventDefault();

    setState({ activeModal: false });
  };

  return (
    <LightLayout title="Lumerin Sockets" data-testid="contracts-container">
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
