import { handleActions } from 'redux-actions';
import get from 'lodash/get';

export const initialState = {
  bestBlockTimestamp: null,
  isIndexerConnected: null,
  isWeb3Connected: null,
  rateLastUpdated: null,
  gasPrice: null,
  height: -1,
  rate: null,
  rateEth: null,
  rateBtc: null,
  networkDifficulty: null
};

const reducer = handleActions(
  {
    'initial-state-received': (state, { payload }) => ({
      ...state,
      ...get(payload, 'meta', initialState),
      isConnected: null, // ignore web3 connection status persisted state
      gasPrice: get(payload, 'meta.gasPrice', payload.config.defaultGasPrice)
    }),

    'indexer-connection-status-changed': (state, { payload }) => ({
      ...state,
      isIndexerConnected: payload.connected
    }),

    'web3-connection-status-changed': (state, { payload }) => ({
      ...state,
      isWeb3Connected: payload.connected
    }),

    'coin-block': (state, { payload }) => ({
      ...state,
      bestBlockTimestamp: payload.timestamp,
      height: payload.number
    }),

    'coin-price-updated': (state, { payload }) => {
      if (payload.token === 'LMR') {
        return {
          ...state,
          rateLastUpdated: parseInt(Date.now() / 1000, 10),
          rate: payload.price
        };
      }

      if (payload.token === 'ETH') {
        return {
          ...state,
          rateEth: payload.price
        };
      }

      if (payload.token === 'BTC') {
        return {
          ...state,
          rateBtc: payload.price
        };
      }
    },

    'network-difficulty-updated': (state, { payload }) => {
      return {
        ...state,
        networkDifficulty: payload.difficulty
      };
    },

    'gas-price-updated': (state, { payload }) => ({
      ...state,
      gasPrice: payload
    }),

    'blockchain-set': (state, { payload }) => ({
      ...state,
      ...payload
    })
  },
  initialState
);

export default reducer;
