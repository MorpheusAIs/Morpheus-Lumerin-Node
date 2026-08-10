import React, { useState, useMemo, useContext } from 'react';
import styled, { keyframes } from 'styled-components';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  IconCopy,
  IconExternalLink,
  IconArrowDownLeft,
  IconArrowUpRight,
  IconChartBar,
  IconLock,
} from '@tabler/icons-react';

import withDashboardState from '../../store/hocs/withDashboardState';

import TransactionModal from './tx-modal';
import TxList from './tx-list/TxList';
import { View } from '../common/View';
import { toUSD } from '../../store/utils/syncAmounts';
import {
  queryKeys,
  computeStakedFunds,
  countOpenSessions,
} from '../../store/queries';
import QueryError from '../common/QueryError';
import WalletSwitcher from './WalletSwitcher';
import { ToastsContext } from '../toasts';
import { abbreviateAddress } from '../../utils';
import BaseLogo from '../icons/BaseLogo';
import { MorpheusLogo } from '../icons/MorpheusLogo';
import { EtherIcon } from '../icons/EtherIcon';

const EMPTY_BALANCE = {
  eth: { value: 0, rate: 0, usd: 0, symbol: 'ETH' },
  mor: { value: 0, rate: 0, usd: 0, symbol: 'MOR' },
};

const fadeUp = keyframes`
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
`;

const pulse = keyframes`
  0%, 100% { opacity: 1; }
  50%      { opacity: 0.35; }
`;

const Page = styled.div`
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
`;

/* ---- Header ---- */
const TopBar = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.6rem;
  flex-wrap: wrap;
  padding-bottom: 1.6rem;
`;

const TitleGroup = styled.div`
  display: flex;
  align-items: center;
  gap: 1.2rem;
`;

const PageTitle = styled.h1`
  margin: 0;
  font-size: 2.4rem;
  line-height: 1;
  font-weight: 600;
  color: ${(p) => p.theme.colors.morMain};
`;

const ChainBadge = styled.div`
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: rgba(255, 255, 255, 0.85);
  font-size: 1.2rem;
  padding: 0.5rem 1rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
`;

const AddressPill = styled.div`
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 0.6rem 0.7rem 0.6rem 1.2rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.92);
  font-size: 1.3rem;
  font-variant-numeric: tabular-nums;
`;

const AddressDot = styled.div`
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: ${(p) => p.theme.colors.morMain};
  box-shadow: 0 0 0 3px rgba(32, 220, 142, 0.18);
`;

const IconBtn = styled.button`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;

  &:hover {
    background: rgba(32, 220, 142, 0.14);
    color: ${(p) => p.theme.colors.morMain};
  }
`;

/* ---- Balance hero ---- */
const HeroGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 1.6rem;
`;

const TokenCard = styled.div`
  position: relative;
  overflow: hidden;
  border-radius: 16px;
  padding: 1.4rem 1.8rem;
  background: ${(p) =>
    p.$accent
      ? 'linear-gradient(135deg, rgba(32,220,142,0.16) 0%, rgba(32,220,142,0.04) 55%, rgba(255,255,255,0.02) 100%)'
      : 'rgba(255,255,255,0.04)'};
  border: 1px solid
    ${(p) =>
      p.$accent ? 'rgba(32,220,142,0.30)' : 'rgba(255,255,255,0.07)'};
  animation: ${fadeUp} 0.35s ease both;
  transition: border-color 0.2s ease, transform 0.2s ease;

  &:hover {
    transform: translateY(-2px);
    border-color: ${(p) =>
      p.$accent ? 'rgba(32,220,142,0.5)' : 'rgba(255,255,255,0.14)'};
  }
`;

const TokenHead = styled.div`
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
`;

const TokenBadge = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  flex-shrink: 0;
  color: ${(p) => (p.$accent ? '#0b1f16' : '#fff')};
  background: ${(p) =>
    p.$accent ? p.theme.colors.morMain : 'rgba(255,255,255,0.08)'};
`;

const TokenName = styled.div`
  display: flex;
  flex-direction: column;
`;

const TokenSymbol = styled.span`
  font-size: 1.5rem;
  font-weight: 600;
  color: #fff;
  letter-spacing: 0.3px;
`;

const TokenSub = styled.span`
  font-size: 1.1rem;
  color: rgba(255, 255, 255, 0.45);
`;

const LiveDot = styled.span`
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 1rem;
  color: rgba(255, 255, 255, 0.4);

  &::before {
    content: '';
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: ${(p) => p.theme.colors.morMain};
    animation: ${pulse} 1.1s ease-in-out infinite;
  }
`;

const TokenBalance = styled.div`
  font-size: 2.7rem;
  line-height: 1.05;
  font-weight: 600;
  color: #fff;
  letter-spacing: -0.5px;
  font-variant-numeric: tabular-nums;
  word-break: break-all;
`;

const TokenUnit = styled.span`
  font-size: 1.6rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.5);
  margin-left: 6px;
