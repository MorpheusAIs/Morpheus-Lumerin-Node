import { BigNumberish } from 'ethers';
import { ethers } from 'hardhat';

import { LumerinDiamond, ProvidersDelegate } from '@/generated-types/ethers';

export const deployProvidersDelegate = async (
  diamond: LumerinDiamond,
  feeTreasury: string,
  fee: BigNumberish,
  name: string,
  endpoint: string,
  deregistrationOpenAt_: bigint | number,
): Promise<ProvidersDelegate> => {
  const [implFactory, proxyFactory] = await Promise.all([
    ethers.getContractFactory('ProvidersDelegate'),
    ethers.getContractFactory('ERC1967Proxy'),
  ]);

  const impl = await implFactory.deploy();
  const proxy = await proxyFactory.deploy(impl, '0x');
  const contract = implFactory.attach(proxy) as ProvidersDelegate;

  await contract.ProvidersDelegate_init(diamond, feeTreasury, fee, name, endpoint, deregistrationOpenAt_);

  return contract;
};
