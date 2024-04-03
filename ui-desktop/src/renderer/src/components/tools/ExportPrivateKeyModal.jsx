import PropTypes from 'prop-types';
import styled from 'styled-components';
import React, { useEffect } from 'react';

import { Modal, BaseBtn } from '../common';
import {
  Container,
  Message,
  DismissBtn,
  ConfirmBtn,
  Row
} from './ConfirmModal.styles';
import { useState } from 'react';
import { Input } from './common';

const PrivateKey = styled.div`
  word-break: break-word;
  color: black;
  opacity: 0.8;
  margin-bottom: 2.4rem;
`;

const ExportPrivateKeyModal = props => {
  // eslint-disable-next-line complexity
  const {
    onRequestClose,
    onExportPrivateKey,
    isOpen,
    privateKey,
    copyToClipboard
  } = props;
  const [password, setPassword] = useState('');

  const closeWrapper = () => {
    setPassword('');
    onRequestClose();
  };

  return (
    <Modal
      shouldReturnFocusAfterClose={false}
      onRequestClose={closeWrapper}
      styleOverrides={{
        width: 450,
        top: '35%'
      }}
      variant="primary"
      isOpen={isOpen}
      title="Show Private Key"
    >
      <Container data-testid="confirm-proxy-config-modal">
        <>
          <Message>
            <b>Warning</b>: Never disclose this key. Anyone with your private
            keys can steal any assets held in your account.
          </Message>
          {privateKey ? (
            <PrivateKey>{privateKey}</PrivateKey>
          ) : (
            <Message>
              <div>Enter password to continue: </div>
              <Input
                type={'password'}
                placeholder={'Make sure nobody is looking'}
                onChange={e => {
                  setPassword(e.value);
                }}
                value={password}
              />
            </Message>
          )}
        </>
        <Row style={{ justifyContent: 'space-around' }}>
          {privateKey ? (
            <ConfirmBtn onClick={() => copyToClipboard(privateKey)}>
              Copy to clipboard
            </ConfirmBtn>
          ) : (
            <ConfirmBtn
              disabled={!password}
              onClick={() => onExportPrivateKey(password)}
            >
              Show
            </ConfirmBtn>
          )}
          <DismissBtn onClick={closeWrapper}>Close</DismissBtn>
        </Row>
      </Container>
    </Modal>
  );
};

ExportPrivateKeyModal.propTypes = {
  onRequestClose: PropTypes.func.isRequired,
  isOpen: PropTypes.bool.isRequired,
  copyToClipboard: PropTypes.func.isRequired
};

export default ExportPrivateKeyModal;
