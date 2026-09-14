import React from 'react';
import styled from 'styled-components';

// The app previously had no error boundary anywhere. Any throw during render —
// e.g. dereferencing a session/model that hadn't loaded yet — unmounted the
// whole React tree, leaving a blank window with no way back short of quitting.
// This catches it, shows what happened, and offers a recovery path.

const Wrapper = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 4rem 3rem;
  text-align: center;
  color: rgba(255, 255, 255, 0.9);
`;

const Title = styled.h2`
  margin: 0 0 1.2rem;
  font-size: 2rem;
  color: ${(p) => p.theme?.colors?.morMain || '#20dc8e'};
`;

const Message = styled.p`
  margin: 0 0 2.4rem;
  max-width: 52rem;
  font-size: 1.4rem;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.6);
`;

const Detail = styled.pre`
  max-width: 60rem;
  max-height: 18rem;
  overflow: auto;
  padding: 1.2rem;
  margin: 0 0 2.4rem;
  text-align: left;
  font-size: 1.1rem;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.55);
  white-space: pre-wrap;
  word-break: break-word;
`;

const Button = styled.button`
  padding: 1rem 2.4rem;
  border-radius: 999px;
  border: 1px solid rgba(32, 220, 142, 0.4);
  background: rgba(32, 220, 142, 0.12);
  color: #fff;
  font-size: 1.3rem;
  cursor: pointer;

  &:hover {
    background: rgba(32, 220, 142, 0.2);
  }
`;

type Props = React.PropsWithChildren<{
  /** Changing this value clears a captured error (e.g. on route change). */
  resetKey?: unknown;
  label?: string;
}>;

type State = {
  error: Error | null;
};

export class ErrorBoundary extends React.Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error(
      `[ErrorBoundary${this.props.label ? `: ${this.props.label}` : ''}]`,
      error,
      info?.componentStack,
    );
  }

  componentDidUpdate(prevProps: Props) {
    // Navigating to a different tab should clear a previous tab's crash so the
    // user isn't stuck on the error screen for the rest of the session.
    if (this.state.error && prevProps.resetKey !== this.props.resetKey) {
      this.setState({ error: null });
    }
  }

  handleReset = () => this.setState({ error: null });

  render() {
    const { error } = this.state;
    if (!error) {
      return this.props.children;
    }

    return (
      <Wrapper>
        <Title>Something went wrong on this screen</Title>
        <Message>
          The rest of the app is still running — your wallet and any open
          sessions are unaffected. You can retry this screen, or switch tabs and
          come back.
        </Message>
        <Detail>{error.message || String(error)}</Detail>
        <Button onClick={this.handleReset}>Retry this screen</Button>
      </Wrapper>
    );
  }
}

export default ErrorBoundary;
