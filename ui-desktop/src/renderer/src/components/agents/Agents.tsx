import { LayoutHeader } from '../common/LayoutHeader';
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
  AgentsView,
  EmptyNote,
  ModalNote,
  ModalErrorBlock,
  TransactionExplorerLink,
} from '@renderer/components/agents/Agents.styles';
import { AgentRowComp } from '@renderer/components/agents/AgentRow';
import { AllowanceRowComp } from '@renderer/components/agents/AllowanceRow';
import QueryError from '../common/QueryError';

export const Agents = (props: ContainerProps & MappedProps) => {
  const {
    pendingAgents,
    activeAgents,
    allowanceRequests,
    txModal,
    setTxModal,
    handleApproveAccess,
    handleApproveAllowance,
    handleDeleteAgent,
    agentsLoading,
    agentsError,
    retryAgents,
  } = props;

  return (
    <AgentsView>
      <LayoutHeader title="Agents" />
      <QueryError error={agentsError} what="agents" onRetry={retryAgents} />
      <ScrollContainer>
        {agentsLoading && <SubHeader role="status">Loading agents…</SubHeader>}
        {pendingAgents.length > 0 && (
          <>
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
                        aria-label={`Decline access for ${agent.username}`}
                        onClick={() => handleApproveAccess(agent, false)}
                      >
                        <TrashIcon fill="#fff" width="2rem" />
                      </AgentDelete>
                    </>
                  }
                />
              ))}
            </AgentList>
          </>
        )}
        {allowanceRequests.length > 0 && (
          <>
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
                      <Button
                        onClick={() => handleApproveAllowance(agent, true)}
                      >
                        Approve allowance
                      </Button>
                      <AgentDelete
                        aria-label={`Decline allowance for ${agent.username}`}
                        onClick={() => handleApproveAllowance(agent, false)}
                      >
                        <TrashIcon fill="#fff" width="2rem" />
                      </AgentDelete>
                    </>
                  }
                />
              ))}
            </AgentList>
          </>
        )}
        {activeAgents.length > 0 && <SubHeader>Connected agents</SubHeader>}
        {!agentsLoading &&
          !agentsError &&
          activeAgents.length === 0 &&
          pendingAgents.length === 0 &&
          allowanceRequests.length === 0 && (
            <EmptyNote>
              No agents connected yet. Agent access requests will appear here
              for you to review before approving.
            </EmptyNote>
          )}
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
                  <AgentDelete
                    aria-label={`Remove agent ${agent.username}`}
                    onClick={() => handleDeleteAgent(agent)}
                  >
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
        {txModal.state === 'loading' && (
          <ModalNote role="status">Loading transactions…</ModalNote>
        )}
        {txModal.state === 'error' && (
          <ModalErrorBlock>
            <p role="alert">{txModal.error}</p>
            <Button
              onClick={() =>
                setTxModal({ state: 'loading', agentName: txModal.agentName })
              }
              type="button"
            >
              Retry transactions
            </Button>
          </ModalErrorBlock>
        )}
        {txModal.state === 'success' && txModal.data.length === 0 && (
          <ModalNote role="status">
            No transactions recorded for this agent yet.
          </ModalNote>
        )}
        <TransactionList>
          {txModal.state === 'success' && (
            <>
              {txModal.data.map((tx) => {
                return (
                  <TransactionRow key={tx}>
                    <TransactionExplorerLink
                      kind="transaction"
                      url={props.txUrlResolver(tx)}
                    >
                      {tx}
                    </TransactionExplorerLink>
                  </TransactionRow>
                );
              })}
            </>
          )}
        </TransactionList>
      </Modal>
    </AgentsView>
  );
};

export default withAgentsState(Agents);
