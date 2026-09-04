import React, { ComponentType } from 'react';
import { connect } from 'react-redux';
import { withClient } from './clientContext';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';
import { pooledMapSettled, withTimeout } from '../utils/concurrency';
import { explainChainError } from '../utils/chainErrors';
import { ApiGateway } from 'src/main/src/client/apiGateway';

const AvailabilityStatus = {
  available: 'available',
  unknown: 'unknown',
  disconnected: 'disconnected',
};

// Availability fan-out tuning. Previously every provider was pinged at once
// with no timeout, and only *successful* results were cached — so each visit
// re-pinged every dead provider and waited for each to time out. Dead providers
// are precisely the slow ones, which is why this tab could take minutes.
const AVAILABILITY_CONCURRENCY = 6;
const AVAILABILITY_PING_TIMEOUT_MS = 4000;
// Successful checks stay valid longer than failures, so a provider that comes
// back online is picked up reasonably quickly without re-pinging dead ones on
// every single render.
const AVAILABILITY_TTL_MS = { available: 15 * 60_000, other: 5 * 60_000 };
// Namespaced so provider addresses don't collide with other localStorage keys.
const AVAILABILITY_KEY_PREFIX = 'provider-availability:';

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
// injects (getProviders, onOpenSession, etc.) — typed loosely as `any` because
// the container builds them dynamically and individual consumers refine them
// in their own prop types.
const withChatState = (WrappedComponent: ComponentType<any>) => {
  class Container extends React.Component<ContainerProps> {
    static contextType = ToastsContext;
    declare context: React.ContextType<typeof ToastsContext>;

    static displayName = `withChatState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    getProviders = async () => {
      try {
        return (await this.props.client.getProviders()) || [];
      } catch (e) {
        console.log('Error', e);
        return [];
      }
    };

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
      try {
        return (await this.props.client.getAllModels()) || [];
      } catch (e) {
        console.log('Error', e);
        return [];
      }
    };

    getLocalModels = async () => {
      try {
        return (await this.props.client.getLocalModels()) || [];
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

    readAvailabilityCache = (address) => {
      try {
        const raw = localStorage.getItem(AVAILABILITY_KEY_PREFIX + address);
        if (!raw) {
          return null;
        }
        const record = JSON.parse(raw);
        const ttl =
          record.status === AvailabilityStatus.available
            ? AVAILABILITY_TTL_MS.available
            : AVAILABILITY_TTL_MS.other;
        if (new Date(record.time).getTime() + ttl < Date.now()) {
          return null;
        }
        return record;
      } catch (e) {
        return null;
      }
    };

    writeAvailabilityCache = (address, record) => {
      try {
        localStorage.setItem(
          AVAILABILITY_KEY_PREFIX + address,
          JSON.stringify({ status: record.status, time: record.time }),
        );
      } catch (e) {
        // Quota exceeded / private mode — availability caching is best-effort.
      }
    };

    getProvidersAvailability = async (providers) => {
      const isValidUrl = (url) => {
        const urlRegex =
          /^(https?:\/\/)?(([a-zA-Z0-9.-]+\.[a-zA-Z]{2,}|localhost)|(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}))(:\d{1,5})?(\/\S*)?$/;
        return urlRegex.test(url);
      };

      // Resolve everything we can from cache first (synchronous, no network),
      // then ping only the genuinely unknown providers — at most
      // AVAILABILITY_CONCURRENCY at a time, each with its own timeout.
      const needsPing: any[] = [];
      const resolved = new Map<string, any>();

      for (const p of providers) {
        const cached = this.readAvailabilityCache(p.Address);
        if (cached) {
          resolved.set(p.Address, { ...cached, id: p.Address });
        } else if (!isValidUrl(p.Endpoint)) {
          const record = {
            id: p.Address,
            status: AvailabilityStatus.disconnected,
            time: new Date(),
          };
          this.writeAvailabilityCache(p.Address, record);
          resolved.set(p.Address, record);
        } else {
          needsPing.push(p);
        }
      }

      const pinged = await pooledMapSettled(
        needsPing,
        async (p: any) => {
          const isValid = await withTimeout(
            this.props.client.checkProviderConnectivity({
              endpoint: p.Endpoint,
              address: p.Address,
            }),
            AVAILABILITY_PING_TIMEOUT_MS,
          );
          const record = {
            id: p.Address,
            status: isValid
              ? AvailabilityStatus.available
              : AvailabilityStatus.disconnected,
            time: new Date(),
          };
          // Cache failures too. The old code only persisted successes, so every
          // load re-pinged (and re-waited on) every unreachable provider.
          this.writeAvailabilityCache(p.Address, record);
          return record;
        },
        (p: any) => {
          const record = {
            id: p.Address,
            status: AvailabilityStatus.unknown,
            time: new Date(),
          };
          this.writeAvailabilityCache(p.Address, record);
          return record;
        },
        AVAILABILITY_CONCURRENCY,
      );

      for (const record of pinged) {
        resolved.set(record.id, record);
      }

      // Preserve the caller's original provider ordering.
      return providers.map((p) => resolved.get(p.Address)).filter(Boolean);
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
      this.context.toast('info', 'Processing...');
      try {
        const failoverSettings = await this.props.client.getFailoverSetting();

        const dataResponse = await this.props.client.openSession({
          modelId,
          failover: failoverSettings?.isEnabled || false,
          duration: +duration,
          directPayment: isDirectPay,
        });
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
