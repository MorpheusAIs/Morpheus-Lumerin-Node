import React, { useState, useContext } from 'react';
import styled from 'styled-components';
import { ToastsContext } from '../../toasts';
import Select from 'react-select';
import { explainChainError } from '../../../store/utils/chainErrors';

import BackIcon from '../../icons/BackIcon';
import { BaseBtn } from '../../common';
import Spinner from '../../common/Spinner';
import {
  HeaderWrapper,
  BackBtn,
  Header,
  Footer,
  FooterRow,
  FooterLabel,
} from './common.styles';

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
  background: #03160e !important;
  outline: none;
  border: none;
  color: ${({ isActive, theme }) =>
    isActive ? theme.colors.morMain : theme.colors.morMain};

  ::placeholder {
    color: ${(p) => p.theme.colors.morMain};
  }

  &[type='number']::-webkit-inner-spin-button,
  &[type='number']::-webkit-outer-spin-button {
    -webkit-appearance: none;
    -moz-appearance: none;
    appearance: none;
    margin: 0;
  }
`;
const AmountSublabel = styled.label`
  color: ${(p) => p.theme.colors.dark};
  font-size: 1.4rem;
  text-align: center;
`;

const SubAmount = styled.div`
  color: ${(p) => p.theme.colors.helpertextGray};
  font-size: 13px;
  text-align: center;
`;

const FeeContainer = styled.div`
  display: flex;
  flex-direction: column;
  padding-top: 5px;
`;

const FeeRow = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
`;

const FeeLabel = styled.div`
  font-size: 1.2rem;
  color: ${(p) => p.theme.colors.dark};
`;

const Column = styled.div`
  display: flex;
  flex-direction: column;
`;
const WalletContainer = styled.label`
  display: block;
  position: relative;
`;
const WalletInputLabel = styled.span`
  position: absolute;
  z-index: 1;
  top: 50%;
  font-weight: bold;
  cursor: text;
  pointer-events: none;
  margin-left: 20px;
  -ms-transform: translateY(-50%);
  transform: translateY(-50%);
  color: ${(p) => p.theme.colors.placeholderGray};
`;

const WalletInput = styled.input`
  width: 100%;
  height: 40px;
  color: ${(p) => p.theme.colors.dark};
  font-weight: 300;
  font-size: 16px;
  background: #03160e !important;
  outline: none;
  border-radius: 5px;
  border-style: solid;
  padding: 8px 20px 6px 60px;
  border: none !important;
`;

const SendBtn = styled(BaseBtn)`
  width: 100%;
  height: 50px;
  border-radius: 5px;
  color: black;
  font-weight: 600;
  background-color: ${({ isActive, theme }) =>
    isActive ? theme.colors.helpertextGray : theme.colors.morMain};
`;

const IconContainer = styled.div`
  margin: 0 auto;
  padding: 5px;
  cursor: pointer;
`;

const SendContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-around;
  width: 100%;
  height: 50px;
  margin: 16px 0 0;
`;

const selectorStyles = {
  singleValue: (provided) => ({
    ...provided,
    color: 'white',
  }),
  control: (base) => ({
    ...base,
    borderColor: '#20dc8e',
    color: '#FFFFFF',
    backgroundColor: '#03160e',
    width: '100%',
  }),
  option: (base, state) => ({
    ...base,
    backgroundColor: state.isSelected ? '#03160e' : undefined,
    color: state.isSelected ? '#FFFFFF' : undefined,
    ':active': {
      ...base[':active'],
      backgroundColor: '#0e435380',
      color: '#FFFFFF',
    },
  }),
};

const ErrorLabel = styled.div`
  color: #ff6b6b;
  font-size: 1.2rem;
  text-align: center;
  min-height: 1.6rem;
  padding-top: 4px;
