import * as utils from '../../store/utils';
import PropTypes from 'prop-types';

import { TextInput, AltLayout, Btn, Sp } from '../common';
import SecondaryBtn from './SecondaryBtn';
import Message from './Message';
import AltLayoutNarrow from '../common/AltLayoutNarrow';
import SetupFeedback from './SetupFeedback';

const VerifyMnemonicStep = (props) => {
  const id = 'mnemonicAgain';
  return (
    <AltLayout title="Recovery Passphrase" data-testid="onboarding-container">
      <form data-testid="mnemonic-form" onSubmit={props.onMnemonicAccepted}>
        <AltLayoutNarrow>
          <Message>
            To verify you have copied the recovery passphrase correctly, enter
            the 12 words provided before in the field below.
          </Message>
        </AltLayoutNarrow>
        <Sp mt={3}>
          <TextInput
            id={id}
            data-testid="mnemonic-field"
            autoFocus
            onChange={props.onInputChange}
            onPaste={(e) => {
              e.preventDefault();
              const value = e.clipboardData.getData('Text').trim();
              props.onInputChange({ value, id });
            }}
            label="Recovery passphrase"
            error={props.errors.mnemonicAgain}
            value={props.mnemonicAgain || ''}
            rows={2}
            disabled={props.isSubmitting}
          />
        </Sp>
        <AltLayoutNarrow>
          <SetupFeedback {...props} />
          <Sp mt={5}>
            <Btn
              data-rh-negative
              data-disabled={!props.shouldSubmit(props.mnemonicAgain)}
              data-rh={props.getTooltip(props.mnemonicAgain)}
              submit
              disabled={
                props.isSubmitting || !props.shouldSubmit(props.mnemonicAgain)
              }
              block
              key="sendMnemonic"
            >
              {props.isSubmitting ? 'Setting up wallet…' : 'Create wallet'}
            </Btn>
          </Sp>
          <Sp mt={2}>
            <SecondaryBtn
              data-testid="goback-btn"
              onClick={props.onMnemonicCopiedToggled}
              disabled={props.isSubmitting}
              block
            >
              Go back
            </SecondaryBtn>
          </Sp>
        </AltLayoutNarrow>
      </form>
    </AltLayout>
  );
};

VerifyMnemonicStep.propTypes = {
  onMnemonicCopiedToggled: PropTypes.func.isRequired,
  onMnemonicAccepted: PropTypes.func.isRequired,
  onInputChange: PropTypes.func.isRequired,
  mnemonicAgain: PropTypes.string,
  shouldSubmit: PropTypes.func.isRequired,
  getTooltip: PropTypes.func.isRequired,
  errors: utils.errorPropTypes('mnemonicAgain'),
};

export default VerifyMnemonicStep;
