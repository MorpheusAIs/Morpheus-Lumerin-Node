import React from 'react';
import styled, { css } from 'styled-components';

import ScanIndicator from './ScanIndicator';
import Filter from './Filter';

const responsiveHeader = width => css`
  @media (min-width: ${width}) {
    align-items: baseline;
    display: flex;
  }
`;

const Container = styled.div`
  position: sticky;
  border-radius: 5px;
  left: 0;
  right: 0;
  z-index: 1;

  ${p => responsiveHeader('800px')}
`;

export default function Header(props) {
  // hasConnections: PropTypes.bool.isRequired,
  // onWalletRefresh: PropTypes.func.isRequired,
  // onFilterChange: PropTypes.func.isRequired,
  // activeFilter: PropTypes.string.isRequired,
  // syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed']).isRequired

  return (
    <>
      <Container>
        <Filter onFilterChange={props.onFilterChange} activeFilter={false} />
      </Container>
    </>
  );
}
