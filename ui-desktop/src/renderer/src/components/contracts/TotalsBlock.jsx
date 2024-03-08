import React, { useState } from 'react';
import styled from 'styled-components';

const relSize = ratio => `calc(100vw / ${ratio})`;

const Container = styled.div`
  margin: 1.6rem 0 1.6rem;
  width: 100%;
  height: 100px;
  border-radius: 5px;
  display: flex;
  flex-direction: row;

  @media (min-width: 1040px) {
  }
`;

const Total = styled.div`
  background-color: ${p => p.theme.colors.xLight};
  display: flex;
  flex-direction: column;
  height: 95%;
  width: 100px;
  justify-content: space-between;
  padding: 1.4rem 2.6rem;
  margin-left: 1.4rem;
  border-radius: 5px;
  @media (min-width: 1040px) {
  }
`;

const TotalRow = styled.div`
  display: flex;
  width: 100%;
  height: 30%;
  flex-direction: row;
  justify-content: space-between;
  border-radius: 5px;
  @media (min-width: 1040px) {
    padding: 0.95em 0;
  }
`;

const TotalLabel = styled.div`
  display: block;
  line-height: 1.5;
  font-weight: 600;
  color: ${p => p.theme.colors.dark}
  white-space: nowrap;
  position: relative;
  top: ${relSize(-400)};
  font-size: ${relSize(76)};

  @media (min-width: 800px) {
    font-size: ${relSize(68)};
  }

  @media (min-width: 1440px) {
    font-size: 2.2rem;
  }
`;

const TotalSubLabel = styled.div`
  display: block;
  line-height: 1.0;
  font-weight: 400;
  color: ${p => p.theme.colors.dark}
  white-space: nowrap;
  position: relative;
  top: ${relSize(-400)};
  font-size: ${relSize(82)};

  @media (min-width: 800px) {
    font-size: ${relSize(82)};
  }

  @media (min-width: 1440px) {
    font-size: 2.2rem;
  }
`;

const TotalValue = styled.div`
  line-height: 1.5;
  font-weight: 600;
  letter-spacing: ${p => (p.large ? '-1px' : 'inherit')};
  color: ${p => p.theme.colors.darker}
  margin: .6rem 0;
  flex-grow: 1;
  position: relative;
  top: ${relSize(-400)};
  font-size: ${relSize(32)};

  @media (min-width: 800px) {
    font-size: ${relSize(44)};
  }

  @media (min-width: 1040px) {
    font-size: ${({ large }) => relSize(large ? 40 : 52)};
  }

  @media (min-width: 1440px) {
    font-size: ${({ large }) => (large ? '3.6rem' : '2.8rem')};
  }
`;

const TotalsBlock = () => {
  // static propTypes = {
  //   coinBalanceUSD: PropTypes.string.isRequired,
  //   coinBalanceWei: PropTypes.string.isRequired,
  //   lmrBalanceWei: PropTypes.string.isRequired,
  //   coinSymbol: PropTypes.string.isRequired
  // };

  return (
    <Container>
      <Total>
        <TotalRow>
          <TotalLabel>Active</TotalLabel>
        </TotalRow>
        <TotalValue>{4}</TotalValue>
      </Total>
      <Total>
        <TotalRow>
          <TotalLabel>Draft</TotalLabel>
        </TotalRow>
        <TotalValue>{1}</TotalValue>
      </Total>
    </Container>
  );
};

export default TotalsBlock;
