import React from 'react';
import styled from 'styled-components';
import { Btn } from '../../common';

export const ContractsRowContainer = styled.div`
  &:hover {
    background-color: rgb(234, 247, 252);
  }
`;

export const SmallAssetContainer = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin: 0 auto;
`;

export const ActionButtons = styled.div`
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
`;

export const ActionButton = styled(Btn)`
  line-height: 1.5rem;
  font-size: 1.2rem;
  letter-spacing: 1px;
  padding: 0.8rem 2.25rem;
`;
