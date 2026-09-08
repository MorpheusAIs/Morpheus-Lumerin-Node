import styled from 'styled-components';

export const View = styled.div`
  height: 100vh;
  max-width: 100%;
  min-width: 0;
  position: relative;
  padding: 2.8rem 3.2rem;
  overflow-y: auto;
  @media (max-width: 799px) {
    padding: 2rem 1.6rem;
  }
`;
