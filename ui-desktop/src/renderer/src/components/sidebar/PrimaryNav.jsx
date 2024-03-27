import React from 'react';
import { NavLink } from 'react-router-dom';
import styled from 'styled-components';
import { IconFileReport } from '@tabler/icons-react';
import { IconBuildingStore } from '@tabler/icons-react';
import { IconChecklist } from '@tabler/icons-react';
import { IconWallet } from '@tabler/icons-react';

const Container = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  max-height: 10%;
`;

const Button = styled(NavLink)`
  display: flex;
  min-height: 6rem;
  align-items: center;
  text-decoration: none;
  color: ${p => p.theme.colors.inactive};
  padding: 1.6rem;
  border-top: 1px solid transparent;

  &:focus {
    outline: none;
  }

  &.active {
    color: ${p => p.theme.colors.primary};
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
        activeClassName="active"
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

      {/* <Button
        onClick={() => setActiveIndex(1)}
        activeClassName="active"
        to="/marketplace"
      >
        <IconWrapper>
          <IconBuildingStore width={iconSize} />
        </IconWrapper>
        <Label parent={parent}>Marketplace</Label>
      </Button>

      <Button
        onClick={() => setActiveIndex(2)}
        activeClassName="active"
        data-testid="auction-nav-btn"
        to="/buyer-hub"
      >
        <IconWrapper>
          <IconFileReport width={iconSize} />
        </IconWrapper>
        <Label parent={parent}>Buyer Hub</Label>
      </Button>

      <Button
        onClick={() => setActiveIndex(3)}
        activeClassName="active"
        data-testid="auction-nav-btn"
        to="/seller-hub"
      >
        <IconWrapper>
          <IconChecklist width={iconSize} />
        </IconWrapper>
        <Label parent={parent}>Seller Hub</Label>
      </Button> */}

      {/* <Button
        onClick={() => setActiveIndex(3)}
        activeClassName="active"
        data-testid="auction-nav-btn"
        to="/reports"
      >
        <IconWrapper>
          <CogIcon isActive={activeIndex === 3} size={iconSize} />
        </IconWrapper>
        <Label parent={parent}>Reports</Label>
      </Button> */}
    </Container>
  );
}
