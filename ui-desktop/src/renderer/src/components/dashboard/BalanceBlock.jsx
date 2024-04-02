import React, { useState, useContext } from 'react';
import withBalanceBlockState from '../../store/hocs/withBalanceBlockState';
import { EtherIcon } from '../icons/EtherIcon';
import { LumerinLogoFull } from '../icons/LumerinLogoFull';
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
import Spinner from '../common/Spinner';
import { ToastsContext } from '../toasts';

const WalletBalance = ({
  lmrBalance,
  lmrBalanceUSD,
  ethBalance,
  ethBalanceUSD,
  symbol,
  symbolEth
}) => (
  <BalanceContainer>
    <CoinsRow>
      <Primary data-testid="lmr-balance">
        <Balance
          currency={symbol}
          value={lmrBalance}
          icon={
            <LumerinLogoFull style={{ color: 'white', height: "2rem"}}/> 
          }
          equivalentUSD={lmrBalanceUSD}
          maxSignificantFractionDigits={0}
        />
      </Primary>
      <Primary data-testid="eth-balance">
        <Balance
          currency={symbolEth}
          value={ethBalance}
          icon={<EtherIcon size="3.3rem" />}
          equivalentUSD={ethBalanceUSD}
          maxSignificantFractionDigits={5}
        />
      </Primary>
    </CoinsRow>
  </BalanceContainer>
);

const BalanceBlock = ({
  lmrBalance,
  lmrBalanceUSD,
  ethBalance,
  ethBalanceUSD,
  sendDisabled,
  sendDisabledReason,
  recaptchaSiteKey,
  faucetUrl,
  showFaucet,
  walletAddress,
  onTabSwitch,
  symbol,
  symbolEth,
  client
}) => {
  const handleTabSwitch = e => {
    e.preventDefault();
    onTabSwitch(e.target.dataset.modal);
  };

  const claimFaucet = e => {
    e.preventDefault();
    const url = new URL(faucetUrl);
    url.searchParams.set('address', walletAddress);
    window.open(url);
  };

  return (
    <GlobalContainer>
      <Container>
        <SecondaryContainer>
          <WalletBalance
            {...{
              lmrBalance,
              lmrBalanceUSD,
              ethBalance,
              ethBalanceUSD,
              symbol,
              symbolEth
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
            <Btn
              data-modal="send"
              data-disabled={true}
              data-rh={sendDisabledReason}
              data-testid="send-btn"
              onClick={sendDisabled ? null : handleTabSwitch}
              block
            >
              Send
            </Btn>

            {showFaucet && (
              <BtnAccent
                data-modal="claim"
                onClick={claimFaucet}
                data-rh={`Payout from the faucet is 2 ${symbol} and 0.01 ${symbolEth} per day.\n
          Wallet addresses are limited to one request every 24 hours.`}
                block
              >
                Get Tokens
              </BtnAccent>
            )}
          </BtnRow>
        </SecondaryContainer>
      </Container>
    </GlobalContainer>
  );
};

export default withBalanceBlockState(BalanceBlock);
