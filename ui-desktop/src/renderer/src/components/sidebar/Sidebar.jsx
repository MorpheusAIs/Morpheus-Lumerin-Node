import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import styled from 'styled-components';
import {
  IconLayoutSidebarLeftCollapse,
  IconLayoutSidebarLeftExpand,
} from '@tabler/icons-react';
import SecondaryNav from './SecondaryNav';
import PrimaryNav from './PrimaryNav';
import { NavAction } from './Nav.styles';
import { LumerinLogoFull } from '../icons/LumerinLogoFull';
import { AddressHeader } from '../common/AddressHeader';
import withSidebarState from '../../store/hocs/withSidebarState';

const Container = styled.aside`
  background: #0a1e15;
  width: 220px;
  flex: 0 0 220px;
  padding: 2.4rem 1.2rem 1.6rem;
  display: flex;
  flex-direction: column;
  height: 100vh;
  border-right: 1px solid var(--border-subtle);
  z-index: 5;
  overflow-y: auto;
  overflow-x: hidden;

  .sidebar-toggle {
    display: none;
  }
  .sidebar-address {
    margin: 1.6rem 0 0;
    width: 100%;
    padding: 0;
    font-family: var(--font-mono);
  }

  @media (max-width: 799px) {
    position: fixed;
    inset: 0 auto 0 0;
    width: 64px;
    padding: 1.6rem 0.8rem;
    .sidebar-toggle {
      display: flex;
      margin: 0.4rem 0 1.6rem;
    }
    .sidebar-address {
      display: none;
    }
    &[data-sidebar-expanded='true'] {
      width: 240px;
      box-shadow: 8px 0 32px rgba(0, 0, 0, 0.32);
      .sidebar-address {
        display: block;
      }
    }
  }
`;

const Brand = styled.div`
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0 1.2rem;
  margin-bottom: 3.6rem;
  min-height: 4rem;
  svg {
    width: 4rem;
    height: 3.6rem;
    flex-shrink: 0;
  }
  strong {
    font-size: 1.7rem;
    font-weight: 650;
    letter-spacing: -0.025em;
  }
  @media (max-width: 799px) {
    padding: 0 0.4rem;
    margin-bottom: 0.8rem;
    [data-sidebar-expanded='false'] & strong {
      display: none;
    }
  }
`;

const Primary = styled.nav`
  flex: 1 0 auto;
  padding-bottom: 3.2rem;
`;
const Secondary = styled.nav`
  border-top: 1px solid var(--border-subtle);
  padding-top: 1.2rem;
`;
const Backdrop = styled.button`
  display: none;
  @media (max-width: 799px) {
    display: block;
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.4);
    border: 0;
    z-index: 4;
  }
`;

export function Sidebar({ address, copyToClipboard, onRouteIntent }) {
  const [expanded, setExpanded] = useState(false);
  const { pathname } = useLocation();
  useEffect(() => {
    setExpanded(false);
  }, [pathname]);
  return (
    <>
      {expanded && (
        <Backdrop
          aria-label="Close navigation"
          onClick={() => setExpanded(false)}
          tabIndex={-1}
        />
      )}
      <Container
        data-sidebar-expanded={String(expanded)}
        aria-label="Morpheus navigation"
        onKeyDown={(event) => {
          if (event.key === 'Escape') setExpanded(false);
        }}
      >
        <Brand>
          <LumerinLogoFull />
          <strong>Morpheus</strong>
        </Brand>
        <NavAction
          className="sidebar-toggle"
          aria-label={expanded ? 'Collapse navigation' : 'Expand navigation'}
          aria-expanded={expanded}
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? (
            <IconLayoutSidebarLeftCollapse />
          ) : (
            <IconLayoutSidebarLeftExpand />
          )}
          <span>Collapse menu</span>
        </NavAction>
        <Primary aria-label="Main">
          <PrimaryNav onRouteIntent={onRouteIntent} />
        </Primary>
        <Secondary aria-label="Support">
          <SecondaryNav onRouteIntent={onRouteIntent} />
        </Secondary>
        <AddressHeader address={address} copyToClipboard={copyToClipboard} />
      </Container>
    </>
  );
}

export default withSidebarState(Sidebar);
