import styled from 'styled-components';
import { Btn } from '../common/Btn';

export const Container = styled.div`
  display: flex;
  flex-direction: column;
  height: 100%;

  .dropdown-toggle::after {
    display: none !important;
  }

  .history-scroll-block {
    display: flex;
    flex-direction: column;
    flex: 1 1 auto;
    min-height: 0;
    overflow: hidden;
  }

  /* react-bootstrap Tabs root */
  #history-tabs {
    flex-shrink: 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);

    .nav-link {
      color: rgba(255, 255, 255, 0.5);
      border: none;
      padding: 0.8rem 1.2rem;
      font-size: 1.3rem;
      font-weight: 500;
      letter-spacing: 0.4px;

      &:hover {
        color: rgba(255, 255, 255, 0.85);
        border: none;
      }
    }

    .nav-link.active {
      color: ${(p) => p.theme.colors.morMain};
      background: transparent;
      border: none;
      border-bottom: 2px solid ${(p) => p.theme.colors.morMain};
    }
  }

  /* The Tabs body — give it the full remaining height so each tab can scroll. */
  .tab-content {
    flex: 1 1 auto;
    min-height: 0;
    overflow: hidden;
  }
  .tab-pane.active {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .history-block,
  .list-container {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    padding: 0 0.4rem 1rem;

    /* Slim scrollbar that doesn't fight the dark theme. */
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.12) transparent;
    &::-webkit-scrollbar {
      width: 6px;
    }
    &::-webkit-scrollbar-thumb {
      background: rgba(255, 255, 255, 0.12);
      border-radius: 3px;
    }
  }
`;

/* Subtle floating label between groups — matches the Claude / ChatGPT
   sidebar idiom. No sticky, no backdrop; just dimmed text with breathing
   room. The first section also gets less top margin so it doesn't feel
   detached from the search bar above. */
export const SectionHeader = styled.div`
  color: rgba(255, 255, 255, 0.38);
  font-size: 1.05rem;
  font-weight: 500;
  letter-spacing: 0.4px;
  padding: 1.2rem 1rem 0.4rem;

  &:first-of-type {
    padding-top: 0.6rem;
  }
`;

export const EmptyState = styled.div`
  color: rgba(255, 255, 255, 0.45);
  text-align: center;
  padding: 4rem 2rem;
  font-size: 1.3rem;
  line-height: 1.5;
`;

export const SearchWrapper = styled.div`
  padding: 1rem 0.4rem 0.6rem;
  flex-shrink: 0;

  .input-group {
    background: rgba(255, 255, 255, 0.05);
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid transparent;
    transition: border-color 0.15s ease;

    &:focus-within {
      border-color: ${(p) => p.theme.colors.morMain};
    }
  }

  .input-group-text {
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.5);
    padding-right: 0;
  }

  input.form-control {
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.9);
    box-shadow: none;
    font-size: 1.3rem;
    padding: 0.6rem 0.8rem;

    &::placeholder {
      color: rgba(255, 255, 255, 0.7);
      opacity: 1; /* Firefox dims placeholders by default; reset. */
    }

    &:focus {
      background: transparent;
      color: rgba(255, 255, 255, 0.95);
      box-shadow: none;
    }
  }
`;

export const Title = styled.div`
  text-align: center;
  margin-bottom: 2.4rem;

  span {
    cursor: pointer;
  }
`;

export const HistoryItem = styled.div`
  color: ${(p) => p.theme.colors.morMain};
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 5px 0 0 0;
`;

export const HistoryEntryContainer = styled.div`
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 8px;
  color: white;
  margin: 0 0.4rem 1rem;
  cursor: pointer;
  padding: 1rem 1.2rem;
  transition: background 0.12s ease, border-color 0.12s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.12);
  }
`;

export const FlexSpaceBetween = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

/* The actual clickable chat-history row.
 * Layout:  [ title (flex:1, truncated) ]  [ icons (auto, hover-revealed) ]
 * Critical bits for the click bug:
 *   - `.title` has `flex: 1 1 auto` and `min-width: 0` so the long text is
 *     forced to shrink + ellipsis instead of overflowing the row.
 *   - The row itself has `overflow: hidden` so nothing leaks past the visible
 *     bounds (which previously made clicks land outside the drawer's
 *     scroll area for long titles).
 *   - Icons are absent from layout flow until hover / active, kept in a fixed
 *     `.icons` slot to avoid layout jitter on hover.
 */
export const HistoryEntryTitle = styled.div`
  display: flex;
  align-items: center;
  gap: 0.6rem;
  color: rgba(255, 255, 255, 0.78);
  margin: 0;
  padding: 0.7rem 1rem;
  border-radius: 8px;
  cursor: pointer;
  overflow: hidden;
  transition: background 0.12s ease, color 0.12s ease;
  position: relative;
  font-size: 1.3rem;
  line-height: 1.35;

  .title {
    flex: 1 1 auto;
    min-width: 0;
    text-overflow: ellipsis;
    overflow: hidden;
    white-space: nowrap;
  }

  .icons {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 0.6rem;
    opacity: 0;
    transition: opacity 0.12s ease;
  }

  &:hover {
    background: rgba(255, 255, 255, 0.06);
    color: rgba(255, 255, 255, 0.95);
    .icons {
      opacity: 1;
    }
  }

  &[data-active='true'] {
    background: rgba(33, 220, 143, 0.1);
    color: ${(p) => p.theme.colors.morMain};
    .icons {
      opacity: 1;
    }
  }
`;

export const ModelName = styled.div`
  text-overflow: ellipsis;
  width: 220px;
  height: 24px;
  overflow: hidden;
  text-wrap: nowrap;
`;

export const CloseBtn = styled(Btn)`
  font-size: 1.4rem;
  padding: 0 1rem;
`;

export const Duration = styled.div`
  color: white;
`;

export const IconsContainer = styled.div`
  display: flex;
  align-items: center;
  gap: 0.6rem;

  svg {
    cursor: pointer;
    opacity: 0.65;
    transition: opacity 0.12s ease, color 0.12s ease;

    &:hover {
      opacity: 1;
    }
  }
`;

export const IconButton = styled.button`
  background: transparent;
  border: none;
  padding: 0.3rem;
  margin: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: inherit;
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.12s ease, color 0.12s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 1);
  }

  &.danger:hover {
    color: #ff7c7c;
  }

  &:focus {
    outline: none;
  }
`;

export const ChangeTitleContainer = styled.div`
  display: flex;
  align-items: center;
  width: 100%;
  gap: 0.6rem;

  /* Override Bootstrap's .form-control + .input-group defaults so the
     rename field blends into the dark drawer instead of flashing white. */
  .input-group {
    background: transparent;
    flex: 1 1 auto;
    min-width: 0;
  }

  .input-group .form-control,
  input.form-control,
  input {
    background-color: transparent !important;
    color: rgba(255, 255, 255, 0.95);
    border: none;
    border-bottom: 1px solid ${(p) => p.theme.colors.morMain}40;
    border-radius: 0;
    box-shadow: none;
    padding: 0.4rem 0.2rem;
    font-size: 1.3rem;
    line-height: 1.35;
  }

  .input-group .form-control:focus,
  input.form-control:focus,
  input:focus {
    background-color: transparent !important;
    outline: none;
    box-shadow: none;
    color: rgba(255, 255, 255, 1);
    border-bottom: 1px solid ${(p) => p.theme.colors.morMain};
  }

  .input-group .form-control::placeholder,
  input.form-control::placeholder,
  input::placeholder {
    color: rgba(255, 255, 255, 0.35);
  }
`;
