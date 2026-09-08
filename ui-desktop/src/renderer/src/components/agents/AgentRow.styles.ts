import styled from 'styled-components';

export const AgentActionsCell = styled.div`
  display: flex;
  justify-content: flex-end;
  gap: 0.8rem;
  align-items: center;
  flex-wrap: wrap;
  min-width: 0;
  button {
    max-width: 100%;
    white-space: normal;
    height: auto;
    min-height: 3.6rem;
    line-height: 1.35;
    padding-block: 0.7rem;
    overflow-wrap: anywhere;
  }
`;

export const AgentLogo = styled.div`
  width: 3em;
  height: 3em;
  background-color: var(--surface-base, #071711);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5em;
`;

export const AgentName = styled.div`
  min-width: 0;
  font-size: 1.7rem;
  font-weight: 600;
  overflow-wrap: anywhere;
`;

export const AgentAllowance = styled.div`
  min-width: 0;
  font-size: 1.3rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
`;

export const AgentPermissions = styled.div`
  min-width: 0;
  font-size: 1.3rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
`;

export const AgentFieldLabel = styled.div`
  color: var(--text-muted, #9ab4a7);
  font-size: 1.2rem;
  font-weight: 500;
`;

export const PermissionList = styled.ul`
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0.4rem 0.8rem;
  margin: 0;
  padding: 0;
  list-style: none;
  overflow-wrap: anywhere;
  li {
    min-width: 0;
  }
`;

export const AllowanceValues = styled.dl<{ $expanded: boolean }>`
  min-width: 0;
  margin: 0;
  max-height: ${(p) => (p.$expanded ? 'none' : '5.4em')};
  overflow: hidden;
  line-height: 1.5;
`;

export const AllowanceEntry = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0 0.4rem;
  min-width: 0;
  overflow-wrap: anywhere;
  dt {
    min-width: 0;
    font-weight: 500;
  }
  dt::after {
    content: ':';
  }
  dd {
    min-width: 0;
    margin: 0;
    font-variant-numeric: tabular-nums;
  }
`;

export const ViewAllButton = styled.button`
  align-self: flex-start;
  display: inline-flex;
  align-items: center;
  min-height: 3.2rem;
  max-width: 100%;
  padding: 0.4rem 0;
  border: 0;
  background: transparent;
  color: var(--accent, #19d695);
  font-family: inherit;
  font-size: 1.2rem;
  line-height: 1.35;
  text-align: start;
  cursor: pointer;
  &:hover {
    text-decoration: underline;
    text-underline-offset: 3px;
  }
  &:focus-visible {
    outline: 2px solid var(--accent, #19d695);
    outline-offset: 3px;
  }
`;

export const AgentRow = styled.div`
  display: grid;
  align-items: center;
  grid-template-columns:
    3em minmax(8rem, 1fr) minmax(10rem, 1.25fr) minmax(10rem, 1fr)
    auto;
  grid-template-areas: 'logo name permissions allowances actions';
  gap: 1.6rem;
  width: 100%;
  min-width: 0;
  min-height: 4em;
  padding: 1.6rem;
  border-radius: 10px;
  background-color: var(--surface-raised, #10271e);
  color: var(--text-primary, #edf7f0);
  ${AgentLogo} {
    grid-area: logo;
  }
  ${AgentName} {
    grid-area: name;
  }
  ${AgentPermissions} {
    grid-area: permissions;
  }
  ${AgentAllowance} {
    grid-area: allowances;
  }
  ${AgentActionsCell} {
    grid-area: actions;
  }

  /* styled-components v4 does not serialize @container rules correctly. */
  @media (max-width: 1180px) {
    grid-template-columns: 3em minmax(0, 1fr) minmax(0, 1fr);
    grid-template-areas: 'logo name actions' '. permissions allowances';
    align-items: start;
  }

  @media (max-width: 680px) {
    grid-template-columns: 3em minmax(0, 1fr);
    grid-template-areas: 'logo name' 'permissions permissions' 'allowances allowances' 'actions actions';
    gap: 1.2rem;
    ${AgentActionsCell} {
      justify-content: flex-start;
    }
  }
`;
