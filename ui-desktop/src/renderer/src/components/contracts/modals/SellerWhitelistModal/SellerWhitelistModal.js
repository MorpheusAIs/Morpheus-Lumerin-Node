import React, { useEffect, useState } from 'react';
import { List as RVList, AutoSizer } from 'react-virtualized';
import {
  Modal,
  Body,
  TitleWrapper,
  Title,
  Subtitle,
  CloseModal,
  RightBtn,
  Row
} from '../CreateContractModal.styles';
import { withClient } from '../../../../store/hocs/clientContext';

function SellerWhitelistModal(props) {
  const { isActive, close, formUrl } = props;

  const handleClose = e => {
    close(e);
  };
  const handlePropagation = e => e.stopPropagation();

  if (!isActive) {
    return <></>;
  }

  return (
    <Modal onClick={handleClose}>
      <Body height={'300px'} onClick={handlePropagation}>
        {CloseModal(handleClose)}
        <TitleWrapper style={{ height: 'auto' }}>
          <Title>You are not whitelisted as Seller</Title>
        </TitleWrapper>
        <p style={{ textAlign: 'justify', marginTop: '10px' }}>
          Lumerin is hand-selecting the first few hashrate sellers for mainnet
          in order to ensure high quality contracts in this initial launch phase
          of the marketplace.
        </p>
        <p style={{ textAlign: 'justify' }}>
          If you are interested in becoming a seller please fill out the form
        </p>

        <Row style={{ justifyContent: 'center' }}>
          <RightBtn
            type="submit"
            onClick={() => {
              window.open(formUrl, '_blank');
              close();
            }}
          >
            Open Form
          </RightBtn>
        </Row>
      </Body>
    </Modal>
  );
}

export default withClient(SellerWhitelistModal);
