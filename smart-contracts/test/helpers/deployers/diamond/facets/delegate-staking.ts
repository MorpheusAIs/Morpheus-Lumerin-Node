import { Fragment } from 'ethers';
import { ethers } from 'hardhat';

import {
  DelegateStaking,
  DelegateStakingFunding,
  IDelegateStakingFunding__factory,
  IDelegateStaking__factory,
  LumerinDiamond,
} from '@/generated-types/ethers';
import { FacetAction } from '@/test/helpers/deployers/diamond/lumerin-diamond';

export const deployFacetDelegateStaking = async (diamond: LumerinDiamond): Promise<DelegateStaking> => {
  let facet: DelegateStaking;

  const LDIDFactory = await ethers.getContractFactory('LinearDistributionIntervalDecrease');
  const LDID = await LDIDFactory.deploy();

  const factory = await ethers.getContractFactory('DelegateStaking', {
    libraries: {
      LinearDistributionIntervalDecrease: LDID,
    },
  });
  facet = await factory.deploy();

  await diamond['diamondCut((address,uint8,bytes4[])[])']([
    {
      facetAddress: facet,
      action: FacetAction.Add,
      functionSelectors: IDelegateStaking__factory.createInterface()
        .fragments.filter(Fragment.isFunction)
        .map((f) => f.selector),
    },
  ]);

  return facet.attach(diamond.target) as DelegateStaking;
};

export const deployFacetDelegateStakingFunding = async (diamond: LumerinDiamond): Promise<DelegateStakingFunding> => {
  let facet: DelegateStakingFunding;

  const factory = await ethers.getContractFactory('DelegateStakingFunding');
  facet = await factory.deploy();

  await diamond['diamondCut((address,uint8,bytes4[])[])']([
    {
      facetAddress: facet,
      action: FacetAction.Add,
      functionSelectors: IDelegateStakingFunding__factory.createInterface()
        .fragments.filter(Fragment.isFunction)
        .map((f) => f.selector),
    },
  ]);

  return facet.attach(diamond.target) as DelegateStakingFunding;
};
