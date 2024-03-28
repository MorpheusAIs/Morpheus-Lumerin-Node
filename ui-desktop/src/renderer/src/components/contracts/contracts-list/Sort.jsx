import React from 'react';
import { IconRefresh, IconSearch } from '@tabler/icons-react';
import styled from 'styled-components';
import Select from 'react-select';

const Container = styled.div`
  /* margin: 10px 0; */
  margin-left: -10px;
  width: fit-content;
  color: ${p => p.theme.colors.primary};
  font-weight: 500;
  display: flex;
  align-self: end;
  align-items: center;
`;

const rangeSelectOptions = [
  {
    label: 'None',
    value: null
  },
  {
    label: 'Price: Low to High',
    value: 'AscPrice'
  },
  {
    label: 'Price: High to Low',
    value: 'DescPrice'
  },
  {
    label: 'Duration: Short to Long',
    value: 'AscDuration'
  },
  {
    label: 'Duration: Long to Short',
    value: 'DescDuration'
  },
  {
    label: 'Speed: Slow to Fast',
    value: 'AscSpeed'
  },
  {
    label: 'Speed: Fast to Slow',
    value: 'DescSpeed'
  },
  {
    label: 'State: Available First',
    value: 'AvailableFirst'
  },
  {
    label: 'State: Running First',
    value: 'RunningFirst'
  },
  {
    label: 'Profit: Below Target First',
    value: 'UnderProfit'
  }
];

export default function Sort(props) {
  return (
    <>
      <Container>
        <Select
          className="sorting"
          classNamePrefix="select"
          name="sorting"
          styles={{
            control: (base, state) => ({
              ...base,
              width: 'auto',
              minWidth: '150px',
              textAlign: 'right',
              cursor: 'pointer',
              color: '#252B34',
              border: state.isFocused ? 0 : 0,
              boxShadow: state.isFocused ? 0 : 0,
              '&:hover': {
                border: state.isFocused ? 0 : 0
              },
              borderColor: state.isFocused ? '#0e4353' : undefined,
              background: 'transparent'
            }),
            placeholder: base => ({
              ...base,
              color: '#252B34',
              fontSize: '1.4rem',
              fontWeight: 700
            }),
            singleValue: base => ({
              ...base,
              color: '#0e4353',
              fontWeight: 600,
              fontSize: '1.4rem'
            }),
            indicatorsContainer: base => ({
              ...base,
              color: '#0e4353'
            }),
            indicatorSeparator: base => ({ ...base, display: 'none' }),
            dropdownIndicator: base => ({
              ...base,
              color: '#0e4353',
              marginLeft: -15
            }),
            option: (base, state) => ({
              ...base,
              cursor: 'pointer',
              backgroundColor: state.isSelected ? '#0e4353' : undefined,
              fontSize: '1.5rem',
              fontWeight: 500,
              lineHeight: '1.6rem',
              color: state.isSelected ? '#FFFFFF' : undefined,
              ':active': {
                ...base[':active'],
                backgroundColor: '#0e435380',
                color: '#FFFFFF'
              },
              ':hover': {
                backgroundColor: '#5ADCE2'
              }
            })
          }}
          onChange={e => (e.value ? props.setSort(e) : props.setSort(null))}
          isSearchable={false}
          placeholder="Sort By"
          value={props.sort}
          options={rangeSelectOptions}
        />
      </Container>
    </>
  );
}
