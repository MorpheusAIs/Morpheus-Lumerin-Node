import styled from 'styled-components';

export const AgentActionsCell = styled.div`
  display: flex;
  justify-content: flex-end;
  gap: 1em;
  align-items: center;
`;

export const AgentRow = styled.div`
  display: flex;
  align-items: center;
  background-color: #0c1f17;
  width: 100%;
  gap: 1em;
  min-height: 4em;
  display: grid;
  grid-template-columns: 3em 1fr 2fr 6em 1.5fr;
  padding: 1em 1em;
`;

export const AgentLogo = styled.div`
  width: 3em;
  height: 3em;
  background-color: #07150f;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5em;
`; // TODO: replace with theme color

export const AgentName = styled.div`
  font-size: 1.7rem;
  font-weight: 600;
  overflow-wrap: anywhere;
`;

export const AgentAllowance = styled.div`
  font-size: 1.2rem;
  display: flex;
  flex-direction: column;
`;

export const AgentPermissions = styled.div`
  font-size: 1.2rem;
  display: flex;
  flex-wrap: wrap;
  column-gap: 1em;
`;
