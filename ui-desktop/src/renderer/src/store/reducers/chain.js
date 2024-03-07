import { combineReducers } from 'redux';
import get from 'lodash/get';

import wallet from './wallet';
import meta from './meta';

const initialState = {
  wallet: {},
  meta: {}
};

const createChainReducer = chain => (state, action) => {
  // ignore messages specifically intended for other chains
  if (get(action, 'payload.chain') && action.payload.chain !== chain) {
    return state;
  }
  return combineReducers({
    wallet,
    meta
  })(state, action);
};

export default function(state = initialState, action) {
  switch (action.type) {
    // init state from persisted state and config
    case 'initial-state-received': {
      if (!get(action, 'payload.config.chain.chainId', '')) {
        throw new Error(
          'config must contain an "chainId" property with one string value.'
        );
      }

      const activeChain = get(action, 'payload.chain');

      return {
        ...state,
        ...createChainReducer(activeChain)(state, action)
      };
    }

    default: {
      const chain = get(action, 'payload.chain');
      return {
        ...state,
        ...createChainReducer(chain)(state, action)
      };
    }
  }
}
