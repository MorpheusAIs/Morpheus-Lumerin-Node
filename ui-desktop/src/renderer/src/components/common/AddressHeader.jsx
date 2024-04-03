import React, { useContext } from 'react';
import styled from 'styled-components';
import { ToastsContext } from '../toasts';
import { BaseBtn } from '.';
import { abbreviateAddress } from '../../utils';
import { IconCopy } from '@tabler/icons-react';

const Container = styled.header`
  padding: 1.6rem;
  display: flex;
  align-items: center;
  justify-content: flex-start;
`;

const AddressContainer = styled.div`
  display: flex;
  align-items: center;
  background-color: #fff;
  border-radius: 12px;
  border: 2px solid rgba(14, 67, 83, 0.28);
  padding: 0.5rem 1.25rem;
  color: ${p => p.theme.colors.dark};
  opacity: 0.8;

  border-radius: 0.375rem;
  background: rgba(255,255,255, 0.04);
  border-width: 1px;
  border: 1px solid rgba(255, 255, 255, 0.04);
  color: white;
`;

const Address = styled.div`
  font-size: 1.3rem;
  margin-right: 1rem;
  cursor: default;
  border-right: 1px;
  font-weight: 600;
  text-overflow: ellipsis;
  overflow: hidden;
  max-width: 240px;
  @media (min-width: 960px) {
    max-width: 100%;
  }
`;

export const AddressHeader = ({ copyToClipboard, address }) => {
  const context = useContext(ToastsContext);

  const onCopyToClipboardClick = () => {
    copyToClipboard(address);
    context.toast('info', 'Address copied to clipboard', {
      autoClose: 1500
    });
  };

  return (
    <Container className="sidebar-address">
      <AddressContainer>
        <Address data-testid="address">{abbreviateAddress(address, 5)}</Address>
        <IconCopy
          style={{ cursor: 'pointer' }}
          onClick={onCopyToClipboardClick}
        />
      </AddressContainer>
    </Container>
  );
};
