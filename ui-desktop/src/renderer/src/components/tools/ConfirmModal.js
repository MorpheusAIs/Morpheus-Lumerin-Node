import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

import { Modal, BaseBtn } from '../common';
import { Container, Message, Button } from './ConfirmModal.styles';

export default class ConfirmModal extends React.Component {
  static propTypes = {
    onRequestClose: PropTypes.func.isRequired,
    onConfirm: PropTypes.func.isRequired,
    isOpen: PropTypes.bool.isRequired
  };

  // eslint-disable-next-line complexity
  render() {
    const { onRequestClose, onConfirm, isOpen } = this.props;

    return (
      <Modal
        shouldReturnFocusAfterClose={false}
        onRequestClose={onRequestClose}
        styleOverrides={{
          width: 304,
          top: '35%'
        }}
        variant="primary"
        isOpen={isOpen}
        title="Confirm Rescan"
      >
        <Container data-testid="confirm-modal">
          <Message>
            Rescanning your transactions will close and re-open the app. You
            will need to log back in.
          </Message>
          <Button onClick={onConfirm}>Confirm and Log Out</Button>
        </Container>
      </Modal>
    );
  }
}
