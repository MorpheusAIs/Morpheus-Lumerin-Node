import React, { useContext } from 'react';
import styled from 'styled-components';
import QRCode from 'qrcode.react';

import { ToastsContext } from '../../toasts';
import BackIcon from '../../icons/BackIcon';
import CopyIcon from '../../icons/CopyIcon';
import { BaseBtn } from '../../common';
import {
  HeaderWrapper,
  Header,
  BackBtn,
  Footer,
  FooterRow,
  FooterLabel,
  FooterBlock,
  FooterSublabel
} from './common.styles';
import { abbreviateAddress } from '../../../utils';
import { BtnAccent } from '../BalanceBlock.styles';
import { useState } from 'react';
const QRContainer = styled.div`
  display: flex;
  align-self: center;
  padding: 3rem 1.6rem 1.6rem 1.6rem;

  & canvas {
    display: block;
  }
`;

export const Divider = styled.div`
  margin-top: 5px
  width:100%;
  height: 0px;
  border: 0.5px solid rgba(0, 0, 0, 0.25);`;

const CopyBtn = styled(BaseBtn)`
  background-color: transparent;
  border-radius: 5px;
  border: 1px;
  padding: 0 !important;
  margin: 0 !important;
  :hover {
    padding: 0 !important;
    margin: 0 !important;
  }
`;

export function ReceiveForm({
  activeTab,
  address,
  onRequestClose,
  lmrBalanceUSD,
  lmrBalanceWei,
  ethBalanceUSD,
  ethBalanceWei,
  copyToClipboard,
  explorerUrl,
  symbol,
  symbolEth
}) {
  const context = useContext(ToastsContext);

  const handleCopyToClipboard = () => {
    copyToClipboard(address);
    context.toast('info', 'Address copied to clipboard', {
      autoClose: 1500
    });
  };

  if (!activeTab) {
    return <></>;
  }

  return (
    <>
      <HeaderWrapper>
        <BackBtn data-modal="send" onClick={onRequestClose}>
          <BackIcon size="2.4rem" fill="white" />
        </BackBtn>
        <Header>You are receiving</Header>
      </HeaderWrapper>
      <QRContainer>
        <QRCode value={address} bgColor='transparent' fgColor='#20dc8e' />
      </QRContainer>
      <Footer>
        <FooterRow>
          <FooterBlock>
            <FooterLabel>{symbol} Address</FooterLabel>
            <FooterSublabel>{abbreviateAddress(address, 8)}</FooterSublabel>
          </FooterBlock>
          <CopyBtn onClick={handleCopyToClipboard}>
            <CopyIcon fill="#20dc8e" size="3.8rem"/>
          </CopyBtn>
        </FooterRow>
        <FooterLabel>{symbol} Balance</FooterLabel>
        <FooterSublabel>
          {lmrBalanceWei.toFixed(6)} {symbol} ≈ {lmrBalanceUSD || 0}
        </FooterSublabel>
        <FooterLabel>{symbolEth} Balance</FooterLabel>
        <FooterSublabel>
          {ethBalanceWei.toFixed(6)} {symbolEth} ≈ {ethBalanceUSD || 0}
        </FooterSublabel>
        <Divider style={{ margin: '2rem 0' }} />
        <BtnAccent
          style={{ marginBottom: '5px' }}
          onClick={() => {
            window.openLink(explorerUrl);
          }}
        >
          View account at {explorerUrl ? new URL(explorerUrl).hostname : ''}
        </BtnAccent>
      </Footer>
    </>
  );
}
