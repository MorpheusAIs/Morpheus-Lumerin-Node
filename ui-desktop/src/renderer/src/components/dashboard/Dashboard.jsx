import React, { useState, useEffect } from 'react'
import styled from 'styled-components'

import withDashboardState from '../../store/hocs/withDashboardState'

import { LayoutHeader } from '../common/LayoutHeader'
import BalanceBlock from './BalanceBlock'
import TransactionModal from './tx-modal'
import TxList from './tx-list/TxList'
import { View } from '../common/View'

const Container = styled.div`
  background-color: ${(p) => p.theme.colors.light};
  height: 100vh;
  max-width: 100vw;
  position: relative;
  padding: 0 2.4rem;
`

const Dashboard = ({
  sendDisabled,
  sendDisabledReason,
  syncStatus,
  address,
  hasTransactions,
  copyToClipboard,
  onWalletRefresh
}) => {
  console.log('dashboard')
  const [activeModal, setActiveModal] = useState(null)

  const onCloseModal = () => setActiveModal(null)
  const onTabSwitch = (modal) => setActiveModal(modal)

  return (
    <View data-testid="dashboard-container">
      <LayoutHeader title="My Wallet" address={address} copyToClipboard={copyToClipboard} />

      <BalanceBlock
        sendDisabled={sendDisabled}
        sendDisabledReason={sendDisabledReason}
        onTabSwitch={onTabSwitch}
      />

      <TxList
        hasTransactions={hasTransactions}
        onWalletRefresh={onWalletRefresh}
        syncStatus={syncStatus}
      />

      <TransactionModal
        onRequestClose={onCloseModal}
        onTabSwitch={onTabSwitch}
        activeTab={activeModal}
      />
    </View>
  )
}

export default withDashboardState(Dashboard)
