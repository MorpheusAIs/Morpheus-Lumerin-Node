import React from 'react';
import {
  BalanceOuterWrap,
  BalanceRow,
  BalanceWrap,
  CurrencySpan,
  EquivalentUSDRow
} from './Balance.styles';

/**
 * Formats as currency and trims if possible the fraction part of the value to display provided number of significant digits
 * @param {Object} params
 * @param {number} params.value currency value
 * @param {string} [params.currency] currency name, if omitted the currency will not be appended
 * @param {number} [params.maxSignificantFractionDigits=5] max number of significant fractdigits
 * @returns {string} formatted string
 */
const formatCurrency = ({
  value,
  currency,
  maxSignificantFractionDigits = 5
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
      maximumSignificantDigits: 5
    }).format(value);
  }

  const integerDigits = value.toFixed(0).toString().length;
  let fractionDigits = maxSignificantFractionDigits - integerDigits;
  if (fractionDigits < 0) {
    fractionDigits = 0;
  }

  return new Intl.NumberFormat(navigator.language, {
    style: style,
    currency: currency,
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits
  }).format(value);
};

export const Balance = ({
  value,
  currency,
  icon,
  equivalentUSD,
  maxSignificantFractionDigits = 5
}) => {
  return (
    <BalanceOuterWrap>
      {icon}
      <BalanceWrap>
        <BalanceRow>
          {formatCurrency({ value, maxSignificantFractionDigits })}&nbsp;
          <CurrencySpan>{currency == "saLMR" ? "saMOR" : currency}</CurrencySpan>
        </BalanceRow>
        {/* <EquivalentUSDRow>≈ {equivalentUSD}</EquivalentUSDRow> */}
      </BalanceWrap>
    </BalanceOuterWrap>
  );
};
