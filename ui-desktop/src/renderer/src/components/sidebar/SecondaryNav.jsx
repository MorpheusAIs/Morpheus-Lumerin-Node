import React from 'react';

import { withClient } from '../../store/hocs/clientContext';
import { NavLink } from 'react-router-dom';
import styled from 'styled-components';
import { IconPlugConnected } from '@tabler/icons';
import { IconCpu2 } from '@tabler/icons';
import { IconHelp } from '@tabler/icons';
import { IconTools } from '@tabler/icons';

const Container = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  max-height: 10%;
  border-top: 1px solid #f4f4f4;
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

const HelpLink = styled.span`
  display: flex;
  min-height: 7.1rem;
  align-items: center;
  text-decoration: none;
  color: ${p => p.theme.colors.darker};
  padding: 1.6rem;
  border-top: 1px solid transparent;
  cursor: pointer;

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
      font-weight: 700;
    }
  }
`;

const NavHeader = styled.h3`
  color: ${p => p.theme.colors.primary};
  padding-left: 2rem;
  text-transform: uppercase;
  font-size: 1.2rem;
  @media (max-width: 800px) {
    display: none;
    ${({ parent }) => parent}:hover & {
      display: block;
    }
  }
`;

const iconSize = '2rem';

function SecondaryNav({
  parent,
  client: { onHelpLinkClick },
  activeIndex,
  setActiveIndex
}) {
  return (
    <Container>
      {/* <NavHeader parent={parent}>Tools</NavHeader>
      <Button
        onClick={() => setActiveIndex(4)}
        activeClassName="active"
        data-testid="auction-nav-btn"
        to="/sockets"
      >
        <IconWrapper>
          <IconPlugConnected width={iconSize} />
        </IconWrapper>
        <Label parent={parent}>Connections</Label>
      </Button>
      <Button
        onClick={() => setActiveIndex(5)}
        activeClassName="active"
        to="/devices"
      >
        <IconWrapper>
          <IconCpu2 width={iconSize} />
        </IconWrapper>
        <Label parent={parent}>Devices</Label>
      </Button>
      <Button
        onClick={() => setActiveIndex(5)}
        activeClassName="active"
        data-testid="tools-nav-btn"
        parent={parent}
        to="/tools"
      >
        <IconWrapper parent={parent}>
          <IconTools width={iconSize} />
        </IconWrapper>
        <Label parent={parent}>Utilities</Label>
      </Button> */}
      <HelpLink data-testid="help-nav-btn" onClick={onHelpLinkClick}>
        <IconWrapper parent={parent}>
          <IconHelp width={iconSize} />
        </IconWrapper>
        <Label parent={parent}>Help</Label>
      </HelpLink>{' '}
    </Container>
  );
}

export default withClient(SecondaryNav);
