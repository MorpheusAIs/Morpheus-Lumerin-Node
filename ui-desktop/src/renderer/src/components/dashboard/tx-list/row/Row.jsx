import React from 'react';
import styled from 'styled-components';

import withTxRowState from '../../../../store/hocs/withTxRowState';
import Details from './Details';
import Amount from './Amount';
import { TxIcon } from './Icon';
import { LumerinDarkIcon } from '../../../icons/LumerinDarkIcon';
import { LumerinLightIcon } from '../../../icons/LumerinLightIcon';
import { EtherIcon } from '../../../icons/EtherIcon';
import { LumerinLogoFull } from '../../../icons/LumerinLogoFull';

const Container = styled.div`
  margin-left: 1.6rem;
  padding: 1.2rem 2.4rem 1.2rem 0;
  display: grid;
  grid-template-columns: 5fr 5fr 5fr 20fr;
  align-items: center;
  box-shadow: 0 -1px 0 0 ${p => p.theme.colors.lightShade} inset;
  cursor: pointer;
  height: 66px;
`;

const IconContainer = styled.div`
  display: block;
  text-align: center;
  margin: 0 auto;
  flex-shrink: 0;
  width: 40px;
`;

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

const Row = ({ tx, explorerUrl, morAddress }) => {
  const morTransaction = tx.isMor;
  const formatedValue = formatCurrency({ value: tx.value, maxSignificantFractionDigits: morTransaction ? 3 : 5 });

  return (
  	<Container onClick={() => window.openLink(explorerUrl)}>
    	<IconContainer>
      	{morTransaction ? (
        	<LumerinLogoFull size="4rem" />
      	) : (
        	<EtherIcon size="4rem"></EtherIcon>
      	)}
    	</IconContainer>
    	<IconContainer>
      	<TxIcon txType={tx.txType} />
    	</IconContainer>
    	<Amount {...tx} symbol={tx.symbol} value={formatedValue} />
    	<Details {...tx} />
  	</Container>
	);
};

export default withTxRowState(Row);
