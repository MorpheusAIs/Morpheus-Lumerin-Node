import React from 'react';
import styled from 'styled-components';

export const Container = styled.div`
  margin-top: 2.4rem;
  height: 100%;

  @media (min-width: 800px) {
  }
  @media (min-width: 1200px) {
  }
`;

export const Contracts = styled.div`
  margin: 1.6rem 0 1.6rem;
  border-radius: 15px;
  background-color: #fff;
`;

export const ListContainer = styled.div`
  height: calc(100vh - ${p => p.offset || '375'}px);

  /*
  ::-webkit-scrollbar {
    display: none;
  }

  @media (min-width: 800px) {
  }
  @media (min-width: 1200px) {
  }
  */
`;

export const FooterLogo = styled.div`
  padding: 4.8rem 0;
  width: 3.2rem;
  margin: 0 auto;
`;

export const Title = styled.div`
  font-size: 2rem;
  line-height: 3rem;
  color: ${p => p.theme.colors.primary};
  white-space: nowrap;
  margin: 0;
  font-weight: 500;
  margin-right: 2.4rem;
  cursor: default;

  @media (min-width: 1140px) {
    margin-right: 0.8rem;
  }

  @media (min-width: 1200px) {
    margin-right: 1.6rem;
  }
`;
