import { validatePassword } from '../validators';
import PropTypes from 'prop-types';
import { withClient } from '../hocs/clientContext';
import React from 'react';

const withLoginState = WrappedComponent => {
  class Container extends React.Component {
    static propTypes = {
      onLoginSubmit: PropTypes.func.isRequired
    };

    static displayName = `withLoginState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      password: null,
      status: 'init',
      errors: {},
      error: null
    };

    onSubmit = e => {
      if (e && e.preventDefault) e.preventDefault();

      const errors = this.validate();
      if (Object.keys(errors).length > 0) return this.setState({ errors });

      this.setState({ status: 'pending', error: null });
      return this.props
        .onLoginSubmit({ password: this.state.password })
        .catch(err =>
          this.setState({
            password: null,
            status: 'failure',
            error: err.message || 'Unknown error'
          })
        );
    };

    logout = () => {
      return this.props.client.logout();
    };

    onInputChange = ({ id, value }) => {
      this.setState(state => ({
        ...state,
        [id]: value,
        errors: {
          ...state.errors,
          [id]: null
        },
        error: null
      }));
    };

    validate = () => {
      const { password } = this.state;
      return { ...validatePassword(password) };
    };

    render() {
      return (
        <WrappedComponent
          onInputChange={this.onInputChange}
          onSubmit={this.onSubmit}
          logout={this.logout}
          {...this.state}
        />
      );
    }
  }

  return withClient(Container);
};

export default withLoginState;
