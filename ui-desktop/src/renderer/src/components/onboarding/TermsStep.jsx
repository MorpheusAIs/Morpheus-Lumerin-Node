import TermsAndConditions from '../../components/common/TermsAndConditions';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

import { AltLayout, AltLayoutNarrow, Btn, Sp } from '../common';
import Message from './Message';

const DisclaimerWarning = styled.div`
  text-align: left;
  color: ${p => p.theme.colors.dark};
  font-size: 16px;
  margin-top: 16px;
  text-align: justify;
`;

const DisclaimerMessage = styled.div`
  width: 100%;
  height: 130px;
  border-radius: 2px;
  background-color: rgba(0, 0, 0, 0.5);
  color: ${p => p.theme.colors.dark};
  overflow: auto;
  font-size: 12px;
  padding: 10px 16px 0 16px;
  margin: 16px 0;
`;

const P = styled.p`
  color: ${p => p.theme.colors.dark};
`;

const Subtext = styled.span`
  color: ${p => p.theme.colors.dark};
  margin-left: 5px;
`;

const TermsStep = props => {
  const onCheckboxToggle = e => {
    props.onInputChange({ id: e.target.id, value: e.target.checked });
  };

  return (
    <AltLayout title="Accept to Continue" data-testid="onboarding-container">
      <AltLayoutNarrow>
        <DisclaimerWarning>
          Please read and accept these terms and conditions.
        </DisclaimerWarning>

        <DisclaimerMessage>
          <TermsAndConditions ParagraphComponent={props => <P {...props} />} />
        </DisclaimerMessage>

        <Message>
          <div style={{ display: 'flex' }}>
            <input
              data-testid="accept-terms-chb"
              onChange={onCheckboxToggle}
              checked={props.termsCheckbox}
              type="checkbox"
              id="termsCheckbox"
            />
            <Subtext>I have read and accept these terms</Subtext>
          </div>
          <div style={{ display: 'flex' }}>
            <input
              data-testid="accept-license-chb"
              onChange={onCheckboxToggle}
              checked={props.licenseCheckbox}
              type="checkbox"
              id="licenseCheckbox"
            />
            <Subtext>I have read and accept the</Subtext>
            <a onClick={props.onTermsLinkClick} style={{ marginLeft: '5px' }}>
              software license
            </a>
          </div>
        </Message>

        <Sp mt={6}>
          <Btn
            data-testid="accept-terms-btn"
            autoFocus
            disabled={!props.licenseCheckbox || !props.termsCheckbox}
            onClick={props.onTermsAccepted}
            block
          >
            Accept
          </Btn>
        </Sp>
      </AltLayoutNarrow>
    </AltLayout>
  );
};

TermsStep.propTypes = {
  onTermsLinkClick: PropTypes.func.isRequired,
  onTermsAccepted: PropTypes.func.isRequired,
  licenseCheckbox: PropTypes.bool.isRequired,
  termsCheckbox: PropTypes.bool.isRequired,
  onInputChange: PropTypes.func.isRequired
};

export default TermsStep;
