import * as validators from '../validators';
import { withClient } from './clientContext';
import * as utils from '../utils';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import React from 'react';
import { ToastsContext } from '../../components/toasts';
import selectors from '../selectors';

const withBidsState = WrappedComponent => {
    class Container extends React.Component {

        static contextType = ToastsContext;

        static displayName = `withBidsState(${WrappedComponent.displayName ||
            WrappedComponent.name})`;

        getBitsByModels = async (modelId) => {
            try {
                const authHeaders = await this.props.client.getAuthHeaders();
                const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/models/${modelId}/bids`
                const response = await fetch(path, {
                    headers: authHeaders
                });
                const data = await response.json();
                return data.bids;
            }
            catch (e) {
                console.log("Error", e)
                return [];
            }
        }

        getProviders = async () => {
            try {
                const authHeaders = await this.props.client.getAuthHeaders();
                const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/providers`
                const response = await fetch(path, {
                    headers: authHeaders
                });
                const data = await response.json();
                return data.providers;
            }
            catch (e) {
                console.log("Error", e)
                return [];
            }
        }

        setBid = async (data) => {
            let signature = '';
            let sessionId = '';

            const authHeaders = await this.props.client.getAuthHeaders();
            try {
                const path = `${this.props.config.chain.localProxyRouterUrl}/proxy/sessions/initiate`;
                const body = {
                    user: this.props.address,
                    provider: data.provider.Address,
                    spend: 10,
                    providerUrl: data.provider.Endpoint.replace("http://", "")
                };
                const response = await fetch(path, {
                    method: "POST",
                    body: JSON.stringify(body),
                    headers: authHeaders
                });
                const dataResponse = await response.json();
                signature = dataResponse.response.result.message;
            }
            catch (e) {
                console.log("Error", e)
                return false;
            }

            try {
                const path = `${this.props.config.chain.localProxyRouterUrl}/blockchain/sessions`;
                const body = {
                    bidId: data.bidId,
                    stake: "380925328193836900000", // TEMP ISSUE SOLVER
                };
                const response = await fetch(path, {
                    method: "POST",
                    body: JSON.stringify(body),
                    headers: authHeaders
                });
                const dataResponse = await response.json();
                sessionId = dataResponse.sessionId;
            }
            catch (e) {
                console.log("Error", e)
                return false;
            }

            this.props.setBidState(data);
            this.props.setActiveSession({ sessionId: sessionId, signature: signature });
            return true;
        }

        render() {

            return (
                <WrappedComponent
                    getProviders={this.getProviders}
                    getBitsByModels={this.getBitsByModels}
                    setBid={this.setBid}
                    {...this.state}
                    {...this.props}
                />
            );
        }
    }

    const mapStateToProps = (state, props) => ({
        address: selectors.getWalletAddress(state),
        config: state.config,
        selectedModel: state.models.selectedModel
    });

    const mapDispatchToProps = dispatch => ({
        setBidState: model => dispatch({ type: 'set-bid', payload: model }),
        setActiveSession: sessionModel => dispatch({ type: 'set-active-session', payload: sessionModel })
    });

    return withClient(connect(mapStateToProps, mapDispatchToProps)(Container));
};

export default withBidsState;
