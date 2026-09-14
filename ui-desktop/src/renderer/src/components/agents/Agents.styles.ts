import styled from 'styled-components';
// TODO: BtnAccent move to common
import { BtnAccent } from '../dashboard/BalanceBlock.styles';
import { BaseBtn } from '../common';
import ExplorerLink from '../common/ExplorerLink';
import { View } from '../common/View';

export const SubHeader = styled.h2`
  font-size: 1.5rem;
  line-height: 1.4;
  font-weight: 600;
  color: var(--text-muted);
  margin: 2.4rem 0 1.2rem;
  &:first-child {
    margin-top: 0;
  }
`;

export const AgentList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.5em;
  padding-bottom: 2rem;
`;

export const Button = styled(BtnAccent)`
  height: 3em;
  padding: 0 0.7em;
`;

export const AgentDelete = styled(BaseBtn)`
  display: flex;
  align-items: center;
  justify-content: center;
  /* The app has one destructive colour, shared with the Workspace panels. */
  background-color: transparent;
  border: 1px solid rgba(255, 131, 119, 0.4);
  color: #ff8377;
  border-radius: 5px;
  width: 3em;
  height: 3em;

  svg {
    fill: currentColor;
  }

  &:hover:not(:disabled) {
    background-color: rgba(255, 131, 119, 0.12);
  }
`;

export const TransactionList = styled.ul`
  display: flex;
  flex-direction: column;
  gap: 1em;
  padding: 1em 1em;
`;

export const TransactionRow = styled.li`
  display: flex;
  flex-direction: row;
  gap: 1em;
  list-style: none;
  justify-content: center;

  a {
    text-decoration: underline;
    color: var(--text-primary);
  }
`;

export const ScrollContainer = styled.div`
  height: 100%;
  overflow-y: auto;
`;

/** The page body. A column so the scroll area can take the remaining height. */
export const AgentsView = styled(View)`
  display: flex;
  flex-direction: column;
`;

/** Explanatory copy shown when a list is empty. */
export const EmptyNote = styled.p`
  color: var(--text-muted);
  font-size: 1.4rem;
  line-height: 1.6;
  max-width: 60ch;
`;

/** Status and error copy inside the transactions modal. */
export const ModalNote = styled.p`
  margin: 0;
  padding: 1.6rem;
`;

export const ModalErrorBlock = styled.div`
  padding: 1.6rem;
`;

/** Lets a long hash shrink inside its row instead of forcing the row wider. */
export const TransactionExplorerLink = styled(ExplorerLink)`
  min-width: 0;
`;
