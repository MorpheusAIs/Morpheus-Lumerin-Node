import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { IconExternalLink } from '@tabler/icons-react';

import withContractsRowState from '../../../store/hocs/withContractsRowState';
import { Btn } from '../../common';
import {
  formatDuration,
  formatSpeed,
  formatPrice,
  getContractState
} from '../utils';
import Spinner from '../../common/Spinner';
import { abbreviateAddress } from '../../../utils';
import ProgressBarWithLabels from '../../common/ProgressBar';

const Container = styled.div`
  padding: 1.2rem 0;
  display: grid;
  grid-template-columns: ${p => p.ratio.map(x => `${x}fr`).join(' ')};
  text-align: center;
  box-shadow: 0 -1px 0 0 ${p => p.theme.colors.lightShade} inset;
  cursor: ${p => p.cursor || 'pointer'};
  height: 66px;
  opacity: ${p => p.opacity || 1};
`;

const Value = styled.label`
  display: flex;
  padding: 0 3rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: ${p => p.theme.colors.primary};
  font-size: 1.2rem;
`;

const ContractValue = styled(Value)`
  cursor: pointer;
  text-decoration: underline;
  flex-direction: row;
  gap: 5px;
`;

const ActionButton = styled(Btn)`
  font-size: 1.2rem;
  letter-spacing: 1px;
  padding: 0.8rem 2.25rem;
  line-height: 1.5rem;
`;

function MarketplaceRow({
  contract,
  ratio,
  explorerUrl,
  onPurchase,
  allowSendTransaction,
  address,
  symbol
}) {
  // TODO: Add better padding
  const [isPending, setIsPending] = useState(false);

  useEffect(() => {
    setIsPending(false);
  }, [contract]);

  const successCount = contract?.stats?.successCount || 0;
  const failCount = contract?.stats?.failCount || 0;
  const isRunning = Number(contract.state) !== 0;
  const iAmSeller = contract.seller === address;

  return (
    <Container
      ratio={ratio}
      opacity={isRunning ? 0.5 : 1}
      data-rh={isRunning ? getContractState(contract) : null}
    >
      <ContractValue onClick={() => window.openLink(explorerUrl)}>
        {abbreviateAddress(contract.id)} <IconExternalLink width={'1.4rem'} />
      </ContractValue>
      <Value>{formatPrice(contract.price, symbol)}</Value>
      <Value>{formatDuration(contract.length)}</Value>
      <Value>{formatSpeed(contract.speed)}</Value>
      <Value>
        <ProgressBarWithLabels
          key={'stats'}
          completed={successCount}
          remaining={failCount}
        />
      </Value>
      {contract.inProgress ? (
        <Value>
          <Spinner size="25px" /> Purchasing..
        </Value>
      ) : !isRunning ? (
        <Value>
          <ActionButton
            disabled={!allowSendTransaction || iAmSeller}
            data-rh={iAmSeller ? 'You are seller' : null}
            onClick={e => {
              e.stopPropagation();
              onPurchase(contract);
            }}
          >
            Purchase
          </ActionButton>
        </Value>
      ) : (
        <></>
      )}
    </Container>
  );
}

export default withContractsRowState(MarketplaceRow);
