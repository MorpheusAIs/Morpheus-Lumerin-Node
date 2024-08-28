import hre from "hardhat";
import { getVar, isAddress } from "../libraries/getConfig";

async function main() {
  const morAddress = await getVar({
    argName: "mor-token-address",
    envName: "MOR_TOKEN_ADDRESS",
    prompt: "Enter MOR token address",
    validator: isAddress,
  });
  const lmrAddress = await getVar({
    argName: "lmr-token-address",
    envName: "LMR_TOKEN_ADDRESS",
    prompt: "Enter LMR token address",
    validator: isAddress,
  });

  console.log("Deploying staking contract ...");
  const staking = await hre.viem.deployContract("StakingMasterChef", [
    lmrAddress,
    morAddress,
  ]);

  console.log("Staking deployed to:", staking.address);

  await new Promise((resolve) => setTimeout(resolve, 10000));

  console.log("Verifying staking contract ...");
  await hre.run("verify", {
    address: staking.address,
    constructorArgsParams: [lmrAddress, morAddress],
  });
  console.log("Staking contract verified");
}

main();
