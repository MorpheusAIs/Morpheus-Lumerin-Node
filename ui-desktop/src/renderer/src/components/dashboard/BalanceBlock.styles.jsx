import styled from 'styled-components';
import { BaseBtn } from '../common';

/**
 * The app button pair. Chat, Models and Agents all pull BtnAccent from here, so
 * these two are the closest thing the app has to a primary and secondary
 * button, and everything in them now comes from the interface tokens.
 *
 * The container and balance styles that used to sit alongside them went with
 * the balance block itself. They were still exported, still carried their own
 * hardcoded whites, and nothing had rendered them for some time.
 */
export const Btn = styled(BaseBtn)`
  font-size: 1.4rem;
  min-height: 4.4rem;
  margin-left: 0;
  padding: 1rem 1.6rem;
  border-radius: 10px;
  border: 1px solid var(--border-strong);
  background-color: var(--surface-raised);
  color: var(--text-primary);

  &:hover:not(:disabled) {
    background-color: var(--surface-hover);
  }

  /* BaseBtn styles a disabled button through a data attribute as well as the
     real one, because some call sites keep the button focusable. */
  &[data-disabled='true'],
  &[disabled] {
    background: transparent;
    border-color: var(--border-subtle);
    color: var(--text-muted);
  }
`;

export const BtnAccent = styled(Btn)`
  background-color: var(--accent);
  border-color: transparent;
  /* Near black rather than the app text colour: this is the one surface in the
     app light enough to need dark type on it. */
  color: #032117;
  font-weight: 600;

  &:hover:not(:disabled) {
    background-color: #48e4ae;
  }
`;
