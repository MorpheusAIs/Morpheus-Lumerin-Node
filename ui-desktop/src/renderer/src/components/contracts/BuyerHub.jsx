import React, { useState } from 'react';

import withContractsState from '../../store/hocs/withContractsState';
import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';
import BuyerHubRow from './contracts-list/BuyerHubRow';
import ContractsList from './contracts-list/ContractsList';
import { ContractsRowContainer } from './contracts-list/ContractsRow.styles';

import HistoryModal from './modals/HistoryModal/HistoryModal';
import HashrateModal from './modals/HashrateModal/HashrateModal';
import { IconHistory } from '@tabler/icons';

import styled from 'styled-components';
import { Btn } from '../common';

const HistoryBtn = styled(Btn)`
  margin: 0 0 0 auto;
  font-weight: 700;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 1.6rem;
  padding: 0.4rem 1.1rem 0.4rem 0.9rem;
  box-shadow: none;

  svg {
    margin-right: 4px;
  }
  color: ${p => p.theme.colors.primary};
  background-color: transparent;
`;

function BuyerHub({
  contracts,
  hasContracts,
  copyToClipboard,
  syncStatus,
  activeCount,
  draftCount,
  address,
  client,
  contractsRefresh,
  allowSendTransaction,
  ...props
}) {
  const contractsToShow = contracts.filter(
    x => x.buyer === address && x.seller !== address
  );

  const tabs = [
    { value: 'id', name: 'Contract', ratio: 3 },
    { value: 'timestamp', name: 'Started', ratio: 3 },
    { name: 'Status', ratio: 1 },
    { value: 'price', name: 'Price', ratio: 2 },
    { value: 'length', name: 'Duration', ratio: 2 },
    { value: 'speed', name: 'Speed', ratio: 2 },
    { value: 'action', name: 'Actions', ratio: 2 }
  ];

  const handleContractCancellation = (e, data) => {
    e.preventDefault();

    client.lockSendTransaction();
    return client
      .cancelContract({
        contractId: data.contractId,
        walletAddress: data.walletAddress,
        closeOutType: data.closeOutType
      })
      .finally(() => {
        client.unlockSendTransaction();
      });
  };

  const rowRenderer = (contractsList, ratio) => ({ key, index, style }) => (
    <ContractsRowContainer style={style} key={`${key}-${index}`}>
      <BuyerHubRow
        key={contractsList[index].id}
        data-testid="BuyerHub-row"
        allowSendTransaction={allowSendTransaction}
        contract={contractsList[index]}
        cancel={handleContractCancellation}
        address={address}
        ratio={ratio}
        onGetHashrate={id => {
          setShowHashrateModal(true);
          setContactToShowHashrate(id);
        }}
      />
    </ContractsRowContainer>
  );

  const [isHistoryModalOpen, setIsHistoryModalOpen] = useState(false);
  const [showHashrateModal, setShowHashrateModal] = useState(false);
  const [contactToShowHashrate, setContactToShowHashrate] = useState();

  const contractsWithHistory = contracts.filter(c => c.history.length);
  const showHistory = contractsWithHistory.length;
  const onHistoryOpen = () => setIsHistoryModalOpen(true);

  return (
    <View data-testid="contracts-container">
      <LayoutHeader
        title="Buyer Hub"
        address={address}
        copyToClipboard={copyToClipboard}
      >
        <HistoryBtn disabled={!showHistory} onClick={onHistoryOpen}>
          <span
            style={{ display: 'flex' }}
            data-rh={showHistory ? null : `You have no purchase history`}
          >
            <IconHistory style={{ display: 'inline-block' }} /> History
          </span>
        </HistoryBtn>
      </LayoutHeader>

      {/* <TotalsBlock /> */}

      <ContractsList
        hasContracts={hasContracts}
        syncStatus={syncStatus}
        contractsRefresh={contractsRefresh}
        address={address}
        contracts={contractsToShow}
        customRowRenderer={rowRenderer}
        noContractsMessage={'You have no contracts.'}
        offset={246}
        tabs={tabs}
      />

      <HistoryModal
        isActive={isHistoryModalOpen}
        historyContracts={contractsWithHistory}
        close={() => {
          setIsHistoryModalOpen(false);
        }}
      />

      <HashrateModal
        isActive={showHashrateModal}
        contractId={contactToShowHashrate}
        close={() => {
          setShowHashrateModal(false);
          setContactToShowHashrate(null);
        }}
      />
    </View>
  );
}

export default withContractsState(BuyerHub);
