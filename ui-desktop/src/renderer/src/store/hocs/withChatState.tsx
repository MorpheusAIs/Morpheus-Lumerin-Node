import React, { ComponentType } from 'react';
import { connect } from 'react-redux';
import { withClient } from './clientContext';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';
import { explainChainError } from '../utils/chainErrors';
import { ApiGateway } from 'src/main/src/client/apiGateway';

export interface ContainerProps {
  client: ApiGateway;
  config: any;
  address: string;
  symbol: string;
  selectedBid?: any;
  model?: any;
  provider?: any;
  activeSession?: any;
  setBid: (model: any) => void;
}

// WrappedComponent receives `ContainerProps` plus all the helper props the HOC
// injects (onOpenSession, getBidsByModelId, etc.) — typed loosely as `any` because
// the container builds them dynamically and individual consumers refine them
// in their own prop types.
const withChatState = (WrappedComponent: ComponentType<any>) => {
  class Container extends React.Component<ContainerProps> {
    static contextType = ToastsContext;
    declare context: React.ContextType<typeof ToastsContext>;

    static displayName = `withChatState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    closeSession = async (sessionId: string) => {
      this.context.toast('info', 'Closing...');
      try {
        const data = await this.props.client.closeSession({ sessionId });
        if (data.error) {
          this.context.toast('error', 'Session not closed');
          throw new Error(data.error);
        }
        if (data.tx) {
          this.context.toast('success', 'Session successfully closed');
        }
      } catch (e) {
        console.log('Error', e);
        this.context.toast('error', 'Failed to close session');
      }
    };

    getAllModels = async () => {
      return (await this.props.client.getAllModels()) || [];
    };

    getLocalModels = async () => {
      return (await this.props.client.getLocalModels()) || [];
    };

    getMetaInfo = async () => {
      const [budget, supply] = await Promise.all([
        this.props.client.getTodaysBudget(),
        this.props.client.getTokenSupply(),
      ]);
      return { budget, supply };
    };

    getSessionsByUser = async (user) => {
      if (!user) {
        return;
      }

      return await this.props.client.getSessionsByUser({ user });
    };

    getBidInfo = async (id) => {
      if (!id) {
        return;
      }

      return await this.props.client.getBidInfo({ id });
    };

    getBidsByModelId = async (modelId) => {
      if (!modelId) {
        return;
      }

      const bids = await this.props.client.getBidsByModel({ modelId });
      return (bids ?? [])
        .filter((b) => +b.DeletedAt === 0)
        .filter((b) => b.Provider != this.props.address);
    };

    onOpenSession = async ({ modelId, duration, isDirectPay = false }) => {
      this.context.toast('info', 'Checking and opening session…');
      try {
        const failoverSettings = await this.props.client.getFailoverSetting();

        const dataResponse = await this.props.client.openSession({
          modelId,
          failover: failoverSettings?.isEnabled || false,
          duration: +duration,
          directPayment: isDirectPay,
        });
        if (dataResponse?.existingSessionID) {
          this.context.toast(
            'info',
            'An open session already exists. Resuming it instead.',
          );
          return { existingSessionID: dataResponse.existingSessionID };
        }
        if (dataResponse?.error) {
          // The proxy-router nests its failures several layers deep
          // ("failed to send transaction: open session failed: failed to send
          // transaction: <real cause>"). Surfacing that verbatim told the user
          // nothing, and hid the fact that the two most common causes — no ETH
          // for gas, and a read-only RPC endpoint — need completely different
          // fixes.
          const { message, hint } = explainChainError(dataResponse.error);
          this.context.toast('error', hint ? `${message} ${hint}` : message, {
            autoClose: 15000,
          });
          console.error('Failed to initiate session:', dataResponse.error);
          return;
        }
        this.context.toast('success', 'Session successfully created');
        return dataResponse.sessionID;
      } catch (e) {
        console.error(e);
        const { message, hint } = explainChainError(e);
        this.context.toast('error', hint ? `${message} ${hint}` : message, {
          autoClose: 15000,
        });
        return;
      }
    };

    getBalances = async () => {
      return await this.props.client.getBalances();
    };

    render() {
      return (
        <WrappedComponent
          getAllModels={this.getAllModels}
          getLocalModels={this.getLocalModels}
          getBidInfo={this.getBidInfo}
          getMetaInfo={this.getMetaInfo}
          getBidsByModelId={this.getBidsByModelId}
          getSessionsByUser={this.getSessionsByUser}
          closeSession={this.closeSession}
          onOpenSession={this.onOpenSession}
          getBalances={this.getBalances}
          toasts={this.context}
          {...this.state}
          {...this.props}
        />
      );
    }
  }

  const mapStateToProps = (state, _props) => ({
    // selectedCurrency: selectors.getSellerSelectedCurrency(state),
    config: state.config,
    selectedBid: state.models.selectedBid,
    model: state.models.selectedModel,
    provider: state.models.selectedProvider,
    activeSession: state.models.activeSession,
    address: selectors.getWalletAddress(state),
    symbol: selectors.getCoinSymbol(state),
  });

  const mapDispatchToProps = (dispatch) => ({
    setBid: (model) => dispatch({ type: 'set-bid', payload: model }),
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withChatState;
