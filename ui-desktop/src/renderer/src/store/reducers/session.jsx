import { handleActions } from 'redux-actions';

const initialState = {
  hasEnoughData: false,
  isLoggedIn: false
};

const reducer = handleActions(
  {
    'session-started': state => ({
      ...state,
      isLoggedIn: true
    }),
    'required-data-gathered': state => ({
      ...state,
      hasEnoughData: true
    })
  },
  initialState
);

export default reducer;
