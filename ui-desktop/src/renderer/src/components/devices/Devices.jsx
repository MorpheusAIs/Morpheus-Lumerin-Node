import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import withDashboardState from '../../store/hocs/withDashboardState';
import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';
import { Btn, Flex, TextInput } from '../common';
import withDevicesState from '../../store/hocs/withDevicesState';
import Selector from '../common/Selector';
import Sp from '../common/Spacing';
import Device from './Device';
import { mapRangeNameToIpRange, RANGE, rangeSelectOptions } from './constants';
import Spinner from '../common/Spinner';

const Container = styled.div`
  display: flex;
  flex-direction: column;
  height: 100vh;
  max-width: 80vw;
  position: relative;

  .discovery-spinner {
    box-shadow: none;
    background-color: transparent;
    svg {
      width: 2em;
      height: 2em;
    }
  }
`;

const DeviceDiscoveryControl = styled.div`
  color: ${p => p.theme.colors.dark};
  .row {
    display: flex;
  }
`;

const DeviceDiscoveryResult = styled.div`
  display: flex;
  flex-wrap: wrap;
  color: ${p => p.theme.colors.dark};
  margin: 1em 0;
  max-height: 470px;
  overflow-y: scroll;
`;

const ConfigureBtn = styled(Btn)`
  margin-left: 10px;
  border-radius: 15px;
  border: 1px solid ${p => p.theme.colors.primary};
  background-color: ${p => p.theme.colors.light};
  color: ${p => p.theme.colors.primary};
`;

const Note = styled.div`
  position: relative;
  min-height: 1em;
  margin: 1em 0em;
  background: #fff;
  padding: 2rem 2.5rem;
  color: ${p => p.theme.colors.primary};
  border-radius: 15px;
`;

const Devices = props => {
  const STRATUM_PROXY_ROUTER_PORT = 3333;
  const proxyRouterHost = new URL(props.proxyRouterUrl).hostname;
  const proxyRouterUrl = `stratum+tcp://${proxyRouterHost}:${STRATUM_PROXY_ROUTER_PORT}`;

  const [range, setRange] = useState(rangeSelectOptions[0].value);
  const [fromIpDefault, toIpDefault] = mapRangeNameToIpRange(RANGE.SUBNET_24);
  const [fromIp, setFromIp] = useState(fromIpDefault);
  const [toIp, setToIp] = useState(toIpDefault);

  const isInputDisabled = [RANGE.SUBNET_16, RANGE.SUBNET_24].includes(range);
  const isInputVisible = range !== RANGE.LOCAL;
  const isDiscovering = props?.devices?.isDiscovering;
  const devices = Object.values(props?.devices?.devices) || [];
  const notConfiguredDevices = devices.filter(
    d => d.poolAddress !== proxyRouterUrl && d.isPrivilegedApiAvailable
  );

  const onRangeChange = e => {
    setRange(e.value);
    if ([RANGE.SUBNET_24, RANGE.SUBNET_16].includes(e.value)) {
      const ipRange = mapRangeNameToIpRange(e.value);
      setFromIp(ipRange[0]);
      setToIp(ipRange[1]);
    }
  };

  const startDiscovery = () => {
    props.resetDevices();

    if (range === RANGE.LOCAL) {
      return props.client.startDiscovery({});
    }
    return props.client.startDiscovery({ fromIp, toIp });
  };

  const setMinerPool = host => {
    return props.client.setMinerPool({ host, pool: proxyRouterUrl });
  };

  const setMinerPoolForAll = () => {
    notConfiguredDevices.forEach(d =>
      props.client.setMinerPool({ host: d.host, pool: proxyRouterUrl })
    );
    return;
  };

  const stopDiscovery = () => props.client.stopDiscovery({});

  return (
    <View data-testid="devices-container">
      <Container>
        <LayoutHeader
          title="Device discovery"
          address={props.address}
          copyToClipboard={props.client.copyToClipboard}
        />
        <Note>
          For manual configuration, please make sure your computer can be
          reached from the miner's network. Point your mining rigs to your
          computer IP on port {props.sellerPort}.
        </Note>
        <DeviceDiscoveryControl>
          <Sp mb={2}>
            <Flex.Row>
              <Sp mr={1}>
                <Selector
                  disabled={false}
                  onChange={onRangeChange}
                  options={rangeSelectOptions}
                  error={null}
                  label="Range"
                  value={range}
                  id="range"
                />
              </Sp>
              {isInputVisible && (
                <>
                  <Sp mr={1}>
                    <TextInput
                      id="from-ip"
                      label="From IP"
                      value={fromIp}
                      onChange={e => setFromIp(e.value)}
                      disabled={isInputDisabled}
                    />
                  </Sp>
                  <Sp>
                    <TextInput
                      id="to-ip"
                      label="To IP"
                      value={toIp}
                      onChange={e => setToIp(e.value)}
                      disabled={isInputDisabled}
                    />
                  </Sp>
                </>
              )}
            </Flex.Row>
          </Sp>
          <Flex.Row align="center">
            {isDiscovering ? (
              <>
                <Btn onClick={stopDiscovery} style={{ marginRight: '20px' }}>
                  Stop Discovery
                </Btn>
                <Spinner className="discovery-spinner" size="25px" />
              </>
            ) : (
              <>
                <Btn onClick={startDiscovery}>Start Discovery</Btn>
                {notConfiguredDevices.length !== 0 && (
                  <ConfigureBtn onClick={setMinerPoolForAll}>
                    Configure all
                  </ConfigureBtn>
                )}
              </>
            )}
          </Flex.Row>
        </DeviceDiscoveryControl>
        <DeviceDiscoveryResult>
          {!isDiscovering && devices.length === 0 && `No devices found`}
          {isDiscovering && devices.length === 0 && `Discovering...`}
          {devices.map(item => (
            <Device
              key={item.host}
              isLoading={!item.isDone}
              ip={item.host}
              deviceModel={item.deviceModel}
              deviceType={item.deviceType}
              hashRateGHS={item.hashRateGHS}
              isApiAvailable={item.isApiAvailable}
              poolAddress={item.poolAddress}
              poolUser={item.poolUser}
              isPrivilegedApiAvailable={item.isPrivilegedApiAvailable}
              proxyRouterUrl={proxyRouterUrl}
              setMinerPool={setMinerPool}
            />
          ))}
        </DeviceDiscoveryResult>
      </Container>
    </View>
  );
};

export default withDevicesState(Devices);
