import React from 'react';
import styled from 'styled-components';

const Container = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 0 0 2.4rem;
  z-index: 2;
  right: 0;
  left: 0;
  top: 0;
`;

const TitleRow = styled.div`
  width: 100%;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1.2rem;
`;

const Title = styled.h1`
  font-size: 2.4rem;
  line-height: 3rem;
  margin: 0;
  font-weight: 650;
  color: var(--text-primary, #edf7f0);
  letter-spacing: -0.025em;
  cursor: default;
`;

export const LayoutHeader = ({ title, children }) => (
  <Container>
    <TitleRow>
      <Title>{title}</Title>
      {children}
    </TitleRow>
  </Container>
);
