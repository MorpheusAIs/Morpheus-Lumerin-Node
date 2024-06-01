import { combineReducers } from 'redux';
import connectivity from './connectivity';
import session from './session';
import chain from './chain';
import config from './config';
import proxyRouter from './proxy-router';
import contracts from './contracts';
import meta from './meta';
import devices from './devices';
import models from './models';

export default combineReducers({
  connectivity,
  session,
  config,
  chain,
  proxyRouter,
  contracts,
  devices,
  models
  // meta
});
