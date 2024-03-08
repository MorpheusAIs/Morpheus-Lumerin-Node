import * as chain from './chain';
import * as config from './config';
import * as connectivity from './connectivity';
import * as contracts from './contracts';
import * as session from './session';
import * as proxyRouter from './proxy-router';
import * as wallet from './wallet';
import * as devices from './devices';

export default {
  ...connectivity,
  ...session,
  ...config,
  ...chain,
  ...wallet,
  ...proxyRouter,
  ...contracts,
  ...devices
};
