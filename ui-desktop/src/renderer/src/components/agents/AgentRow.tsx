import { AgentUser } from 'src/main/src/client/api.types';
import {
  formatTokenNameValue,
  getAbbreviation,
} from '@renderer/components/agents/utils';
import {
  AgentActionsCell,
  AgentLogo,
  AgentName,
  AgentRow,
  AgentAllowance,
  AgentPermissions,
} from '@renderer/components/agents/AgentRow.styles';
import { Field } from '@renderer/components/agents/Field';
import { useRef, useState } from 'react';
import { useIsOverflow } from '@renderer/hooks/useIsOverflow';
import { Button } from '@renderer/components/agents/Agents.styles';
import Modal from '@renderer/components/common/Modal';
import styled from 'styled-components';
const ViewAllButton = styled(Button)`
  margin: 0.5rem 0 0 0;
  padding: 0.4rem 0.5rem;
  font-size: 1.1rem;
  line-height: 1;
  height: unset;
`;

export const AgentRowComp: React.FC<{
  agent: AgentUser;
  actions: React.ReactNode;
  cfg: { symbol: string; symbolEth: string; morTokenAddress: string };
}> = ({ agent, actions, cfg: props }) => {
  const allowancesRef = useRef<HTMLDivElement>(null);
  const { x, y } = useIsOverflow(allowancesRef);
  const isOverflow = x || y;
  const [isAllowancesModalOpen, setIsAllowancesModalOpen] = useState(false);

  return (
    <AgentRow key={agent.username}>
      <AgentLogo>{getAbbreviation(agent.username)}</AgentLogo>
      <AgentName>{agent.username}</AgentName>
      <AgentPermissions>
        <Field title="Permissions">
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
            {agent.perms.map((permission) => (
              <div key={permission}>{permission}</div>
            ))}
          </div>
        </Field>
      </AgentPermissions>
      <AgentAllowance>
        <Field title="Allowances" ref={allowancesRef}>
          {isOverflow ||
            Object.entries(agent.allowances).map(([token, val]) => {
              const { name, value } = formatTokenNameValue(token, val, props);
              return (
                <div key={token}>
                  {name}: {value}
                </div>
              );
            })}
          {isOverflow && (
            <ViewAllButton onClick={() => setIsAllowancesModalOpen(true)}>
              View all
            </ViewAllButton>
          )}
        </Field>
      </AgentAllowance>
      <AgentActionsCell>{actions}</AgentActionsCell>
      <Modal
        isOpen={isAllowancesModalOpen}
        onRequestClose={() => setIsAllowancesModalOpen(false)}
        variant="primary"
        title="View allowances"
        styleOverrides={{
          width: '500px',
        }}
      >
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: '1rem',
            padding: '1em',
          }}
        >
          {Object.entries(agent.allowances).map(([token, val]) => {
            const { name, value } = formatTokenNameValue(token, val, props);
            return (
              <div key={token}>
                {name}: {value}
              </div>
            );
          })}
        </div>
      </Modal>
    </AgentRow>
  );
};
