import { ComponentType, useState, useEffect, useContext } from 'react';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import { ToastsContext } from '../../components/toasts';
import { ApiGateway } from 'src/main/src/client/apiGateway';
import {
  AgentUser,
  AgentAllowanceRequest,
} from 'src/main/src/client/api.types';

export interface ContainerProps {
  client: ApiGateway;
  pendingAgents: AgentUser[];
  activeAgents: AgentUser[];
  allowanceRequests: AgentAllowanceRequest[];
  isLoading: boolean;
  loadError: string | null;
  txModal: TxModal;
  setTxModal: (txModal: TxModal) => void;
  handleApproveAccess: (agent: AgentUser, approve: boolean) => Promise<void>;
  handleApproveAllowance: (
    data: { username: string; token: string },
    approve: boolean,
  ) => Promise<void>;
  handleDeleteAgent: (agent: AgentUser) => Promise<void>;
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
  const Container = (props: ContainerProps) => {
    const [pendingAgents, setPendingAgents] = useState<AgentUser[]>([]);
    const [activeAgents, setActiveAgents] = useState<AgentUser[]>([]);
    const [allowanceRequests, setAllowanceRequests] = useState<
      AgentAllowanceRequest[]
    >([]);
    const [isLoading, setIsLoading] = useState(true);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [refresh, setRefresh] = useState(0);
    const context = useContext(ToastsContext);

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

    function refreshPage() {
      setRefresh((prev) => prev + 1);
    }

    useEffect(() => {
      void fetchPageData();
    }, [refresh]);

    async function fetchPageData() {
      setIsLoading(true);
      setLoadError(null);

      try {
        const [agentUsersResponse, allowanceRequestsResponse] =
          await Promise.all([
            props.client.getAgentUsers(),
            props.client.getAgentAllowanceRequests(),
          ]);

        if (!agentUsersResponse || !allowanceRequestsResponse) {
          throw new Error('One or more agent API requests failed');
        }

        const nextPendingAgents: AgentUser[] = [];
        const nextActiveAgents: AgentUser[] = [];

        for (const agent of agentUsersResponse.agents) {
          if (agent.isConfirmed) {
            nextActiveAgents.push(agent);
          } else {
            nextPendingAgents.push(agent);
          }
        }

        setPendingAgents(nextPendingAgents);
        setActiveAgents(nextActiveAgents);
        setAllowanceRequests(allowanceRequestsResponse.requests);
      } catch (error) {
        console.error('Failed to fetch agent data', error);
        setLoadError(
          'Unable to load agent data. Check that the proxy router is running and try again.',
        );
      } finally {
        setIsLoading(false);
      }
    }

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
        refreshPage();
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
        refreshPage();
      }
    }

    async function handleDeleteAgent(agent: AgentUser) {
      const res = await props.client.removeAgentUser({
        username: agent.username,
      });
      if (res) {
        context.toast('success', `Agent "${agent.username}" deleted`);
        refreshPage();
      }
    }

    return (
      <WrappedComponent
        {...props}
        pendingAgents={pendingAgents}
        activeAgents={activeAgents}
        allowanceRequests={allowanceRequests}
        isLoading={isLoading}
        loadError={loadError}
        txModal={txModal}
        setTxModal={setTxModal}
        handleApproveAccess={handleApproveAccess}
        handleApproveAllowance={handleApproveAllowance}
        handleDeleteAgent={handleDeleteAgent}
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

export default withAgentsState;
