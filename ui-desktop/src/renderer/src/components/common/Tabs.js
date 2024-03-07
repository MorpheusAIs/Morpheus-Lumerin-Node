import React from 'react';
import styled from 'styled-components';

const Container = styled.div`
  display: flex;
  flex-shrink: 0;
`;

const Tab = styled.button`
  font: inherit;
  cursor: ${p => (p.isDisabled ? 'not-allowed' : 'pointer')};
  flex-grow: 1;
  border: none;
  border-bottom: 2px solid;
  border-bottom-color: ${p =>
    p.isActive ? p.theme.colors.primary : p.theme.colors.darkShade};
  transition: 0.5s;
  padding: 2rem;
  color: ${p =>
    p.isActive
      ? p.theme.colors.light
      : `rgba(255,255,255,${p.isDisabled ? '0.2' : '0.5'})`};
  border-radius: 0;
  background: ${p =>
    p.isActive
      ? 'linear-gradient(253deg, rgba(66, 53, 119, 0.4), rgba(126, 97, 248, 0.1))'
      : 'transparent'};
  line-height: 2rem;
  font-size: 1.6rem;
  font-weight: 600;
  letter-spacing: 1.6px;
  text-align: center;
  outline: none;

  @media (min-height: 700px) {
    padding: 2.5rem;
  }
`;

export default function Tabs({ onClick, active, items }) {
  onClick = ({ currentTarget }) => onClick(currentTarget.dataset.tab);

  return (
    <Container>
      {items.map(({ label, id, disabled, ...other }) => (
        <Tab
          data-testid={`${id}-tab`}
          isDisabled={disabled}
          isActive={active === id}
          data-tab={id}
          onClick={disabled ? null : this.onClick}
          key={id}
          {...other}
        >
          {label}
        </Tab>
      ))}
    </Container>
  );
}
