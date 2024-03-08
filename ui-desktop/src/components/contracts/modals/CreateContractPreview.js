import React from 'react';
import {
  TitleWrapper,
  Title,
  Row,
  RightBtn,
  LeftBtn
} from './CreateContractModal.styles';
import PriceIcon from '../../icons/PriceIcon';
import SpeedIcon from '../../icons/SpeedIcon';
import DurationIcon from '../../icons/DurationIcon';
import {
  Divider,
  HeaderFlex,
  SmallTitle,
  Values,
  UpperCaseTitle,
  ContractInfoContainer
} from './PurchaseModal/common.styles';
import Spinner from '../../common/Spinner';
import { fromTokenBaseUnitsToETH } from '../../../utils/coinValue';

export const CreateContractPreview = ({
  data: { address, time, speed, price },
  submit,
  isCreating,
  isEditMode,
  close,
  symbol,
  marketplaceFee
}) => (
  <>
    <TitleWrapper>
      <Title>Review Contract</Title>
      <HeaderFlex>
        <UpperCaseTitle>SUMMARY</UpperCaseTitle>
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
            {speed} TH/s
          </Values>
        </div>
        <div>
          <SmallTitle>Duration</SmallTitle>
          <Values>
            <DurationIcon
              key={'duration'}
              style={{ marginRight: '4px', height: '1.4rem' }}
            />
            {time} hours
          </Values>
        </div>
        <div>
          <SmallTitle>Price</SmallTitle>
          <Values>
            <PriceIcon
              key={'price'}
              style={{ marginRight: '4px', height: '1.4rem' }}
            />
            {price} {symbol}
          </Values>
        </div>
      </ContractInfoContainer>
      {!isEditMode ? (
        <SmallTitle style={{ marginTop: '10px' }}>
          All proceeds are subject to a non-refundable{' '}
          {fromTokenBaseUnitsToETH(marketplaceFee)} ETH marketplace fee,
          including contract creation and claiming funds
        </SmallTitle>
      ) : (
        <></>
      )}
    </TitleWrapper>
    {isCreating ? (
      <Row style={{ justifyContent: 'center', marginTop: '3rem' }}>
        <Spinner size="16px" />
      </Row>
    ) : (
      <Row style={{ marginTop: '3rem' }}>
        <LeftBtn onClick={close}>Edit Contract</LeftBtn>
        <RightBtn onClick={submit}>Confirm</RightBtn>
      </Row>
    )}
  </>
);
