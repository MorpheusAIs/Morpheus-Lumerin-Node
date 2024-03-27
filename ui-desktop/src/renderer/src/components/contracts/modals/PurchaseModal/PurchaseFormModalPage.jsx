import React, { useState } from 'react';
import {
  TitleWrapper,
  Title,
  Form,
  InputGroup,
  Row,
  Input,
  RightBtn,
  ContractLink,
  LeftBtn,
  Sublabel,
  ErrorLabel
} from '../CreateContractModal.styles';
import { fromTokenBaseUnitsToETH } from '../../../../utils/coinValue';

import {
  Divider,
  HeaderFlex,
  SmallTitle,
  UrlContainer,
  Values,
  EditBtn,
  PoolInfoContainer,
  UpperCaseTitle,
  ActionsGroup,
  ContractInfoContainer
} from './common.styles';

import { IconExternalLink, IconQuestionCircle } from '@tabler/icons-react';
import { formatDuration, formatSpeed } from '../../utils';
import { fromTokenBaseUnitsToLMR } from '../../../../utils/coinValue';

export const PurchaseFormModalPage = ({
  inputs,
  setInputs,
  onFinished,
  contract,
  rate,
  pool,
  explorerUrl,
  onEditPool,
  handleSubmit,
  register,
  close,
  formState,
  symbol,
  marketplaceFee
}) => {
  const [isEditPool, setIsEditPool] = useState(false);

  const validateAddress = address => {
    const regexP = /^[a-zA-Z0-9.-]+:\d+$/;
    if (!regexP.test(address)) return false;

    const regexPortNumber = /:\d+/;
    const portMatch = address.match(regexPortNumber);
    if (!portMatch) return false;

    const port = portMatch[0].replace(':', '');
    if (Number(port) < 0 || Number(port) > 65536) return false;

    return true;
  };

  const poolParts = pool ? pool.replace('stratum+tcp://', '').split(':@') : [];

  const handleClose = e => {
    e.preventDefault();
    close();
  };

  const submit = data => {
    onFinished();
  };

  return (
    <>
      <TitleWrapper>
        <Title>Purchase Hashpower</Title>
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
            <SmallTitle>Terms</SmallTitle>
            <Values>
              {formatSpeed(contract.speed)} for{' '}
              {formatDuration(contract.length)}
            </Values>
          </div>
          <div>
            <SmallTitle>Price</SmallTitle>
            <Values>
              {fromTokenBaseUnitsToLMR(contract.price)} {symbol} (≈ $
              {(fromTokenBaseUnitsToLMR(contract.price) * rate).toFixed(2)} USD)
            </Values>
            <SmallTitle>
              + {fromTokenBaseUnitsToETH(marketplaceFee)} ETH fee
              <IconQuestionCircle
                data-rh={`All proceeds are subject to a non-refundable ${fromTokenBaseUnitsToETH(
                  marketplaceFee
                )} ETH marketplace fee`}
                width={'1.7rem'}
                style={{ padding: '0 0 1px 4px' }}
              />
            </SmallTitle>
          </div>
        </ContractInfoContainer>
      </TitleWrapper>
      <Form onSubmit={handleSubmit(submit)}>
        <UrlContainer>
          <UpperCaseTitle>validator address (lumerin node)</UpperCaseTitle>
          <Divider />
          {isEditPool ? (
            <Row key="addressRow">
              <InputGroup key="addressGroup">
                <Input
                  {...register('address', {
                    required: true,
                    validate: validateAddress
                  })}
                  placeholder={'HOST_IP:PORT'}
                  type="text"
                  name="address"
                  key="address"
                  id="address"
                />
                {formState?.errors?.address?.type === 'validate' && (
                  <ErrorLabel>Address should match HOST_IP:PORT</ErrorLabel>
                )}
              </InputGroup>
            </Row>
          ) : (
            <PoolInfoContainer>
              <Values>{inputs.address}</Values>
              <EditBtn onClick={() => setIsEditPool(true)}>Edit</EditBtn>
            </PoolInfoContainer>
          )}
        </UrlContainer>
        <UrlContainer style={{ marginTop: '30px' }}>
          <SmallTitle>Worker Name</SmallTitle>
          <Values
            key={contract?.id}
            style={{ width: '85%', wordBreak: 'break-all' }}
          >
            {contract?.id}
          </Values>
        </UrlContainer>
        <UrlContainer style={{ marginTop: '50px' }}>
          <UpperCaseTitle>Forwarding to (mining pool)</UpperCaseTitle>
          <Divider />
          <PoolInfoContainer>
            <div>
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
            <EditBtn onClick={() => onEditPool()}>Edit</EditBtn>
          </PoolInfoContainer>
        </UrlContainer>
        <ActionsGroup>
          <Row style={{ justifyContent: 'space-between', marginTop: '3rem' }}>
            <LeftBtn onClick={handleClose}>Cancel</LeftBtn>
            <RightBtn type="submit" disabled={!formState?.isValid}>
              Review Order
            </RightBtn>
          </Row>
        </ActionsGroup>
      </Form>
    </>
  );
};
