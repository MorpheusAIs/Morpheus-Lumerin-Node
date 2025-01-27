import { Deployer } from '@solarity/hardhat-migrate';

import { parseConfig } from './helpers/config-parser';

import {
  DelegateFactory__factory,
  ERC1967Proxy__factory,
  ProvidersDelegate__factory,
} from '@/generated-types/ethers';
import { wei } from '@/scripts/utils/utils';

module.exports = async function (deployer: Deployer) {
  const config = parseConfig();

  const providersDelegatorImpl = await deployer.deploy(ProvidersDelegate__factory);
  const delegatorFactoryImpl = await deployer.deploy(DelegateFactory__factory);
  const proxy = await deployer.deploy(ERC1967Proxy__factory, [await delegatorFactoryImpl.getAddress(), '0x']);

  const delegatorFactory = await deployer.deployed(DelegateFactory__factory, await proxy.getAddress());

  await delegatorFactory.DelegateFactory_init(config.lumerinProtocol, providersDelegatorImpl, 60 * 60);

  await delegatorFactory.deployProxy(
    '0x19ec1E4b714990620edf41fE28e9a1552953a7F4',
    wei(0.2, 25),
    'First Subnet',
    'Custom endpoint',
    Math.floor((Date.now() / 1000)) + 24 * 60 * 60,
  );
};

// npx hardhat migrate --only 3

// npx hardhat migrate --network arbitrum_sepolia --only 3 --verify
// npx hardhat migrate --network arbitrum_sepolia --only 3 --verify --continue

// npx hardhat migrate --network arbitrum --only 3 --verify
// npx hardhat migrate --network arbitrum --only 3 --verify --continue
