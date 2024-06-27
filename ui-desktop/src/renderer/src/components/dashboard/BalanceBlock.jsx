import React, { useState, useContext } from 'react';
import withBalanceBlockState from '../../store/hocs/withBalanceBlockState';
import { EtherIcon } from '../icons/EtherIcon';
import { LumerinLogoFull } from '../icons/LumerinLogoFull';
import { toUSD } from '../../store/utils/syncAmounts';
import { Balance } from './Balance';
import {
  WalletBalanceHeader,
  Btn,
  BtnAccent,
  BtnRow,
  SecondaryContainer,
  Container,
  Primary,
  CoinsRow,
  BalanceContainer,
  GlobalContainer
} from './BalanceBlock.styles';

const WalletBalance = ({
  eth, mor
}) => (
  <BalanceContainer>
    <CoinsRow>
      <Primary data-testid="mor-balance">
        <Balance
          currency={mor.symbol}
          value={+mor.value}
          icon={
            <LumerinLogoFull style={{ color: 'white', height: "2rem"}}/> 
          }
          equivalentUSD={mor.usd}
          maxSignificantFractionDigits={0}
        />
      </Primary>
      <Primary data-testid="eth-balance">
        <Balance
          currency={eth.symbol}
          value={+eth.value}
          icon={<EtherIcon size="3.3rem" />}
          equivalentUSD={eth.usd}
          maxSignificantFractionDigits={5}
        />
      </Primary>
    </CoinsRow>
  </BalanceContainer>
);

const BalanceBlock = ({
  onTabSwitch,
  ...props
}) => {
  const handleTabSwitch = e => {
    e.preventDefault();
    onTabSwitch(e.target.dataset.modal);
  };

  return (
    <GlobalContainer>
      <Container>
        <SecondaryContainer>
          <WalletBalance
            {...{
              eth: props.eth,
              mor: props.mor
            }}
          />
          <BtnRow>
            <BtnAccent
              data-modal="receive"
              data-testid="receive-btn"
              onClick={handleTabSwitch}
              block
            >
              Receive
            </BtnAccent>
            <BtnAccent
              data-modal="send"
              data-testid="send-btn"
              onClick={handleTabSwitch}
              block
            >
              Send
            </BtnAccent>
          </BtnRow>
        </SecondaryContainer>
      </Container>
    </GlobalContainer>
  );
};

export default withBalanceBlockState(BalanceBlock);
