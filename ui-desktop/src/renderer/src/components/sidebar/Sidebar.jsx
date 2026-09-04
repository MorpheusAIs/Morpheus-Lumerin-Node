import React, { useState } from 'react';
import styled from 'styled-components';

import SecondaryNav from './SecondaryNav';
import PrimaryNav from './PrimaryNav';

import { LumerinLogoFull } from '../icons/LumerinLogoFull';
import { AddressHeader } from '../common/AddressHeader';
import withSidebarState from '../../store/hocs/withSidebarState';

const Container = styled.div`
  background: #03160e !important;
  width: 250px;
  padding-bottom: 4.5rem;
  display: flex;
  flex-direction: column;
  clip-path: inset(0 calc(250px - 7rem) 0 0);
  transition: clip-path 0.2s cubic-bezier(0.22, 1, 0.36, 1);
  position: absolute;
  overflow: hidden;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 3;
  border-right: 1px solid rgba(255, 255, 255, 0.16);

  .sidebar-address {
    display: none;
  }

  &:hover {
    clip-path: inset(0);
    box-shadow: 0 0 16px 0 rgba(0, 0, 0, 0.2);

    .sidebar-address {
      display: block;
    }
  }
  @media (min-width: 800px) {
    position: relative;
    min-width: 250px;
    width: 250px;
    clip-path: none;
    transition: none;

    .sidebar-address {
      display: block;
    }

    &:hover {
      box-shadow: none;
    }
  }
`;

const FullLogoContainer = styled.div`
  padding: 4rem 2.2rem 2.8rem 2.2rem;
  height: 100px;
  display: none;
  flex-shrink: 0;

  ${({ parent }) => parent}:hover & {
    display: block;
  }
  @media (min-width: 800px) {
    display: block;
  }
`;

const IconLogoContainer = styled.div`
  padding: 40px 0.8rem 2rem 0.8rem;
  height: 100px;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-shrink: 0;

  ${({ parent }) => parent}:hover & {
    display: none;
  }
  @media (min-width: 800px) {
    display: none;
  }
`;

const NavContainer = styled.div`
  display: flex;
  justify-content: space-between;
  flex-direction: column;
  height: 100%;
`;

const PrimaryNavContainer = styled.nav`
  flex-grow: 1;
  margin-top: 3rem;

  @media (max-width: 800px) {
    padding-left: 0.5rem;
  }
`;

function Sidebar(props) {
  const { address, copyToClipboard, onRouteIntent } = props;
  const [activeIndex, setActiveIndex] = useState(0);
  return (
    <Container>
      <FullLogoContainer parent={Container}>
        <LumerinLogoFull />
      </FullLogoContainer>

      <IconLogoContainer parent={Container}>
        <LumerinLogoFull />
      </IconLogoContainer>
      <NavContainer>
        <PrimaryNavContainer>
          <PrimaryNav
            parent={Container}
            activeIndex={activeIndex}
            setActiveIndex={setActiveIndex}
            onRouteIntent={onRouteIntent}
          />
        </PrimaryNavContainer>

        <nav style={{ borderTop: '1px solid rgba(255, 255, 255, 0.16)' }}>
          <SecondaryNav
            activeIndex={activeIndex}
            setActiveIndex={setActiveIndex}
            parent={Container}
            onRouteIntent={onRouteIntent}
          />
        </nav>
        <AddressHeader address={address} copyToClipboard={copyToClipboard} />
      </NavContainer>
    </Container>
  );
}

export default withSidebarState(Sidebar);
