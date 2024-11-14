import React, { useState } from 'react';
import styled from 'styled-components';
import PropTypes from 'prop-types';
import 'react-hint/css/index.css';
import * as utils from '../../store/utils';
import {
  PasswordStrengthMeter,
  TextInput,
  AltLayout,
  AltLayoutNarrow,
  Btn,
  Sp,
  Tooltip
} from '../common';
import Message from './Message';

const PasswordMessage = styled(Message)`
  text-align: left;
  color: ${p => p.theme.colors.dark};
  text-align: justify;
`;

const Green = styled.div`
  display: inline-block;
  color: ${p => p.theme.colors.success};
`;

const PasswordInputWrap = styled.div`
  position: relative;
`;

const SecondaryBtn = styled(Btn)`
    border: 1px solid #20dc8e;
    color: #20dc8e;
    background: transparent;
`

const PasswordStep = props => {
  const [typed, setTyped] = useState(false);
  const [suggestion, setSuggestion] = useState('');
  const onPasswordSubmit = (e, useImportFlow) => {
    e.preventDefault();
    props.onPasswordSubmit({ clearOnError: false, useImportFlow });
  };
  let tooltipTimeout;

  return (
    <AltLayout title="Let`s get started" data-testid="onboarding-container">
      <AltLayoutNarrow>
        <form data-testid="pass-form">
          <PasswordMessage>
            Enter a strong password until the meter turns <Green>green</Green>.
          </PasswordMessage>
          <PasswordInputWrap>
            <Sp mt={2}>
              <Tooltip
                content={suggestion}
                show={typed && props.password && suggestion.length}
              />
              <TextInput
                data-testid="pass-field"
                autoFocus
                onChange={e => {
                  if (!typed) {
                    tooltipTimeout && clearTimeout(tooltipTimeout);
                    setTyped(true);
                    tooltipTimeout = setTimeout(() => setTyped(false), 5000);
                  }
                  return props.onInputChange(e);
                }}
                error={props.errors.password}
                label="Password"
                value={props.password}
                type="password"
                id="password"
              />
              {!props.errors.password && (
                <PasswordStrengthMeter
                  password={props.password}
                  onChange={res => {
                    const string = res?.suggestions?.join(`\n`);
                    setSuggestion(string);
                  }}
                />
              )}
            </Sp>
          </PasswordInputWrap>
          <Sp mt={3}>
            <TextInput
              data-testid="pass-again-field"
              onChange={props.onInputChange}
              error={props.errors.passwordAgain}
              label="Repeat password"
              value={props.passwordAgain}
              type="password"
              id="passwordAgain"
            />
          </Sp>
          <Sp mt={6}>
            <Btn block onClick={(e) => onPasswordSubmit(e, false)}>
              Create a new wallet
            </Btn>
          </Sp>
          <Sp style={{ marginTop: '10px'}}>
            <SecondaryBtn block onClick={(e) => onPasswordSubmit(e, true)}>
              Import an existing wallet
            </SecondaryBtn>
          </Sp>
        </form>
      </AltLayoutNarrow>
    </AltLayout>
  );
};

PasswordStep.propTypes = {
  onPasswordSubmit: PropTypes.func.isRequired,
  onInputChange: PropTypes.func.isRequired,
  passwordAgain: PropTypes.string,
  password: PropTypes.string,
  errors: utils.errorPropTypes('passwordAgain', 'password')
};

export default PasswordStep;
