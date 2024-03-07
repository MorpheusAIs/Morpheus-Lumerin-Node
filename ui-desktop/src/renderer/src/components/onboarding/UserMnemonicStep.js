import PropTypes from 'prop-types';
import React from 'react';

import * as utils from '../../store/utils';
import { TextInput, AltLayout, Btn, Sp } from '../common';
import SecondaryBtn from './SecondaryBtn';
import Message from './Message';
import AltLayoutNarrow from '../common/AltLayoutNarrow';

const UserMnemonic = props => {
  const id = 'userMnemonic';
  return (
    <AltLayout title="Recovery Mnemonic" data-testid="onboarding-container">
      <form data-testid="mnemonic-form" onSubmit={props.onMnemonicAccepted}>
        <AltLayoutNarrow>
          <Message>
            Enter a valid 12 word mnemonic to recover a previously created
            wallet.
          </Message>
        </AltLayoutNarrow>
        <Sp mt={3}>
          <TextInput
            data-testid="mnemonic-field"
            autoFocus
            onChange={props.onInputChange}
            onPaste={e => {
              e.preventDefault();
              const value = e.clipboardData.getData('Text').trim();
              console.log(value);
              props.onInputChange({ value, id });
            }}
            label="Recovery Mnemonic"
            error={props.errors.userMnemonic}
            value={props.userMnemonic || ''}
            rows={2}
            id={id}
          />
        </Sp>
        <AltLayoutNarrow>
          <Sp mt={5}>
            <Btn
              data-rh-negative
              data-disabled={!props.shouldSubmit(props.userMnemonic)}
              data-rh={props.getTooltip(props.userMnemonic)}
              submit={props.shouldSubmit(props.userMnemonic)}
              block
            >
              Recover
            </Btn>
          </Sp>
          <Sp mt={2}>
            <SecondaryBtn
              data-testid="cancel-btn"
              onClick={props.onUseUserMnemonicToggled}
              block
            >
              Cancel
            </SecondaryBtn>
          </Sp>
        </AltLayoutNarrow>
      </form>
    </AltLayout>
  );
};

UserMnemonic.propTypes = {
  onUseUserMnemonicToggled: PropTypes.func.isRequired,
  onMnemonicAccepted: PropTypes.func.isRequired,
  onInputChange: PropTypes.func.isRequired,
  userMnemonic: PropTypes.string,
  shouldSubmit: PropTypes.func.isRequired,
  getTooltip: PropTypes.func.isRequired,
  errors: utils.errorPropTypes('userMnemonic')
};

export default UserMnemonic;
