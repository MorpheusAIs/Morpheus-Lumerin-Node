import { Deployer } from '@solarity/hardhat-migrate';

import { Marketplace__factory } from '@/generated-types/ethers';

module.exports = async function (deployer: Deployer) {
  // const marketplaceFacet = await deployer.deployed(Marketplace__factory, '0xb8C55cD613af947E73E262F0d3C54b7211Af16CF');
  const marketplaceFacet = await deployer.deployed(Marketplace__factory, '0xDE819AaEE474626E3f34Ef0263373357e5a6C71b');

  console.log(await marketplaceFacet.getMinMaxBidPricePerSecond());

  // await marketplaceFacet.setMinMaxBidPricePerSecond('10000000000', '10000000000000000');
};

// npx hardhat migrate --only 2

// npx hardhat migrate --network arbitrum_sepolia --only 2 --verify
// npx hardhat migrate --network arbitrum_sepolia --only 2 --verify --continue

// npx hardhat migrate --network arbitrum --only 1 --verify
// npx hardhat migrate --network arbitrum --only 1 --verify --continue
