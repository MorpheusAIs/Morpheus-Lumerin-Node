import React, { ComponentType } from 'react';
import { connect } from 'react-redux';
import { withClient } from './clientContext';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';
import {
  getSessionsByUser,
  getBidsByModelId,
  getBidInfoById,
} from '../utils/apiCallsHelper';
import { ApiGateway } from 'src/main/src/client/apiGateway';

const AvailabilityStatus = {
  available: 'available',
  unknown: 'unknown',
  disconnected: 'disconnected',
};

export interface ContainerProps {
  client: ApiGateway;
  config: any;
}

const withChatState = (WrappedComponent: ComponentType<ContainerProps>) => {
  class Container extends React.Component<ContainerProps> {
    static contextType = ToastsContext;
    declare context: React.ContextType<typeof ToastsContext>;

    static displayName = `withChatState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    getProviders = async () => {
      try {
        const authHeaders = await this.props.client.getAuthHeaders();
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/providers`;
        const response = await fetch(path, {
          headers: authHeaders,
        });
        const data = await response.json();
        if (data.error) {
          console.error(data.error);
          return [];
        }
        return data.providers;
      } catch (e) {
        console.log('Error', e);
        return [];
      }
    };

    closeSession = async (sessionId: string) => {
      this.context.toast('info', 'Closing...');
      try {
        const authHeaders = await this.props.client.getAuthHeaders();
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/sessions/${sessionId}/close`;
        const response = await fetch(path, {
          method: 'POST',
          headers: authHeaders,
        });
        const data = await response.json();
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
      try {
        const authHeaders = await this.props.client.getAuthHeaders();
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/models`;
        const response = await fetch(path, {
          headers: authHeaders,
          method: 'GET',
        });
        const data = await response.json();
        if (data.error) {
          console.error(data.error);
          return [];
        }
        return data.models;
      } catch (e) {
        console.log('Error', e);
        return [];
      }
    };

    getLocalModels = async () => {
      try {
        const authHeaders = await this.props.client.getAuthHeaders();
        const path = `${this.props.config.chain.localProxyRouterUrl}/v1/models`;
        const response = await fetch(path, {
          headers: authHeaders,
        });
        if (!response.ok) {
          return [];
        }
        return await response.json();
      } catch (e) {
        console.log('Error', e);
        return [];
      }
    };

    getModelsData = async () => {
      const [localModels, modelsResp, providersResp, meta, userBalances] =
        await Promise.all([
          this.getLocalModels(),
          this.getAllModels(),
          this.getProviders(),
          this.getMetaInfo(),
          this.getBalances(),
        ]);

      const models = modelsResp.filter((m) => !m.IsDeleted);
      const providers = providersResp.filter((m) => !m.IsDeleted);

      const result = [
        ...localModels.map((m) => ({ ...m, isLocal: true })),
        ...models,
      ];

      return { models: result, providers, meta, userBalances };
    };

    getProvidersAvailability = async (providers) => {
      const isValidUrl = (url) => {
        const urlRegex =
          /^(https?:\/\/)?(([a-zA-Z0-9.-]+\.[a-zA-Z]{2,}|localhost)|(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}))(:\d{1,5})?(\/\S*)?$/;
        return urlRegex.test(url);
      };

      const availabilityResults = await Promise.all(
        providers.map(async (p) => {
          try {
            const recordString = localStorage.getItem(p.Address);
            const storedRecord = recordString && JSON.parse(recordString);
            if (
              storedRecord &&
              storedRecord.status == AvailabilityStatus.available
            ) {
              const lastUpdatedAt = new Date(storedRecord.time);
              const cacheMinutes = 15;
              const timestampBefore = new Date(
                new Date().getTime() - cacheMinutes * 60 * 1000,
              );

              if (lastUpdatedAt > timestampBefore) {
                return { ...storedRecord, id: p.Address };
              }
            }

            if (!isValidUrl(p.Endpoint)) {
              return {
                id: p.Address,
                status: AvailabilityStatus.disconnected,
                time: new Date(),
              };
            }

            const isValid = await this.props.client.checkProviderConnectivity({
              endpoint: p.Endpoint,
              address: p.Address,
            });

            const record = {
              id: p.Address,
              status: isValid
                ? AvailabilityStatus.available
                : AvailabilityStatus.disconnected,
              time: new Date(),
            };
            localStorage.setItem(
              record.id,
              JSON.stringify({ status: record.status, time: record.time }),
            );
            return record;
          } catch (e) {
            return {
              id: p.Address,
              status: AvailabilityStatus.unknown,
              time: new Date(),
            };
          }
        }),
      );
      return availabilityResults;
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

      const authHeaders = await this.props.client.getAuthHeaders();
      return await getSessionsByUser(
        this.props.config.chain.localProxyRouterUrl,
        user,
        authHeaders,
      );
    };

    getBidInfo = async (id) => {
      if (!id) {
        return;
      }

      const authHeaders = await this.props.client.getAuthHeaders();
      return await getBidInfoById(
        this.props.config.chain.localProxyRouterUrl,
        id,
        authHeaders,
      );
    };

    getBidsByModelId = async (modelId) => {
      if (!modelId) {
        return;
      }

      const authHeaders = await this.props.client.getAuthHeaders();
      const bids = await getBidsByModelId(
        this.props.config.chain.localProxyRouterUrl,
        modelId,
        authHeaders,
      );
      return bids
        .filter((b) => +b.DeletedAt === 0)
        .filter((b) => b.Provider != this.props.address);
    };

    onOpenSession = async ({ modelId, duration, isDirectPay = false }) => {
      this.context.toast('info', 'Processing...');
      try {
        const failoverSettings = await this.props.client.getFailoverSetting();

        const authHeaders = await this.props.client.getAuthHeaders();
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/models/${modelId}/session`;
        const body = {
          failover: failoverSettings?.isEnabled || false,
          sessionDuration: +duration, // convert to seconds
          directPayment: isDirectPay,
        };
        const response = await fetch(path, {
          method: 'POST',
          body: JSON.stringify(body),
          headers: authHeaders,
        });
        const dataResponse = await response.json();
        if (!response.ok) {
          this.context.toast(
            'error',
            `Failed to open session: "${dataResponse.error}"`,
          );
          console.log('Failed initiate session', dataResponse);
          return;
        }
        this.context.toast('success', 'Session successfully created');
        return dataResponse.sessionID;
      } catch (e) {
        console.error(e);
        this.context.toast('error', 'Failed to open session');
        return;
      }
    };

    getBalances = async () => {
      return await this.props.client.getBalances();
    };

    render() {
      return (
        <WrappedComponent
          getProviders={this.getProviders}
          getProvidersAvailability={this.getProvidersAvailability}
          getBidInfo={this.getBidInfo}
          getMetaInfo={this.getMetaInfo}
          getBidsByModelId={this.getBidsByModelId}
          getModelsData={this.getModelsData}
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

  const mapStateToProps = (state, props) => ({
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
