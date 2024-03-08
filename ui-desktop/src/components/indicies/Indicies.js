import React from 'react';
import styled from 'styled-components';

import withIndiciesState from '../../store/hocs/withIndiciesState';
import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';

const Title = styled.h3`
  line-height: 3rem;
  color: ${p => p.theme.colors.darker}
  white-space: nowrap;
  margin: 2.4rem 0 0 0;
  cursor: default;
`;

const Description = styled.p`
  color: ${p => p.theme.colors.dark};
  margin: 1rem 2rem;
`;
// const Title = styled.h1`
//   font-size: 2.4rem;
//   line-height: 3rem;
//   color: ${p => p.theme.colors.darker}
//   white-space: nowrap;
//   margin: 0;
//   cursor: default;
// `

const Indicies = ({ address, copyToClipboard }) => (
  <View data-testid="indicies-container">
    <LayoutHeader
      title="Swap"
      address={address}
      copyToClipboard={copyToClipboard}
    />
    <Title>Uniswap</Title>
    <Description>
      Swap, earn, and build on the leading decentralized trading protocol. A
      growing network of DeFi Apps.
    </Description>

    <Title>SushiSwap</Title>
    <Description>
      Swap, earn, borrow, leverage all on one decentralized, community driven
      platform. Be a DeFi Chef with Sushi. Welcome home to DeFi.
    </Description>

    <Title>PancakeSwap</Title>
    <Description>
      Trade, earn, and win crypto on the most popular decentralized platform in
      the galaxy.
    </Description>
  </View>
);

export default withIndiciesState(Indicies);
