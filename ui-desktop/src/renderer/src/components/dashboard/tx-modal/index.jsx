import React from 'react';
import styled from 'styled-components';

import { ReceiveForm } from './ReceiveForm';
import { SendForm } from './SendForm';
import { SuccessForm } from './SuccessForm';
import withTransactionModalState from '../../../store/hocs/withTransactionModalState';

const Modal = styled.div`
  display: flex;
  flex-direction: column;
  position: fixed;
  z-index: 10;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  overflow: auto;
  background-color: rgb(0, 0, 0);
  background-color: rgba(0, 0, 0, 0.4);
  align-items: center;
  justify-content: center;
`;

const Body = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  position: fixed;
  z-index: 20;
  background-color: ${p => p.theme.colors.morLight};
  width: 400px;
  height: 500px;
  border-radius: 5px;
  padding: 2rem 3rem 2rem 3rem;
`;

function TransactionModal(props) {
  const handlePropagation = e => e.stopPropagation();

  // Form state (amount, destination, selected currency, validation errors) all
  // lives in withTransactionModalState now. It used to be split between here
  // and the HOC, with the two copies drifting out of sync — the modal tracked a
  // `destinationAddress` that the submit path never actually read.
  const handleClose = () => {
    props.resetForm();
    props.onRequestClose();
  };

  if (!props.activeTab) {
    return <></>;
  }

  return (
    <Modal onClick={handleClose}>
      <Body onClick={handlePropagation}>
        {props.activeTab === 'receive' && (
          <ReceiveForm {...props} onRequestClose={handleClose} />
        )}
        {props.activeTab === 'send' && (
          <SendForm {...props} onRequestClose={handleClose} />
        )}
        {props.activeTab === 'success' && (
          <SuccessForm
            {...props}
            onRequestClose={handleClose}
            amountInput={props.coinAmount}
            symbol={props.selectedCurrency?.label}
          />
        )}
      </Body>
    </Modal>
  );
}

export default withTransactionModalState(TransactionModal);
