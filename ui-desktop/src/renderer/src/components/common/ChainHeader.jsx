
import styled from 'styled-components';

import ArbLogo from '../icons/ArbLogo';

const Container = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  width: 100%;
  padding: 0 0 1.5rem 0;
  z-index: 2;
  right: 0;
  left: 0;
  top: 0;
`;

const TitleRow = styled.div`
  width: 100%;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
`;

const Title = styled.div`
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 2.4rem;
  line-height: 3rem;
  white-space: nowrap;
  margin: 0;
  font-weight: 600;
  color: ${p => p.theme.colors.morMain};
  margin-right: 2.4rem;
  cursor: default;
  /* width: 100%; */

  @media (min-width: 1140px) {
  }

  @media (min-width: 1200px) {
  }
`;

const ChainContainer = styled.div`
    display: flex;
    align-items: center;
    gap: 10px;
    color: white;
    font-size: 1.2rem;
    padding: 0.6rem 1.2rem;
    background: rgba(255,255,255,0.04);
    border-width: 1px;
    border: 1px solid rgba(255,255,255,0.04);
`

const getChainLogo = (chainId) => {
  const arbLogo = (<ArbLogo style={{ width: '20px'}} />);
  if(chainId === 42161 || chainId === 421614) {
    return arbLogo;
  }
  // Handle other icons (ETH, Base, etc.)

  return arbLogo;
}

export const ChainHeader = ({ title, children, chain }) => (
    <Container>
      <TitleRow>
        <Title>
          <div>{title}</div>
          <ChainContainer>
            {getChainLogo(Number(chain?.chainId || 42161))}
            <div>{chain?.displayName || "Arbitrum"}</div>
          </ChainContainer>
        </Title>
        {children}
      </TitleRow>
    </Container>
);
