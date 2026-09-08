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
    <View
      style={{
        display: 'flex',
        flexDirection: 'column',
      }}
    >
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
            <p
              style={{
                color: 'var(--text-muted)',
                fontSize: '1.4rem',
                lineHeight: 1.6,
                maxWidth: '60ch',
              }}
            >
              No agents connected yet. Agent access requests will appear here
              for you to review before approving.
            </p>
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
          <p role="status" style={{ padding: '1.6rem', margin: 0 }}>
            Loading transactions…
          </p>
        )}
        {txModal.state === 'error' && (
          <div style={{ padding: '1.6rem' }}>
            <p role="alert">{txModal.error}</p>
            <Button
              onClick={() =>
                setTxModal({ state: 'loading', agentName: txModal.agentName })
              }
              type="button"
            >
              Retry transactions
            </Button>
          </div>
        )}
        {txModal.state === 'success' && txModal.data.length === 0 && (
          <p role="status" style={{ padding: '1.6rem', margin: 0 }}>
            No transactions recorded for this agent yet.
          </p>
        )}
        <TransactionList>
          {txModal.state === 'success' && (
            <>
              {txModal.data.map((tx) => {
                return (
                  <TransactionRow key={tx}>
                    <a
                      target="_blank"
                      rel="noopener noreferrer"
                      href={props.txUrlResolver(tx)}
                      style={{ overflowWrap: 'anywhere', minWidth: 0 }}
                    >
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
