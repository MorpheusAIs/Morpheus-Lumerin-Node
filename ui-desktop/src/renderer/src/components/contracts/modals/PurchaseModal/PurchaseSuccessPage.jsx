import React from 'react';
import styled from 'styled-components';

import { BaseBtn } from '../../../common';
import { SuccessLayer } from '../../../dashboard/tx-modal/SuccessLayer';
import { abbreviateAddress } from '../../../../utils';
import { fromTokenBaseUnitsToLMR } from '../../../../utils/coinValue';
const SuccessImage = styled.div`
  margin: 0 auto;
  margin-bottom: 40px;
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

export function PurchaseSuccessPage(props) {
  const onDone = () => {
    props.close();
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

      <Column></Column>

      <Footer>
        <FooterLabel>
          You have successfully purchased {abbreviateAddress(props.contractId)}{' '}
          contract with {fromTokenBaseUnitsToLMR(props.price)} {props.symbol}
          {/* {abbreviateAddress(props.toAddress)} */}
        </FooterLabel>
        <DoneBtn data-modal={null} onClick={onDone}>
          Done
        </DoneBtn>
      </Footer>
    </>
  );
}
