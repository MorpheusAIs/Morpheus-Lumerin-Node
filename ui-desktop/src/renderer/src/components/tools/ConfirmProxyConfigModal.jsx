import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

import { Modal, BaseBtn } from '../common';
import {
  Container,
  Message,
  DismissBtn,
  ConfirmBtn,
  Row
} from './ConfirmModal.styles';

export const RestartNowBtn = styled(ConfirmBtn)``;
export const LaterBtn = styled(DismissBtn)``;

const ConfirmProxyConfigModal = props => {
  // eslint-disable-next-line complexity
  const { onRequestClose, onConfirm, onLater, isOpen, message, title } = props;
  return (
    <Modal
      shouldReturnFocusAfterClose={false}
      onRequestClose={onRequestClose}
      styleOverrides={{
        width: 450,
        top: '35%'
      }}
      variant="primary"
      isOpen={isOpen}
      title={title || 'Proxy-router restart'}
    >
      <Container data-testid="confirm-proxy-config-modal">
        {message || (
          <>
            <Message>
              You are going to restart Proxy Router. It may affect your running
              contracts.
            </Message>
            <Message>You can restart right now or later.</Message>
          </>
        )}
        <Row style={{ justifyContent: 'space-around' }}>
          <RestartNowBtn onClick={onConfirm}>Restart now</RestartNowBtn>
          <LaterBtn onClick={onLater}>Later</LaterBtn>
        </Row>
      </Container>
    </Modal>
  );
};

ConfirmProxyConfigModal.propTypes = {
  onRequestClose: PropTypes.func.isRequired,
  onConfirm: PropTypes.func.isRequired,
  onLater: PropTypes.func.isRequired,
  isOpen: PropTypes.bool.isRequired
};

export default ConfirmProxyConfigModal;
