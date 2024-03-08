import React from 'react';
import styled from 'styled-components';

const relSize = ratio => `calc(100vw / ${ratio})`;

const Container = styled.div`
  margin: 1.6rem 0 1.6rem;
  width: 100%;
  height: 120px;
  border-radius: 5px;
  display: flex;
  flex-direction: row;

  @media (min-width: 1040px) {
  }
`;

const Total = styled.div`
  background-color: white;
  flex-direction: column;
  height: 95%;
  justify-content: space-between;
  padding: 1.4rem 2.6rem;
  border-radius: 15px;
  min-width: 200px;
  @media (min-width: 1040px) {
  }
`;

const TotalLabel = styled.div`
  display: block;
  line-height: 1.5;
  font-weight: 600;
  color: ${p => p.theme.colors.primary};
  white-space: nowrap;
  position: relative;
  top: ${relSize(-400)};

  @media (min-width: 1440px) {
    font-size: 1.8rem;
  }
`;

const TotalSubLabel = styled.div`
  display: block;
  line-height: 1;
  font-weight: 400;
  color: ${p => p.theme.colors.primary};
  white-space: nowrap;
  position: relative;
  top: ${relSize(-400)};

  @media (min-width: 1440px) {
    font-size: 1.8rem;
  }
`;

const TotalValue = styled.div`
  line-height: 1.5;
  font-weight: 600;
  letter-spacing: ${p => (p.large ? '-1px' : 'inherit')};
  color: ${p => p.theme.colors.primary};
  margin: 0.6rem 0;
  flex-grow: 1;
  position: relative;
  font-size: 2.5rem;

  @media (min-width: 1440px) {
    font-size: ${({ large }) => (large ? '3.6rem' : '2.5rem')};
  }
`;

const TotalsBlock = ({ incoming, outgoing, routed }) => {
  // static propTypes = {
  //   coinBalanceUSD: PropTypes.string.isRequired,
  //   coinBalanceWei: PropTypes.string.isRequired,
  //   lmrBalanceWei: PropTypes.string.isRequired,
  //   coinSymbol: PropTypes.string.isRequired
  // };

  return (
    <>
      <Container>
        <Total>
          <TotalLabel>My Miners</TotalLabel>
          <TotalSubLabel>Incoming Connections</TotalSubLabel>
          <TotalValue>{incoming}</TotalValue>
        </Total>
        {/* <Total>
          <TotalLabel>Lumerin Pool</TotalLabel>
          <TotalSubLabel>Default Outgoing</TotalSubLabel>
          <TotalValue>{outgoing}</TotalValue>
        </Total>
        <Total>
          <TotalLabel>Alt Pool</TotalLabel>
          <TotalSubLabel>Routed</TotalSubLabel>
          <TotalValue>{routed}</TotalValue>
        </Total> */}
      </Container>
    </>
  );
};

export default TotalsBlock;
