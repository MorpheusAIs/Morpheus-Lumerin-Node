import React from 'react';
import styled, { keyframes } from 'styled-components';

const Container = styled.div`
  display: inline-block;
  background-color: #ffffff;
  border-radius: ${({ size }) => size};
  padding: 2px;
  box-shadow: 0 1px 1px 0 ${p => p.theme.colors.darkShade};
`;

const rotate = keyframes`
  100% {
    transform: rotate(360deg);
  }
`;

const Svg = styled.svg`
  transform-origin: center center;
  animation: ${rotate} 2s linear infinite;
  display: block;
  height: ${({ size }) => size};
  width: ${({ size }) => size};
`;

const dash = keyframes`
  0% {
    stroke-dasharray: 1, 200;
    stroke-dashoffset: 0;
  }
  50% {
    stroke-dasharray: 89, 200;
    stroke-dashoffset: -35px;
  }
  100% {
    stroke-dasharray: 89, 200;
    stroke-dashoffset: -124px;
  }
`;

const Circle = styled.circle`
  stroke-dasharray: 1, 200;
  stroke-dashoffset: 0;
  animation: ${dash} 1.5s ease-in-out infinite;
  stroke-linecap: round;
  stroke: ${p => p.theme.colors.primary};
`;

export default function Spinner({ size = '12px', ...rest }) {
  return (
    <Container size={size}>
      <Svg viewBox="25 25 50 50" size={size} {...rest}>
        <Circle
          strokeMiterlimit="10"
          strokeWidth="6"
          fill="none"
          cx="50"
          cy="50"
          r="20"
        />
      </Svg>
    </Container>
  );
}
