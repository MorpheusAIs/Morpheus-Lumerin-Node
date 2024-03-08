import React, { useState } from 'react';
import styled from 'styled-components';

import withReportsState from '../../store/hocs/withReportsState';
import { LayoutHeader } from '../common/LayoutHeader';
import { Toggle } from './ToggleWrapper';
import { BaseBtn } from '../common';
import { View } from '../common/View';

const Container = styled.div`
  background-color: ${p => p.theme.colors.light};
  min-height: 100%;
  width: 100%;
  position: relative;
  padding: 0 2.4rem;
`;

const Subtitle = styled.h3`
  line-height: 3rem;
  color: ${p => p.theme.colors.darker}
  white-space: nowrap;
  margin: 0 2rem;
  cursor: default;
`;

const ToggleContainer = styled.div`
  margin: 2.4rem 4rem;
  height: 400px;
  background-color: ${p => p.theme.colors.light};
  display: flex;
  flex-direction: column;
  justify-content: space-between;

  @media (min-width: 960px) {
  }
`;
const BtnContainer = styled.div`
  width: 80%;
`;

const GenerateBtn = styled(BaseBtn)`
  width: 120px;
  font-size: 1.5rem;
  padding: .6rem 1rem;
  max-height: 60px;
  margin-right: 2.6rem;
  border-radius: 5px;
  border: 1px solid ${p => p.theme.colors.primary};
  border-radius: 5px;
  background-color: ${p => p.theme.colors.primary}
  color: ${p => p.theme.colors.light}
  float: right;

  @media (min-width: 1040px) {
  }
`;

const Reports = ({ address, copyToClipboard }) => {
  const [activeToggles, setActiveToggles] = useState({
    first: false,
    second: false,
    third: false,
    fourth: false,
    fifth: false,
    sixth: false,
    seventh: false
    // eighth: false,
    // ninth: false
  });

  const handleToggles = e =>
    setActiveToggles({
      ...activeToggles,
      [e.target.name]: !activeToggles[e.target.name]
    });
  const handleGenerateReport = () => {};

  return (
    <View data-testid="reports-container">
      <LayoutHeader
        title="Reports"
        address={address}
        copyToClipboard={copyToClipboard}
      />

      <Subtitle>Generate a report:</Subtitle>
      <ToggleContainer>
        <Toggle
          name="first"
          active={activeToggles['first']}
          toggle={handleToggles}
        />
        <Toggle
          name="second"
          active={activeToggles['second']}
          toggle={handleToggles}
        />
        <Toggle
          name="third"
          active={activeToggles['third']}
          toggle={handleToggles}
        />
        <Toggle
          name="fourth"
          active={activeToggles['fourth']}
          toggle={handleToggles}
        />
        <Toggle
          name="fifth"
          active={activeToggles['fifth']}
          toggle={handleToggles}
        />
        <Toggle
          name="sixth"
          active={activeToggles['sixth']}
          toggle={handleToggles}
        />
        <Toggle
          name="seventh"
          active={activeToggles['seventh']}
          toggle={handleToggles}
        />
        {/* <Toggle name="eighth" active={activeToggles["eighth"]} toggle={handleToggles} />
        <Toggle name="ninth" active={activeToggles["ninth"]} toggle={handleToggles} /> */}
      </ToggleContainer>

      <BtnContainer>
        <GenerateBtn onClick={handleGenerateReport}>
          Generate Report
        </GenerateBtn>
      </BtnContainer>
    </View>
  );
};

export default withReportsState(Reports);
