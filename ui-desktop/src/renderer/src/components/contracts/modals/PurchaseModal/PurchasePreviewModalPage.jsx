import React, { useEffect, useState, useContext } from 'react';
import styled from 'styled-components';
import {
  TitleWrapper,
  Title,
  InputGroup,
  Row,
  RightBtn,
  ContractLink,
  LeftBtn,
  InstructionLink
} from '../CreateContractModal.styles';
import Spinner from '../../../common/Spinner';
import { IconExternalLink, IconAlertTriangle, IconCheck } from '@tabler/icons';
import { formatDuration, formatSpeed, formatPrice } from '../../utils';
import PriceIcon from '../../../icons/PriceIcon';
import SpeedIcon from '../../../icons/SpeedIcon';
import DurationIcon from '../../../icons/DurationIcon';
import { fromTokenBaseUnitsToLMR } from '../../../../utils/coinValue';
import {
  Divider,
  HeaderFlex,
  SmallTitle,
  UrlContainer,
  Values,
  PreviewCont,
  PoolInfoContainer,
  UpperCaseTitle,
  ActionsGroup,
  ContractInfoContainer
} from './common.styles';
import { fromTokenBaseUnitsToETH } from '../../../../utils/coinValue';
import logger from '@lumerin/wallet-core/src/logger';
import { ToastsContext } from '../../../toasts';
import theme from '../../../../ui/theme';

const calculateAddress = (address, contractId) => {
  if (!address || !contractId) return null;
  const noProtocolPart = address.replace('stratum+tcp://', '');
  return `stratum+tcp://${contractId}:@${noProtocolPart}`;
};

const FlexCenter = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`;

export const PurchasePreviewModalPage = ({
  explorerUrl,
  contract,
  pool,
  inputs,
  isPurchasing,
  onBackToForm,
  onPurchase,
  symbol,
  marketplaceFee,
  isProxyPortPublic,
  portCheckErrorLink
}) => {
  const context = useContext(ToastsContext);

  const [isPortPublic, setPortPublic] = useState(null);
  const [loading, setLoading] = useState(true);

  const [ip, port] = inputs.address.replace('stratum+tcp://', '').split(':');
  const poolParts = pool ? pool.replace('stratum+tcp://', '').split(':@') : [];

  useEffect(() => {
    isProxyPortPublic({
      ip,
      port: +port
    })
      .then(data => {
        setLoading(false);
        setPortPublic(data);
      })
      .catch(err => {
        logger.error(
          `Failed to check port availability: ${err.message || err}`
        );
        context.toast('error', 'Failed to check port availability');
        setLoading(false);
      });
  }, []);

  return (
    <>
      <TitleWrapper>
        <Title>Review Purchase</Title>
        <HeaderFlex>
          <UpperCaseTitle>Order summary</UpperCaseTitle>
          <ContractLink onClick={() => window.openLink(explorerUrl)}>
            <span style={{ marginRight: '4px' }}>View contract</span>
            <IconExternalLink width={'1.4rem'} />
          </ContractLink>
        </HeaderFlex>
        <Divider />
        <ContractInfoContainer>
          <div>
            <SmallTitle>Speed</SmallTitle>
            <Values>
              <SpeedIcon
                key={'speed'}
                style={{ marginRight: '4px', height: '1.4rem' }}
              />
              {formatSpeed(contract.speed)}
            </Values>
          </div>
          <div>
            <SmallTitle>Duration</SmallTitle>
            <Values>
              <DurationIcon
                key={'duration'}
                style={{ marginRight: '4px', height: '1.4rem' }}
              />
              {formatDuration(contract.length)}
            </Values>
          </div>
          <div>
            <SmallTitle>Price</SmallTitle>
            <Values>
              <PriceIcon
                key={'price'}
                style={{ marginRight: '4px', height: '1.4rem' }}
              />
              {fromTokenBaseUnitsToLMR(contract.price)} {symbol}
              <span style={{ fontSize: '1rem', marginLeft: '3px' }}>
                (+ {fromTokenBaseUnitsToETH(marketplaceFee)} ETH fee)
              </span>
            </Values>
          </div>
        </ContractInfoContainer>
      </TitleWrapper>
      {loading ? (
        <FlexCenter>
          <Spinner size="20px"></Spinner>
        </FlexCenter>
      ) : (
        <PreviewCont>
          <UrlContainer>
            <UpperCaseTitle>validator address (lumerin node)</UpperCaseTitle>
            <Divider />
            <PoolInfoContainer>
              <Values style={{ wordBreak: 'break-all' }}>
                {calculateAddress(inputs.address, contract.id)}
              </Values>
            </PoolInfoContainer>
            {!isPortPublic ? (
              <>
                <InstructionLink
                  onClick={() => window.openLink(portCheckErrorLink)}
                >
                  <IconAlertTriangle
                    width={'18px'}
                    color={theme.colors.warning}
                    overflow={'visible'}
                  />
                  <span style={{ marginRight: '4px' }}>
                    Your node cannot receive inbound hashrate.
                  </span>{' '}
                </InstructionLink>
                <InstructionLink
                  onClick={() => window.openLink(portCheckErrorLink)}
                >
                  <span>
                    {' '}
                    Please check your network before confirming purchase!
                  </span>
                  <IconExternalLink width={'14px'} overflow={'visible'} />
                </InstructionLink>
              </>
            ) : (
              <>
                <InstructionLink style={{ cursor: 'default' }}>
                  <IconCheck
                    width={'18px'}
                    color={theme.colors.primary}
                    overflow={'visible'}
                  />
                  <span>
                    Your node is configured properly to receive inbound
                    hashrate, please confirm your purchase!
                  </span>
                </InstructionLink>
              </>
            )}
          </UrlContainer>
          <UrlContainer>
            <UpperCaseTitle>forwarding to (mining pool)</UpperCaseTitle>
            <Divider />
            <div style={{ marginTop: '10px' }}>
              <SmallTitle>Pool Address</SmallTitle>
              <Values style={{ wordBreak: 'break-all' }}>
                {decodeURIComponent(
                  poolParts[1] || 'Validation node default pool address'
                )}
              </Values>
              <br />
              <SmallTitle>Account</SmallTitle>
              <Values style={{ wordBreak: 'break-all' }}>
                {decodeURIComponent(poolParts[0] || '')}
              </Values>
            </div>
          </UrlContainer>
          <ActionsGroup>
            {isPurchasing ? (
              <Row style={{ justifyContent: 'center', marginTop: '3rem' }}>
                <Spinner size="16px" />
              </Row>
            ) : (
              <Row
                style={{ justifyContent: 'space-between', marginTop: '3rem' }}
              >
                <LeftBtn onClick={onBackToForm}>Edit Order</LeftBtn>
                <RightBtn onClick={onPurchase}>Confirm Purchase</RightBtn>
              </Row>
            )}
          </ActionsGroup>
        </PreviewCont>
      )}
    </>
  );
};
