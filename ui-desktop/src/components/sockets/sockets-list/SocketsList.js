import React, { useState, useEffect } from 'react';
import { List as RVList, AutoSizer, WindowScroller } from 'react-virtualized';
import withSocketsListState from '../../../store/hocs/withSocketsListState';
import styled from 'styled-components';

import ScanningSocketsPlaceholder from './ScanningSocketsPlaceholder';
import NoSocketsPlaceholder from './NoSocketsPlaceholder';
import { ItemFilter } from '../../common';
import Header from './Header';
import SocketsRow from './Row';

const Container = styled.div`
  margin-top: 2.4rem;
  background-color: ${p => p.theme.colors.light};
  border-radius: 15px;

  @media (min-width: 960px) {
  }
`;

const Sockets = styled.div`
  margin-top: 1.6rem;
  border-radius: 15px;
`;

const ListContainer = styled.div`
  height: calc(100vh - 355px);
`;

const SocketsRowContainer = styled.div`
  &:hover {
    background-color: rgba(126, 97, 248, 0.1);
  }
`;

const Subtitle = styled.div`
  font-size: 1.4rem;
  align-self: end;
  line-height: 2rem;
  white-space: nowrap;
  margin: 0 1.2rem;
  display: inline;
  font-weight: 400;
  color: ${p => p.theme.colors.primary};
  cursor: default;

  @media (min-width: 1140px) {
    margin-right: 0.8rem;
  }

  @media (min-width: 1200px) {
    margin-right: 1.6rem;
  }
`;

const SocketsList = props => {
  // static propTypes = {
  //   hasSockets: PropTypes.bool.isRequired,
  //   onWalletRefresh: PropTypes.func.isRequired,
  //   syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed']).isRequired,
  //   items: PropTypes.arrayOf(
  //     PropTypes.shape({
  //       SocketsType: PropTypes.string.isRequired,
  //       hash: PropTypes.string.isRequired
  //     })
  //   ).isRequired
  // };

  const rowRenderer = sockets => ({ key, index, style }) => (
    <SocketsRowContainer style={style} key={`${key}-${index}`}>
      <SocketsRow
        data-testid="Sockets-row"
        // onClick={props.onSocketsClicked}
        socket={sockets[index]}
      />
    </SocketsRowContainer>
  );

  const filterExtractValue = ({ Status }) => Status;

  return (
    <Container data-testid="Sockets-list">
      <Sockets>
        <ItemFilter extractValue={filterExtractValue} items={props.connections}>
          {({ filteredItems, onFilterChange, activeFilter }) => (
            <React.Fragment>
              <Header
                onFilterChange={onFilterChange}
                activeFilter={activeFilter}
                syncStatus={props.syncStatus}
              />

              <ListContainer>
                {!props.hasConnections &&
                  (props.syncStatus === 'syncing' &&
                  props.isLocalProxyRouter ? (
                    <ScanningSocketsPlaceholder />
                  ) : (
                    <NoSocketsPlaceholder />
                  ))}
                <AutoSizer>
                  {({ width, height }) => (
                    <RVList
                      rowRenderer={rowRenderer(filteredItems)}
                      rowHeight={66}
                      rowCount={filteredItems.length}
                      height={height || 500} // defaults for tests
                      width={width || 500} // defaults for tests
                    />
                  )}
                </AutoSizer>
              </ListContainer>
            </React.Fragment>
          )}
        </ItemFilter>
      </Sockets>
    </Container>
  );
};

export default withSocketsListState(SocketsList);
