import { ethers } from 'hardhat';

import { DelegateRegistry } from '@/generated-types/ethers';

export const deployDelegateRegistry = async (): Promise<DelegateRegistry> => {
  const factory = await ethers.getContractFactory('DelegateRegistry');
  const contract = await factory.deploy();

  return contract;
};
