import styled from 'styled-components';
import React from 'react';

import Flex from './Flex';
import Sp from './Spacing';
import { LumerinLogoFull } from '../icons/LumerinLogoFull';

const Container = styled(Flex.Column)`
  min-height: 100vh;
  padding: 4rem 2rem;
  justify-content: center;
  background: transparent;
`;

const Body = styled.div`
  background: var(--surface-raised, #10271e);
  border: 1px solid var(--border-subtle);
  border-radius: 16px;
  padding: 3.2rem;
  max-width: 53rem;
  width: 100%;
  margin-top: 2.4rem;
  @media (max-width: 480px) {
    padding: 2.4rem;
  }
`;

const Title = styled.h1`
  line-height: 3rem;
  font-size: 1.8rem;
  font-weight: 650;
  letter-spacing: -0.025em;
  text-align: center;
  cursor: default;
  color: ${(p) => p.theme.colors.dark};
  @media (min-height: 600px) {
    font-size: 2.4rem;
  }
`;

const LogoContainer = styled.div`
  display: flex;
  justify-content: center;
  svg {
    width: 7.2rem;
    height: 4rem;
  }
`;

export default function AltLayout({ title, children, ...other }) {
  return (
    <Container align="center" {...other}>
      <LogoContainer>
        <LumerinLogoFull />
      </LogoContainer>
      <Body>
        {title && <Title>{title}</Title>}
        <Sp mt={2}>{children}</Sp>
      </Body>
    </Container>
  );
}
