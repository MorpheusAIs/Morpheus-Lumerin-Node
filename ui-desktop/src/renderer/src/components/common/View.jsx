import React from 'react';
import styled from 'styled-components';

const Container = styled.div`
  height: 100vh;
  max-width: 100%;
  min-width: 600px;
  position: relative;
  padding: 0 2.4rem;
  padding-top: 2rem;

  @media (min-width: 800px) {
  }

  @media (min-width: 1200px) {
  }
`;

export const View = ({ children }) => (
  <>
    <Container>{children}</Container>
  </>
);
