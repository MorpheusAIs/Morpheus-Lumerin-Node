import { BigNumberish } from 'ethers';
import { ethers } from 'hardhat';

import { LumerinDiamond, ProvidersDelegator } from '@/generated-types/ethers';

export const deployProvidersDelegator = async (
  diamond: LumerinDiamond,
  feeTreasury: string,
  fee: BigNumberish,
  name: string,
  endpoint: string,
): Promise<ProvidersDelegator> => {
  const [implFactory, proxyFactory] = await Promise.all([
    ethers.getContractFactory('ProvidersDelegator'),
    ethers.getContractFactory('ERC1967Proxy'),
  ]);

  const impl = await implFactory.deploy();
  const proxy = await proxyFactory.deploy(impl, '0x');
  const contract = implFactory.attach(proxy) as ProvidersDelegator;

  await contract.ProvidersDelegator_init(diamond, feeTreasury, fee, name, endpoint);

  return contract;
};
