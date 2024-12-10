import { ethers } from 'hardhat';

import { DelegatorFactory, LumerinDiamond } from '@/generated-types/ethers';

export const deployDelegatorFactory = async (diamond: LumerinDiamond): Promise<DelegatorFactory> => {
  const [providersDelegatorImplFactory, delegatorFactoryImplFactory, proxyFactory] = await Promise.all([
    ethers.getContractFactory('ProvidersDelegator'),
    ethers.getContractFactory('DelegatorFactory'),
    ethers.getContractFactory('ERC1967Proxy'),
  ]);

  const delegatorFactoryImpl = await delegatorFactoryImplFactory.deploy();
  const proxy = await proxyFactory.deploy(delegatorFactoryImpl, '0x');
  const delegatorFactory = delegatorFactoryImpl.attach(proxy) as DelegatorFactory;

  const providersDelegatorImpl = await providersDelegatorImplFactory.deploy();
  await delegatorFactory.DelegatorFactory_init(diamond, providersDelegatorImpl);

  return delegatorFactory;
};
