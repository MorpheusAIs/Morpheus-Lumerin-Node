import React, { useContext } from 'react';
import styled from 'styled-components';
import QRCode from 'qrcode.react';
import { IconCopy } from '@tabler/icons-react';

import { ToastsContext } from '../../toasts';
import BackIcon from '../../icons/BackIcon';
import { BaseBtn } from '../../common';
import {
  HeaderWrapper,
  Header,
  BackBtn,
  Footer,
  FooterRow,
  FooterLabel,
  FooterBlock,
  FooterSublabel,
} from './common.styles';
import { copyWalletAddress } from '../../../utils/clipboard';
import { BtnAccent } from '../BalanceBlock.styles';
const QRContainer = styled.div`
  display: flex;
  align-self: center;
  padding: 3rem 1.6rem 1.6rem 1.6rem;

  & canvas {
    display: block;
  }
`;

export const Divider = styled.div`
  margin-top: 5px;
  width: 100%;
  height: 1px;
  background: rgba(255, 255, 255, 0.12);
`;

const AddressBlock = styled(FooterBlock)`
  flex: 1;
  min-width: 0;
`;

const FullAddress = styled(FooterSublabel)`
  user-select: all;
  overflow-wrap: anywhere;
  line-height: 1.6;
`;

const CopyBtn = styled(BaseBtn)`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  align-self: center;
  flex: 0 0 auto;
  width: 4rem;
  height: 4rem;
  background-color: transparent;
  color: ${(p) => p.theme.colors.morMain};
  border-radius: 0.6rem;
  margin-inline-start: 1rem;

  &:hover:not(:disabled) {
    background: rgba(32, 220, 142, 0.1);
  }

  &:focus-visible {
    outline: 2px solid ${(p) => p.theme.colors.morMain};
    outline-offset: 2px;
  }
`;

export function ReceiveForm({
  activeTab,
  address,
  onRequestClose,
  copyToClipboard,
  explorerUrl,
  eth,
  mor,
}) {
  const context = useContext(ToastsContext);

  const handleCopyToClipboard = () =>
    copyWalletAddress(address, copyToClipboard, context.toast);

  if (!activeTab) {
    return <></>;
  }

  return (
    <>
      <HeaderWrapper>
        <BackBtn
          data-modal="send"
          aria-label="Close receive window"
          onClick={onRequestClose}
        >
          <BackIcon size="2.4rem" fill="white" />
        </BackBtn>
        <Header>You are receiving</Header>
      </HeaderWrapper>
      <QRContainer>
        <QRCode value={address} bgColor="transparent" fgColor="#20dc8e" />
      </QRContainer>
      <Footer style={{ padding: '0 2rem 2rem' }}>
        <FooterRow>
          <AddressBlock>
            <FooterLabel>{mor.symbol} Address</FooterLabel>
            <FullAddress as="span">{address}</FullAddress>
          </AddressBlock>
          <CopyBtn
            aria-label="Copy wallet address"
            title="Copy wallet address"
            disabled={!address}
            onClick={handleCopyToClipboard}
          >
            <IconCopy size={20} aria-hidden="true" />
          </CopyBtn>
        </FooterRow>
        <FooterLabel>{mor.symbol} Balance</FooterLabel>
        <FooterSublabel>
          {mor.value.toFixed(6)} {mor.symbol} ≈ {mor.usd || 0}
        </FooterSublabel>
        <FooterLabel>{eth.symbol} Balance</FooterLabel>
        <FooterSublabel>
          {eth.value.toFixed(6)} {eth.symbol} ≈ {eth.usd || 0}
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
