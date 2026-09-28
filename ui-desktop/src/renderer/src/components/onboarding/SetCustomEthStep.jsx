import TermsAndConditions from '../../components/common/TermsAndConditions';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';
import SecondaryBtn from './SecondaryBtn';
import { AltLayout, AltLayoutNarrow, Btn, Sp, TextInput } from '../common';
import Message from './Message';
import SetupFeedback from './SetupFeedback';

const DisclaimerWarning = styled.div`
  text-align: left;
  color: ${(p) => p.theme.colors.dark};
  font-size: 16px;
  margin-top: 16px;
  text-align: justify;
`;

const DisclaimerMessage = styled.div`
  width: 100%;
  height: 130px;
  border-radius: 2px;
  background-color: rgba(0, 0, 0, 0.1);
  color: ${(p) => p.theme.colors.dark};
  overflow: auto;
  font-size: 12px;
  padding: 10px 16px 0 16px;
  margin: 16px 0;
`;

const P = styled.p`
  color: ${(p) => p.theme.colors.dark};
`;

const Subtext = styled.span`
  color: ${(p) => p.theme.colors.dark};
`;

export const SetCustomEthStep = (props) => {
  return (
    <AltLayout
      title="Choose your connection"
      data-testid="onboarding-container"
    >
      <AltLayoutNarrow>
        <DisclaimerWarning>
          Use the default blockchain connection, or enter your own node URL. You
          can change this later in Settings.
        </DisclaimerWarning>

        <Sp mt={3}>
          <TextInput
            data-testid="ethNode-field"
            autoFocus
            onChange={props.onInputChange}
            placeholder={'{wss|https}://{url}'}
            onPaste={(e) => {
              e.preventDefault();
              const value = e.clipboardData.getData('Text').trim();
              props.onInputChange({ value, id: 'customEthNode' });
            }}
            label="Custom ETH node URL (optional)"
            disabled={props.isSubmitting}
            error={props.errors.customEthNode}
            value={props.customEthNode || ''}
            id={'customEthNode'}
          />
        </Sp>

        <SetupFeedback {...props} />
        <Sp mt={6}>
          <Btn
            data-testid="accept-btn"
            autoFocus
            onClick={(e) => props.onEthNodeSet(e)}
            disabled={props.isSubmitting || !props.customEthNode}
            block
          >
            Use custom connection
          </Btn>
        </Sp>
        <Sp mt={2}>
          <SecondaryBtn
            data-testid="skip-btn"
            onClick={(e) => {
              e.preventDefault();
              props.onEthNodeSet(e, true);
            }}
            disabled={props.isSubmitting}
            block
          >
            {props.isSubmitting
              ? 'Setting up wallet…'
              : 'Use default connection'}
          </SecondaryBtn>
        </Sp>
      </AltLayoutNarrow>
    </AltLayout>
  );
};
