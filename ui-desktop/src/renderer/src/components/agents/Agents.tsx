import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';
import { TrashIcon } from '@renderer/components/icons/TrashIcon';
import Modal from '../common/Modal';
import withAgentsState, {
  MappedProps,
  ContainerProps,
} from '@renderer/store/hocs/withAgentsState';
import {
  AgentDelete,
  AgentList,
  Button,
  SubHeader,
  TransactionList,
  TransactionRow,
  ScrollContainer,
} from '@renderer/components/agents/Agents.styles';
import { AgentRowComp } from '@renderer/components/agents/AgentRow';
import { AllowanceRowComp } from '@renderer/components/agents/AllowanceRow';

const Agents = (props: ContainerProps & MappedProps) => {
  const {
    pendingAgents,
    activeAgents,
    allowanceRequests,
    txModal,
    setTxModal,
    handleApproveAccess,
    handleApproveAllowance,
    handleDeleteAgent,
  } = props;

  return (
    <View
      style={{
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <LayoutHeader title="Agents" />
      <ScrollContainer>
        <SubHeader>Access requests</SubHeader>
        <AgentList>
          {pendingAgents.map((agent) => (
            <AgentRowComp
              key={agent.username}
              agent={agent}
              cfg={{
                symbol: props.symbol,
                symbolEth: props.symbolEth,
                morTokenAddress: props.morTokenAddress,
              }}
              actions={
                <>
                  <Button onClick={() => handleApproveAccess(agent, true)}>
                    Approve access
                  </Button>
                  <AgentDelete
                    onClick={() => handleApproveAccess(agent, false)}
                  >
                    <TrashIcon fill="#fff" width="2rem" />
                  </AgentDelete>
                </>
              }
            />
          ))}
        </AgentList>
        <SubHeader>Allowance requests</SubHeader>
        <AgentList>
          {allowanceRequests.map((agent) => (
            <AllowanceRowComp
              key={`${agent.username}-${agent.token}`}
              agent={{
                token: agent.token,
                allowance: agent.allowance,
                username: agent.username,
              }}
              props={{
                symbol: props.symbol,
                symbolEth: props.symbolEth,
                morTokenAddress: props.morTokenAddress,
              }}
              actions={
                <>
                  <Button onClick={() => handleApproveAllowance(agent, true)}>
                    Approve allowance
                  </Button>
                  <AgentDelete
                    onClick={() => handleApproveAllowance(agent, false)}
                  >
                    <TrashIcon fill="#fff" width="2rem" />
                  </AgentDelete>
                </>
              }
            />
          ))}
        </AgentList>
        <SubHeader>All Agents</SubHeader>
        <AgentList>
          {activeAgents.map((agent) => (
            <AgentRowComp
              key={agent.username}
              agent={agent}
              cfg={{
                symbol: props.symbol,
                symbolEth: props.symbolEth,
                morTokenAddress: props.morTokenAddress,
              }}
              actions={
                <>
                  <Button
                    onClick={() =>
                      setTxModal({
                        state: 'loading',
                        agentName: agent.username,
                      })
                    }
                  >
                    Transactions
                  </Button>
                  <AgentDelete onClick={() => handleDeleteAgent(agent)}>
                    <TrashIcon fill="#fff" width="2rem" />
                  </AgentDelete>
                </>
              }
            />
          ))}
        </AgentList>
      </ScrollContainer>
      <Modal
        isOpen={txModal.state !== 'pending'}
        onRequestClose={() => setTxModal({ state: 'pending' })}
        variant="primary"
        title="View transactions"
      >
        <TransactionList>
          {txModal.state === 'success' && (
            <>
              {txModal.data.map((tx) => {
                return (
                  <TransactionRow key={tx}>
                    <a target="_blank" href={props.txUrlResolver(tx)}>
                      {tx}
                    </a>
                  </TransactionRow>
                );
              })}
            </>
          )}
        </TransactionList>
      </Modal>
    </View>
  );
};

export default withAgentsState(Agents);
