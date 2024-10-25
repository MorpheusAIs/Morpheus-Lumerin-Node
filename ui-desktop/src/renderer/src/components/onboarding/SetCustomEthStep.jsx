import TermsAndConditions from '../../components/common/TermsAndConditions';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';
import SecondaryBtn from './SecondaryBtn';
import { AltLayout, AltLayoutNarrow, Btn, Sp, TextInput } from '../common';
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
  background-color: rgba(0, 0, 0, 0.1);
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
`;

export const SetCustomEthStep = props => {

  return (
    <AltLayout title="ETH Node Url" data-testid="onboarding-container">
      <AltLayoutNarrow>
        <DisclaimerWarning>
          Set Custom ETH node url that will be used for blockchain interactions instead of default. This can be set later in Settings
        </DisclaimerWarning>

        <Sp mt={3}>
          <TextInput

            data-testid="ethNode-field"
            autoFocus
            onChange={props.onInputChange}
            placeholder={"{wss|https}://{url}"}
            onPaste={e => {
              e.preventDefault();
              const value = e.clipboardData.getData('Text').trim();
              console.log(value);
              props.onInputChange({ value, id: 'customEthNode' });
            }}
            label="Custom ETH Node Url"
            error={props.errors.customEthNode}
            value={props.customEthNode || ''}
            id={'customEthNode'}
          />
        </Sp>

        <Sp mt={6}>
          <Btn
            data-testid="accept-btn"
            autoFocus
            onClick={props.onEthNodeSet}
            block
          >
            Accept
          </Btn>
        </Sp>
        <Sp mt={2}>
            <SecondaryBtn
              data-testid="skip-btn"
              onClick={(e) => {
                props.onInputChange({ value: "", id: 'customEthNode' });
                props.onEthNodeSet(e)
              }}
              block
            >
              Skip
            </SecondaryBtn>
          </Sp>
      </AltLayoutNarrow>
    </AltLayout>
  );
};
