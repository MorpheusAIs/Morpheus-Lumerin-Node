import React, { useContext } from 'react';
import { List as RVList, AutoSizer } from 'react-virtualized';
import {
  Modal,
  Body,
  TitleWrapper,
  Title,
  Subtitle,
  CloseModal
} from '../CreateContractModal.styles';
import ArchiveRow from './ArchiveRow';
import { withClient } from '../../../../store/hocs/clientContext';
import { CLOSEOUT_TYPE } from '../../../../enums';
import { ToastsContext } from '../../../../components/toasts';

function ArchiveModal(props) {
  const { isActive, close, address, deletedContracts, client } = props;
  const context = useContext(ToastsContext);
  const handleClose = e => {
    close(e);
  };
  const handlePropagation = e => e.stopPropagation();

  if (!isActive) {
    return <></>;
  }

  const handleClaim = contractId => {
    client.lockSendTransaction();
    return client
      .cancelContract({
        contractId: contractId,
        walletAddress: address,
        closeOutType: CLOSEOUT_TYPE.Claim
      })
      .catch(e => {
        context.toast('error', `Failed to claim funds: ${e.message}`);
      })
      .finally(() => {
        client.unlockSendTransaction();
      });
  };

  const handleRestore = contract => {
    client.lockSendTransaction();
    return client
      .setDeleteContractStatus({
        contractId: contract.id,
        walletAddress: contract.seller,
        deleteContract: false
      })
      .finally(() => {
        client.unlockSendTransaction();
      });
  };

  const rowRenderer = deletedContracts => ({ key, index, style }) => (
    <div style={style}>
      <ArchiveRow
        style={style}
        key={deletedContracts[index].id}
        contract={deletedContracts[index]}
        handleRestore={handleRestore}
        handleClaim={handleClaim}
      />
    </div>
  );

  return (
    <Modal onClick={handleClose}>
      <Body
        height={'500px'}
        width={'70%'}
        maxWidth={'100%'}
        onClick={handlePropagation}
      >
        {CloseModal(handleClose)}
        <TitleWrapper>
          <Title>Archived contracts</Title>
        </TitleWrapper>
        <AutoSizer width={400}>
          {({ width, height }) => (
            <RVList
              rowRenderer={rowRenderer(deletedContracts)}
              rowHeight={50}
              rowCount={deletedContracts.length}
              height={height || 500} // defaults for tests
              width={width || 500} // defaults for tests
            />
          )}
        </AutoSizer>
      </Body>
    </Modal>
  );
}

export default withClient(ArchiveModal);
