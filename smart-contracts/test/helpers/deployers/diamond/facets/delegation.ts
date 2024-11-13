import { Fragment } from 'ethers';
import { ethers } from 'hardhat';

import { Delegation, IDelegation__factory, LumerinDiamond } from '@/generated-types/ethers';
import { DelegateRegistry } from '@/generated-types/ethers/contracts/mock/delegate-registry/src';
import { FacetAction } from '@/test/helpers/deployers/diamond/lumerin-diamond';

export const deployFacetDelegation = async (
  diamond: LumerinDiamond,
  delegateRegistry: DelegateRegistry,
): Promise<Delegation> => {
  let facet: Delegation;

  const factory = await ethers.getContractFactory('Delegation');
  facet = await factory.deploy();

  await diamond['diamondCut((address,uint8,bytes4[])[])']([
    {
      facetAddress: facet,
      action: FacetAction.Add,
      functionSelectors: IDelegation__factory.createInterface()
        .fragments.filter(Fragment.isFunction)
        .map((f) => f.selector),
    },
  ]);

  facet = facet.attach(diamond.target) as Delegation;
  await facet.__Delegation_init(delegateRegistry);

  return facet;
};
