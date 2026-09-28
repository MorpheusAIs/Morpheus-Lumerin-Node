import { AgentUser } from 'src/main/src/client/api.types';
import { formatTokenNameValue, getAbbreviation } from './utils';
import {
  AgentActionsCell,
  AgentLogo,
  AgentName,
  AgentRow,
  AgentAllowance,
  AgentPermissions,
  AgentFieldLabel,
  PermissionList,
  AllowanceValues,
  AllowanceEntry,
  ViewAllButton,
} from './AgentRow.styles';
import { useId, useRef, useState } from 'react';
import { useIsOverflow } from '@renderer/hooks/useIsOverflow';

export const AgentRowComp: React.FC<{
  agent: AgentUser;
  actions: React.ReactNode;
  cfg: { symbol: string; symbolEth: string; morTokenAddress: string };
}> = ({ agent, actions, cfg: props }) => {
  const allowancesRef = useRef<HTMLDListElement>(null);
  const { x, y } = useIsOverflow(allowancesRef);
  const isOverflow = x || y;
  const [isExpanded, setIsExpanded] = useState(false);
  const allowancesId = useId();
  const permissionsId = useId();
  const allowanceEntries = Object.entries(agent.allowances || {});

  return (
    <AgentRow role="group" aria-label={`Agent ${agent.username}`}>
      <AgentLogo>{getAbbreviation(agent.username)}</AgentLogo>
      <AgentName>{agent.username}</AgentName>
      <AgentPermissions aria-labelledby={permissionsId}>
        <AgentFieldLabel id={permissionsId}>Permissions</AgentFieldLabel>
        <PermissionList>
          {agent.perms.map((permission) => (
            <li key={permission}>{permission}</li>
          ))}
          {agent.perms.length === 0 && <li>None</li>}
        </PermissionList>
      </AgentPermissions>
      <AgentAllowance>
        <AgentFieldLabel>Allowances</AgentFieldLabel>
        <AllowanceValues
          $expanded={isExpanded}
          aria-label={`Allowances for ${agent.username}`}
          id={allowancesId}
          ref={allowancesRef}
        >
          {allowanceEntries.map(([token, val]) => {
            const { name, value } = formatTokenNameValue(token, val, props);
            return (
              <AllowanceEntry key={token}>
                <dt>{name}</dt>
                <dd>{value}</dd>
              </AllowanceEntry>
            );
          })}
        </AllowanceValues>
        {allowanceEntries.length === 0 && <span>None</span>}
        {(isOverflow || isExpanded) && (
          <ViewAllButton
            aria-controls={allowancesId}
            aria-expanded={isExpanded}
            onClick={() => setIsExpanded((expanded) => !expanded)}
            type="button"
          >
            {isExpanded ? 'Show less' : 'Show all allowances'}
          </ViewAllButton>
        )}
      </AgentAllowance>
      <AgentActionsCell>{actions}</AgentActionsCell>
    </AgentRow>
  );
};
