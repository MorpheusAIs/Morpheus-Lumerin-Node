import FilteredMessage from '../../components/common/FilteredMessage';
import * as validators from '../../store/validators';
import { withClient } from '../../store/hocs/clientContext';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import theme from '../../ui/theme';
import React, { useState } from 'react';

import { LoadingBar, TextInput, BaseBtn, Flex, Btn, Sp } from './index';
import CheckIcon from '../icons/CheckIcon';
import CloseIcon from '../icons/CloseIcon';

const ConfirmationTitle = styled.h1`
  font-size: 1.6rem;
  font-weight: 600;
  margin: 0 0 1.6rem 0;
`;

const Title = styled.div`
  line-height: 3rem;
  font-size: 2.4rem;
  font-weight: bold;
  text-align: center;
  cursor: default;
  text-shadow: 0 1px 1px ${p => p.theme.colors.darkShade};
`;

const Message = styled.div`
  line-height: 1.6rem;
  font-size: 1.3rem;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-align: center;
  text-shadow: 0 1px 1px ${p => p.theme.colors.darkShade};
`;

const EditBtn = styled(BaseBtn)`
  margin: 1.6rem auto;
  display: block;
  font-size: 1.4rem;
  opacity: 0.7;
  font-weight: 600;
  letter-spacing: 1.4px;
  line-height: 1.8rem;
  text-transform: uppercase;
`;

const TryAgainBtn = styled(BaseBtn)`
  color: ${p => p.theme.colors.primary};
  margin-top: 1.6rem;
  font-size: 1.4rem;
`;

const BtnContainer = styled.div`
  background-image: linear-gradient(to bottom, #272727, #323232);
  padding: 3.2rem 2.4rem;
`;

const Disclaimer = styled.div`
  font-size: 1.1rem;
  line-height: 1.64;
  padding: 2.4rem;
  letter-spacing: 0.5px;
  opacity: 0.7;
  text-align: justify;
`;

const Focusable = styled.div.attrs({
  tabIndex: '-1'
})`
  &:focus {
    outline: none;
  }
`;

const ConfirmationWizard = props => {
  let initialState = {
    password: null,
    errors: {},
    status: 'init', // init | confirm | pending | success | failure
    error: null
  };

  let focusable = null;

  const [state, setState] = useState(initialState);

  const goToReview = ev => {
    ev.preventDefault();
    const isValid = !props.validate || props.validate();
    if (isValid) setState({ ...state, status: 'confirm', password: null });
  };

  const onCancelClick = () => setState(...state, initialState);

  const onConfirmClick = ev => {
    ev.preventDefault();
    validateConfirmation()
      .then(isValid => {
        if (isValid) {
          submitWizard();
          return;
        }
        setState({ ...state, errors: { password: 'Invalid password' } });
      })
      .catch(err => setState({ ...state, errors: { password: err.message } }));
  };

  const validateConfirmation = () => {
    const errors = validators.validatePassword(state.password);
    const hasErrors = Object.keys(errors).length > 0;
    if (hasErrors) {
      setState({ ...state, errors });
      return Promise.reject(new Error(errors.password));
    }
    return props.client.validatePassword(state.password);
  };

  const submitWizard = () => {
    setState({ ...state, status: 'pending' }, () =>
      focusable ? focusable.focus() : null
    );
    props
      .onWizardSubmit(state.password)
      .then(() => setState({ ...state, status: 'success' }))
      .then(() => (focusable ? focusable.focus() : null))
      .catch(err =>
        setState({ ...state, status: 'failure', error: err.message })
      );
  };

  const onPasswordChange = ({ value }) =>
    setState({ ...state, password: value, errors: {} });

  // eslint-disable-next-line complexity
  const { password, errors, status, error } = state;

  if (status === 'init') {
    return props.RenderForm(goToReview);
  } else if (status === 'confirm') {
    return (
      <form onSubmit={onConfirmClick} data-testid="confirm-form">
        <Sp py={4} px={3} style={props.styles.confirmation || {}}>
          {props.confirmationTitle && (
            <ConfirmationTitle>{props.confirmationTitle}</ConfirmationTitle>
          )}
          {props.renderConfirmation()}
          <Sp mt={2}>
            <TextInput
              data-testid="pass-field"
              autoFocus
              onChange={onPasswordChange}
              error={errors.password}
              value={password}
              label="Enter your password to confirm"
              type="password"
              id="password"
            />
          </Sp>
        </Sp>
        <BtnContainer style={props.styles.btns || {}}>
          <Btn submit block>
            Confirm
          </Btn>
          {!props.noCancel && (
            <EditBtn onClick={onCancelClick} data-testid="cancel-btn">
              {props.editLabel}
            </EditBtn>
          )}
        </BtnContainer>
        {props.disclaimer && <Disclaimer>{props.disclaimer}</Disclaimer>}
      </form>
    );
  } else if (status === 'success') {
    return (
      <Sp my={19} mx={12} data-testid="success">
        <Focusable ref={element => (focusable = element)}>
          <Flex.Column align="center">
            <CheckIcon color={theme.colors.success} />
            <Sp my={2}>
              <Title>{props.successTitle}</Title>
            </Sp>
            {props.successText && <Message>{props.successText}</Message>}
          </Flex.Column>
        </Focusable>
      </Sp>
    );
  } else if (status === 'failure') {
    return (
      <Sp my={19} mx={12} data-testid="failure">
        <Flex.Column align="center">
          <CloseIcon color={theme.colors.danger} size="4.8rem" />
          <Sp my={2}>
            <Title>{props.failureTitle}</Title>
          </Sp>
          {error && <FilteredMessage>{error}</FilteredMessage>}
          <TryAgainBtn
            data-testid="try-again-btn"
            onClick={onCancelClick}
            autoFocus
          >
            Try again
          </TryAgainBtn>
        </Flex.Column>
      </Sp>
    );
  } else {
    return (
      <Sp my={19} mx={12} data-testid="waiting">
        <Focusable ref={element => (focusable = element)}>
          <Flex.Column align="center">
            <Sp mb={2}>
              <Title>{props.pendingTitle}</Title>
            </Sp>
            <LoadingBar />
            {props.pendingText && (
              <Sp mt={2}>
                <Message>{props.pendingText}</Message>
              </Sp>
            )}
          </Flex.Column>
        </Focusable>
      </Sp>
    );
  }
};

ConfirmationWizard.defaultProps = {
  confirmationTitle: 'Transaction Preview',
  successTitle: 'Success!',
  successText:
    'You can view the status of this transaction in the transaction list.',
  failureTitle: 'Error',
  pendingTitle: 'Sending...',
  editLabel: 'Edit this transaction',
  styles: {}
};

ConfirmationWizard.propTypes = {
  renderConfirmation: PropTypes.func.isRequired,
  confirmationTitle: PropTypes.string,
  onWizardSubmit: PropTypes.func.isRequired,
  successTitle: PropTypes.string,
  failureTitle: PropTypes.string,
  pendingTitle: PropTypes.string,
  pendingText: PropTypes.string,
  successText: PropTypes.string,
  renderForm: PropTypes.func.isRequired,
  disclaimer: PropTypes.string,
  editLabel: PropTypes.string,
  noCancel: PropTypes.bool,
  validate: PropTypes.func,
  styles: PropTypes.object,
  client: PropTypes.shape({
    validatePassword: PropTypes.func.isRequired
  }).isRequired
};

export default withClient(ConfirmationWizard);
