import styled from 'styled-components';
// TODO: BtnAccent move to common
import { BtnAccent } from '../dashboard/BalanceBlock.styles';
import { BaseBtn } from '../common';

export const SubHeader = styled.h2`
  font-size: 2rem;
  white-space: nowrap;
  font-weight: 600;
  color: ${(p) => p.theme.colors.morMain};
  margin: 2em 0 1em;
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
  background-color: #970000;
  svg {
    fill: #fff;
  }
  color: #fff;
  border-radius: 5px;
  width: 3em;
  height: 3em;
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
    color: #fff;
  }
`;

export const ScrollContainer = styled.div`
  height: 100%;
  overflow-y: scroll;
`;
