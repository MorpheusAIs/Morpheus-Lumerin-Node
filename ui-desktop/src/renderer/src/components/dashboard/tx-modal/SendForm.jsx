import React, { useState, useContext } from 'react';
import styled from 'styled-components';
import { ToastsContext } from '../../toasts';
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

// The token choice used to be a react-select dropdown carrying its own
// hardcoded palette, which is the one control in this modal that looked like it
// came from a different app. There are exactly two assets, so a segmented
// control shows both at once and needs no styling escape hatch.
const CurrencyToggle = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin: 4px 0 8px;
`;

const CurrencyOption = styled.button`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 10px 8px;
  cursor: pointer;
  border-radius: 5px;
  font-family: inherit;
  background: ${(p) =>
    p.$selected ? 'rgba(32, 220, 142, 0.14)' : 'rgba(255, 255, 255, 0.04)'};
  border: 1px solid
    ${(p) => (p.$selected ? p.theme.colors.morMain : 'rgba(255, 255, 255, 0.1)')};
  color: ${(p) =>
    p.$selected ? p.theme.colors.morMain : 'rgba(255, 255, 255, 0.6)'};
  transition: background 0.15s ease, border-color 0.15s ease, color 0.15s ease;

  &:hover:not(:disabled) {
    border-color: ${(p) => p.theme.colors.morMain};
    color: ${(p) => p.theme.colors.morMain};
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.55;
  }

  &:focus {
    outline: none;
  }
`;

const CurrencyName = styled.span`
  font-size: 1.5rem;
  font-weight: 600;
  letter-spacing: 0.4px;
`;

const CurrencyBalance = styled.span`
  font-size: 1.1rem;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.45);
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
  background: transparent;
  outline: none;
  border: none;
  color: ${(p) => p.theme.colors.morMain};

  ::placeholder {
    color: ${(p) => p.theme.colors.morMain};
    opacity: 0.4;
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
  background: rgba(255, 255, 255, 0.04);
  outline: none;
  border-radius: 5px;
  padding: 8px 20px 6px 60px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: border-color 0.15s ease;

  &:focus {
    border-color: ${(p) => p.theme.colors.morMain};
  }
`;

const SendBtn = styled(BaseBtn)`
  width: 100%;
  height: 50px;
  border-radius: 5px;
  color: ${(p) => p.theme.colors.primaryDark};
  font-weight: 600;
  background-color: ${(p) => p.theme.colors.morMain};

  &:disabled {
    background-color: ${(p) => p.theme.colors.helpertextGray};
    cursor: not-allowed;
  }
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

const ErrorLabel = styled.div`
  color: ${(p) => p.theme.colors.danger};
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
    balanceFor,
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

      <CurrencyToggle role="group" aria-label="Asset to send">
        {(currencyOptions || []).map((option) => {
          const selected = selectedCurrency?.value === option.value;
          return (
            <CurrencyOption
              key={option.value}
              type="button"
              $selected={selected}
              aria-pressed={selected}
              disabled={isPending}
              onClick={() => props.setSelectedCurrency(option)}
            >
              <CurrencyName>{option.label}</CurrencyName>
              <CurrencyBalance>
                {Number(
                  balanceFor ? balanceFor(option.value) : 0,
                ).toLocaleString(undefined, { maximumFractionDigits: 6 })}
              </CurrencyBalance>
            </CurrencyOption>
          );
        })}
      </CurrencyToggle>

      <Column>
        <AmountContainer>
          <AmountInput
            type="number"
            min="0"
            step="any"
            placeholder="0"
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
