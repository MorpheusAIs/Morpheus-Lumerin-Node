import { useState } from 'react';
import styled from 'styled-components';

import withLoginState from '../store/hocs/withLoginState';

import { TextInput, AltLayout, BaseBtn, Sp, AltLayoutNarrow } from './common';
import ConfirmWalletReset from './common/ConfirmWalletReset';

const LoginBtn = styled(BaseBtn)`
  font-size: 1.5rem;
  font-weight: bold;
  min-height: 44px;
  padding: 1rem 1.6rem;
  border-radius: 10px;
  background-color: ${(p) => p.theme.colors.morMain};
  color: black;

  @media (min-width: 1040px) {
    margin-left: 0;
    margin-top: 1.6rem;
  }
`;

const SecondaryBtn = styled(BaseBtn)`
  font-size: 1.3rem;
  line-height: 1.5;
  min-height: 32px;
  color: ${(p) => p.theme.colors.dark};
  :hover {
    opacity: 0.75;
  }
`;

function Login({
  onInputChange,
  onSubmit,
  password,
  errors,
  status,
  error,
  logout,
}) {
  const [confirmReset, setConfirmReset] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  return (
    <AltLayout title="Enter your password">
      <p
        style={{
          color: 'var(--text-muted)',
          textAlign: 'center',
          fontSize: '1.4rem',
          margin: '0 0 2rem',
        }}
      >
        Unlock your wallet to continue. Your projects stay on this device.
      </p>
      <ConfirmWalletReset
        open={confirmReset}
        fromLogin
        onCancel={() => setConfirmReset(false)}
        onConfirm={() => logout({})}
      />
      <AltLayoutNarrow>
        <form onSubmit={onSubmit} data-testid="login-form">
          <Sp mt={4}>
            <TextInput
              id="password"
              type={showPassword ? 'text' : 'password'}
              label="Password"
              value={password}
              data-testid="pass-field"
              autoFocus
              autoComplete="current-password"
              onChange={onInputChange}
              error={errors.password || error}
            />
          </Sp>
          <Sp mt={1}>
            <SecondaryBtn
              aria-pressed={showPassword}
              onClick={() => setShowPassword(!showPassword)}
            >
              {showPassword ? 'Hide password' : 'Show password'}
            </SecondaryBtn>
          </Sp>
          <Sp mt={2}>
            {/* Was labelled "Or setup new wallet", which reads as additive.
                It deletes the existing wallet and restarts into onboarding. */}
            <SecondaryBtn
              type="button"
              onClick={() => setConfirmReset(true)}
              block
            >
              Forgot password? Erase wallet and start over
            </SecondaryBtn>
          </Sp>
          <Sp mt={4}>
            <LoginBtn block submit disabled={status === 'pending'}>
              {status === 'pending' ? 'Unlocking wallet…' : 'Login'}
            </LoginBtn>
          </Sp>
        </form>
      </AltLayoutNarrow>
    </AltLayout>
  );
}

export default withLoginState(Login);
