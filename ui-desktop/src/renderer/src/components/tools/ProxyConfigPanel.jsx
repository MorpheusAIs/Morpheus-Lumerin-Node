//ts-check

import React, { useState } from 'react';
import { Sp } from '../common';
import Spinner from '../common/Spinner';
import ConfirmProxyConfigModal from './ConfirmProxyConfigModal';
import { generatePoolUrl } from '../../utils';
import { validatePoolAddress } from '../../store/validators';

import 'react-tabs/style/react-tabs.css';
import './styles.css';

import {
  ErrorLabel,
  StyledBtn,
  Subtitle,
  StyledParagraph,
  Input
} from './common';

function RemoteProxyConfig() {
  return (
    <StyledParagraph>
      You are running Wallet wihout Proxy-Router. Reset wallet to setup
      validator node.
    </StyledParagraph>
  );
}

function TitanLightningProxyPanel(props) {
  const { titanLightningDashboard, sellerPoolParts } = props;
  return (
    <>
      <StyledParagraph>
        <div>
          <span>Titan Lightning Address:</span> {sellerPoolParts?.account}{' '}
        </div>
        <p
          style={{
            textDecoration: 'underline',
            cursor: 'pointer'
          }}
          data-tooltip={titanLightningDashboard}
          onClick={() => {
            window.open(titanLightningDashboard, '_blank');
          }}
        >
          Dashboard for Lightning users
        </p>
      </StyledParagraph>
    </>
  );
}

function ProxyConfigView(props) {
  const { isTitanLightning, sellerPoolParts, proxyRouterEditClick } = props;
  return (
    <>
      {isTitanLightning ? (
        <TitanLightningProxyPanel {...props} />
      ) : (
        <StyledParagraph>
          <div>
            <span>Proxy Default Pool:</span> {sellerPoolParts?.pool}{' '}
          </div>
          <div>
            <span>Proxy Default Account:</span> {sellerPoolParts?.account}{' '}
          </div>
        </StyledParagraph>
      )}
      <StyledBtn onClick={proxyRouterEditClick}>Edit</StyledBtn>
    </>
  );
}

function ProxyConfigEdit(props) {
  const {
    isTitanLightning,
    sellerPoolParts,
    setSellerPoolParts,
    errors,
    setErrors
  } = props;

  const onChangePoolAddess = address => {
    const result = validatePoolAddress(address, {});

    setErrors({
      ...errors,
      proxyDefaultPool: result.proxyDefaultPool || null
    });

    setSellerPoolParts({
      ...sellerPoolParts,
      pool: address,
      isTitanLightning
    });
  };

  return (
    <>
      {isTitanLightning ? (
        <StyledParagraph>
          Titan Lightning Address:
          <Input
            placeholder={'bob@getalby.com'}
            onChange={e =>
              setSellerPoolParts({
                ...sellerPoolParts,
                account: e.value
              })
            }
            value={sellerPoolParts?.account}
          />
        </StyledParagraph>
      ) : (
        <>
          <StyledParagraph>
            Proxy Default Pool Host & Port:{' '}
            <Input
              placeholder="example: btc.global.luxor.tech:8888"
              onChange={e => onChangePoolAddess(e.value)}
              value={sellerPoolParts?.pool}
            />
            {errors?.proxyDefaultPool && (
              <ErrorLabel>{errors?.proxyDefaultPool}</ErrorLabel>
            )}
          </StyledParagraph>
          <StyledParagraph>
            Proxy Default Account:
            <Input
              placeholder="bob@getalby.com"
              onChange={e =>
                setSellerPoolParts({
                  ...sellerPoolParts,
                  account: e.value
                })
              }
              value={sellerPoolParts?.account}
            />
          </StyledParagraph>
        </>
      )}
    </>
  );
}

export function ProxyConfigPanel(props) {
  const [errors, setErrors] = useState({});

  return !props.isLocalProxyRouter ? (
    <RemoteProxyConfig />
  ) : (
    <>
      <Sp mt={5}>
        <Subtitle>Proxy-Router Configuration</Subtitle>
        {props.proxyRouterSettings.isFetching ? (
          <Spinner />
        ) : props.proxyRouterSettings.proxyRouterEditMode ? (
          <>
            <div style={{ display: 'flex' }}>
              <span>Use Titan Pool for Lightning Payouts</span>
              <input
                style={{ marginLeft: '10px' }}
                data-testid="use-titan-lightning"
                onChange={() => {
                  props.toggleIsLightning();
                }}
                checked={props.isTitanLightning}
                type="checkbox"
                id="isTitanLightning"
              />
            </div>
            <ProxyConfigEdit {...props} errors={errors} setErrors={setErrors} />
            <hr></hr>
            <StyledBtn
              disabled={!!errors?.proxyDefaultPool}
              onClick={() => {
                props.setProxyRouterSettings({
                  ...props.proxyRouterSettings,
                  isTitanLightning: props.isTitanLightning,
                  sellerDefaultPool: generatePoolUrl(
                    props.sellerPoolParts.account,
                    !props.isTitanLightning
                      ? props.sellerPoolParts.pool
                      : props.titanLightningPool
                  )
                });
                props.onActiveModalClick('confirm-proxy-restart');
              }}
            >
              Save
            </StyledBtn>
          </>
        ) : (
          <ProxyConfigView {...props} />
        )}

        <ConfirmProxyConfigModal
          onRequestClose={props.onCloseModal}
          onConfirm={props.confirmProxyRouterRestart}
          onLater={props.saveProxyRouterConfig}
          isOpen={props.state.activeModal === 'confirm-proxy-restart'}
        />
      </Sp>
      <Sp mt={5}>
        <Subtitle>Restart Proxy Router</Subtitle>
        <StyledParagraph>Restart the connected Proxy Router.</StyledParagraph>
        {props.isRestarting ? (
          <Spinner size="20px" />
        ) : (
          <StyledBtn
            onClick={() =>
              props.onActiveModalClick('confirm-proxy-direct-restart')
            }
          >
            Restart
          </StyledBtn>
        )}
        <ConfirmProxyConfigModal
          onRequestClose={props.onCloseModal}
          onConfirm={props.onRestartClick}
          onLater={props.onCloseModal}
          isOpen={props.state.activeModal === 'confirm-proxy-direct-restart'}
        />
      </Sp>
    </>
  );
}
