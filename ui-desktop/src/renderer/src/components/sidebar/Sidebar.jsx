import React, { useState } from 'react'
import styled from 'styled-components'

import SecondaryNav from './SecondaryNav'
import PrimaryNav from './PrimaryNav'

import { LumerinLogoFull } from '../icons/LumerinLogoFull'
import { AddressHeader } from '../common/AddressHeader'
import withSidebarState from '../../store/hocs/withSidebarState'

const Container = styled.div`
  background: transparent;
  width: 7rem;
  padding-bottom: 4.5rem;
  display: flex;
  flex-direction: column;
  transition: width 0.2s;
  position: absolute;
  overflow: hidden;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 3;
  border-right: 1px solid rgba(255, 255, 255, 0.16);
  background: #03160e!important

  .sidebar-address {
    display: none;
  }

  &:hover {
    width: 210px;
    box-shadow: 0 0 16px 0 rgba(0, 0, 0, 0.2);

    .sidebar-address {
      display: block;
    }
  }
  @media (min-width: 800px) {
    position: relative;
    min-width: 210px;
    width: 210px;

    .sidebar-address {
      display: block;
    }

    &:hover {
      box-shadow: none;
    }
  }
`

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
`

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
`

const NavContainer = styled.div`
  display: flex;
  justify-content: space-between;
  flex-direction: column;
  height: 100%;
`

const PrimaryNavContainer = styled.nav`
  flex-grow: 1;
  margin-top: 3rem;

  @media (max-width: 800px) {
    padding-left: 0.5rem;
  }
`

function Sidebar(props) {
  const { address, copyToClipboard } = props;
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
          />
        </PrimaryNavContainer>

        <nav style={{ borderTop: '1px solid rgba(255, 255, 255, 0.16)'}}>
          <SecondaryNav
            activeIndex={activeIndex}
            setActiveIndex={setActiveIndex}
            parent={Container}
          />
        </nav>
        <AddressHeader address={address} copyToClipboard={copyToClipboard} />
      </NavContainer>
    </Container>
  )
}

export default withSidebarState(Sidebar)
