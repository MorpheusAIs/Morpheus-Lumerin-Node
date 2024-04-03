import React from 'react';
import styled from 'styled-components';

const columnCount = 3;
const calcWidth = n => 100 / n;

const Container = styled.div`
  display: flex;
  justify-content: start;
  width: 100%;
`;

const Tab = styled.button`
  width: ${calcWidth(columnCount)}%;
  font: inherit;
  line-height: 1.2rem;
  font-size: 1.2rem;
  font-weight: bold;
  color: ${p => p.theme.colors.primary};
  letter-spacing: 1.4px;
  text-align: center;
  opacity: ${p => (p.isActive ? '1' : '0.75')};
  padding: 1.6rem 1rem;
  background: transparent;
  border: none;
  cursor: pointer;
  border-bottom: 2px solid ${p => (p.isActive ? 'white' : 'transparent')};
  margin-bottom: 1px;
  transition: 0.3s;

  &:focus {
    outline: none;
  }

  @media (min-width: 800px) {
    width: ${calcWidth(columnCount)}%;
    font-size: 1.4rem;
  }
`;

const Spacer = styled.div`
  width: ${calcWidth(columnCount)}%;
`;

export default function Filter({ onFilterChange, activeFilter }) {
  // static propTypes = {
  //   onFilterChange: PropTypes.func.isRequired,
  //   activeFilter: PropTypes.oneOf([
  //     'converted',
  //     'received',
  //     'auction',
  //     'ported',
  //     'sent',
  //     ''
  //   ]).isRequired
  // }

  return (
    <Container>
      <Tab
        isActive={activeFilter === 'ipAddress'}
        onClick={() => onFilterChange('ipAddress')}
      >
        Worker Address
      </Tab>
      {/* <Spacer /> */}
      <Tab
        isActive={activeFilter === 'status'}
        onClick={() => onFilterChange('status')}
      >
        Status
      </Tab>
      <Tab
        isActive={activeFilter === 'socketAddress'}
        onClick={() => onFilterChange('socketAddress')}
      >
        Socket
      </Tab>
      {/* <Tab
        isActive={activeFilter === 'total'}
        onClick={() => onFilterChange('totalShares')}
      >
        Total
      </Tab>
      <Tab
        isActive={activeFilter === 'accepted'}
        onClick={() => onFilterChange('accepted')}
      >
        Accepted
      </Tab>
      <Tab
        isActive={activeFilter === 'rejected'}
        onClick={() => onFilterChange('rejected')}
      >
        Rejected
      </Tab> */}
    </Container>
  );
}
