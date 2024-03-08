import React from 'react';
import styled from 'styled-components';

import withSocketsRowState from '../../../store/hocs/withSocketsRowState';
import theme from '../../../ui/theme';
import { SocketIcon } from '../../icons/SocketIcon';

const columnCount = 3;

const calcWidth = n => 100 / n;

const Container = styled.div`
  padding: 1.2rem 0;
  display: flex;
  text-align: center;
  box-shadow: 0 -1px 0 0 ${p => p.theme.colors.lightShade} inset;
  cursor: pointer;
  height: 66px;
`;

const Value = styled.label`
  display: flex;
  flex-direction: column;
  justify-content: center;
  width: ${calcWidth(columnCount)}%;
  color: black;
  font-size: 1.2rem;
`;

const AssetContainer = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: center;
  text-align: center;
  width: ${calcWidth(columnCount)}%;
`;

const SmallAssetContainer = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: center;
  text-align: center;
  /* width: ${calcWidth(columnCount)}%; */
`;

const StatusPill = styled.span`
  border-radius: 5px;
  background-color: ${({ color }) => color};
  font-size: 1rem;
  padding: 0.6rem 1rem;
  width: 75%;
  display: block;
  margin: 0 auto;
`;

const stateColors = {
  Running: theme.colors.primaryLight,
  Available: theme.colors.success,
  Vetting: theme.colors.warning
};

const statusToState = status => {
  switch (status) {
    case 'free':
      return 'Available';
    case 'vetting':
      return 'Vetting';
    default:
      return 'Running';
  }
};

const Row = ({ socket }) => {
  return (
    <Container>
      <Value>{socket.ID}</Value>
      <AssetContainer>
        <StatusPill color={stateColors[statusToState(socket.Status)]}>
          {statusToState(socket.Status)}
        </StatusPill>
      </AssetContainer>
      <Value>{socket.CurrentDestination}</Value>
    </Container>
  );
};

export default withSocketsRowState(Row);
