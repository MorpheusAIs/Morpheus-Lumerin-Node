import styled from 'styled-components';
import { AgentRow } from '@renderer/components/agents/AgentRow.styles';

export const AllowanceRow = styled(AgentRow)`
  grid-template-columns: 3em 1fr 1.25fr 1.25fr 1.5fr;
`; // TODO: replace with theme color

export const AgentAllowanceToken = styled.div`
  font-size: 1.2rem;
  display: flex;
  flex-direction: column;
`;

export const AgentAllowanceValue = styled.div`
  font-size: 1.2rem;
  display: flex;
  flex-direction: column;
`;
