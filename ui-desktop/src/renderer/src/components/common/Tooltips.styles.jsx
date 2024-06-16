import styled, { createGlobalStyle, keyframes } from 'styled-components';
import 'react-hint/css/index.css';

export const GlobalStyles = createGlobalStyle`
  .react-hint {
    &:after { display: none !important; }
  }
`;

const trans = keyframes`
  from { transform: translateY(-10px); }
  to { transform: translateY(-5px); }
`;

export const Container = styled.div`
  animation: 0.5s ${trans};
  animation-fill-mode: forwards;
  background-color: ${p =>
    p.negative
      ? p.theme.colors.morMain
      : p.darker
      ? p.theme.colors.darker
      : p.theme.colors.dark};
  max-width: ${p => p.maxWidth || 'auto'};
  font-size: 1.3rem;
  padding: 8px 12px;
  border-radius: 4px;
  box-shadow: 0 0px 3px 0px #323232;
  position: relative;
  color: ${p => p.theme.colors.light};

  &:after {
    content: '';
    width: 0;
    height: 0;
    margin: auto;
    display: block;
    position: absolute;
    top: auto;
    bottom: -5px;
    left: 0;
    right: 0;
    border: 5px solid transparent;
    z-index: 1;
    border-bottom: none;
    border-top-color: ${p =>
      p.negative
        ? p.theme.colors.morMain
        : p.darker
        ? p.theme.colors.darker
        : p.theme.colors.dark};
  }
`;

export const ContainerLocal = styled(Container)`
  display: ${p => (p.show ? 'block' : 'none')}
  position: absolute;
  left: 50%;
  transform: translate(-50%, -5px);
  white-space: pre;
  animation: 0.5s 
    ${keyframes`
    from { 
      transform: translate(-50%, -10px); 
      opacity: 0;
    }
    to { 
      transform: translate(-50%, -5px);
      opacity: 1;
    }
  `}
  bottom: calc(100% - 2em);
`;
