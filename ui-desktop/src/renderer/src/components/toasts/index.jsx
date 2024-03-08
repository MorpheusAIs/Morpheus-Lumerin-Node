import React from 'react';
import { TransitionMotion, spring } from 'react-motion';
import PropTypes from 'prop-types';

import ToastsContainer from './ToastsContainer';
import Toast from './Toast';
import Timer from './Timer';

export const ToastsContext = React.createContext({});

const defaults = {
  messagesPerToast: 1,
  autoClose: 6000
};

export class ToastsProvider extends React.Component {
  static propTypes = {
    messagesPerToast: PropTypes.number,
    autoClose: PropTypes.number,
    children: PropTypes.node.isRequired
  };

  timers = {};

  addToast = (type, message, options = {}) => {
    if (!type || !message) return;

    const autoClose =
      typeof options.autoClose === 'number'
        ? options.autoClose
        : typeof this.props.autoClose === 'number'
        ? this.props.autoClose
        : defaults.autoClose;

    // check if requested type is already visible
    const typeGroup = this.state.stack.find(([typeName]) => typeName === type);

    // only set timer if first message in toast or if toast is not fixed
    if (
      autoClose > 0 &&
      (!typeGroup || (this.timers[type] && this.timers[type].timerId))
    ) {
      this.clearTimeout(type);
      this.timers[type] = new Timer(() => this.removeToast(type), autoClose);
    }

    this.setState(state => ({
      ...state,
      stack: typeGroup
        ? // if type group exists, append message to it
          state.stack.map(([typeName, ...messages]) =>
            typeName === type
              ? // append using Set to automatically avoid duplicates
                [typeName, ...new Set([...messages, message])]
              : [typeName, ...messages]
          )
        : // if not, append a new type group with the new message
          [...state.stack, [type, message]]
    }));
  };

  state = {
    stack: []
  };

  componentDidMount() {
    window.ipcRenderer.on('wallet-error', (_, { message }) =>
      this.addToast('error', message, { autoClose: 15000 })
    );
  }

  removeToast = type => {
    this.setState(state => ({
      ...state,
      stack: state.stack.filter(([typeName]) => typeName !== type)
    }));
  };

  clearTimeout = type => {
    if (this.timers[type]) this.timers[type].stop();
  };

  handleDismiss = type => this.removeToast(type);

  handleShowMore = type => this.clearTimeout(type);

  handleMouseEnter = e => {
    const type = e.currentTarget.dataset.type;
    if (this.timers[type] && this.timers[type].timerId) {
      this.timers[type].pause();
    }
  };

  handleMouseLeave = e => {
    const type = e.currentTarget.dataset.type;
    if (this.timers[type] && this.timers[type].timerId) {
      this.timers[type].resume();
    }
  };

  willEnter = () => ({ maxHeight: 0, opacity: 0, translate: -45 });

  willLeave = () => ({
    translate: spring(-45),
    maxHeight: spring(0),
    opacity: spring(0)
  });

  contextValue = { toast: this.addToast };

  render() {
    return (
      <ToastsContext.Provider value={this.contextValue}>
        {this.props.children}
        <TransitionMotion
          willLeave={this.willLeave}
          willEnter={this.willEnter}
          styles={this.state.stack.map(([type, ...messages]) => ({
            style: {
              maxHeight: spring(450, { stiffness: 150, damping: 20 }),
              translate: spring(0, { stiffness: 170, damping: 15 }),
              opacity: spring(1)
            },
            data: messages,
            key: type
          }))}
        >
          {interpolatedStyles => (
            <ToastsContainer>
              {interpolatedStyles.map(
                ({ key, data, style: { translate, ...other } }) => (
                  <div
                    onMouseEnter={this.handleMouseEnter}
                    onMouseLeave={this.handleMouseLeave}
                    data-type={key}
                    style={{
                      ...other,
                      transform: `translateY(${translate || 0}px)`
                    }}
                    key={`toast-${key}`}
                  >
                    <Toast
                      messagesPerToast={
                        this.props.messagesPerToast || defaults.messagesPerToast
                      }
                      onShowMore={this.handleShowMore}
                      onDismiss={this.handleDismiss}
                      messages={data}
                      type={key}
                    />
                  </div>
                )
              )}
            </ToastsContainer>
          )}
        </TransitionMotion>
      </ToastsContext.Provider>
    );
  }
}
