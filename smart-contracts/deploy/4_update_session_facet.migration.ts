import { Deployer } from '@solarity/hardhat-migrate';
import { Fragment } from 'ethers';
import { ethers } from 'hardhat';

import { parseConfig } from './helpers/config-parser';

import {
  ISessionRouter__factory,
  IStatsStorage__factory,
  LinearDistributionIntervalDecrease__factory,
  LumerinDiamond__factory,
  SessionRouter__factory,
} from '@/generated-types/ethers';
import { FacetAction } from '@/test/helpers/deployers';

module.exports = async function (deployer: Deployer) {
  const config = parseConfig();

  const ldid = await deployer.deploy(LinearDistributionIntervalDecrease__factory);
  const newSessionRouterFacet = await deployer.deploy(SessionRouter__factory, {
    libraries: {
      LinearDistributionIntervalDecrease: ldid,
    },
  });

  const lumerinDiamond = await deployer.deployed(LumerinDiamond__factory, config.lumerinProtocol);

  // ONLY FOR TESTS
  // const testSigner = await ethers.getImpersonatedSigner(await lumerinDiamond.owner());
  // END

  const oldSessionRouterFacet = '0xCc48cB2DbA21A5D36C16f6f64e5B5E138EA1ba13';
  const oldSelectors = await lumerinDiamond.facetFunctionSelectors(oldSessionRouterFacet);

  // ONLY FOR TESTS - remove or add `.connect(testSigner)`
  await lumerinDiamond['diamondCut((address,uint8,bytes4[])[])']([
    {
      facetAddress: oldSessionRouterFacet,
      action: FacetAction.Remove,
      functionSelectors: [...oldSelectors],
    },
    {
      facetAddress: newSessionRouterFacet,
      action: FacetAction.Add,
      functionSelectors: ISessionRouter__factory.createInterface()
        .fragments.filter(Fragment.isFunction)
        .map((f) => f.selector),
    },
    {
      facetAddress: newSessionRouterFacet,
      action: FacetAction.Add,
      functionSelectors: IStatsStorage__factory.createInterface()
        .fragments.filter(Fragment.isFunction)
        .map((f) => f.selector),
    },
  ]);
};

// npx hardhat migrate --only 4

// npx hardhat migrate --network arbitrum_sepolia --only 4 --verify
// npx hardhat migrate --network arbitrum_sepolia --only 4 --verify --continue

// npx hardhat migrate --network arbitrum --only 4 --verify
// npx hardhat migrate --network arbitrum --only 4 --verify --continue
