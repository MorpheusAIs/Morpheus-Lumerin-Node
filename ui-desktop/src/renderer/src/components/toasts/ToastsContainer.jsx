import styled from 'styled-components';

const ToastsContainer = styled.div`
  position: fixed;
  top: 10px;
  left: 50%;
  width: min(360px, calc(100vw - 32px));
  transform: translateX(-50%);
  z-index: 200;
`;

export default ToastsContainer;
