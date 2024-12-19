import { Deployer } from '@solarity/hardhat-migrate';

import { parseConfig } from './helpers/config-parser';

import { Marketplace__factory } from '@/generated-types/ethers';

module.exports = async function (deployer: Deployer) {
  const config = parseConfig();

  const marketplaceFacet = await deployer.deployed(Marketplace__factory, config.lumerinProtocol);

  console.log(await marketplaceFacet.getMinMaxBidPricePerSecond());

  // await marketplaceFacet.setMinMaxBidPricePerSecond('10000000000', '10000000000000000');
};

// npx hardhat migrate --only 2

// npx hardhat migrate --network arbitrum_sepolia --only 2 --verify
// npx hardhat migrate --network arbitrum_sepolia --only 2 --verify --continue

// npx hardhat migrate --network arbitrum --only 1 --verify
// npx hardhat migrate --network arbitrum --only 1 --verify --continue
