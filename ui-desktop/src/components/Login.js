import React from 'react';
import styled from 'styled-components';

import withLoginState from '../store/hocs/withLoginState';

import { TextInput, AltLayout, BaseBtn, Sp, AltLayoutNarrow } from './common';

const LoginBtn = styled(BaseBtn)`
  font-size: 1.5rem;
  font-weight: bold;
  height: 40px;
  border-radius: 5px;
  background-color: ${p => p.theme.colors.primary};
  color: ${p => p.theme.colors.light};

  @media (min-width: 1040px) {
    margin-left: 0;
    margin-top: 1.6rem;
  }
`;

const SecondaryBtn = styled(BaseBtn)`
  font-size: 1.2rem;
  color: ${p => p.theme.colors.dark};
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
  logout
}) {
  return (
    <AltLayout title="Enter your password">
      <AltLayoutNarrow>
        <form onSubmit={onSubmit} data-testid="login-form">
          <Sp mt={4}>
            <TextInput
              id="password"
              type="password"
              label="Password"
              value={password}
              data-testid="pass-field"
              autoFocus
              onChange={onInputChange}
              error={errors.password || error}
            />
          </Sp>
          <Sp mt={2}>
            <SecondaryBtn onClick={() => logout({})} block>
              Or setup new wallet
            </SecondaryBtn>
          </Sp>
          <Sp mt={4}>
            <LoginBtn block submit disabled={status === 'pending'}>
              Login
            </LoginBtn>
          </Sp>
        </form>
      </AltLayoutNarrow>
    </AltLayout>
  );
}

export default withLoginState(Login);
