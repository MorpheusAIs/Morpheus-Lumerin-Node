import React, { useState, useEffect } from 'react';
import styled from 'styled-components';

import withDashboardState from '../../store/hocs/withDashboardState';

import { ChainHeader } from '../common/ChainHeader';
import BalanceBlock from './BalanceBlock';
import TransactionModal from './tx-modal';
import TxList from './tx-list/TxList';
import { View } from '../common/View';
import { toUSD } from '../../store/utils/syncAmounts';
import { BtnAccent } from './BalanceBlock.styles';

const CustomBtn = styled(BtnAccent)`
  margin-left: 0;
  padding: 1.5rem 1rem;
`;
const WidjetsContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: left;
  gap: 1.6rem;
`;

const WidjetItem = styled.div`
  margin: 1.6rem 0 1.6rem;
  padding: 1.6rem 3.2rem;
  border-radius: 0.375rem;
  color: white;
  max-width: 720px;

  color: white;
`;

const StakingWidjet = styled(WidjetItem)`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.04);
  border-width: 1px;
  border: 1px solid rgba(255, 255, 255, 0.04);
`;

const Dashboard = ({
  sendDisabled,
  sendDisabledReason,
  syncStatus,
  address,
  copyToClipboard,
  onWalletRefresh,
  getBalances,
  ethCoinPrice,
  loadTransactions,
  getStakedFunds,
  explorerUrl,
  morTokenAddr,
  ...props
}) => {
  const [activeModal, setActiveModal] = useState(null);

  const onCloseModal = () => setActiveModal(null);
  const onTabSwitch = (modal) => setActiveModal(modal);

  const [balanceData, setBalanceData] = useState({
    eth: {
      value: 0,
      rate: 0,
      usd: 0,
      symbol: 'ETH',
    },
    mor: {
      value: 0,
      rate: 0,
      usd: 0,
      symbol: 'MOR',
    },
  });
  const [transactions, setTransactions] = useState([]);
  // const [pagging, setPagging] = useState({ page: 1, pageSize: 50, hasNextPage: true })
  const [staked, setStaked] = useState(0);

  const loadBalances = async () => {
    const data = await getBalances();
    const eth = data.balances.eth / 10 ** 18;
    const mor = data.balances.mor / 10 ** 18;
    const ethUsd = toUSD(eth, ethCoinPrice);
    const morUsd = toUSD(mor, +data.rate);

    const balances = {
      eth: {
        value: eth,
        rate: ethCoinPrice,
        usd: ethUsd,
        symbol: props.symbolEth,
      },
      mor: {
        value: mor,
        rate: +data.rate,
        usd: morUsd,
        symbol: props.symbol,
      },
    };
    setBalanceData(balances);
  };

  const getTransactions = async () => {
    const pageTransactions = await loadTransactions(1, 15);
    // const hasNextPage = !!pageTransactions.length;
    // setPagging({ ...pagging, page: pagging.page + 1, hasNextPage });
    setTransactions([...pageTransactions]);
  };

  useEffect(() => {
    loadBalances();
    getTransactions();
    getStakedFunds(address).then((data) => {
      setStaked(data);
    });

    const interval = setInterval(() => {
      console.log('Update balances...');
      loadBalances();
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    loadBalances();
  }, [ethCoinPrice]);

  return (
    <View data-testid="dashboard-container">
      <ChainHeader
        title="My Wallet"
        chain={props.config.chain}
        address={address}
        copyToClipboard={copyToClipboard}
      />

      <BalanceBlock
        {...balanceData}
        sendDisabled={sendDisabled}
        sendDisabledReason={sendDisabledReason}
        onTabSwitch={onTabSwitch}
      />

      <WidjetsContainer>
        <StakingWidjet className="staking">
          <div>Staked Balance</div>
          <div>
            {staked} {props.symbol}
          </div>
        </StakingWidjet>
        <WidjetItem>
          <CustomBtn onClick={() => window.openLink(explorerUrl)} block>
            Transaction Explorer
          </CustomBtn>
        </WidjetItem>
        <WidjetItem>
          <CustomBtn
            onClick={() => window.openLink('https://staking.mor.lumerin.io')}
            block
          >
            Staking Dashboard
          </CustomBtn>
        </WidjetItem>
      </WidjetsContainer>

      <TxList
        // {...pagging}
        // hasNextPage={pagging.hasNextPage}
        loadNextTransactions={() => {}}
        hasTransactions={!!transactions.length}
        syncStatus={syncStatus}
        transactions={transactions}
      />

      <TransactionModal
        {...balanceData}
        onRequestClose={onCloseModal}
        onTabSwitch={onTabSwitch}
        activeTab={activeModal}
      />
    </View>
  );
};

export default withDashboardState(Dashboard);
