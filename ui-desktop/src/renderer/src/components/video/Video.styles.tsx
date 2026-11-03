import styled, { keyframes } from 'styled-components';
import { BaseBtn } from '../common/Btn';

export const Layout = styled.div`
  display: grid;
  grid-template-columns: minmax(32rem, 52rem) minmax(0, 1fr);
  gap: 2.4rem;
  align-items: start;

  @container route-space (max-width: 960px) {
    grid-template-columns: minmax(0, 1fr);
  }
`;

export const Card = styled.section`
  display: flex;
  flex-direction: column;
  gap: 1.6rem;
  min-width: 0;
  padding: 2rem;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--surface-raised);
`;

export const CardTitle = styled.h2`
  margin: 0;
  font-size: 1.6rem;
  font-weight: 650;
  color: var(--text-primary);
`;

export const Field = styled.label`
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  margin: 0;
  min-width: 0;
  color: var(--text-muted);
  font-size: 1.3rem;

  select,
  textarea {
    width: 100%;
    padding: 1rem 1.2rem;
    color: var(--text-primary);
    background: var(--surface-base);
    border: 1px solid var(--border-strong);
    border-radius: 8px;
    font: inherit;
    font-size: 1.4rem;
  }

  select {
    cursor: pointer;
  }

  textarea {
    min-height: 11rem;
    resize: vertical;
    line-height: 1.5;
  }

  select:focus-visible,
  textarea:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
`;

export const Fieldset = styled.fieldset`
  display: flex;
  flex-direction: column;
  gap: 1.6rem;
  min-width: 0;
  margin: 0;
  padding: 0;
  border: 0;
`;

export const FieldRow = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
  gap: 1.2rem;
`;

export const Hint = styled.p<{ $tone?: 'warning' | 'error' }>`
  margin: 0;
  font-size: 1.3rem;
  line-height: 1.5;
  color: ${(p) =>
    p.$tone === 'warning'
      ? '#f2bd6d'
      : p.$tone === 'error'
        ? '#ff8f8f'
        : 'var(--text-muted)'};
`;

export const Checkbox = styled.label`
  display: flex;
  align-items: center;
  gap: 0.8rem;
  margin: 0;
  color: var(--text-primary);
  font-size: 1.4rem;
  cursor: pointer;

  input {
    width: 1.6rem;
    height: 1.6rem;
    accent-color: var(--accent);
  }
`;

export const SecondaryBtn = styled(BaseBtn)`
  min-height: 3.6rem;
  padding: 0.6rem 1.4rem;
  border: 1px solid var(--border-strong);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 1.3rem;
  font-weight: 600;

  &:not([disabled]):hover,
  &:not([disabled]):focus-visible {
    background: var(--surface-hover);
  }
`;

export const FramePreview = styled.div`
  display: flex;
  align-items: center;
  gap: 1.2rem;

  img {
    width: 9.6rem;
    height: 5.4rem;
    object-fit: cover;
    border-radius: 6px;
    border: 1px solid var(--border-subtle);
  }

  span {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text-primary);
    font-size: 1.3rem;
  }
`;

export const QuoteList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
`;

export const QuoteRow = styled.label<{ $disabled?: boolean }>`
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 1rem;
  margin: 0;
  padding: 1rem 1.2rem;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--surface-base);
  color: var(--text-primary);
  font-size: 1.3rem;
  cursor: ${(p) => (p.$disabled ? 'default' : 'pointer')};
  opacity: ${(p) => (p.$disabled ? 0.6 : 1)};

  &:has(input:checked) {
    border-color: var(--accent);
  }

  input {
    accent-color: var(--accent);
  }

  small {
    display: block;
    color: var(--text-muted);
    font-size: 1.2rem;
  }

  strong {
    font-weight: 650;
    white-space: nowrap;
  }
`;

export const Mono = styled.span`
  font-family: var(--font-mono);
`;

export const SessionBox = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1.2rem;
  padding: 1.2rem 1.4rem;
  border: 1px solid var(--border-strong);
  border-radius: 10px;
  background: var(--surface-base);
  font-size: 1.3rem;

  div {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    min-width: 0;
  }

  small {
    color: var(--text-muted);
    font-size: 1.2rem;
  }
`;

export const JobList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
`;

export const JobCard = styled.article`
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.6rem;
  border: 1px solid var(--border-subtle);
  border-radius: 10px;
  background: var(--surface-base);
`;

export const JobHeader = styled.div`
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1.2rem;

  p {
    margin: 0;
    color: var(--text-primary);
    font-size: 1.4rem;
    line-height: 1.5;
    overflow-wrap: anywhere;
  }
`;

export const JobMeta = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 1.2rem;
  color: var(--text-muted);
  font-size: 1.2rem;
`;

const spin = keyframes`
  to {
    transform: rotate(360deg);
  }
`;

/** Turns while a clip is queued or rendering, so a long render reads as alive. */
export const Spinner = styled.span`
  width: 1rem;
  height: 1rem;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: ${spin} 0.9s linear infinite;

  @media (prefers-reduced-motion: reduce) {
    animation-duration: 3s;
  }
`;

export const StatusBadge = styled.span<{ $status: string }>`
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  flex-shrink: 0;
  font-variant-numeric: tabular-nums;
  padding: 0.2rem 0.8rem;
  border-radius: 999px;
  font-size: 1.2rem;
  font-weight: 650;
  color: ${(p) =>
    p.$status === 'succeeded'
      ? '#032117'
      : p.$status === 'failed' ||
          p.$status === 'cancelled' ||
          p.$status === 'expired'
        ? '#ffd2d2'
        : 'var(--text-primary)'};
  background: ${(p) =>
    p.$status === 'succeeded'
      ? 'var(--accent)'
      : p.$status === 'failed' ||
          p.$status === 'cancelled' ||
          p.$status === 'expired'
        ? 'rgba(255, 107, 107, 0.22)'
        : 'var(--surface-hover)'};
`;

export const Player = styled.video`
  width: 100%;
  max-height: 48rem;
  border-radius: 8px;
  background: #000;
`;

export const Actions = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.8rem;
`;

export const EmptyState = styled.div`
  padding: 2.4rem 0;
  color: var(--text-muted);
  font-size: 1.4rem;
  line-height: 1.5;
  text-align: center;
`;