`;

const MaxBtn = styled.button`
  background: none;
  border: none;
  color: ${(p) => p.theme.colors.morMain};
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0 4px;
`;

export function SendForm(props) {
  const [isPending, setIsPending] = useState(false);
  const context = useContext(ToastsContext);

  const {
    selectedCurrency,
    currencyOptions,
    availableBalance,
    errors = {},
  } = props;

  const handleSend = async (e) => {
    e.preventDefault();

    if (isPending) {
      return;
    }

    // Client-side validation first, so obviously-bad input never costs gas.
    if (props.validate()) {
      return;
    }

    setIsPending(true);
    try {
      const txHash = await props.onSubmit();
      if (!txHash) {
        throw new Error('Transfer did not return a transaction hash');
      }
      props.onTabSwitch('success');
    } catch (err) {
      // The main-process transfer handler throws on a non-2xx response, so the
      // real proxy-router message (insufficient funds, bad address, nonce
      // problems) reaches the user instead of a silent no-op.
      //
      // A timeout is the dangerous case: the transaction may already be on
      // chain. Say so explicitly rather than letting the user assume it failed
      // and send a second time.
      const { message, hint } = explainChainError(err);
      const timedOut = /timed out/i.test(String(err?.message ?? ''));
      context.toast(
        'error',
        timedOut
          ? 'Timed out waiting for confirmation. The transaction may still go through — ' +
              'check your transaction list or the block explorer BEFORE sending again.'
          : hint
            ? `${message} ${hint}`
            : message,
        { autoClose: 15000 },
      );
    } finally {
      setIsPending(false);
    }
  };

  const handleDestinationAddressInput = (e) =>
    props.onInputChange({ id: 'toAddress', value: e.target.value });

  const handleAmountInput = (e) =>
    props.onInputChange({ id: 'coinAmount', value: e.target.value });

  if (!props.activeTab) {
    return <></>;
  }

  return (
    <>
      <HeaderWrapper>
        <BackBtn data-modal="send" onClick={props.onRequestClose}>
          <BackIcon size="2.4rem" fill="white" />
        </BackBtn>
        <Header>You are sending</Header>
      </HeaderWrapper>

      <div style={{ color: 'black' }}>
        <Select
          className="basic-single"
          classNamePrefix="select"
          name="currency"
          styles={selectorStyles}
          onChange={props.setSelectedCurrency}
          value={selectedCurrency}
          options={currencyOptions}
          isDisabled={isPending}
        />
      </div>

      <Column>
        <AmountContainer>
          <AmountInput
            type="number"
            min="0"
            step="any"
            placeholder="0"
            isActive={true}
            disabled={isPending}
            onChange={handleAmountInput}
            value={props.coinAmount}
          />
        </AmountContainer>
        <AmountSublabel>{selectedCurrency?.label}</AmountSublabel>
        <ErrorLabel>{errors.coinAmount}</ErrorLabel>

        <FeeContainer>
          <FeeRow>
            <FeeLabel>
              Network fee is paid in {props.symbolEth} and deducted separately.
            </FeeLabel>
          </FeeRow>
        </FeeContainer>
      </Column>

      <WalletContainer>
        <WalletInputLabel>To: </WalletInputLabel>
        <WalletInput
          id="toAddress"
          aria-label="Recipient wallet address"
          placeholder="0x…"
          autoComplete="off"
          autoCapitalize="none"
          spellCheck={false}
          disabled={isPending}
          onChange={handleDestinationAddressInput}
          value={props.toAddress}
        />
      </WalletContainer>
      <ErrorLabel>{errors.toAddress}</ErrorLabel>

      <Footer>
        <FooterRow>
          <FooterLabel>{selectedCurrency?.label} Balance</FooterLabel>
          <FooterLabel>
            {Number(availableBalance || 0).toFixed(6)}
            <MaxBtn
              type="button"
              disabled={isPending}
              onClick={props.onMaxClick}
            >
              MAX
            </MaxBtn>
          </FooterLabel>
        </FooterRow>
        <FooterRow>
          <SendContainer>
            {isPending && <Spinner size="16px" />}
            {!isPending && (
              <SendBtn data-modal="success" onClick={handleSend}>
                Send now
              </SendBtn>
            )}
          </SendContainer>
        </FooterRow>
      </Footer>
    </>
  );
}
