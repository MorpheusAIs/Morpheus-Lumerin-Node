import hre from "hardhat";
import { parseEventLogs } from "viem";
import { getTxTimestamp } from "../utils";

export async function getPoolId(poolTx: `0x${string}`) {
  const publicClient = await hre.viem.getPublicClient();
  const receipt = await publicClient.waitForTransactionReceipt({
    hash: poolTx,
  });
  const artifact = await hre.artifacts.readArtifact("StakingMasterChef");
  const events = parseEventLogs({
    abi: artifact.abi,
    logs: receipt.logs,
    eventName: "PoolAdded",
  });

  if (events.length === 0) {
    throw new Error("AddPool event not found");
  }
  if (events.length > 1) {
    throw new Error("Multiple AddPool events found");
  }
  return events[0].args.poolId;
}

export async function getStakeId(stakeTx: `0x${string}`) {
  const publicClient = await hre.viem.getPublicClient();

  const receipt = await publicClient.getTransactionReceipt({ hash: stakeTx });
  const artifact = await hre.artifacts.readArtifact("StakingMasterChef");
  const events = parseEventLogs({
    abi: artifact.abi,
    logs: receipt.logs,
    eventName: "Stake",
  });

  if (events.length === 0) {
    throw new Error("Stake event not found");
  }
  if (events.length > 1) {
    throw new Error("Multiple Stake events found");
  }
  return events[0].args.stakeId;
}

/** Elapsed time between two transactions */
export async function elapsedTxs(
  tx1: `0x${string}`,
  tx2: `0x${string}`,
): Promise<bigint> {
  const pc = await hre.viem.getPublicClient();
  return (await getTxTimestamp(pc, tx2)) - (await getTxTimestamp(pc, tx1));
}
