import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';

import { withClient } from './clientContext';
import selectors from '../selectors';
import { IsPasswordStrong } from '../../lib/PasswordStrength';

const withChangePasswordState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      client: PropTypes.shape({
        changePassword: PropTypes.func.isRequired
      }).isRequired
    };

    static displayName = `withChangePasswordState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      newPasswordAgain: null,
      newPassword: null,
      oldPassword: null,
      status: 'init',
      errors: {},
      error: null
    };

    onInputChange = ({ id, value }) => {
      this.setState(state => ({
        ...state,
        [id]: value,
        errors: {
          ...state.errors,
          [id]: null
        }
      }));
    };

    // eslint-disable-next-line complexity
    validate = clearOnError => {
      const { oldPassword, newPassword, newPasswordAgain } = this.state;
      const { config, client } = this.props;

      const errors = {};

      if (!oldPassword) {
        errors.oldPassword = 'Current password is required';
      } else if (!newPassword) {
        errors.newPassword = 'New password is required';
      }
      // else if (!IsPasswordStrong(newPassword)) {
      //   errors.password = 'Password is not strong enough';
      // }
      else if (!errors.password && !newPasswordAgain) {
        errors.newPasswordAgain = `Repeat the ${
          clearOnError ? 'PIN' : 'password'
        }`;
      } else if (!errors.password && newPasswordAgain !== newPassword) {
        errors.newPasswordAgain = `${
          clearOnError ? 'PINs' : 'Passwords'
        } don't match`;
      }

      const hasErrors = Object.keys(errors).length > 0;
      if (hasErrors) {
        this.setState({
          newPasswordAgain: clearOnError ? '' : newPasswordAgain,
          status: 'failure',
          errors
        });
      }
      return !hasErrors;
    };

    onSubmit = (clearOnError = false) => {
      if (!this.validate(clearOnError)) return;
      this.setState({ status: 'pending', error: null, errors: {} });
      this.props.client
        .changePassword({
          oldPassword: this.state.oldPassword,
          newPassword: this.state.newPassword
        })
        .then(isValid => {
          this.setState({
            status: isValid ? 'success' : 'failure',
            errors: isValid ? {} : { oldPassword: 'Invalid password' }
          });
        })
        .catch(err => {
          this.setState({ status: 'failure', error: err.message });
        });
    };

    render() {
      return (
        <WrappedComponent
          onInputChange={this.onInputChange}
          onSubmit={this.onSubmit}
          validate={this.validate}
          {...this.state}
          {...this.props}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    config: selectors.getConfig(state)
  });

  return connect(mapStateToProps)(withClient(Container));
};

export default withChangePasswordState;
