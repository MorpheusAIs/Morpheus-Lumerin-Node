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

  // Toasts are grouped by type, so these are keyed by type too. `hovered`
  // means "the cursor is on it right now"; `pinned` means "the user expanded
  // it and asked to keep reading", which is the only reason a group should
  // ever stop counting down on its own.
  hovered = {};

  pinned = {};

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

    // A group that left the screen takes its pin with it.
    if (!typeGroup) this.pinned[type] = false;

    // Re-arm on every message unless the user pinned this group. The previous
    // guard also required a live `timerId`, which is exactly what `pause()`
    // clears — so once the cursor had touched an error toast, no later error
    // could ever re-arm the timer, and `handleMouseLeave` checked the same
    // null `timerId` and never resumed either. The error group then sat there
    // permanently, which is the "errors just don't disappear" report.
    if (autoClose > 0 && !this.pinned[type]) {
      this.clearTimeout(type);
      const timer = new Timer(() => this.removeToast(type), autoClose);
      // Don't start counting down underneath the cursor.
      if (this.hovered[type]) timer.pause();
      this.timers[type] = timer;
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
    window.ipcRenderer.on('wallet-error', ({ message }) =>
      this.addToast('error', message, { autoClose: 15000 }),
    );
  }

  removeToast = (type) => {
    this.clearTimeout(type);
    this.hovered[type] = false;
    this.pinned[type] = false;
    this.setState((state) => ({
      ...state,
      stack: state.stack.filter(([typeName]) => typeName !== type),
    }));
  };

  clearTimeout = (type) => {
    if (this.timers[type]) this.timers[type].stop();
  };

  handleDismiss = (type) => this.removeToast(type);

  handleShowMore = (type) => {
    this.pinned[type] = true;
    this.clearTimeout(type);
  };

  handleMouseEnter = (e) => {
    const type = e.currentTarget.dataset.type;
    this.hovered[type] = true;
    if (this.timers[type]) this.timers[type].pause();
  };

  handleMouseLeave = (e) => {
    const type = e.currentTarget.dataset.type;
    this.hovered[type] = false;
    if (this.timers[type] && !this.pinned[type]) this.timers[type].resume();
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
                initial={{ maxHeight: 0, opacity: 0, y: -8 }}
                animate={{ maxHeight: 450, opacity: 1, y: 0 }}
                exit={{ maxHeight: 0, opacity: 0, y: -4 }}
                transition={{
                  duration: 0.18,
                  ease: 'easeOut',
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
