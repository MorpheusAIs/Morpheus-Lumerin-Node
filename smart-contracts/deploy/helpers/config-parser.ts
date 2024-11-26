import { readFileSync } from 'fs';

import { ISessionStorage } from '@/generated-types/ethers';

export type Config = {
  MOR: string;
  fundingAccount: string;
  pools: ISessionStorage.PoolStruct[];
  providerMinStake: string;
  modelMinStake: string;
  marketplaceBidFee: string;
  marketplaceMinBidPricePerSecond: string;
  marketplaceMaxBidPricePerSecond: string;
  delegateRegistry: string;
  owner: string;
};

export function parseConfig(): Config {
  const configPath = `deploy/data/config_arbitrum_mainnet.json`;

  return JSON.parse(readFileSync(configPath, 'utf-8')) as Config;
}
