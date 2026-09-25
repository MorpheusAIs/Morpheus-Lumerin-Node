import styled from 'styled-components';

export const ModelActionButton = styled.button`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 3.6rem;
  height: 3.6rem;
  padding: 0.8rem;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-muted, #9ab4a7);
  cursor: pointer;

  &:hover:not(:disabled) {
    background: var(--surface-hover, #18372a);
    color: var(--accent, #19d695);
  }

  &:focus-visible {
    outline: 2px solid var(--accent, #19d695);
    outline-offset: 2px;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;
