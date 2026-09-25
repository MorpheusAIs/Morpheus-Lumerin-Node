import styled from 'styled-components';
import Card from 'react-bootstrap/Card';

/**
 * The card surface shared by the model registry and the pinned files list.
 *
 * Both screens had their own copy of this block, each with its own hardcoded
 * greens and its own stack of !important. One definition on the app tokens
 * means the two lists cannot drift apart again, and Bootstrap 5.3 reads its
 * own variables off the component, so nothing here has to shout at the
 * framework stylesheet to be applied.
 */
export const ModelCardSurface = styled(Card)`
  --bs-card-bg: var(--surface-raised);
  --bs-card-color: var(--text-primary);
  --bs-card-border-color: var(--border-subtle);
  --bs-card-border-radius: 12px;
  --bs-card-title-color: var(--accent);
  --bs-card-subtitle-color: var(--text-muted);

  position: relative;
  width: 36rem;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition:
    transform 150ms var(--ease-out),
    border-color 150ms var(--ease-out),
    box-shadow 150ms var(--ease-out);

  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.25);
    --bs-card-border-color: var(--border-strong);
  }

  p {
    color: var(--text-primary);
  }

  .card-body {
    padding: 1.5rem;
  }

  .card-title {
    margin-bottom: 5px;
    font-weight: 600;
    font-size: 1.3rem;
    letter-spacing: 0.02em;
  }

  .card-subtitle {
    font-size: 0.85rem;
    margin-bottom: 16px;
  }

  .model-card-title {
    align-items: center;
    display: flex;
    justify-content: space-between;
  }

  .model-card-name {
    max-width: 90%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .model-info-section {
    display: flex;
    flex-direction: column;
    gap: 3px;
    padding-top: 8px;
    border-top: 1px solid var(--border-subtle);
  }

  .model-info-item {
    display: flex;
    align-items: center;
    font-size: 1.1rem;
    padding: 4px 0;
  }

  .info-label {
    font-weight: 600;
    min-width: 90px;
    color: var(--text-muted);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .info-value {
    color: var(--text-primary);
    display: flex;
    align-items: center;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tag-container {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }

  .tag-item {
    background: var(--surface-hover);
    color: var(--text-primary);
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 1rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 22px;
    line-height: 1;
    border: 1px solid var(--border-subtle);
    transition:
      border-color 150ms var(--ease-out),
      transform 150ms var(--ease-out);

    &:hover {
      border-color: var(--border-strong);
      transform: translateY(-2px);
    }
  }

  .monospace {
    font-family: var(--font-mono);
    font-size: 0.85rem;
    letter-spacing: -0.03em;
  }

  .hash-container {
    background: var(--surface-base);
    border: 1px solid var(--border-subtle);
    border-radius: 6px;
    padding: 6px 10px;
    display: flex;
    align-items: center;
    font-size: 1.1rem;
  }
`;

/** The wrapping grid both card lists sit in. */
export const ModelCardGrid = styled.div`
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  gap: 28px;
  max-height: 75vh;
  padding: 8px 4px;
  overflow-y: auto;
`;

/** Shown in place of the grid contents when a list has nothing in it. */
export const ModelCardEmptyState = styled.div`
  align-items: center;
  color: var(--text-muted);
  display: flex;
  font-size: 1.4rem;
  justify-content: center;
  padding: 3.2rem 0;
  width: 100%;
`;
