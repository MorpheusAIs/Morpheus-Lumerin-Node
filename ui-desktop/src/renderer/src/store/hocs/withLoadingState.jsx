import React from 'react';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import every from 'lodash/every';

import selectors from '../selectors';

// Time to wait before updating checklist status (in ms)
// The idea is to prevent fast-loading checklists which would look like a glitch
const MIN_CADENCE = 200;

// Time to wait before exiting the loading screen (in ms)
const ON_COMPLETE_DELAY = 20;

const withLoadingState = WrappedComponent => {
  class Container extends React.Component {
    // static propTypes = {
    //   chainStatus: PropTypes.objectOf(
    //     PropTypes.shape({
    //       hasBlockHeight: PropTypes.bool.isRequired,
    //       hasCoinBalance: PropTypes.bool.isRequired,
    //       hasLmrBalance: PropTypes.bool.isRequired,
    //       displayName: PropTypes.string.isRequired,
    //       hasCoinRate: PropTypes.bool.isRequired,
    //       symbol: PropTypes.string.isRequired
    //     })
    //   ).isRequired,
    //   onComplete: PropTypes.func.isRequired
    // }

    static displayName = `withLoadingState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    state = {
      ...this.props.chainStatus,
      hasBlockHeight: false,
      hasCoinBalance: false,
      hasLmrBalance: false,
      hasCoinRate: false
    };

    checkFinished = () => {
      if (every(this.state, every)) {
        clearInterval(this.interval);
        setTimeout(this.props.onComplete, ON_COMPLETE_DELAY);
      }
    };

    checkTasks = () => {
      const { chainStatus } = this.props;
      const prevStatus = this.state || {};
      if (chainStatus.hasBlockHeight && !prevStatus.hasBlockHeight) {
        this.setState(
          state => ({
            ...state,
            hasBlockHeight: true
          }),
          this.checkFinished
        );
        return;
      }
      if (chainStatus.hasCoinRate && !prevStatus.hasCoinRate) {
        this.setState(
          state => ({
            ...state,
            hasCoinRate: true
          }),
          this.checkFinished,
          this.checkFinished
        );
        return;
      }
      if (chainStatus.hasCoinBalance && !prevStatus.hasCoinBalance) {
        this.setState(
          state => ({
            ...state,
            hasCoinBalance: true
          }),
          this.checkFinished,
          this.checkFinished
        );
        return;
      }
      if (chainStatus.hasLmrBalance && !prevStatus.hasLmrBalance) {
        this.setState(
          state => ({
            ...state,
            hasLmrBalance: true
          }),
          this.checkFinished,
          this.checkFinished
        );
      }
    };

    componentDidMount() {
      this.interval = setInterval(this.checkTasks, MIN_CADENCE);
    }

    componentWillUnmount() {
      if (this.interval) clearInterval(this.interval);
    }

    render() {
      return <WrappedComponent chainStatus={this.state} />;
    }
  }

  const mapStateToProps = state => ({
    chainStatus: selectors.getChainReadyStatus(state)
  });

  const mapDispatchToProps = dispatch => ({
    onComplete: () => dispatch({ type: 'required-data-gathered' })
  });

  return connect(mapStateToProps, mapDispatchToProps)(Container);
};

export default withLoadingState;
