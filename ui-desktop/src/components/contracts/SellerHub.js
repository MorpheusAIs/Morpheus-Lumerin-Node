import React, { useState, useContext, useEffect, useCallback } from 'react';
import styled from 'styled-components';
import { uniqueId, debounce } from 'lodash';

import withContractsState from '../../store/hocs/withContractsState';
import { Btn } from '../common';
import { LayoutHeader } from '../common/LayoutHeader';
import ContractsList from './contracts-list/ContractsList';
import CreateContractModal from './modals/CreateContractModal';
import { View } from '../common/View';
import { ToastsContext } from '../toasts';
import { CONTRACT_STATE } from '../../enums';
import { lmrDecimals } from '../../utils/coinValue';
import { formatBtcPerTh, calculateSuggestedPrice } from './utils';
import ArchiveModal from './modals/ArchiveModal/ArchiveModal';
import { IconArchive } from '@tabler/icons';
import SellerWhitelistModal from './modals/SellerWhitelistModal/SellerWhitelistModal';
import AdjustProfitModal from './modals/AdjustProfitModal/AdjustProfitModal';

const Container = styled.div`
  background-color: ${p => p.theme.colors.light};
  min-height: 100%;
  width: 100%;
  position: relative;
  padding: 0 2.4rem;
`;

const Title = styled.div`
  font-size: 2.4rem;
  line-height: 3rem;
  white-space: nowrap;
  margin: 0;
  font-weight: 600;
  color: ${p => p.theme.colors.dark};
  margin-bottom: 4.8px;
  margin-right: 2.4rem;
  cursor: default;

  @media (min-width: 800px) {
  }
  @media (min-width: 1200px) {
    margin-right: 1.6rem;
  }
`;

const ContractBtn = styled(Btn)`
  font-size: 1.3rem;
  padding: 0.6rem 1.4rem;

  @media (min-width: 1040px) {
    margin-left: 0;
  }
`;

