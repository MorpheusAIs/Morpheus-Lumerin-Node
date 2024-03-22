import React from 'react';
import styled from 'styled-components';

import { BaseBtn } from '../../common';
import { abbreviateAddress } from '../../../utils';
import { SuccessLayer } from './SuccessLayer';
import { toUSD } from '../../../store/utils/syncAmounts';

const SuccessImage = styled.div`
  margin: 0 auto;
`;

const HeaderWrapper = styled.div`
  display: flex;
  position: relative;
  height: 10%;
  align-content: center;
  margin-bottom: 40px;
`;

const Header = styled.div`
  font-size: 1.6rem;
  font-weight: bold;
  color: ${p => p.theme.colors.dark};
  text-align: center;
  width: 100%;
`;

const AmountContainer = styled.label`
  display: block;
  position: relative;
  font-weight: bold;
`;

const AmountInput = styled.input`
  display: flex;
  font-weight: bold;
  font-size: 4rem;
  width: 100%;
  text-align: center;
  outline: none;
  border: none;
  color: ${({ isActive, theme }) =>
    isActive ? theme.colors.primary : theme.colors.dark};

  ::placeholder {
    color: ${p => p.theme.colors.dark};
  }
`;

const DoneBtn = styled(BaseBtn)`
  width: 100%;
  height: 50px;
  border-radius: 5px;
  background-color: ${({ isActive, theme }) =>
    isActive ? theme.colors.lumerin.helpertextGray : theme.colors.primary};
`;

const Column = styled.div`
  display: flex;
  flex-direction: column;
`;

const Footer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: left;
  text-align: center;
`;

const FooterLabel = styled.label`
  color: ${p => p.theme.colors.dark};
  font-size: 1.2rem;
  font-weight: 600;
  margin-bottom: 5px;
`;

const SubAmount = styled.div`
  color: ${p => p.theme.colors.lumerin.helpertextGray};
  font-size: 13px;
  text-align: center;
`;

export function SuccessForm(props) {
  const LMRtoUSD = val => {
    return toUSD(val, props.coinPrice);
  };

  if (!props.activeTab) {
    return <></>;
  }

  const onDone = () => {
    props.onRequestClose();
    props.resetForm();
  };

  return (
    <>
      <Column>
        <HeaderWrapper>
          <Header>Success</Header>
        </HeaderWrapper>
        <SuccessImage>
          <SuccessLayer />
        </SuccessImage>
      </Column>

      <Column>
        <AmountContainer>
          {/* <Currency isActive={props.amountInput > 0}>$</Currency> */}
          <AmountInput
            placeholder={0}
            isActive={props.coinAmount > 0}
            value={props.coinAmount}
          />
        </AmountContainer>
        {/* <SubAmount>≈ {LMRtoUSD(props.coinAmount)}</SubAmount> */}
      </Column>

      <Footer>
        <FooterLabel>
          You have successfully transferred {props.symbol} to{' '}
          {abbreviateAddress(props.toAddress)}
        </FooterLabel>
        <DoneBtn data-modal={null} onClick={onDone}>
          Done
        </DoneBtn>
      </Footer>
    </>
  );
}
