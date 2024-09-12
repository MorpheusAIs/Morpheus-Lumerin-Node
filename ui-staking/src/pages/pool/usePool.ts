import { useNavigate, useParams } from "react-router-dom";
import { useStopwatch } from "react-timer-hook";
import { useAccount, useBalance, useBlock, usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { stakingMasterChefAbi } from "../../blockchain/abi.ts";
import { erc20Abi } from "viem";
import { mapPoolDataAndDerive } from "../../helpers/pool.ts";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { filterPoolQuery, filterStakeQuery, filterUserBalanceQuery } from "../../helpers/invalidators.ts";
import { useTxModal } from "../../hooks/useTxModal.ts";

export function usePool(onUpdate: () => void) {
  const writeContract = useWriteContract();

  const { poolId: poolIdString } = useParams();
  const poolId = poolIdString !== "" ? Number(poolIdString) : undefined;
  const navigate = useNavigate();

  const { address, chain } = useAccount();

  const block = useBlock();
  const { totalSeconds } = useStopwatch({ autoStart: true });
  const timestamp = block.isSuccess ? block.data?.timestamp + BigInt(totalSeconds) : 0n;

  const qc = useQueryClient();

  const poolsCount = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getPoolsCount",
    args: [],
  });

  const shouldQueryPool = poolId !== undefined && poolsCount.isSuccess && poolId < poolsCount.data;
  const poolNotFound =
    (poolId !== undefined && poolsCount.isSuccess && poolId >= poolsCount.data) || poolsCount.isError;

  const poolDataArr = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "pools",
    args: [BigInt(poolId as number)],
    query: {
      enabled: shouldQueryPool,
    },
  });

  const locks = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getLockDurations",
    args: [BigInt(poolId as number)],
    query: {
      enabled: shouldQueryPool,
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const stakes = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getStakes",
    args: [address as `0x${string}`, BigInt(poolId as number)],
    query: {
      enabled: address && shouldQueryPool,
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
      retry: true,
    },
  });

  const ethBalance = useBalance({
    address,
    query: { refetchOnMount: false, refetchOnReconnect: false, refetchOnWindowFocus: false },
  });

  const lmrBalance = useReadContract({
    abi: erc20Abi,
    address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
    functionName: "balanceOf",
    args: [address as `0x${string}`],
    query: {
      enabled: address !== undefined,
    },
  });

  const morBalance = useReadContract({
    abi: erc20Abi,
    address: process.env.REACT_APP_MOR_ADDR as `0x${string}`,
    functionName: "balanceOf",
    args: [address as `0x${string}`],
    query: {
      enabled: address !== undefined,
    },
  });

  const precision = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "PRECISION",
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const withdrawModal = useTxModal();
  const unstakeModal = useTxModal();

  const locksMap = new Map<bigint, bigint>(
    locks.data?.map(({ durationSeconds, multiplierScaled }) => [durationSeconds, multiplierScaled]),
  );

  const poolData = mapPoolDataAndDerive(poolDataArr.data, timestamp, precision.data);

  async function unstake(stakeId: bigint) {
    if (poolId === undefined) {
      console.error("No poolId");
      return;
    }

    await unstakeModal.start({
      txCall: async () =>
        writeContract.writeContractAsync({
          abi: stakingMasterChefAbi,
          address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
          functionName: "unstake",
          args: [BigInt(poolId), stakeId],
        }),
      onSuccess: async () => {
        await qc.invalidateQueries({
          predicate: filterPoolQuery(BigInt(poolId)),
          refetchType: "all",
        });
        await qc.invalidateQueries({
          predicate: filterStakeQuery(BigInt(poolId)),
          refetchType: "all",
        });
        if (address) {
          await qc.invalidateQueries({
            predicate: filterUserBalanceQuery(address),
            refetchType: "all",
          });
        }
        onUpdate();
      },
    });
  }

  async function withdraw(stakeId: bigint) {
    if (poolId === undefined) {
      console.error("No poolId");
      return;
    }

    await withdrawModal.start({
      txCall: async () =>
        writeContract.writeContractAsync({
          abi: stakingMasterChefAbi,
          address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
          functionName: "withdrawReward",
          args: [BigInt(poolId), stakeId],
        }),
      onSuccess: async () => {
        await qc.invalidateQueries({
          predicate: filterPoolQuery(BigInt(poolId)),
          refetchType: "all",
        });
        await qc.invalidateQueries({
          predicate: filterStakeQuery(BigInt(poolId)),
          refetchType: "all",
        });
        if (address) {
          await qc.invalidateQueries({
            predicate: filterUserBalanceQuery(address),
            refetchType: "all",
          });
        }
        writeContract.reset();
        onUpdate();
      },
    });
  }

  return {
    poolId,
    precision,
    chain,
    unstake,
    withdraw,
    timestamp,
    poolsCount,
    stakes,
    poolData,
    poolIsLoading: poolDataArr.isLoading,
    poolError: poolDataArr.error,
    poolNotFound,
    locks,
    locksMap,
    ethBalance,
    lmrBalance,
    morBalance,
    navigate,
    withdrawModal,
    unstakeModal,
  };
}
