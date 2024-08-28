import { useAccount, usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { erc20Abi, stakingMasterChefAbi } from "../../blockchain/abi.ts";
import { useNavigate, useParams } from "react-router-dom";
import { useState } from "react";
import { getStakeId } from "./utils.ts";

export function useStake(onStakeCb?: (id: bigint) => void) {
  const { poolId: poolIdString } = useParams();
  const { address } = useAccount();
  const poolId = Number(poolIdString);
  const navigate = useNavigate();

  const [lockIndex, setLockIndex] = useState(0);
  const [stakeAmount, setStakeAmount] = useState("");

  const locks = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getLockDurations",
    args: [BigInt(poolId)],
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const decimal = useReadContract({
    abi: erc20Abi,
    address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
    functionName: "decimals",
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const multiplier = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "PRECISION",
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const pubClient = usePublicClient();
  const { writeContractAsync } = useWriteContract();

  async function onStake() {
    if (!pubClient) {
      console.error("Public client not initialized");
      return;
    }

    if (!address) {
      console.error("No address");
      return;
    }

    const stakeAmountDecimals = BigInt(stakeAmount) * BigInt(1e8);

    const tx = await writeContractAsync({
      abi: erc20Abi,
      address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
      functionName: "approve",
      args: [process.env.REACT_APP_STAKING_ADDR as `0x${string}`, stakeAmountDecimals],
    });

    await pubClient?.waitForTransactionReceipt({
      hash: tx,
      confirmations: 1,
      timeout: 10000,
    });

    const tx2 = await writeContractAsync({
      abi: [...stakingMasterChefAbi, ...erc20Abi],
      address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
      functionName: "stake",
      args: [BigInt(poolId), stakeAmountDecimals, lockIndex],
    });
    const receipt = await pubClient.waitForTransactionReceipt({
      hash: tx2,
      confirmations: 1,
      timeout: 10000,
    });
    const stakeId = getStakeId(receipt.logs, address, BigInt(poolId));
    onStakeCb?.(stakeId);
  }

  return {
    poolId,
    locks,
    decimal,
    pubClient,
    navigate,
    multiplier,
    lockIndex,
    setLockIndex,
    onStake,
    stakeAmount,
    setStakeAmount,
  };
}
