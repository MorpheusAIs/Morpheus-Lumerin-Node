import { Deployer, Reporter } from '@solarity/hardhat-migrate';

import { LinearDistributionIntervalDecrease__factory, SessionRouter__factory } from '@/generated-types/ethers';

// Step 1 of the two-step, owner-signed SessionRouter day-lock upgrade (#830).
// Deploys and verifies new facet bytecode from ANY funded EOA — facet deploy is
// permissionless; only the diamondCut in step 2 is signed by the diamond owner
// (EOA on BASE Sepolia, Safe on mainnet).
//
// Performs NO diamondCut and touches NO live state.
// After this completes, generate the owner's cut payload with:
//   NEW_SESSION_ROUTER_FACET=0x... \
//     npx hardhat run scripts/session-day-lock-cut-calldata.ts --network <network>
//
// See smart-contracts/docs/session-day-lock-upgrade-runbook.md
module.exports = async function (deployer: Deployer) {
  const ldid = await deployer.deploy(LinearDistributionIntervalDecrease__factory);
  const newSessionRouterFacet = await deployer.deploy(SessionRouter__factory, {
    libraries: {
      LinearDistributionIntervalDecrease: ldid,
    },
  });

  Reporter.reportContracts(
    ['LinearDistributionIntervalDecrease (library)', await ldid.getAddress()],
    ['SessionRouter Facet (new)', await newSessionRouterFacet.getAddress()],
  );
};

// npx hardhat migrate --network base_sepolia --only 5 --verify
// npx hardhat migrate --network base --only 5 --verify
