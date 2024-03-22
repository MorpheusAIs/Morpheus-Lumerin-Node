import React from 'react';
import { IconRefresh, IconSearch } from '@tabler/icons';
import styled from 'styled-components';

import debounce from 'lodash/debounce';

const Container = styled.div`
  background-color: white;
  padding: 0 15px;
  display: flex;
  justify-content: start;
  color: ${p => p.theme.colors.primary};
  align-items: center;
  font-weight: 100;
  width: 230px;
  font-size: 1.5rem;
  border-radius: 12px;
`;

const SearchInput = styled.input`
  outline: 0;
  border: 0px;
  background: white;
  width: 100%;

  ::placeholder {
    color: #565656;
  }
`;

export default function Search(props) {
  const onChange = debounce(v => props.onSearch(v), 500);

  return (
    <>
      <Container style={{ display: 'flex', alignItems: 'center' }}>
        <IconSearch
          onClick={props.refresh}
          style={{
            width: '16px',
            marginRight: '6px',
            cursor: 'pointer'
          }}
        />
        <SearchInput
          placeholder="Search Сontracts"
          onChange={e => onChange(e.target.value)}
        />
      </Container>
    </>
  );
}
