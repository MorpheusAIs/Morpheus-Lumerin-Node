import { useState } from 'react';
import styled from 'styled-components';
import { TextInput, AltLayout, AltLayoutNarrow, Btn, Sp } from '../common';
import PasswordStrengthMeter from '../common/PasswordStrengthMeter';
import SecondaryBtn from './SecondaryBtn';
import SetupFeedback from './SetupFeedback';

const Description = styled.p`
  color: var(--text-muted);
  font-size: 1.4rem;
  line-height: 1.6;
`;
export default function PasswordStep(props) {
  const [showPassword, setShowPassword] = useState(false);
  const [suggestions, setSuggestions] = useState('');
  return (
    <AltLayout
      title="Create your app password"
      data-testid="onboarding-container"
    >
      <AltLayoutNarrow>
        <Description>
          Choose a new password to unlock Morpheus on this device. This is not
          your recovery phrase or a password from another wallet app.
        </Description>
        <form
          data-testid="pass-form"
          onSubmit={(event) => {
            event.preventDefault();
            props.onPasswordSubmit();
          }}
        >
          <Sp mt={3}>
            <TextInput
              id="password"
              label="New password"
              type={showPassword ? 'text' : 'password'}
              value={props.password}
              onChange={props.onInputChange}
              error={props.errors.password}
              autoFocus
              autoComplete="new-password"
              data-testid="pass-field"
              disabled={props.isPreparingWallet}
            />
            <PasswordStrengthMeter
              password={props.password}
              onChange={(result) =>
                setSuggestions(result?.suggestions?.join(' ') || '')
              }
            />
            {suggestions && <Description>{suggestions}</Description>}
          </Sp>
          <Sp mt={2}>
            <TextInput
              id="passwordAgain"
              label="Repeat new password"
              type={showPassword ? 'text' : 'password'}
              value={props.passwordAgain}
              onChange={props.onInputChange}
              error={props.errors.passwordAgain}
              autoComplete="new-password"
              data-testid="pass-again-field"
              disabled={props.isPreparingWallet}
            />
          </Sp>
          <Sp mt={2}>
            <SecondaryBtn
              block
              aria-pressed={showPassword}
              onClick={() => setShowPassword(!showPassword)}
            >
              {showPassword ? 'Hide passwords' : 'Show passwords'}
            </SecondaryBtn>
          </Sp>
          <SetupFeedback setupError={props.setupError} />
          <Sp mt={3}>
            <Btn block submit disabled={props.isPreparingWallet}>
              {props.isPreparingWallet
                ? 'Preparing recovery phrase…'
                : props.useImportFlow
                  ? 'Continue to wallet import'
                  : 'Continue to recovery phrase'}
            </Btn>
          </Sp>
          <Sp mt={2}>
            <SecondaryBtn
              block
              onClick={props.onChooseWallet}
              disabled={props.isPreparingWallet}
            >
              Back to wallet options
            </SecondaryBtn>
          </Sp>
        </form>
      </AltLayoutNarrow>
    </AltLayout>
  );
}
