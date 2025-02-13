import {
  AgentActionsCell,
  AgentLogo,
  AgentName,
} from '@renderer/components/agents/AgentRow.styles';
import {
  AgentAllowanceToken,
  AgentAllowanceValue,
  AllowanceRow,
} from '@renderer/components/agents/AllowanceRow.styles';
import { Field } from '@renderer/components/agents/Field';
import {
  formatTokenNameValue,
  getAbbreviation,
} from '@renderer/components/agents/utils';

export const AllowanceRowComp: React.FC<{
  agent: { username: string; token: string; allowance: string };
  actions: React.ReactNode;
  props: { symbol: string; symbolEth: string; morTokenAddress: string };
}> = ({ agent, actions, props }) => {
  const { name, value } = formatTokenNameValue(
    agent.token,
    agent.allowance,
    props,
  );
  return (
    <AllowanceRow key={agent.username}>
      <AgentLogo>{getAbbreviation(agent.username)}</AgentLogo>
      <AgentName>{agent.username}</AgentName>
      <AgentAllowanceToken>
        <Field title="Token" children={name} />
      </AgentAllowanceToken>
      <AgentAllowanceValue>
        <Field title="Allowance" children={value} />
      </AgentAllowanceValue>

      <AgentActionsCell>{actions}</AgentActionsCell>
    </AllowanceRow>
  );
};
