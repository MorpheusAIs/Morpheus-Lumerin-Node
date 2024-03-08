import React from 'react';
import { AddressHeader } from './AddressHeader';
import styled from 'styled-components';

const Container = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  width: 100%;
  padding: 1.5rem 0;
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
`;

const Title = styled.label`
  font-size: 2.4rem;
  line-height: 3rem;
  white-space: nowrap;
  margin: 0;
  font-weight: 600;
  color: ${p => p.theme.colors.primary};
  margin-bottom: 4.8px;
  margin-right: 2.4rem;
  cursor: default;
  /* width: 100%; */

  @media (min-width: 1140px) {
  }

  @media (min-width: 1200px) {
  }
`;

export const LayoutHeader = ({ title, children }) => (
  <>
    <Container>
      <TitleRow>
        <Title>{title}</Title>
        {children}
      </TitleRow>
    </Container>
  </>
);
