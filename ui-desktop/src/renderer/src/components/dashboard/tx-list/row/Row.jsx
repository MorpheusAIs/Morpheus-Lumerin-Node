//@ts-check

import React from 'react';
import styled from 'styled-components';
import withTxRowState from '../../../../store/hocs/withTxRowState';
import { EtherIcon } from '../../../icons/EtherIcon';
import { MorpheusLogo } from '../../../icons/MorpheusLogo';
import { UnknownContractIcon } from '../../../icons/UnknownContractIcon';
import { Field } from './Field';
import { defaultMeta, metaComponentMap } from './RowMeta';
import { MorpheusMarketplaceTxIcon } from '../../../icons/MorpheusMarketplaceTx';

const Container = styled.div`
  margin-left: 1.6rem;
  padding: 1.2rem 2.4rem 1.2rem 0;
  display: flex;
  gap: 1rem;
  align-items: center;
  box-shadow: 0 -1px 0 0 ${(p) => p.theme.colors.lightShade} inset;
  cursor: pointer;
  height: 66px;
`;

const IconContainer = styled.div`
  display: block;
  text-align: center;
  width: 10%;
  flex-shrink: 0;
`;

const Action = styled.div`
  font-size: 1.6rem;
  color: #fff;
  width: 20%;
  flex-shrink: 0;
  text-wrap: nowrap;
  overflow: hidden;
}`;

const MetaComponentWrap = styled.div`
  display: flex;
  gap: 1rem;
  width: 40%;
  flex-shrink: 0;
  overflow: hidden;
`;

const DateWrap = styled.div`
  display: flex;
  width: 30%;
  flex-shrink: 0;
`;

/** @param {{tx: import("./tx").Tx, explorerUrl: string, walletAddress: string, morTokenAddress: string, diamondAddress: string}} props */
const Row = ({
  tx,
  explorerUrl,
  walletAddress,
  morTokenAddress,
  diamondAddress,
}) => {
  const tokens = {
    [morTokenAddress.toLowerCase()]: {
      name: 'Morpheus',
      symbol: 'MOR',
      decimals: 18,
    },
    '0x0000000000000000000000000000000000000000': {
      name: 'Ether',
      symbol: 'ETH',
      decimals: 18,
    },
  };

  const contractIcons = {
    [morTokenAddress.toLowerCase()]: MorpheusLogo, //MorpheusTokenIcon,
    [diamondAddress.toLowerCase()]: MorpheusMarketplaceTxIcon, //MorpheusMarketplace,
    '0x0000000000000000000000000000000000000000': EtherIcon,
  };

  let Icon = tx.contract
    ? contractIcons[tx.contract.contractAddress.toLowerCase()] ||
      UnknownContractIcon
    : UnknownContractIcon;

  let action;
  if (tx.contract) {
    action = tx.contract.methodName || 'Smart Contract Call';
  } else if (tx.transfers?.length > 0) {
    action = 'ETH Transfer';
  } else {
    action = 'Unknown';
  }

  if (action === 'ETH Transfer') {
    Icon = EtherIcon;
  }

  if (action === 'openSession') {
    console.log('diamondAddress,', diamondAddress);
    console.log('tx,', tx);
  }

  const MetaComponent = metaComponentMap[action] || defaultMeta;

  return (
    <Container onClick={() => window.openLink(explorerUrl)}>
      <IconContainer>
        <Icon width="2em" />
      </IconContainer>
      <Action>{action}</Action>
      <MetaComponentWrap>
        <MetaComponent tx={tx} walletAddress={walletAddress} tokens={tokens} />
      </MetaComponentWrap>
      <DateWrap>
        <DateCell timestamp={new Date(tx.timestamp)} />
      </DateWrap>
    </Container>
  );
};

/** @param {{timestamp: Date}} props */
const DateCell = ({ timestamp }) => (
  <Field title="Date">{timestamp.toLocaleString()}</Field>
);

export default withTxRowState(Row);
