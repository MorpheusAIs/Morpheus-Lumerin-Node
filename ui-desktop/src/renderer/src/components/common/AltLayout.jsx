import styled from 'styled-components'
import React from 'react'

import Flex from './Flex'
import Sp from './Spacing'

const Container = styled(Flex.Column)`
  min-height: 100vh;
  padding: 3.2rem;
  background: transparent;
  top: center;
`

const Body = styled.div`
  background: rgba(255,255,255, 0.04);
  border-width: 1px;
  border: 1px solid rgba(255, 255, 255, 0.04);
  padding: 3rem;
  max-width: 53rem;
  width: 100%;
  margin-top: 4rem;
  @media (min-height: 800px) {
    margin-top: 8rem;
  }
`

const Title = styled.div`
  line-height: 3rem;
  font-size: 1.8rem;
  font-weight: bold;
  text-align: center;
  cursor: default;
  color: ${(p) => p.theme.colors.dark};
  @media (min-height: 600px) {
    font-size: 2.4rem;
  }
`

const LogoContainer = styled.div`
  padding-top: 80px;
  padding-bottom: 20px;
`

export default function AltLayout({ title, children, ...other }) {
  return (
    <Container align="center" {...other}>
      <LogoContainer>
        {/* <LumerinLogoFull height="80px" width="250px" /> */}
      </LogoContainer>
      <Body>
        {title && <Title>{title}</Title>}
        <Sp mt={2}>{children}</Sp>
      </Body>
    </Container>
  )
}
