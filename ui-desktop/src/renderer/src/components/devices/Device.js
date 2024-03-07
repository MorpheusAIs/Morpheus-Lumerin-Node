import React from 'react';
import styled from 'styled-components';
import { BaseBtn } from '../common';
import Spinner from '../common/Spinner';

const DeviceContainer = styled.div`
  border: 1px solid #ccc;
  width: 11em;
  margin: 1em 0.5em 0 0;
  color: ${p => p.theme.colors.primary};
  text-align: right;
  border-radius: 5px;

  .row {
    font-size: 0.8em;
    margin: 0;
    padding: 0.4em 0.2em;
  }

  .row:not(:last-child) {
    border-bottom: 1px solid #ccc;
  }

  dl {
    display: flex;
    justify-content: space-between;
  }

  dd {
    font-weight: bold;
    margin: 0;
  }

  .image-row {
    position: relative;
    div {
      display: block;
      padding: 0;
    }
  }

  .image-row img {
    max-width: 100%;
    display: block;
  }

  .image-row .spinner {
    position: absolute;
    bottom: 0.5em;
    right: 0.3em;
  }

  .status-row {
    font-weight: bold;
  }

  .statistics-row {
    .hash-rate {
      margin: 0;
      padding: 0;
      font-size: 1.5em;
    }

    .ip {
      margin: 0;
      padding: 0;
    }
  }

  .pool-row {
    overflow-x: scroll;
  }

  .pool-user-row {
    overflow-x: scroll;
  }
`;

const Btn = styled(BaseBtn)`
  display: block;
  font-size: 1.2rem;
  padding: 5px;
  margin: 5px auto;

  border-radius: 5px;
  border: 1px solid ${p => p.theme.colors.primary};
  background-color: ${p => p.theme.colors.light};
  color: ${p => p.theme.colors.primary};
`;

const Device = ({
  imageSrc = 'images/ant-miner.jpg',
  deviceModel = 'Model unavailable',
  deviceType = 'Type unavailable',
  isApiAvailable,
  ip = 'IP address unavailable',
  hashRateGHS = 0,
  poolAddress = 'Pool address not available',
  poolUser = 'Username not available',
  isLoading,
  status = 'Miner status unavailable',
  isPrivilegedApiAvailable,
  proxyRouterUrl,
  setMinerPool
}) => {
  return (
    <DeviceContainer>
      <div className="image-row">
        <img src={imageSrc} alt="miner-img" />
        {isLoading && <Spinner className="spinner" size="1em"></Spinner>}
      </div>
      {/* <div className="status-row row">{status}</div> */}
      <div className="statistics-row row">
        <p className="hash-rate">{hashRateGHS} Gh/s</p>
        <p className="ip">{ip}</p>
      </div>
      <dl className="type-row row">
        <dd>Type </dd>
        <dt>{deviceType}</dt>
      </dl>
      <dl className="name-row row">
        <dd>Model </dd>
        <dt>{deviceModel}</dt>
      </dl>
      <dl className="api-row row">
        <dd>Api access </dd>
        <dt>{isApiAvailable ? 'true' : 'false'}</dt>
      </dl>
      <dl className="api-row row">
        <dd>Priviliged access </dd>
        <dt>{isPrivilegedApiAvailable ? 'true' : 'false'}</dt>
      </dl>
      <div className="pool-row row">{poolAddress}</div>
      <div className="pool-user-row row">{poolUser}</div>
      {poolAddress !== proxyRouterUrl && !isLoading && (
        <Btn
          data-disabled={!isPrivilegedApiAvailable}
          data-testid="configure-miner-btn"
          data-rh={
            isPrivilegedApiAvailable
              ? null
              : "Your local network doesn't have privileged API access. (W:192.168.0.0/24, W:127.0.0.1/24)"
          }
          onClick={() => setMinerPool(ip)}
        >
          Set ProxyRouter pool
        </Btn>
      )}
    </DeviceContainer>
  );
};

export default Device;
