import React from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import PropTypes from 'prop-types';

import ToastsContainer from './ToastsContainer';
import Toast from './Toast';
import Timer from './Timer';

export type ToastsContextType = {
  toast: (
    type: string,
    message: string,
    options?: { autoClose?: number },
  ) => void;
};

export const ToastsContext = React.createContext<ToastsContextType>({
  toast: () => {},
});

const defaults = {
  messagesPerToast: 1,
  autoClose: 6000,
};

type ToastStackItem = [string, ...string[]];

type ToastsProviderProps = React.PropsWithChildren<{
  messagesPerToast?: number;
  autoClose?: number;
}>;

type ToastsProviderState = {
  stack: ToastStackItem[];
};

export class ToastsProvider extends React.Component<
  ToastsProviderProps,
  ToastsProviderState
> {
  static propTypes = {
    messagesPerToast: PropTypes.number,
    autoClose: PropTypes.number,
    children: PropTypes.node.isRequired,
  };

  timers = {};

  addToast = (
    type: string,
    message: string,
    options: { autoClose?: number } = {},
  ) => {
    if (!type || !message) return;

    const autoClose =
      typeof options.autoClose === 'number'
        ? options.autoClose
        : typeof this.props.autoClose === 'number'
          ? this.props.autoClose
          : defaults.autoClose;

    const typeGroup = this.state.stack.find(([typeName]) => typeName === type);

    if (
      autoClose > 0 &&
      (!typeGroup || (this.timers[type] && this.timers[type].timerId))
    ) {
      this.clearTimeout(type);
      this.timers[type] = new Timer(() => this.removeToast(type), autoClose);
    }

    this.setState((state) => ({
      ...state,
      stack: typeGroup
        ? state.stack.map(([typeName, ...messages]) =>
            typeName === type
              ? [typeName, ...new Set([...messages, message])]
              : [typeName, ...messages],
          )
        : [...state.stack, [type, message]],
    }));
  };

  state: ToastsProviderState = {
    stack: [],
  };

  componentDidMount() {
    window.ipcRenderer.on('wallet-error', (_, { message }) =>
      this.addToast('error', message, { autoClose: 15000 }),
    );
  }

  removeToast = (type) => {
    this.setState((state) => ({
      ...state,
      stack: state.stack.filter(([typeName]) => typeName !== type),
    }));
  };

  clearTimeout = (type) => {
    if (this.timers[type]) this.timers[type].stop();
  };

  handleDismiss = (type) => this.removeToast(type);

  handleShowMore = (type) => this.clearTimeout(type);

  handleMouseEnter = (e) => {
    const type = e.currentTarget.dataset.type;
    if (this.timers[type] && this.timers[type].timerId) {
      this.timers[type].pause();
    }
  };

  handleMouseLeave = (e) => {
    const type = e.currentTarget.dataset.type;
    if (this.timers[type] && this.timers[type].timerId) {
      this.timers[type].resume();
    }
  };

  contextValue = { toast: this.addToast };

  render() {
    return (
      <ToastsContext.Provider value={this.contextValue}>
        {this.props.children}
        <ToastsContainer>
          <AnimatePresence initial={false}>
            {this.state.stack.map(([type, ...messages]) => (
              <motion.div
                key={type as string}
                data-type={type}
                onMouseEnter={this.handleMouseEnter}
                onMouseLeave={this.handleMouseLeave}
                initial={{ maxHeight: 0, opacity: 0, y: -45 }}
                animate={{ maxHeight: 450, opacity: 1, y: 0 }}
                exit={{ maxHeight: 0, opacity: 0, y: -45 }}
                transition={{
                  type: 'spring',
                  stiffness: 170,
                  damping: 15,
                }}
                style={{ overflow: 'hidden' }}
              >
                <Toast
                  messagesPerToast={
                    this.props.messagesPerToast || defaults.messagesPerToast
                  }
                  onShowMore={this.handleShowMore}
                  onDismiss={this.handleDismiss}
                  messages={messages}
                  type={type}
                />
              </motion.div>
            ))}
          </AnimatePresence>
        </ToastsContainer>
      </ToastsContext.Provider>
    );
  }
}
