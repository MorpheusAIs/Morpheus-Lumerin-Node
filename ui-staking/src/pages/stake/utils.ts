import { parseEventLogs } from "viem/utils";
import { stakingMasterChefAbi } from "../../blockchain/abi.ts";
import type { Log } from "viem";

export function getStakeId(logs: Log<bigint, number, false>[], address: `0x${string}`, poolId: bigint): bigint {
  const events = parseEventLogs({
    abi: stakingMasterChefAbi,
    logs,
    eventName: "Stake",
    args: {
      poolId,
      userAddress: address,
    },
  });

  if (events.length === 0) {
    throw new Error("Stake event not found");
  }
  if (events.length > 1) {
    throw new Error("Multiple Stake events found");
  }
  return events[0].args.stakeId;
}
