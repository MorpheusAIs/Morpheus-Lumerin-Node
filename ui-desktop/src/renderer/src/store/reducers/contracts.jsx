import { handleActions } from 'redux-actions';
import get from 'lodash/get';
import { v4 as uuidv4 } from 'uuid';
import { keyBy } from 'lodash';

// TODO: remove dummy data
const initialState = {
  lastUpdated: null,
  syncStatus: null,
  drafts: [],
  actives: {}
};

const ZERO_ADDRESS = '0x0000000000000000000000000000000000000000';

const reducer = handleActions(
  {
    'initial-state-received': (state, { payload }) => ({
      ...state,
      ...get(payload, 'contracts', {})
    }),

    'contracts-scan-started': state => ({
      ...state,
      lastUpdated: parseInt(Date.now() / 1000, 10),
      syncStatus: 'syncing'
    }),

    'contracts-scan-failed': (state, { payload }) => ({
      ...state,
      actives: {},
      lastUpdated: parseInt(Date.now() / 1000, 10),
      syncStatus: 'failed'
    }),

    'contracts-scan-finished': (state, { payload }) => {
      const idContractMap = keyBy(payload.actives, 'id');

      return {
        ...state,
        actives: { ...state.actives, ...idContractMap },
        lastUpdated: parseInt(Date.now() / 1000, 10),
        syncStatus: 'up-to-date'
      };
    },

    'contract-updated': (state, { payload }) => {
      const idContractMap = keyBy(payload.actives, 'id');

      return {
        ...state,
        actives: { ...state.actives, ...idContractMap },
        lastUpdated: parseInt(Date.now() / 1000, 10)
      };
    },

    'remove-draft': (state, { payload }) => ({
      ...state,
      drafts: Object.assign(state.drafts, []).filter(
        draft => draft.id !== payload.id
      ),
      lastUpdated: parseInt(Date.now() / 1000, 10),
      syncStatus: payload.syncStatus
    }),

    'create-draft': (state, { payload }) => {
      const draftWithId = { ...payload.draft, id: uuidv4() };

      return {
        ...state,
        drafts: [...state.drafts, draftWithId],
        lastUpdated: parseInt(Date.now() / 1000, 10),
        syncStatus: payload.syncStatus
      };
    },

    'purchase-temp-contract': (state, { payload }) => {
      const contract = {
        ...state.actives[payload.id],
        inProgress: true,
        buyer: payload.address,
        timestamp: parseInt(Date.now() / 1000, 10)
      };

      return {
        ...state,
        actives: {
          ...state.actives,
          [contract.id]: contract
        }
      };
    },

    'purchase-contract-success': (state, { payload }) => {
      const contract = {
        ...state.actives[payload.id],
        inProgress: false
      };

      return {
        ...state,
        actives: {
          ...state.actives,
          [contract.id]: contract
        }
      };
    },

    'purchase-contract-failed': (state, { payload }) => {
      const contract = {
        ...state.actives[payload.id],
        inProgress: false,
        buyer: ZERO_ADDRESS
      };

      return {
        ...state,
        actives: {
          ...state.actives,
          [contract.id]: contract
        }
      };
    },

    'create-temp-contract': (state, { payload }) => {
      return {
        ...state,
        actives: {
          ...state.actives,
          [payload.id]: payload
        }
      };
    },

    'remove-contract': (state, { payload }) => {
      const { [payload.id]: _, ...filtered } = state.actives;
      return {
        ...state,
        actives: filtered
      };
    },

    'edit-contract-state': (state, { payload }) => {
      const { [payload.id]: contract, ...filtered } = state.actives;
      return {
        ...state,
        actives: {
          ...filtered,
          [payload.id]: {
            ...contract,
            ...payload
          }
        }
      };
    }
  },
  initialState
);

export default reducer;
