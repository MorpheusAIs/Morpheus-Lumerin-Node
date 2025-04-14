import { handleActions } from 'redux-actions';
import { LoadingState } from 'src/main/orchestrator.types';

// TODO: update only changed entries in the arrays
const initialState: LoadingState = {
  download: [],
  startup: [],
};

const reducer = handleActions(
  {
    'services-state': (state, { payload }) => ({
      ...state,
      ...payload,
    }),
  },
  initialState,
);

export default reducer;
