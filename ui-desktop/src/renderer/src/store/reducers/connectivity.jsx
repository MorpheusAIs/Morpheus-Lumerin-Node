import { handleActions, combineActions } from 'redux-actions';

const initialState = {
  isOnline: true
};

// Set connectivity status to 'online' if any of these actions is received
const CONNECTIVITY_PROOF_ACTIONS = [
  'wallet-state-changed',
  'coin-price-updated',
  'coin-block'
];

const reducer = handleActions(
  {
    'connectivity-state-changed': (state, action) => ({
      ...state,
      isOnline: Boolean(action.payload.ok)
    }),
    [combineActions(...CONNECTIVITY_PROOF_ACTIONS)]: state => ({
      ...state,
      isOnline: true
    })
  },
  initialState
);

export default reducer;
