import React from 'react';
import styled, { keyframes } from 'styled-components';

const loading = keyframes`
    from {left: -200px; width: 30%;}
    50% {width: 30%;}
    70% {width: 70%;}
    80% { left: 50%;}
    95% {left: 120%;}
    to {left: 100%;}
`;

const Container = styled.div`
  width: 100%;
  background-color: ${p => p.theme.colors.translucentPrimary};
  padding: 0.2rem;
  border-radius: 0.8rem;
`;

const Bar = styled.div`
  position: relative;
  overflow: hidden;
  height: 0.4rem;
  border-radius: 0.2rem;

  &:before {
    border-radius: 0.4rem;
    height: 0.4rem;
    display: block;
    position: absolute;
    content: '';
    left: -200px;
    width: 200px;
    background-color: ${p => p.theme.colors.primary};
    animation: ${loading} 2s linear infinite;
  }
`;

export default function LoadingBar() {
  return (
    <Container>
      <Bar />
    </Container>
  );
}
