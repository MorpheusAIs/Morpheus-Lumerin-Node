import { ComponentType, useState, useEffect, useContext } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import { ToastsContext } from '../../components/toasts';
import { ApiGateway } from 'src/main/src/client/apiGateway';
import {
  AgentUser,
  AgentAllowanceRequest,
} from 'src/main/src/client/api.types';
import { queryKeys } from '../queries';

export interface ContainerProps {
  client: ApiGateway;
  pendingAgents: AgentUser[];
  activeAgents: AgentUser[];
  allowanceRequests: AgentAllowanceRequest[];
  txModal: TxModal;
  setTxModal: (txModal: TxModal) => void;
  handleApproveAccess: (agent: AgentUser, approve: boolean) => Promise<void>;
  handleApproveAllowance: (
    data: { username: string; token: string },
    approve: boolean,
  ) => Promise<void>;
  handleDeleteAgent: (agent: AgentUser) => Promise<void>;
  agentsLoading: boolean;
  agentsError: unknown;
  retryAgents: () => void;
}

type TxModal =
  | {
      state: 'pending';
    }
  | {
      state: 'loading';
      agentName: string;
    }
  | {
      state: 'success';
      agentName: string;
      data: string[];
    }
  | {
      state: 'error';
      agentName: string;
      error: string;
    };

export interface MappedProps {
  config: any; // Replace 'any' with actual config type
  syncStatus: boolean;
  address: string;
  symbol: string;
  symbolEth: string;
  txUrlResolver: (hash: string) => string;
  morTokenAddress: string;
}

// `MappedProps` get injected later by `connect()`; the Container itself only
// receives `ContainerProps`. Type the wrapped component loosely so callers can
// declare their own prop shapes without fighting HOC composition.
const withAgentsState = (WrappedComponent: ComponentType<any>) => {
  const Container = (props: ContainerProps & MappedProps) => {
    const context = useContext(ToastsContext);
    const queryClient = useQueryClient();

    const [txModal, setTxModal] = useState<TxModal>({ state: 'pending' });

    useEffect(() => {
      if (txModal.state !== 'pending') {
        props.client
          .getAgentTxs({ username: txModal.agentName, cursor: '', limit: 10 })
          .then((res) => {
            if (!res) {
              setTxModal({
                state: 'error',
                agentName: txModal.agentName,
                error: 'Failed to fetch transactions',
              });
            } else {
              setTxModal({
                state: 'success',
                agentName: txModal.agentName,
                data: res.txHashes,
              });
            }
          });
      }
    }, [txModal.state !== 'pending' && txModal.agentName]);

    const agentsQuery = useQuery({
      queryKey: queryKeys.agents(props.address),
      queryFn: () => loadAgentsPageData(props.client),
    });
    const pendingAgents = agentsQuery.data?.pendingAgents ?? [];
    const activeAgents = agentsQuery.data?.activeAgents ?? [];
    const allowanceRequests = agentsQuery.data?.allowanceRequests ?? [];

    async function handleApproveAccess(agent: AgentUser, approve: boolean) {
      const res = await props.client.confirmDeclineAgentUser({
        username: agent.username,
        confirm: approve,
      });
      if (res) {
        context.toast(
          'success',
          `Agent "${agent.username}" ${approve ? 'approved' : 'declined'}`,
        );
        await queryClient.invalidateQueries({
          queryKey: queryKeys.agents(props.address),
        });
      }
    }

    async function handleApproveAllowance(
      data: { username: string; token: string },
      approve: boolean,
    ) {
      const res = await props.client.confirmDeclineAgentAllowanceRequest({
        username: data.username,
        token: data.token,
        confirm: approve,
      });
      if (res) {
        context.toast(
          'success',
          `Allowance for "${data.username}" ${approve ? 'approved' : 'declined'}`,
        );
        await queryClient.invalidateQueries({
          queryKey: queryKeys.agents(props.address),
        });
      }
    }

    async function handleDeleteAgent(agent: AgentUser) {
      const res = await props.client.removeAgentUser({
        username: agent.username,
      });
      if (res) {
        context.toast('success', `Agent "${agent.username}" deleted`);
        await queryClient.invalidateQueries({
          queryKey: queryKeys.agents(props.address),
        });
      }
    }

    return (
      <WrappedComponent
        {...props}
        pendingAgents={pendingAgents}
        activeAgents={activeAgents}
        allowanceRequests={allowanceRequests}
        txModal={txModal}
        setTxModal={setTxModal}
        handleApproveAccess={handleApproveAccess}
        handleApproveAllowance={handleApproveAllowance}
        handleDeleteAgent={handleDeleteAgent}
        agentsLoading={agentsQuery.isPending}
        agentsError={agentsQuery.error}
        retryAgents={() => void agentsQuery.refetch()}
      />
    );
  };

  const mapStateToProps = (state): MappedProps => ({
    config: state.config,
    syncStatus: selectors.getTxSyncStatus(state),
    address: selectors.getWalletAddress(state),
    symbol: selectors.getCoinSymbol(state),
    symbolEth: selectors.getSymbolEth(state),
    txUrlResolver: selectors.getTransactionExplorerUrlResolver(state),
    morTokenAddress: state.config.chain.mainTokenAddress,
  });

  return withClient(connect(mapStateToProps)(Container));
};

export const loadAgentsPageData = async (client: ApiGateway) => {
  // These endpoints are independent. Starting both before awaiting either
  // removes an unnecessary round trip from the first Agents render.
  const [agentUsers, allowanceRequests] = await Promise.all([
    client.getAgentUsers(),
    client.getAgentAllowanceRequests(),
  ]);
  if (!agentUsers) throw new Error('Failed to fetch agent access requests.');
  if (!allowanceRequests) {
    throw new Error('Failed to fetch agent allowance requests.');
  }

  const pendingAgents: AgentUser[] = [];
  const activeAgents: AgentUser[] = [];
  for (const agent of agentUsers.agents) {
    (agent.isConfirmed ? activeAgents : pendingAgents).push(agent);
  }

  return {
    pendingAgents,
    activeAgents,
    allowanceRequests: allowanceRequests.requests as AgentAllowanceRequest[],
  };
};

export default withAgentsState;
