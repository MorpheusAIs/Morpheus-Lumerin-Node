import PropTypes from 'prop-types';
import moment from 'moment';
import React from 'react';
import 'moment-precise-range-plugin';

export default class CountDownProvider extends React.Component {
  static propTypes = {
    targetTimestamp: PropTypes.number.isRequired,
    children: PropTypes.func.isRequired
  };

  constructor(props) {
    super(props);
    this.state = this.getCountdownValues(props.targetTimestamp);
  }

  componentDidMount() {
    this.intervalId = window.setInterval(this.updateCountdown, 1000);
  }

  componentWillUnmount() {
    if (this.intervalId) window.clearInterval(this.intervalId);
  }

  updateCountdown = () => {
    this.setState(this.getCountdownValues(this.props.targetTimestamp));
  };

  getCountdownValues = targetTime =>
    moment.preciseDiff(moment.unix(targetTime), moment(), true);

  render() {
    const { firstDateWasLater: inFuture, ...other } = this.state;

    return this.props.children({ inFuture, ...other });
  }
}
