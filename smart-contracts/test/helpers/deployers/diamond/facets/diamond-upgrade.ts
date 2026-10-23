import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { Fragment, Interface } from 'ethers';
import { ethers } from 'hardhat';

import {
  DelegateStaking,
  DelegateStakingFunding,
  IDelegateStakingFunding__factory,
  IDelegateStaking__factory,
  IProviderPayouts__factory,
  ISessionRouterLegacy__factory,
  ISessionRouter__factory,
  IStatsStorage__factory,
  LumerinDiamond,
  ProviderPayouts,
  SessionRouter,
  SessionRouterLegacy,
} from '@/generated-types/ethers';
import { FacetAction } from '@/test/helpers/deployers/diamond/lumerin-diamond';
import { getDefaultPools } from '@/test/helpers/pool-helper';
import { DAY } from '@/utils/time';

export const getSelectors = (contractInterface: Interface): string[] =>
  contractInterface.fragments.filter(Fragment.isFunction).map((f) => f.selector);

// Pre-upgrade SessionRouter (copy of HEAD in contracts/mock/legacy), cut the same way as before the upgrade.
export const deployFacetSessionRouterLegacy = async (
  diamond: LumerinDiamond,
  fundingAccount: SignerWithAddress,
): Promise<SessionRouterLegacy> => {
  let facet: SessionRouterLegacy;

  const LDIDFactory = await ethers.getContractFactory('LinearDistributionIntervalDecrease');
  const LDID = await LDIDFactory.deploy();

  const factory = await ethers.getContractFactory('SessionRouterLegacy', {
    libraries: {
      LinearDistributionIntervalDecrease: LDID,
    },
  });
  facet = await factory.deploy();

  await diamond['diamondCut((address,uint8,bytes4[])[])']([
    {
      facetAddress: facet,
      action: FacetAction.Add,
      functionSelectors: getSelectors(ISessionRouterLegacy__factory.createInterface()),
    },
    {
      facetAddress: facet,
      action: FacetAction.Add,
      functionSelectors: getSelectors(IStatsStorage__factory.createInterface()),
    },
  ]);

  facet = facet.attach(diamond.target) as SessionRouterLegacy;
  await facet.__SessionRouter_init(fundingAccount, DAY, getDefaultPools());

  return facet;
};

export interface UpgradedFacets {
  sessionRouter: SessionRouter;
  providerPayouts: ProviderPayouts;
  delegateStaking: DelegateStaking;
  delegateStakingFunding: DelegateStakingFunding;
  facetAddresses: {
    sessionRouter: string;
    providerPayouts: string;
    delegateStaking: string;
    delegateStakingFunding: string;
  };
}

// Forward-only upgrade cut over a diamond that runs the legacy SessionRouter.
export const upgradeDiamond = async (diamond: LumerinDiamond): Promise<UpgradedFacets> => {
  const LDIDFactory = await ethers.getContractFactory('LinearDistributionIntervalDecrease');
  const LDID = await LDIDFactory.deploy();
  const libraries = { LinearDistributionIntervalDecrease: LDID };

  const sessionRouter = await (await ethers.getContractFactory('SessionRouter', { libraries })).deploy();
  const providerPayouts = await (await ethers.getContractFactory('ProviderPayouts', { libraries })).deploy();
  const delegateStaking = await (await ethers.getContractFactory('DelegateStaking', { libraries })).deploy();
  const delegateStakingFunding = await (await ethers.getContractFactory('DelegateStakingFunding')).deploy();

  const legacySelectors = getSelectors(ISessionRouterLegacy__factory.createInterface());
  const routerSelectors = [
    ...getSelectors(ISessionRouter__factory.createInterface()),
    ...getSelectors(IStatsStorage__factory.createInterface()),
  ];
  const payoutSelectors = getSelectors(IProviderPayouts__factory.createInterface());
  const movedSelectors = payoutSelectors.filter((s) => legacySelectors.includes(s));
  const newPayoutSelectors = payoutSelectors.filter((s) => !legacySelectors.includes(s));

  await diamond['diamondCut((address,uint8,bytes4[])[],address,bytes)'](
    [
      { facetAddress: sessionRouter, action: FacetAction.Replace, functionSelectors: routerSelectors },
      { facetAddress: providerPayouts, action: FacetAction.Replace, functionSelectors: movedSelectors },
      { facetAddress: providerPayouts, action: FacetAction.Add, functionSelectors: newPayoutSelectors },
      {
        facetAddress: delegateStaking,
        action: FacetAction.Add,
        functionSelectors: getSelectors(IDelegateStaking__factory.createInterface()),
      },
      {
        facetAddress: delegateStakingFunding,
        action: FacetAction.Add,
        functionSelectors: getSelectors(IDelegateStakingFunding__factory.createInterface()),
      },
    ],
    providerPayouts,
    providerPayouts.interface.encodeFunctionData('__ProviderPayouts_init'),
  );

  return {
    sessionRouter: sessionRouter.attach(diamond.target) as SessionRouter,
    providerPayouts: providerPayouts.attach(diamond.target) as ProviderPayouts,
    delegateStaking: delegateStaking.attach(diamond.target) as DelegateStaking,
    delegateStakingFunding: delegateStakingFunding.attach(diamond.target) as DelegateStakingFunding,
    facetAddresses: {
      sessionRouter: await sessionRouter.getAddress(),
      providerPayouts: await providerPayouts.getAddress(),
      delegateStaking: await delegateStaking.getAddress(),
      delegateStakingFunding: await delegateStakingFunding.getAddress(),
    },
  };
};
