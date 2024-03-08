import React from 'react';

import { FieldBtn } from './Btn';
import TextInput from './TextInput';
import TxIcon from '../icons/TxIcon';
import Flex from './Flex';
import Sp from './Spacing';

export default function AmountFields({
  coinPlaceholder,
  usdPlaceholder,
  coinSymbol,
  onMaxClick,
  coinAmount,
  usdAmount,
  autoFocus,
  onChange,
  errors
}) {
  return (
    <Flex.Row justify="space-between">
      <Flex.Item grow="1" basis="0">
        <FieldBtn
          data-testid="max-btn"
          tabIndex="-1"
          onClick={onMaxClick}
          float
        >
          MAX
        </FieldBtn>
        <TextInput
          placeholder={coinPlaceholder}
          data-testid="coinAmount-field"
          autoFocus={autoFocus}
          onChange={onChange}
          error={errors.coinAmount}
          value={coinAmount}
          label={`Amount (${coinSymbol})`}
          id="coinAmount"
        />
      </Flex.Item>
      <Sp mt={6} mx={1}>
        <TxIcon />
      </Sp>
      <Flex.Item grow="1" basis="0">
        <TextInput
          placeholder={usdPlaceholder}
          data-testid="usdAmount-field"
          onChange={onChange}
          error={errors.usdAmount}
          value={usdAmount}
          label="Amount (USD)"
          id="usdAmount"
        />
      </Flex.Item>
    </Flex.Row>
  );
}