const ArchiveBtn = styled(Btn)`
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

function SellerHub({
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
  networkDifficulty,
  selectedCurrency,
  formUrl,
  ...props
}) {
  const [isModalActive, setIsModalActive] = useState(false);
  const [isArchiveModalActive, setIsArchiveModalActive] = useState(false);
  const [showSellerWhitelistForm, setShowSellerWhitelistForm] = useState(false);
  const [showAdjustForm, setShowAdjustForm] = useState(false);
  const [isEditModalActive, setIsEditModalActive] = useState(false);
  const [editContractData, setEditContractData] = useState({});
  const context = useContext(ToastsContext);
  const [showSuccess, setShowSuccess] = useState(false);
  const [underProfitContracts, setUnderProfitContracts] = useState([]);
  const [profitSettings, setProfitSettings] = useState({});
  const [autoAdjustPriceData, setAutoAdjustPriceData] = useState({});

  const refreshAutoAdjustPriceData = () => {
    client.getAutoAdjustPriceData().then(data => {
      setAutoAdjustPriceData(data);
    });
  };

  useEffect(() => {
    client.getProfitSettings().then(settings => {
      setProfitSettings(settings);
    });
    refreshAutoAdjustPriceData();
  }, []);

  useEffect(() => {
    if (contracts.length) {
      verify(
        contracts,
        props.lmrCoinPrice,
        props.btcCoinPrice,
        address,
        networkDifficulty
      );
    }
  }, [contracts]);

  const verify = useCallback(
    debounce((...param) => {
      client.getProfitSettings().then(settings => {
        if (!settings) {
          return;
        }
        const contracts = param[0].filter(
          c => c.seller === param[3] && !c.isDead
        );
        const lmrCoinPrice = param[1];
        const btcCoinPrice = param[2];
        const reward = formatBtcPerTh(param[4]);
        const deviation = +settings.deviation;
        const result = contracts.reduce((acc, contract) => {
          const contractProfitTarget =
            +contract.futureTerms?.profitTarget || +contract.profitTarget;
          const profitTarget =
            contractProfitTarget !== 0
              ? contractProfitTarget
              : settings?.adaptExisting
              ? +settings?.target
              : 0;
          if (+profitTarget === 0) {
            return acc;
          }

          const left =
            1 +
            (+profitTarget > 0
              ? deviation - profitTarget
              : profitTarget - deviation) /
              100;
          const right = 1 + (deviation + profitTarget) / 100;

          const length =
            (contract.futureTerms?.length || contract.length) / 3600;
          const speed =
            (contract.futureTerms?.speed || contract.speed) / 10 ** 12;

          const estimatedLeft = calculateSuggestedPrice(
            length,
            speed,
            btcCoinPrice,
            lmrCoinPrice,
            reward,
            left
          );
          const estimatedRight = calculateSuggestedPrice(
            length,
            speed,
            btcCoinPrice,
            lmrCoinPrice,
            reward,
            right
          );
          const targetEstimate = calculateSuggestedPrice(
            length,
            speed,
            btcCoinPrice,
            lmrCoinPrice,
            reward,
            1 + profitTarget / 100
          );

          const zeroBasedEstimate = calculateSuggestedPrice(
            length,
            speed,
            btcCoinPrice,
            lmrCoinPrice,
            reward,
            1
          );
          const currPrice =
            (contract.futureTerms?.price || contract.price) / lmrDecimals;
          const isPriceWithinRange =
            estimatedLeft <= currPrice && estimatedRight >= currPrice;

          if (!isPriceWithinRange) {
            return [
              ...acc,
              {
                id: contract.id,
                profit: profitTarget,
                zeroBasedEstimate: zeroBasedEstimate,
                zeroRate: (currPrice / zeroBasedEstimate).toFixed(2),
                currentRate: (currPrice / targetEstimate).toFixed(2),
                estimatedPrice: targetEstimate,
                price: currPrice
              }
            ];
          }
          return acc;
        }, []);

        setUnderProfitContracts(result);
      });
    }, 300),
    []
  );

  const tabs = [
    { name: 'Status', ratio: 1 },
    {
      value: 'price',
      name: 'Price',
      ratio: 1,
      options: [
        {
          label: 'Price (BTC)',
          value: 'BTC',
          selected: selectedCurrency === 'BTC'
        },
        {
          label: 'Price (LMR)',
          value: 'LMR',
          selected: selectedCurrency === 'LMR'
        }
      ]
    },
    { value: 'btc-th', name: 'BTC/TH/day', ratio: 1 },
    { value: 'length', name: 'Duration', ratio: 1 },
    { value: 'speed', name: 'Speed', ratio: 1 },
    { value: 'history', name: 'History', ratio: 1 },
    {
      value: 'claimable',
      name: 'Claimable',
      ratio: 1,
      options: [
        {
          label: 'Claimable (BTC)',
          value: 'BTC',
          selected: selectedCurrency === 'BTC'
        },
        {
          label: 'Claimable (LMR)',
          value: 'LMR',
          selected: selectedCurrency === 'LMR'
        }
      ]
    },
    { value: 'action', name: 'Actions', ratio: 1 }
  ];

  const handleOpenModal = () => setIsModalActive(true);

  const handleCloseModal = e => {
    setIsModalActive(false);
    setIsEditModalActive(false);
    setShowSuccess(false);
  };

  const handleEditModal = contract => {
    setEditContractData(contract);
    setIsEditModalActive(true);
    setShowSuccess(false);
  };

  const updateContractAutoAdjustSettings = (address, profit, isEnabled) => {
    if (isEnabled === undefined) {
      return;
    }
    return client
      .setAutoAdjustPriceData({
        [address?.toLowerCase()]: {
          enabled: isEnabled,
          profitTarget: profit
        }
      })
      .then(() => refreshAutoAdjustPriceData());
  };

  const createTempContract = (id, contract) => {
    client.store.dispatch({
      type: 'create-temp-contract',
      payload: {
        id,
        ...contract,
        length: contract.duration,
        seller: contract.sellerAddress,
        state: CONTRACT_STATE.Avaliable,
        timestamp: 0,
        isDeploying: true
      }
    });
  };

  const dispatchEditContract = (id, contract) => {
    client.store.dispatch({
      type: 'edit-contract-state',
      payload: {
        id,
        ...contract,
        length: contract.duration,
        seller: contract.sellerAddress
      }
    });
  };

  const removeTempContract = (id, contract) => {
    client.store.dispatch({
      type: 'remove-contract',
      payload: {
        id,
        ...contract
      }
    });
  };

  const handleContractUpdate = async (
    e,
    contractDetails,
    contractId,
    initialContractData,
    autoAdjustPrice
  ) => {
    e.preventDefault();

    const newPrice = contractDetails.price * lmrDecimals;
    const newSpeed = contractDetails.speed * 10 ** 12;
    const newDuration = contractDetails.time * 3600;

    const contract = {
      id: contractId,
      price: newPrice,
      speed: newSpeed, // THs
      duration: newDuration, // Hours to seconds
      sellerAddress: contractDetails.address,
      profit: contractDetails.profitTarget
    };

    const shouldCallBlockchain =
      !initialContractData ||
      initialContractData.price != newPrice ||
      initialContractData.speed != newSpeed ||
      initialContractData.length != newDuration ||
      initialContractData.profitTarget != contractDetails.profitTarget;

    if (shouldCallBlockchain) {
      await client.lockSendTransaction();
      await client
        .editContract(contract)
        .then(() => {
          setShowSuccess(true);
          updateContractAutoAdjustSettings(
            contractId,
            contractDetails.profitTarget,
            autoAdjustPrice
          );
          // dispatchEditContract(contract.id, contract); // TODO: investigate rows are not rerendering
        })
        .catch(error => {
          context.toast('error', error.message || error);
          setIsModalActive(false);
        })
        .finally(() => {
          client.unlockSendTransaction();
        });
    } else {
      updateContractAutoAdjustSettings(
        contractId,
        contractDetails.profitTarget,
        autoAdjustPrice
      ).then(() => {
        setShowSuccess(true);
        setIsModalActive(false);
      });
    }
  };

  const handleContractDeploy = async (
    e,
    contractDetails,
    autoAdjustPrice = false
  ) => {
    e.preventDefault();

    const contract = {
      price: contractDetails.price * lmrDecimals,
      speed: contractDetails.speed * 10 ** 12, // THs
      duration: contractDetails.time * 3600, // Hours to seconds
      sellerAddress: contractDetails.address,
      profit: contractDetails.profitTarget || profitSettings?.target
    };

    const tempContractId = uniqueId();
    createTempContract(tempContractId, contract);

    await client.lockSendTransaction();
    await client
      .createContract(contract)
      .then(result => {
        const contractAddress = result?.events && result?.events[0]?.address;
        if (autoAdjustPrice) {
          updateContractAutoAdjustSettings(
            contractAddress,
            contract.profit,
            true
          );
        }
        setShowSuccess(true);
      })
      .catch(error => {
        setIsModalActive(false);
        if (error.message == 'seller is not whitelisted') {
          setShowSellerWhitelistForm(true);
          return;
        }
        context.toast('error', error.message || error);
      })
      .finally(() => {
        removeTempContract(tempContractId, contract);
        client.unlockSendTransaction();
      });
  };

  const handleContractCancellation = data => {
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

  const handleDeleteContractStateChange = data => {
    client.lockSendTransaction();
    return client
      .setDeleteContractStatus({
        contractId: data.contractId,
        walletAddress: data.walletAddress,
        deleteContract: data.deleteContract
      })
      .finally(() => {
        client.unlockSendTransaction();
      });
  };

  const handleContractSave = e => {
    e.preventDefault();
  };
  const contractsToShow = contracts.filter(
    c => c.seller === address && !c.isDead
  );

  const deadContracts = contracts
    .filter(c => c.seller === address && c.isDead)
    .sort((a, b) => b.balance - a.balance);

  const rentedContracts =
    contractsToShow?.filter(x => Number(x.state) === 1) ?? [];
  const speedReducer = (acc, c) => acc + Number(c.speed) / 10 ** 12;
  const sellerStats = {
    count: contractsToShow.length ?? 0,
    rented: rentedContracts.reduce(speedReducer, 0),
    totalPosted: contractsToShow.reduce(speedReducer, 0),
    networkReward: formatBtcPerTh(networkDifficulty)
  };
  const showArchive = deadContracts?.length;
  const onArchiveOpen = () => setIsArchiveModalActive(true);

  const adjustPrice = data => {
    const c = contracts.find(x => x.id == data.id);
    return handleContractUpdate(
      { preventDefault: () => {} },
      {
        price: data.price,
        speed: c.speed / 10 ** 12,
        time: c.length / 3600,
        address: c.address,
        profitTarget: c.profitTarget
      },
      data.id
    );
  };

  const applyAllSuggested = async contractsToUpdate => {
    try {
      for (let i = 0; i < contractsToUpdate.length; i++) {
        const item = contractsToUpdate[i];
        await adjustPrice(item);
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    } catch (err) {
      context.toast('error', err.message || err);
    }
  };

  return (
    <View data-testid="contracts-container">
      <LayoutHeader
        title="Seller Hub"
        address={address}
        copyToClipboard={copyToClipboard}
      >
        <ArchiveBtn disabled={!showArchive} onClick={onArchiveOpen}>
          <span
            style={{ display: 'flex' }}
            data-rh={showArchive ? null : `You have no archived contracts`}
          >
            <IconArchive style={{ display: 'inline-block' }} /> Archived
          </span>
        </ArchiveBtn>
      </LayoutHeader>

      <ContractsList
        hasContracts={hasContracts}
        syncStatus={syncStatus}
        cancel={handleContractCancellation}
        deleteContract={handleDeleteContractStateChange}
        createContract={handleOpenModal}
        contractsRefresh={contractsRefresh}
        address={address}
        contracts={contractsToShow}
        allowSendTransaction={allowSendTransaction}
        noContractsMessage={'You have no contracts.'}
        tabs={tabs}
        edit={handleEditModal}
        setEditContractData={setEditContractData}
        isSellerTab={true}
        sellerStats={sellerStats}
        offset={394}
        underProfitContracts={underProfitContracts}
        onAdjustFormOpen={() => setShowAdjustForm(true)}
      />

      <CreateContractModal
        isActive={isModalActive}
        save={handleContractSave}
        deploy={handleContractDeploy}
        close={handleCloseModal}
        showSuccess={showSuccess}
        networkReward={sellerStats.networkReward}
        editContractData={{}}
        profitSettings={profitSettings}
      />

      <ArchiveModal
        isActive={isArchiveModalActive}
        deletedContracts={deadContracts}
        handlePurchase={() => {}}
        close={() => {
          setIsArchiveModalActive(false);
        }}
        restore={handleDeleteContractStateChange}
        address={address}
        showSuccess={false}
      />

      <SellerWhitelistModal
        isActive={showSellerWhitelistForm}
        formUrl={formUrl}
        close={() => {
          setShowSellerWhitelistForm(false);
        }}
      />

      <AdjustProfitModal
        isActive={showAdjustForm}
        contracts={[...underProfitContracts]}
        close={() => {
          setShowAdjustForm(false);
        }}
        onAdjust={adjustPrice}
        onApplySuggested={applyAllSuggested}
      />

      <CreateContractModal
        isActive={isEditModalActive}
        isEditMode={true}
        editContractData={editContractData}
        autoAdjustPriceData={autoAdjustPriceData}
        edit={handleContractUpdate}
        showSuccess={showSuccess}
        close={() => {
          setIsEditModalActive(false);
        }}
      ></CreateContractModal>
    </View>
  );
}

export default withContractsState(SellerHub);
