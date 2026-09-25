import styled from 'styled-components';

const Feedback = styled.p`
  color: var(--text-primary);
  font-size: 1.4rem;
  line-height: 1.5;
  margin: 1.6rem 0;
  overflow-wrap: anywhere;
`;

export default function SetupFeedback({
  setupError = '',
  isSubmitting = false,
}) {
  if (setupError) return <Feedback role="alert">{setupError}</Feedback>;
  if (isSubmitting)
    return <Feedback role="status">Setting up your wallet…</Feedback>;
  return null;
}
