import styled from 'styled-components';

// Shown when a data query fails.
//
// Before this existed, the main-process handlers swallowed every fetch error
// and returned `[]` / `null`, so a dead proxy-router looked exactly like an
// empty wallet: zero balance, no models, no transactions, no explanation.
// Handlers now throw (see subscriptions/handlers.ts), and this is what the user
// sees instead of a convincing but wrong zero.

const Banner = styled.div`
  display: flex;
  align-items: flex-start;
  gap: 1.2rem;
  padding: 1.4rem 1.6rem;
  margin-bottom: 1.6rem;
  border-radius: 12px;
  background: rgba(255, 107, 107, 0.08);
  border: 1px solid rgba(255, 107, 107, 0.3);
  color: rgba(255, 255, 255, 0.9);
`;

const Icon = styled.div`
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.3rem;
  font-weight: 700;
  background: rgba(255, 107, 107, 0.9);
  color: #1a0000;
`;

const Body = styled.div`
  flex: 1;
  min-width: 0;
`;

const Title = styled.div`
  font-size: 1.4rem;
  font-weight: 600;
  margin-bottom: 0.3rem;
`;

const Detail = styled.div`
  font-size: 1.25rem;
  line-height: 1.5;
  color: rgba(255, 255, 255, 0.65);
  word-break: break-word;
`;

const RetryBtn = styled.button`
  flex-shrink: 0;
  align-self: center;
  padding: 0.7rem 1.6rem;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: transparent;
  color: #fff;
  font-size: 1.2rem;
  cursor: pointer;

  &:hover {
    background: rgba(255, 255, 255, 0.08);
  }
`;

type Props = {
  /** The thrown error, if any. Renders nothing when null/undefined. */
  error?: unknown;
  /** What failed to load, e.g. "balances". Used in the headline. */
  what?: string;
  /** Wire to react-query's `refetch`. */
  onRetry?: () => void;
};

/**
 * Errors originating from ProxyRouterError carry a message that is already
 * written for a user. Anything else gets a generic headline with the raw
 * message underneath so it is still diagnosable.
 */
export function QueryError({ error, what = 'data', onRetry }: Props) {
  if (!error) {
    return null;
  }

  const message =
    (error as any)?.message ??
    (typeof error === 'string' ? error : 'Unknown error');

  // Main-process errors cross the IPC boundary as plain objects, so instanceof
  // is unavailable — match on the wording the handler produced instead.
  const looksUnreachable = /cannot reach|cannot authenticate/i.test(
    String(message),
  );

  return (
    <Banner role="alert" data-testid="query-error">
      <Icon aria-hidden>!</Icon>
      <Body>
        <Title>
          {looksUnreachable
            ? 'Not connected to your node'
            : `Couldn't load ${what}`}
        </Title>
        <Detail>
          {message}
          {looksUnreachable && (
            <>
              {' '}
              Your funds are unaffected — this is a connection problem, not a
              balance of zero.
            </>
          )}
        </Detail>
      </Body>
      {onRetry && <RetryBtn onClick={onRetry}>Retry</RetryBtn>}
    </Banner>
  );
}

export default QueryError;
