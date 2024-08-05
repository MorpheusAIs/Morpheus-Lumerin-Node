import { getVar, isAddress, isBigInt } from "../libraries/getConfig";
import hre from "hardhat";
import { getDefaultDurations } from "../test/Staking/fixtures";
import { getPoolId } from "../test/Staking/utils";

async function main() {
  const morAddress = await getVar({
    argName: "mor-token-address",
    envName: "MOR_TOKEN_ADDRESS",
    prompt: "Enter MOR token address",
    validator: isAddress,
  });

  const stakingAddress = await getVar({
    argName: "staking-contract-address",
    envName: "STAKING_CONTRACT_ADDRESS",
    prompt: "Enter staking contract address",
    validator: isAddress,
  });

  const startDate = await getVar({
    argName: "start-date",
    prompt: "Enter pool start date",
    validator: isBigInt,
  });

  const duration = await getVar({
    argName: "duration",
    prompt: "Enter pool duration",
    validator: isBigInt,
  });

  const totalReward = await getVar({
    argName: "total-reward",
    prompt: "Enter total reward",
    validator: isBigInt,
  });

  const staking = await hre.viem.getContractAt(
    "StakingMasterChef",
    stakingAddress,
  );
  const morToken = await hre.viem.getContractAt("MorpheusToken", morAddress);
  await morToken.write.approve([stakingAddress, totalReward]);

  const precision = await staking.read.PRECISION();
  const defaultDurations = getDefaultDurations(precision);

  console.log("Adding pool ...");
  const tx = await staking.write.addPool([
    startDate,
    duration,
    totalReward,
    defaultDurations,
  ]);

  const poolId = await getPoolId(tx);
  console.log("Pool added with id:", poolId);
}

main();
