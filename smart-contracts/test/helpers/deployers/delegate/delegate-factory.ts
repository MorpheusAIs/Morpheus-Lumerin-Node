import { ethers } from 'hardhat';

import { DelegateFactory, LumerinDiamond } from '@/generated-types/ethers';

export const deployDelegateFactory = async (
  diamond: LumerinDiamond,
  minDeregistrationTimeout: number,
): Promise<DelegateFactory> => {
  const [providersDelegateImplFactory, delegateFactoryImplFactory, proxyFactory] = await Promise.all([
    ethers.getContractFactory('ProvidersDelegate'),
    ethers.getContractFactory('DelegateFactory'),
    ethers.getContractFactory('ERC1967Proxy'),
  ]);

  const delegatorFactoryImpl = await delegateFactoryImplFactory.deploy();
  const proxy = await proxyFactory.deploy(delegatorFactoryImpl, '0x');
  const delegatorFactory = delegatorFactoryImpl.attach(proxy) as DelegateFactory;

  const providersDelegateImpl = await providersDelegateImplFactory.deploy();
  await delegatorFactory.DelegateFactory_init(diamond, providersDelegateImpl, minDeregistrationTimeout);

  return delegatorFactory;
};