`;

const TokenUsd = styled.div`
  margin-top: 0.6rem;
  font-size: 1.35rem;
  color: rgba(255, 255, 255, 0.55);
  font-variant-numeric: tabular-nums;
`;

/* ---- Stats / actions ---- */
const StatsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.6rem;
  margin: 1.2rem 0;
`;

const StatCard = styled.div`
  display: flex;
  align-items: center;
  gap: 1.4rem;
  padding: 1.6rem 1.8rem;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.07);
`;

const StatIcon = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  flex-shrink: 0;
  color: ${(p) => p.theme.colors.morMain};
  background: rgba(32, 220, 142, 0.12);
`;

const StatText = styled.div`
  display: flex;
  flex-direction: column;
  min-width: 0;
`;

const StatLabel = styled.span`
  font-size: 1.15rem;
  color: rgba(255, 255, 255, 0.5);
  text-transform: uppercase;
  letter-spacing: 0.6px;
`;

const StatValue = styled.span`
  font-size: 1.9rem;
  font-weight: 600;
  color: #fff;
  font-variant-numeric: tabular-nums;
`;

const ActionTile = styled.button`
  display: flex;
  align-items: center;
  gap: 1.4rem;
  padding: 1.6rem 1.8rem;
  border-radius: 14px;
  text-align: left;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.07);
  color: #fff;
  transition: background 0.15s ease, border-color 0.15s ease,
    transform 0.15s ease;

  &:hover {
    background: rgba(32, 220, 142, 0.08);
    border-color: rgba(32, 220, 142, 0.35);
    transform: translateY(-2px);
  }
`;

const ActionText = styled.div`
  display: flex;
  flex-direction: column;
`;

const ActionTitle = styled.span`
  font-size: 1.5rem;
  font-weight: 600;
`;

const ActionSub = styled.span`
  font-size: 1.1rem;
  color: rgba(255, 255, 255, 0.45);
