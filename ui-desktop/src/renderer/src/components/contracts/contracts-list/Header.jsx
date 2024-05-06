import React from 'react';
import styled, { css } from 'styled-components';

import ScanIndicator from './ScanIndicator';
import Filter from './Filter';

const responsiveHeader = width => css`
  @media (min-width: ${width}) {
    align-items: baseline;
    display: flex;
    top: 6.8rem;
  }
`;

const Container = styled.div`
  // position: sticky;
  border-radius: 5px;
  // top: 4.1rem;
  // left: 0;
  // right: 0;
  z-index: 1;
  
  ${p => responsiveHeader('800px')}
`;

export default function Header(props) {
  return (
    <>
      <Container>
        <Filter
          onFilterChange={props.onFilterChange}
          onColumnOptionChange={props.onColumnOptionChange}
          activeFilter={false}
          tabs={props.tabs}
        />
      </Container>
    </>
  );
}
