import React from 'react';
import { IconRefresh } from '@tabler/icons';
import styled from 'styled-components';

const Container = styled.div`
  background-color: white;
  padding: 5px 15px;
  display: flex;
  justify-content: center;
  color: ${p => p.theme.colors.primary};
  align-items: center;
  font-weight: 500;
  font-size: 1.3rem;
  cursor: pointer;
  border-radius: 12px;
`;

export default function StatusHeader(props) {
  const onRefresh = () => {
    if (props.syncStatus === 'syncing') return;
    props.refresh();
  };

  const iconStyles = {
    width: '20px',
    paddingBottom: '2px',
    paddingTop: '2px',
    marginRight: '0.75rem',
    cursor: 'pointer'
  };

  return (
    <>
      <Container
        onClick={onRefresh}
        style={{ display: 'flex', alignItems: 'center' }}
      >
        {props.syncStatus === 'syncing' ? (
          <span>Syncing Contracts...</span>
        ) : (
          <>
            <IconRefresh style={iconStyles} data-rh={'Refresh Contracts'} />
            Refresh Contracts
            {/* <span>Refresh Contracts</span> */}
          </>
        )}
      </Container>
    </>
  );
}
