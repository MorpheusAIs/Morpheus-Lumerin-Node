import React from 'react';
import styled from 'styled-components';

const calcWidth = n => 100 / n;

const Container = styled.div`
  display: flex;
  justify-content: start;
  width: 100%;
`;

const Tab = styled.button`
  width: ${calcWidth(3)}%;
  font: inherit;
  line-height: 1.2rem;
  font-size: 1.2rem;
  font-weight: bold;
  color: ${p => p.theme.colors.primary};
  letter-spacing: 1.4px;
  text-align: center;
  opacity: ${p => (p.isActive ? '1' : '0.5')};
  padding: 1.6rem 1rem;
  background: transparent;
  border: none;
  cursor: pointer;
  text-decoration: ${p => (p.isActive ? 'underline' : 'none')};
  margin-bottom: 1px;
  transition: 0.3s;

  &:focus {
    outline: none;
  }

  @media (min-width: 800px) {
    width: ${calcWidth(3)}%;
    font-size: 1.4rem;
  }
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
      <Tab isActive={activeFilter === ''} onClick={() => onFilterChange('')}>
        All
      </Tab>
      <Tab
        isActive={activeFilter === 'sent'}
        onClick={() => onFilterChange('sent')}
      >
        Sent
      </Tab>
      <Tab
        isActive={activeFilter === 'received'}
        onClick={() => onFilterChange('received')}
      >
        Received
      </Tab>
    </Container>
  );
}
