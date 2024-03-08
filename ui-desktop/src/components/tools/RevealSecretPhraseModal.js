import PropTypes from 'prop-types';
import styled from 'styled-components';
import React, { useState } from 'react';

import { Modal } from '../common';
import {
  Container,
  Message,
  Row,
  DismissBtn,
  ConfirmBtn
} from './ConfirmModal.styles';
import { Input } from './common';

const Mnemonic = styled.div`
  padding: 10px 0;
  font-size: 1.8rem;
  font-weight: 600;
  line-height: 2;
  text-align: center;
  color: ${p => p.theme.colors.primary};
  word-spacing: 1.6rem;
`;

const RevealSecretPhraseModal = props => {
  // eslint-disable-next-line complexity
  const {
    onRequestClose,
    onShowMnemonic,
    isOpen,
    mnemonic,
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
      title="Reveal Secret Recovery Phrase"
    >
      <Container data-testid="confirm-proxy-config-modal">
        {mnemonic ? (
          <Mnemonic>{mnemonic}</Mnemonic>
        ) : (
          <>
            <Message>
              The Secret Recovery Phrase provides full access to your wallet and
              funds.
            </Message>
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
          </>
        )}

        <Row style={{ justifyContent: 'space-around' }}>
          {mnemonic ? (
            <ConfirmBtn onClick={() => copyToClipboard(mnemonic)}>
              Copy to clipboard
            </ConfirmBtn>
          ) : (
            <ConfirmBtn
              disabled={!password}
              onClick={() => onShowMnemonic(password)}
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

RevealSecretPhraseModal.propTypes = {
  onRequestClose: PropTypes.func.isRequired,
  isOpen: PropTypes.bool.isRequired,
  onShowMnemonic: PropTypes.func.isRequired,
  copyToClipboard: PropTypes.func.isRequired
};

export default RevealSecretPhraseModal;
