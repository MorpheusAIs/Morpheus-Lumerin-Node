import React, { useState } from 'react';
import { abbreviateAddress } from '../../../../utils';
import styled from 'styled-components';
import { RightBtn } from '../CreateContractModal.styles';

const RowContainer = styled.div`
  padding: 1.2rem 0;
  display: grid;
  grid-template-columns: 0.5fr 1fr 1fr 1fr;
  text-align: center;
  box-shadow: 0 -1px 0 0 ${p => p.theme.colors.lightShade} inset;
  color: ${p => p.theme.colors.primary};
  height: 50px;
  font-size: 1.4rem;
`;

const ContractValue = styled.label`
  display: flex;
  padding: 0 1.5rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: ${p => p.theme.colors.primary};
  cursor: pointer;
  text-decoration: underline;
  flex-direction: row;
  gap: 5px;
  font-size: 1.4rem;
`;

const Circle = styled.div`
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: inline-block;

  background: ${p => p.color};
  color: #fff;
  text-align: center;
`;

const FlexCenter = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`;

export const Input = styled.input`
  font-size: 1.4rem;
  padding: 4px 8px;
  outline: 0;
  border: 0px;
  background: #eaf7fc;
  margin: 0 12px;
  width: 35%;
  font-size: 16px;
  font-weight: 400;
  color: #0e4353;
  min-width: 60px;
  ::placeholder {
    color: rgba(1, 67, 83, 0.56);
  }
`;

function AdjustContractRow({ item, onAdjust }) {
  const [price, setPrice] = useState(item.estimatedPrice);
  const profit = item.profit;

  return (
    <RowContainer>
      <FlexCenter>
        <ContractValue>{abbreviateAddress(item.id, 4)}</ContractValue>
      </FlexCenter>
      <FlexCenter>
        <span style={{ color: item.zeroRate > 1 ? 'green' : 'darkred' }}>
          {item.zeroRate > 1
            ? ((item.zeroRate - 1) * 100).toFixed(0)
            : (-(1 - item.zeroRate) * 100).toFixed(0)}{' '}
          %{' '}
        </span>
        {` / ${profit} %`}
      </FlexCenter>
      <FlexCenter>
        {item.price} LMR /
        <Input
          min={0}
          type="number"
          value={price}
          onChange={e => setPrice(e.target.value)}
        ></Input>{' '}
        LMR
      </FlexCenter>
      <FlexCenter>
        <RightBtn
          style={{ height: '90%', 'font-size': '1.4rem' }}
          onClick={() => onAdjust({ id: item.id, price })}
        >
          Apply
        </RightBtn>
      </FlexCenter>
    </RowContainer>
  );
}

export default AdjustContractRow;
