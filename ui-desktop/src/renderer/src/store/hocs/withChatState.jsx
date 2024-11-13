import { withClient } from './clientContext';
import { connect } from 'react-redux';
import React from 'react';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';
import axios from 'axios';

const AvailabilityStatus = {
  available: "available",
  unknown: "unknown",
  disconnected: "disconnected"
}

const withChatState = WrappedComponent => {
  class Container extends React.Component {

    static contextType = ToastsContext;

    static displayName = `withChatState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    getBidsByModels = async (modelId) => {
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/models/${modelId}/bids`
        const response = await fetch(path);
        const data = await response.json();
        if (data.error) {
          console.error(data.error);
          return [];
        }
        return data.bids;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    }

    getProviders = async () => {
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/providers`
        const response = await fetch(path);
        const data = await response.json();
        if (data.error) {
          console.error(data.error);
          return [];
        }
        return data.providers;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    }

    closeSession = async (sessionId) => {
      this.context.toast('info', 'Closing...');
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/sessions/${sessionId}/close`;
        const response = await fetch(path, {
          method: "POST"
        });
        const data = await response.json();
        if (data.error) {
          this.context.toast('error', 'Session not closed');
          throw new Error(data.error);
        }
        if(data.tx) {
          this.context.toast('success', 'Session successfully closed');
        }
      }
      catch (e) {
        console.log("Error", e)
        this.context.toast('error', 'Failed to close session');
        return [];
      }
    }

    getBidsRatingByModel = async (modelId) => {
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/models/${modelId}/bids/rated`;
        const response = await fetch(path);
        const data = await response.json();
        if (data.error) {
          console.error(data.error);
          return [];
        }
        return data.bids;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    }

    getAllModels = async () => {
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/models`;
        const response = await fetch(path);
        const data = await response.json();
        if (data.error) {
          console.error(data.error);
          return [];
        }
        return data.models;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    }

    getLocalModels = async () => {
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/v1/models`;
        const response = await fetch(path);
        if (!response.ok) {
          return [];
        }
        return await response.json();
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    }

    getModelsData = async () => {
      const [localModels, modelsResp, providersResp] = await Promise.all([
        this.getLocalModels(),
        this.getAllModels(),
        this.getProviders()]);

      const models = modelsResp.filter(m => !m.IsDeleted);
      const providers = providersResp.filter(m => !m.IsDeleted);

      const availabilityResults = await this.getProvidersAvailability(providers);
      availabilityResults.forEach(ar => {
        const provider = providers.find(p => p.Address == ar.id);
        if(!provider)
          return;

        provider.availabilityStatus = ar.status;
        provider.availabilityUpdatedAt = ar.time;
      });

      const providersMap = providers.reduce((a, b) => ({ ...a, [b.Address.toLowerCase()]: b }), {});

      const responses = (await Promise.all(
        models.map(async m => {
          const id = m.Id;
          const bids = (await this.getBidsByModels(id))
            .filter(b => +b.DeletedAt === 0)
            .map(b => ({ ...b, ProviderData: providersMap[b.Provider.toLowerCase()], Model: m }))
            .filter(b => b.ProviderData);
          return { id, bids }
        })
      )).reduce((a,b) => ({...a, [b.id]: b.bids}), {});

      const result = [];

      for (const model of models) {
        const id = model.Id;
        const bids = responses[id];
        
        const localModel = localModels.find(lm => lm.Id == id);

        result.push({ ...model, bids, hasLocal: Boolean(localModel) })
      }

      return { models: result.filter(r => r.bids.length || r.hasLocal), providers }
    }

    getProvidersAvailability = async (providers) => {
      const availabilityResults = await Promise.all(providers.map(async p => {
        try {
          const storedRecord = JSON.parse(localStorage.getItem(p.Address));
          if(storedRecord && storedRecord.status == AvailabilityStatus.available) {
            const lastUpdatedAt = new Date(storedRecord.time);
            const cacheMinutes = 15;
            const timestampBefore = new Date(new Date().getTime() - (cacheMinutes * 60 * 1000));

            if(lastUpdatedAt > timestampBefore) {
              return ({...storedRecord, id: p.Address});
            }
          }

          const endpoint = p.Endpoint;
          const [domain, port] = endpoint.split(":");
          const { data } = await axios.post("https://portchecker.io/api/v1/query", {
            host: domain,
            ports: [port],
          });
    
          const isValid = !!data.check?.find((c) => c.port == port && c.status == true);
          const record = ({id: p.Address, status: isValid ? AvailabilityStatus.available : AvailabilityStatus.disconnected, time: new Date() });
          localStorage.setItem(record.id, JSON.stringify({ status: record.status, time: record.time }));
          return record;
        }
        catch(e) {
          return ({id: p.Address, status: AvailabilityStatus.unknown, time: new Date() })
        }
      }));
      return availabilityResults;
    }

    getMetaInfo = async () => {
      const [budget, supply] = await Promise.all([
        this.props.client.getTodaysBudget(),
        this.props.client.getTokenSupply()]);
      return { budget, supply };
    }

    getSessionsByUser = async (user) => {
      if(!user) {
        return;
      }
      try {
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/sessions/user?user=${user}`;
        const response = await fetch(path);
        const data = await response.json();
        return data.sessions;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    }

    onOpenSession = async ({ modelId, duration }) => {
      this.context.toast('info', 'Processing...');
      try {
        const failoverSettings = await this.props.client.getFailoverSetting();
        
        const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/models/${modelId}/session`;
        const body = {
          failover: failoverSettings?.isEnabled || false,
          sessionDuration: +duration // convert to seconds
        };
        const response = await fetch(path, {
          method: "POST",
          body: JSON.stringify(body)
        });
        const dataResponse = await response.json();
        if (!response.ok) {
          this.context.toast('error', 'Failed to open session');
          console.log("Failed initiate session", dataResponse);
          return;
        }
        this.context.toast('success', 'Session successfully created');
        return dataResponse.sessionID;
      }
      catch (e) {
        console.error(e);
        this.context.toast('error', 'Failed to open session');
        return;
      }
    }

    getBalances = async () => {
      return await this.props.client.getBalances();
    }

    render() {

      return (
        <WrappedComponent
          getProviders={this.getProviders}
          getBidsByModels={this.getBidsByModels}
          getMetaInfo={this.getMetaInfo}
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
  });

  const mapDispatchToProps = dispatch => ({
    setBid: model => dispatch({ type: 'set-bid', payload: model })
  });

  return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withChatState;
