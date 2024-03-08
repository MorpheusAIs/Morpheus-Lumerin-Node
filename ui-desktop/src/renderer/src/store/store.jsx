import { createStore as reduxCreateStore } from 'redux';

import reducer from './reducers';

export { Provider } from 'react-redux';

export function createStore(reduxDevtoolsOptions, initialState) {
  const enhancers =
    typeof window !== 'undefined' &&
    window.__REDUX_DEVTOOLS_EXTENSION__ &&
    window.__REDUX_DEVTOOLS_EXTENSION__(reduxDevtoolsOptions);

  return reduxCreateStore(reducer, initialState, enhancers);
}
