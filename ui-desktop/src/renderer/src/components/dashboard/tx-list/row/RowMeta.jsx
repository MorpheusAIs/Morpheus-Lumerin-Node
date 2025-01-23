//@ts-check

import React from 'react';
import { Field } from './Field';

/** @param {{tx: import("./tx").Tx, tokens: Record<string,{decimals: number, symbol:string}>}} props */
export const ApproveMeta = ({ tx, tokens }) => {
  const token = tokens[tx.contract.contractAddress.toLowerCase()];

  const spender = tx.contract.decodedInput[0].value;
  const amount = Number(tx.contract.decodedInput[1].value);

  const displayAmount = token
    ? formatCurrency({
        value: amount / 10 ** token.decimals,
        currency: token.symbol,
      })
    : tx.contract.decodedInput[1].value;

  return (
    <>
      <Field title="Amount">{displayAmount}</Field>
      <Field title="Spender">{formatAddress(spender)}</Field>
    </>
  );
};

/** @param {{tx: import("./tx").Tx, walletAddress: string}} props */
export const defaultMeta = ({ tx, walletAddress }) => {
  // if no specific meta component is defined for the action
  // display transfers meta if there are any
  if (tx.transfers?.length > 0) {
    return (
      <TransfersMeta transfers={tx.transfers} walletAddress={walletAddress} />
    );
  }

  // if no transfers display all contract arguments
  return (
    <>
      {tx.contract?.decodedInput.map((arg) => (
        <div key={arg.key}>
          <div>{arg.key}</div>
          <div>{arg.value}</div>
        </div>
      ))}
    </>
  );
};

/** @param {{transfers: import("./tx").Transfer[], walletAddress: string}} props */
const TransfersMeta = ({ transfers, walletAddress }) => {
  if (transfers.length > 0) {
    const transfer = transfers[0];
    const direction =
      transfer.from.toLowerCase() === walletAddress.toLowerCase()
        ? 'sent'
        : 'received';
    const sign = direction === 'sent' ? '-' : '+';
    const displayAmount = Number(transfer.value) / 10 ** transfer.tokenDecimals;
    const formatedValue = formatCurrency({
      value: displayAmount,
      currency: transfer.tokenSymbol,
    });

    return (
      <>
        <Field title="Transfer">{`${sign}${formatedValue}`}</Field>
        {direction === 'sent' && (
          <Field title="To">{formatAddress(transfer.to)}</Field>
        )}
        {direction === 'received' && (
          <Field title="From">{formatAddress(transfer.from)}</Field>
        )}
      </>
    );
  }
};

/** @param {string} address */
const formatAddress = (address) => {
  return `${address.slice(0, 6)}...${address.slice(-4)}`;
};

const formatCurrency = ({
  value,
  currency,
  maxSignificantFractionDigits = 5,
}) => {
  let style = 'currency';

  if (!currency) {
    currency = undefined;
    style = 'decimal';
  }

  if (value < 1) {
    return new Intl.NumberFormat(navigator.language, {
      style: style,
      currency: currency,
      maximumSignificantDigits: 5,
    }).format(value);
  }

  const integerDigits = value?.toFixed(0).toString().length;
  let fractionDigits = maxSignificantFractionDigits - integerDigits;
  if (fractionDigits < 0) {
    fractionDigits = 0;
  }

  return new Intl.NumberFormat(navigator.language, {
    style: style,
    currency: currency,
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(value);
};

export const metaComponentMap = {
  approve: ApproveMeta,
  increaseAllowance: ApproveMeta,
  // put here other actions that need a specific meta component
};
