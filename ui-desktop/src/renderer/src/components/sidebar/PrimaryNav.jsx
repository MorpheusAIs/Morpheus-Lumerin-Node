import React from 'react';
import { NavLink } from 'react-router-dom';
import styled from 'styled-components';
import { IconFileReport, IconMessage } from '@tabler/icons-react';
import { IconBuildingStore } from '@tabler/icons-react';
import { IconBrandStackshare } from '@tabler/icons-react';
import {
  IconWallet,
  IconPhoto,
  IconPackages,
  IconUsers,
} from '@tabler/icons-react';

const Container = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: space-between;

  max-height: 10%;

  @media (min-width: 800px) {
    padding-left: 2.2rem;
  }
`;

const Button = styled(NavLink)`
  display: flex;
  min-height: 6rem;
  align-items: center;
  text-decoration: none;
  color: white;
  padding: 1.6rem;
  border-top: 1px solid transparent;

  &:focus {
    outline: none;
  }

  &.active {
    color: ${(p) => p.theme.colors.morMain};
    pointer-events: none;
  }
`;

const IconWrapper = styled.div`
  margin-right: 0.75rem;
  margin-left: 0.3rem;
  width: 3rem;
  opacity: 0.5;

  ${Button}.active & {
    opacity: 1;
  }
`;

const Label = styled.span`
  opacity: 0;
  flex-grow: 1;
  font-weight: 500;
  text-align: left;
  padding-bottom: 2px;

  ${({ parent }) => parent}:hover ${Button}.active & {
    opacity: 1;
  }

  ${({ parent }) => parent}:hover & {
    opacity: 1;
  }

  @media (min-width: 800px) {
    opacity: 0.9;

    ${Button}.active & {
      opacity: 1;
      font-weight: 600;
    }
  }
`;

const iconSize = '2rem';

export default function PrimaryNav({ parent, activeIndex, setActiveIndex }) {
  return (
    <Container>
      <Button
        onClick={() => setActiveIndex(0)}
        className={(navData) => (navData.isActive ? 'active-style' : 'none')}
        data-testid="wallet-nav-btn"
        to="/wallet"
      >
        <IconWrapper>
          <IconWallet width={iconSize} />
        </IconWrapper>
        <Label active={activeIndex === 0} parent={parent}>
          Wallet
        </Label>
      </Button>

      <Button onClick={() => setActiveIndex(1)} to="/chat">
        <IconWrapper>
          <IconMessage width={iconSize} />
        </IconWrapper>
        <Label active={activeIndex === 1} parent={parent}>
          Chat
        </Label>
      </Button>

      <Button onClick={() => setActiveIndex(2)} to="/models">
        <IconWrapper>
          <IconPackages width={iconSize} />
        </IconWrapper>
        <Label active={activeIndex === 2} parent={parent}>
          Models
        </Label>
      </Button>

      <Button onClick={() => setActiveIndex(3)} to="/agents">
        <IconWrapper>
          <IconUsers width={iconSize} />
        </IconWrapper>
        <Label active={activeIndex === 3} parent={parent}>
          Agents
        </Label>
      </Button>

      <Button onClick={() => setActiveIndex(4)} to="/providers">
        <IconWrapper>
          <IconBrandStackshare width={iconSize} />
        </IconWrapper>
        <Label active={activeIndex === 4} parent={parent}>
          Provider Hub
        </Label>
      </Button>
    </Container>
  );
}