`;

const formatAmount = (value, maxFrac = 4) => {
  const n = Number(value) || 0;
  if (n === 0) return '0';
  if (n > 0 && n < 0.0001) return '< 0.0001';
  return new Intl.NumberFormat(navigator.language, {
    minimumFractionDigits: 0,
    maximumFractionDigits: maxFrac,
  }).format(n);
};

const formatUsd = (usd) => (usd && usd !== 0 ? `≈ ${usd}` : '≈ $0.00');

const Dashboard = ({
  syncStatus,
  address,
  copyToClipboard,
  getBalances,
  ethCoinPrice,
  loadTransactions,
  getSessionsByUser,
  explorerUrl,
  ...props
}) => {
  const [activeModal, setActiveModal] = useState(null);
  const context = useContext(ToastsContext);
  const queryClient = useQueryClient();

  const onCloseModal = () => {
    // A send may have just gone through, so drop the cached balances and
    // transaction list rather than showing pre-transfer numbers for up to the
    // 30s staleTime window.
    if (activeModal === 'success') {
      queryClient.invalidateQueries({ queryKey: queryKeys.balances(address) });
      queryClient.invalidateQueries({
        queryKey: queryKeys.transactions(address),
      });
    }
    setActiveModal(null);
  };
  const onTabSwitch = (modal) => setActiveModal(modal);

  // Cached, stale-while-revalidate data. Revisiting the wallet tab renders the
  // last-known balances/transactions/staked instantly and refreshes silently.
  const balancesQuery = useQuery({
    queryKey: queryKeys.balances(address),
    queryFn: getBalances,
    enabled: !!address,
    refetchInterval: 30000,
  });

  const transactionsQuery = useQuery({
    queryKey: queryKeys.transactions(address),
    queryFn: () => loadTransactions(1, 15),
    enabled: !!address,
  });

  // Shares the exact cache key/shape used by the Chat tab, so the (expensive,
  // paginated) sessions fetch is reused across tabs.
  const sessionsQuery = useQuery({
    queryKey: queryKeys.sessions(address),
    queryFn: () => getSessionsByUser(address),
    enabled: !!address,
  });

  const balanceData = useMemo(() => {
    const data = balancesQuery.data;
    if (!data || !data.balances) {
      return EMPTY_BALANCE;
    }
    const eth = data.balances.eth / 10 ** 18;
    const mor = data.balances.mor / 10 ** 18;
    return {
      eth: {
        value: eth,
        rate: ethCoinPrice,
        usd: toUSD(eth, ethCoinPrice),
        symbol: props.symbolEth,
      },
      mor: {
        value: mor,
        rate: +data.rate,
        usd: toUSD(mor, +data.rate),
        symbol: props.symbol,
      },
    };
  }, [balancesQuery.data, ethCoinPrice, props.symbol, props.symbolEth]);

  const transactions = transactionsQuery.data ?? [];
  const staked = useMemo(
    () => computeStakedFunds(sessionsQuery.data),
    [sessionsQuery.data],
  );

  // Drives the switch confirmation: switching wallets stops any in-flight chat,
  // so the user should know a session is live before doing it.
  const openSessionCount = useMemo(
    () => countOpenSessions(sessionsQuery.data),
    [sessionsQuery.data],
  );

  const { eth, mor } = balanceData;
  const balancesLive = balancesQuery.isFetching;
  const morSymbol = mor.symbol || 'MOR';
  const ethSymbol = eth.symbol || 'ETH';
  const explorerHost = explorerUrl ? new URL(explorerUrl).hostname : 'explorer';

  const handleCopy = () => {
    if (!address) return;
    copyToClipboard(address);
    context.toast('info', 'Address copied to clipboard', { autoClose: 1500 });
  };

  return (
    <View data-testid="dashboard-container">
      <Page>
        <TopBar>
          <TitleGroup>
            <PageTitle>My Wallet</PageTitle>
            <ChainBadge>
              <BaseLogo style={{ width: '18px' }} />
              {props.config?.chain?.displayName || 'Base'}
            </ChainBadge>
          </TitleGroup>

          {address && (
            <div
              style={{ display: 'flex', alignItems: 'center', gap: '0.8rem' }}
            >
              <WalletSwitcher
                activeAddress={address}
                openSessionCount={openSessionCount}
              />
              <AddressPill>
                <AddressDot />
                {abbreviateAddress(address, 6)}
                <IconBtn title="Copy address" onClick={handleCopy}>
                  <IconCopy size={16} />
                </IconBtn>
                <IconBtn
                  title={`View on ${explorerHost}`}
                  onClick={() => explorerUrl && window.openLink(explorerUrl)}
                >
                  <IconExternalLink size={16} />
                </IconBtn>
              </AddressPill>
            </div>
          )}
        </TopBar>

        <QueryError
          error={balancesQuery.error}
          what="balances"
          onRetry={() => balancesQuery.refetch()}
        />

        <HeroGrid>
          <TokenCard $accent>
            <TokenHead>
              <TokenBadge $accent>
                <MorpheusLogo style={{ width: '22px', height: '22px' }} />
              </TokenBadge>
              <TokenName>
                <TokenSymbol>{morSymbol}</TokenSymbol>
                <TokenSub>Morpheus</TokenSub>
              </TokenName>
              {balancesLive && <LiveDot>live</LiveDot>}
            </TokenHead>
            <TokenBalance>
              {formatAmount(mor.value, 4)}
              <TokenUnit>{morSymbol}</TokenUnit>
            </TokenBalance>
            <TokenUsd>{formatUsd(mor.usd)}</TokenUsd>
          </TokenCard>

          <TokenCard>
            <TokenHead>
              <TokenBadge>
                <EtherIcon size="22px" />
              </TokenBadge>
              <TokenName>
                <TokenSymbol>{ethSymbol}</TokenSymbol>
                <TokenSub>Ethereum</TokenSub>
              </TokenName>
              {balancesLive && <LiveDot>live</LiveDot>}
            </TokenHead>
            <TokenBalance>
              {formatAmount(eth.value, 5)}
              <TokenUnit>{ethSymbol}</TokenUnit>
            </TokenBalance>
            <TokenUsd>{formatUsd(eth.usd)}</TokenUsd>
          </TokenCard>
        </HeroGrid>

        <StatsRow>
          <StatCard>
            <StatIcon>
              <IconLock size={22} />
            </StatIcon>
            <StatText>
              <StatLabel>Staked Balance</StatLabel>
              <StatValue>
                {staked} {morSymbol}
              </StatValue>
            </StatText>
          </StatCard>

          <ActionTile onClick={() => onTabSwitch('send')}>
            <StatIcon>
              <IconArrowUpRight size={22} />
            </StatIcon>
            <ActionText>
              <ActionTitle>Send</ActionTitle>
              <ActionSub>Transfer {morSymbol} or {ethSymbol}</ActionSub>
            </ActionText>
          </ActionTile>

          <ActionTile onClick={() => onTabSwitch('receive')}>
            <StatIcon>
              <IconArrowDownLeft size={22} />
            </StatIcon>
            <ActionText>
              <ActionTitle>Receive</ActionTitle>
              <ActionSub>Show address & QR</ActionSub>
            </ActionText>
          </ActionTile>

          <ActionTile
            onClick={() => window.openLink('https://staking.mor.lumerin.io')}
          >
            <StatIcon>
              <IconChartBar size={22} />
            </StatIcon>
            <ActionText>
              <ActionTitle>Staking Dashboard</ActionTitle>
              <ActionSub>staking.mor.lumerin.io</ActionSub>
            </ActionText>
          </ActionTile>
        </StatsRow>

        <TxList
          loadNextTransactions={() => {}}
          hasTransactions={!!transactions.length}
          syncStatus={syncStatus}
          loading={transactionsQuery.isLoading}
          isRefreshing={transactionsQuery.isFetching}
          transactions={transactions}
        />
      </Page>

      <TransactionModal
        {...balanceData}
        onRequestClose={onCloseModal}
        onTabSwitch={onTabSwitch}
        activeTab={activeModal}
        address={address}
        copyToClipboard={copyToClipboard}
        explorerUrl={explorerUrl}
      />
    </View>
  );
};

export default withDashboardState(Dashboard);
