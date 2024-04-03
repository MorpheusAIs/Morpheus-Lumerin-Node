import React, { useState } from 'react';

import withSocketsState from '../../store/hocs/withSocketsState';
import TotalsBlock from './TotalsBlock';
import SocketsList from './sockets-list/SocketsList';
import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';

// const Title = styled.h1`
//   font-size: 2.4rem;
//   line-height: 3rem;
//   color: ${p => p.theme.colors.darker}
//   white-space: nowrap;
//   margin: 0;
//   cursor: default;
// `

const Sockets = ({
  address,
  syncStatus,
  copyToClipboard,
  incomingCount,
  outgoingCount,
  routedCount
}) => {
  const [activeModal, setActiveModal] = useState('');
  const ipAddress = '127.0.0.1';
  const port = '3000';
  // static propTypes = {
  //   sendDisabledReason: PropTypes.string,
  //   hasSockets: PropTypes.bool.isRequired,
  //   copyToClipboard: PropTypes.func.isRequired,
  //   onWalletRefresh: PropTypes.func.isRequired,
  //   sendDisabled: PropTypes.bool.isRequired,
  //   syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed']).isRequired,
  //   address: PropTypes.string.isRequired
  // };

  const onOpenModal = e => setActiveModal(e.target.dataset.modal);

  const onCloseModal = () => setActiveModal(null);

  return (
    <View data-testid="sockets-container">
      <LayoutHeader
        title="Connections"
        address={address}
        copyToClipboard={copyToClipboard}
      />

      <TotalsBlock
        incoming={incomingCount}
        outgoing={outgoingCount}
        routed={routedCount}
      />

      <SocketsList
        ipAddress={ipAddress}
        port={port}
        // onWalletRefresh={props.onWalletRefresh}
        syncStatus={syncStatus}
      />
    </View>
  );
};

export default withSocketsState(Sockets);
