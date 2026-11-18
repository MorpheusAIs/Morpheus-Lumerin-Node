import { Fragment } from 'ethers';
import { ethers } from 'hardhat';

import { IProviderPayouts__factory, LumerinDiamond, ProviderPayouts } from '@/generated-types/ethers';
import { FacetAction } from '@/test/helpers/deployers/diamond/lumerin-diamond';

export const deployFacetProviderPayouts = async (diamond: LumerinDiamond): Promise<ProviderPayouts> => {
  let facet: ProviderPayouts;

  const LDIDFactory = await ethers.getContractFactory('LinearDistributionIntervalDecrease');
  const LDID = await LDIDFactory.deploy();

  const factory = await ethers.getContractFactory('ProviderPayouts', {
    libraries: {
      LinearDistributionIntervalDecrease: LDID,
    },
  });
  facet = await factory.deploy();

  await diamond['diamondCut((address,uint8,bytes4[])[])']([
    {
      facetAddress: facet,
      action: FacetAction.Add,
      functionSelectors: IProviderPayouts__factory.createInterface()
        .fragments.filter(Fragment.isFunction)
        .map((f) => f.selector),
    },
  ]);

  facet = facet.attach(diamond.target) as ProviderPayouts;
  await facet.__ProviderPayouts_init();

  return facet;
};
