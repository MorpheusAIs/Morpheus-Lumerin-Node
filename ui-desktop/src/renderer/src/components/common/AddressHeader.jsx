import React, { useContext } from 'react';
import styled from 'styled-components';
import { ToastsContext } from '../toasts';
import { BaseBtn } from '.';
import { abbreviateAddress } from '../../utils';
import { IconCopy } from '@tabler/icons-react';
import { copyWalletAddress } from '../../utils/clipboard';

const Container = styled.header`
  padding: 1.6rem;
  display: flex;
  align-items: center;
  justify-content: flex-start;
`;

const AddressContainer = styled.div`
  display: flex;
  align-items: center;
  min-width: 0;
  width: 100%;
  gap: 0.8rem;
  padding: 0.4rem 0.6rem 0.4rem 1.2rem;
  border-radius: 0.8rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: ${(p) => p.theme.colors.light};
`;

const Address = styled.div`
  font-size: 1.3rem;
  flex: 1;
  min-width: 0;
  cursor: default;
  font-weight: 600;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
`;

const CopyButton = styled(BaseBtn)`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 3.6rem;
  height: 3.6rem;
  border-radius: 0.6rem;
  color: ${(p) => p.theme.colors.morMain};

  &:hover:not(:disabled) {
    background: rgba(32, 220, 142, 0.1);
  }

  &:focus-visible {
    outline: 2px solid ${(p) => p.theme.colors.morMain};
    outline-offset: 2px;
  }
`;

export const AddressHeader = ({ copyToClipboard, address }) => {
  const context = useContext(ToastsContext);

  const onCopyToClipboardClick = () =>
    copyWalletAddress(address, copyToClipboard, context.toast);

  return (
    <Container className="sidebar-address">
      <AddressContainer>
        <Address data-testid="address" title={address}>
          {address ? abbreviateAddress(address, 5) : 'Wallet not connected'}
        </Address>
        <CopyButton
          aria-label="Copy wallet address"
          title="Copy wallet address"
          disabled={!address}
          onClick={onCopyToClipboardClick}
        >
          <IconCopy size={18} aria-hidden="true" />
        </CopyButton>
      </AddressContainer>
    </Container>
  );
};
