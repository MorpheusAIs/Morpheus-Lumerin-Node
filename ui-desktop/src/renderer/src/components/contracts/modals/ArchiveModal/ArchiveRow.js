import React, { useState } from 'react';
import { abbreviateAddress } from '../../../../utils';
import { IconTrashOff } from '@tabler/icons';
import styled from 'styled-components';

import { formatDuration, formatSpeed, formatPrice } from '../../utils';
import withContractsRowState from '../../../../store/hocs/withContractsRowState';
import Spinner from '../../../common/Spinner';
import { fromTokenBaseUnitsToLMR } from '../../../../utils/coinValue';

const RowContainer = styled.div`
  padding: 1.2rem 0;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 0.5fr 1fr;
  text-align: center;
  box-shadow: 0 -1px 0 0 ${p => p.theme.colors.lightShade} inset;
  color: ${p => p.theme.colors.primary}
  height: 50px;
`;

const ContractValue = styled.label`
  display: flex;
  padding: 0 1.5rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: ${p => p.theme.colors.primary};
  cursor: pointer;
  text-decoration: underline;
  flex-direction: row;
  gap: 5px;
`;

const Circle = styled.div`
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: inline-block;

  background: ${p => p.color};
  color: #fff;
  text-align: center;
`;

const FlexCenter = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`;

function ArchiveRow(props) {
  const { explorerUrl, contract, handleClaim, handleRestore, symbol } = props;

  const [isProcessing, setIsProcessing] = useState(false);

  const wrapRestoreContract = async () => {
    setIsProcessing(true);
    await handleRestore(contract);
    setIsProcessing(false);
  };

  return (
    <RowContainer>
      <FlexCenter>
        <ContractValue onClick={() => window.openLink(explorerUrl)}>
          {abbreviateAddress(contract.id, 4)}
        </ContractValue>
      </FlexCenter>
      <FlexCenter style={{ padding: '0 5px' }}>
        {formatPrice(contract.price, symbol)}
      </FlexCenter>
      <FlexCenter style={{ padding: '0 5px' }}>
        {formatDuration(contract.length)}
      </FlexCenter>
      <FlexCenter style={{ padding: '0 5px' }}>
        {formatSpeed(contract.speed)}
      </FlexCenter>
      <FlexCenter style={{ justifyContent: 'space-evenly' }}>
        <Circle
          data-rh={`${contract?.stats?.successCount || 0} Completed`}
          color={'green'}
        >
          {contract?.stats?.successCount || 0}
        </Circle>
        <Circle
          data-rh={`${contract?.stats?.failCount || 0} Cancelled`}
          color={'red'}
        >
          {contract?.stats?.failCount || 0}
        </Circle>
      </FlexCenter>
      <FlexCenter
        style={{
          justifyContent: contract.balance !== '0' ? 'space-evenly' : null
        }}
      >
        {isProcessing ? (
          <Spinner size="18px" />
        ) : (
          <>
            {contract.balance !== '0' && (
              <div
                data-rh={fromTokenBaseUnitsToLMR(contract.balance) + ' LMR'}
                style={{ cursor: 'pointer' }}
                onClick={() => handleClaim(contract.id)}
              >
                Claim
              </div>
            )}
            <IconTrashOff
              data-rh={`Restore contract`}
              onClick={wrapRestoreContract}
              style={{ cursor: 'pointer' }}
            />
          </>
        )}
      </FlexCenter>
    </RowContainer>
  );
}

export default withContractsRowState(ArchiveRow);
