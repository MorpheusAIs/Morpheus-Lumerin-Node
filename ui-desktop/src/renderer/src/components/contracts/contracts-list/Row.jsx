import React, { useState, useEffect, useContext } from 'react';
import { IconCircle } from '@tabler/icons';
import {
  IconTriangleInverted,
  IconTriangle,
  IconAlertTriangle
} from '@tabler/icons';
import { ToastsContext } from '../../toasts';
import styled from 'styled-components';

import withContractsRowState from '../../../store/hocs/withContractsRowState';

import { CLOSEOUT_TYPE, CONTRACT_STATE } from '../../../enums';
import Spinner from '../../common/Spinner';
import { ClockIcon } from '../../icons/ClockIcon';
import theme from '../../../ui/theme';
import {
  formatDuration,
  formatSpeed,
  formatTimestamp,
  formatPrice,
  convertLmrToBtc,
  getContractState,
  getContractEndTimestamp,
  getContractRewardBtcPerTh,
  formatExpNumber
} from '../utils';
import {
  ActionButton,
  ActionButtons,
  SmallAssetContainer
} from './ContractsRow.styles';
import ContractActions from '../../common/ContractActions';
import ProgressBarWithLabels from '../../common/ProgressBar';

const Container = styled.div`
  padding: 1.2rem 0;
  display: grid;
  grid-template-columns: ${p => p.ratio.map(x => `${x}fr`).join(' ')};
  text-align: center;
  box-shadow: 0 -1px 0 0 ${p => p.theme.colors.lightShade} inset;
  cursor: pointer;
  height: 66px;
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

const STATE_COLOR = {
  [CONTRACT_STATE.Running]: theme.colors.warning,
  [CONTRACT_STATE.Avaliable]: theme.colors.success
};

function Row({
  contract,
  cancel,
  edit,
  deleteContract,
  address,
  ratio,
  explorerUrl,
  allowSendTransaction,
  lmrRate,
  btcRate,
  symbol,
  converters,
  selectedCurrency,
  underProfitContracts
}) {
  // TODO: Add better padding
  const context = useContext(ToastsContext);
  const [isPending, setIsPending] = useState(false);
  const speed = contract.futureTerms?.speed || contract.speed;
  const length = contract.futureTerms?.length || contract.length;
  const price = contract.futureTerms?.price || contract.price;
  const limit = contract.futureTerms?.limit || contract.limit;
  const profitTarget =
    contract.futureTerms?.profitTarget || contract.profitTarget;
  const underProfit = underProfitContracts.find(x => x.id == contract.id);

  useEffect(() => {
    setIsPending(false);
  }, [contract]);

  const handleCancel = closeOutType => {
    setIsPending(true);
    cancel({
      contractId: contract.id,
      walletAddress: contract.seller,
      closeOutType
    })
      .catch(e => {
        const action =
          closeOutType === CLOSEOUT_TYPE.Claim
            ? 'claim funds'
            : 'close contract';
        context.toast('error', `Failed to ${action}: ${e.message}`);
      })
      .finally(() => {
        setIsPending(false);
      });
  };

  const handleEdit = () => {
    edit({
      ...contract,
      price,
      length,
      speed,
      limit,
      profitTarget
    });
  };

  const handleDelete = () => {
    setIsPending(true);
    deleteContract({
      contractId: contract.id,
      walletAddress: contract.seller,
      deleteContract: true
    })
      .catch(e => {
        context.toast('error', `Failed to delete contract: ${e.message}`);
      })
      .finally(() => {
        setIsPending(false);
      });
  };

  const contractEndTimestamp = getContractEndTimestamp(contract);

  const isContractExpired = () => {
    return (
      contract.state !== CONTRACT_STATE.Avaliable &&
      Date.now() > contractEndTimestamp
    );
  };

  const getClaimDisabledReason = () => {
    if (contract.balance === '0') {
      return 'Balance is empty';
    }
    return null;
  };

  const isClaimBtnDisabled = () => {
    if (!allowSendTransaction) {
      return true;
    }
    return contract.balance === '0';
  };

  const getClockColor = contract => {
    return STATE_COLOR[contract.state];
  };

  const handleActionSelector = value => {
    if (value === 1) {
      return window.openLink(explorerUrl);
    }
    if (value === 2) {
      return handleCancel(CLOSEOUT_TYPE.Claim);
    }
    if (value === 3) {
      return handleEdit();
    }
    if (value === 4) {
      return handleCancel(CLOSEOUT_TYPE.Close);
    }
    if (value === 5) {
      return handleDelete();
    }
  };

  const btcPerThReward = getContractRewardBtcPerTh(contract, btcRate, lmrRate);

  const successCount = contract?.stats?.successCount || 0;
  const failCount = contract?.stats?.failCount || 0;
  const isLmrSelected = (balance, selectedCurrency) => {
    return balance ? balance === 'LMR' : selectedCurrency === 'LMR';
  };

  return (
    <Container ratio={ratio}>
      {/* <Value>
        {formatTimestamp(contract.timestamp, timer, contract.state)}
      </Value> */}
      {contract.state === CONTRACT_STATE.Avaliable ? (
        <SmallAssetContainer data-rh={getContractState(contract)}>
          <IconCircle
            data-rh={getContractState(contract)}
            fill={getClockColor(contract)}
            size="3rem"
            stroke="currentColor"
          ></IconCircle>
        </SmallAssetContainer>
      ) : (
        <SmallAssetContainer data-rh={getContractState(contract)}>
          <ClockIcon size="3rem" fill={getClockColor(contract)} />
        </SmallAssetContainer>
      )}

      <Value style={{ flexDirection: 'row', alignItems: 'center' }}>
        <div style={{ height: '18px', marginRight: '3px' }}>
          {isLmrSelected(converters.price, selectedCurrency)
            ? `${formatPrice(contract.price, 'LMR')}`
            : `${convertLmrToBtc(contract.price, btcRate, lmrRate).toFixed(
                10
              )} BTC`}
        </div>
        {underProfit ? (
          <div>
            {+underProfit.zeroRate < 1 ? (
              <IconTriangleInverted
                data-rh={
                  'Negative profit is ' +
                  ((1 - underProfit.zeroRate) * 100).toFixed(0) +
                  '%'
                }
                style={{ width: '14px', color: 'darkred', fill: 'darkred' }}
              ></IconTriangleInverted>
            ) : (
              <IconTriangle
                data-rh={
                  'Excess profit is ' +
                  ((+underProfit.zeroRate - 1) * 100).toFixed(0) +
                  '%'
                }
                style={{ width: '14px', color: 'darkgreen', fill: 'darkgreen' }}
              ></IconTriangle>
            )}
          </div>
        ) : (
          <></>
        )}
      </Value>
      <Value
      // data-rh={`${formatExpNumber(btcPerThReward)} BTC/TH/day`}
      >
        {formatExpNumber(btcPerThReward)} BTC/TH/day
      </Value>
      <Value>{formatDuration(length)}</Value>
      <Value>{formatSpeed(speed)}</Value>
      <Value>
        <ProgressBarWithLabels
          key={'stats'}
          completed={successCount}
          remaining={failCount}
        />
      </Value>
      <Value>
        {isLmrSelected(converters.claimable, selectedCurrency)
          ? `${formatPrice(contract.balance, 'LMR')}`
          : `${convertLmrToBtc(contract.balance, btcRate, lmrRate).toFixed(
              contract.balance == 0 ? 0 : 10
            )} BTC`}
      </Value>
      {contract.seller === address &&
        (isPending ? (
          <Value>
            <Spinner size="25px" />
          </Value>
        ) : contract.isDeploying ? (
          <Value>
            <Spinner size="25px" /> Deploying...
          </Value>
        ) : (
          <ActionButtons>
            <ContractActions
              onChange={e => handleActionSelector(e.value)}
              options={[
                {
                  label: 'Actions',
                  value: 0,
                  hidden: true
                },
                {
                  label: 'View',
                  value: 1
                },
                {
                  label: 'Claim Funds',
                  value: 2,
                  disabled: !allowSendTransaction || isClaimBtnDisabled()
                },
                {
                  label: 'Edit',
                  value: 3,
                  disabled: !allowSendTransaction
                },
                {
                  label: 'Close',
                  value: 4,
                  disabled: !(allowSendTransaction && isContractExpired())
                },
                {
                  label: 'Archive',
                  value: 5,
                  disabled: !allowSendTransaction || contract.isDead,
                  message:
                    getContractState(contract) === CONTRACT_STATE.Running
                      ? 'Will not affect hashrate delivery of running contract'
                      : null
                }
              ]}
              value={0}
              id="range"
            />
          </ActionButtons>
        ))}
    </Container>
  );
}

export default withContractsRowState(Row);
