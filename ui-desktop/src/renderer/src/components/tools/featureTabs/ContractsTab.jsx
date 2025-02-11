import React, { useState } from 'react';
import styled from 'styled-components';
import { TextInput, BaseBtn } from '../../common';
const StyledParagraph = styled.p`
  color: ${(p) => p.theme.colors.dark};

  span {
    font-weight: bold;
  }
`;

const StyledBtn = styled(BaseBtn)`
  width: 40%;
  height: 40px;
  font-size: 1.5rem;
  border-radius: 5px;
  padding: 0 0.6rem;
  background-color: ${(p) => p.theme.colors.primary};
  color: ${(p) => p.theme.colors.light};

  @media (min-width: 1040px) {
    width: 35%;
    height: 40px;
    margin-left: 0;
    margin-top: 1.6rem;
  }
`;

const Input = styled(TextInput)`
  outline: 0;
  border: 0px;
  max-width: 10%;
  background: #eaf7fc;
  border-radius: 15px;
  padding: 1.2rem 1.2rem;
  margin-top: 0.25rem;
`;

const Subtitle = styled.h3`
  color: ${(p) => p.theme.colors.dark};
`;

export const ContractsTab = ({ settings, onCommit }) => {
  const [profitSettings, setProfitSettings] = useState(settings);
  const [adaptExisting, setAdaptExisting] = useState(settings?.adaptExisting);
  return (
    <div>
      <Subtitle>Profit Targets</Subtitle>
      <div>
        Profit settings required to track actual contact prices for Sellers.
      </div>
      <StyledParagraph>
        Default level of contract margin profit (%):{' '}
        <Input
          type="number"
          min={0}
          max={50}
          placeholder={10}
          onChange={(e) =>
            setProfitSettings({ ...profitSettings, target: e.value })
          }
          value={profitSettings?.target}
        />
      </StyledParagraph>
      <StyledParagraph>
        Deviation level of target profit acceptable to snooze notification (%):{' '}
        <Input
          type="number"
          min={0}
          max={10}
          placeholder={2}
          onChange={(e) => {
            setProfitSettings({ ...profitSettings, deviation: e.value });
            return;
          }}
          value={profitSettings?.deviation}
        />
      </StyledParagraph>
      <div style={{ display: 'flex' }}>
        <span>Apply Target Profit For Existed Contracts</span>
        <input
          style={{ marginLeft: '10px' }}
          data-testid="use-titan-lightning"
          onChange={() => {
            setAdaptExisting(!adaptExisting);
            setProfitSettings({
              ...profitSettings,
              adaptExisting: !adaptExisting,
            });
          }}
          checked={adaptExisting}
          type="checkbox"
          id="isTitanLightning"
        />
      </div>

      <StyledBtn onClick={() => onCommit(profitSettings)}>Save</StyledBtn>
    </div>
  );
};
