import withWalletInfoState from '../../store/hocs/withWalletInfoState';
import styled from 'styled-components';
import React from 'react';

import { Flex, Sp } from '../common';
import LastUpdated, { Label } from '../common/LastUpdated';

const Text = styled.div`
  color: ${p => p.theme.colors.dark};
  font-size: 1.3rem;
  margin: 0.8rem 0;
  display: flex;
  align-items: center;
`;

const MinedAgo = styled(Label)`
  margin-left: 0.8rem;
`;

const IndicatorLed = styled.div`
  width: 10px;
  height: 10px;
  background-color: ${({ isOnline, isConnected, theme }) =>
    isOnline
      ? isConnected
        ? theme.colors.success
        : theme.colors.danger
      : 'rgba(119, 132, 125, 0.68)'};
  border: 1px solid white;
  border-radius: 10px;
  margin: 5px 8px 2px 1px;
`;

const WalletStatus = function(props) {
  return (
    <div>
      <Text>Version {props.appVersion}</Text>

      <Text>Connected to {props.chainName} chain</Text>

      <LastUpdated
        timestamp={props.bestBlockTimestamp}
        render={({ timeAgo, diff }) => (
          <Text>
            Block height {props.height}
            <MinedAgo diff={diff} as="span">
              mined {timeAgo}
            </MinedAgo>
          </Text>
        )}
      />

      <Flex.Row align="center">
        <Text>
          <IndicatorLed
            isConnected={props.isWeb3Connected}
            isOnline={props.isOnline}
          />
          Web3 connection
        </Text>
        <Sp px={2} />
        {/* <Text>
          <IndicatorLed
            isConnected={props.isIndexerConnected}
            isOnline={props.isOnline}
          />
          Indexer connection
        </Text>
        <Sp px={2} /> */}
        <Text>
          <IndicatorLed
            isConnected={props.isProxyRouterConnected}
            isOnline={props.isOnline}
          />
          Proxy-Router connection
        </Text>
      </Flex.Row>
    </div>
  );
};

export default withWalletInfoState(WalletStatus);
