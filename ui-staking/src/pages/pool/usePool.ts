import { useNavigate, useParams } from "react-router-dom";
import { useStopwatch } from "react-timer-hook";
import { useAccount, useBlock, usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { stakingMasterChefAbi } from "../../blockchain/abi.ts";
import { erc20Abi } from "viem";
import { mapPoolDataAndDerive } from "../../helpers/pool.ts";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

export function usePool(onUpdate: () => void) {
  const pubClient = usePublicClient();
  const writeContract = useWriteContract();

  const { poolId: poolIdString } = useParams();
  const poolId = poolIdString !== "" ? Number(poolIdString) : undefined;
  const navigate = useNavigate();

  const { address, chain } = useAccount();

  const block = useBlock();
  const { totalSeconds, reset } = useStopwatch({ autoStart: true });
  const timestamp = block.isSuccess ? block.data?.timestamp + BigInt(totalSeconds) : 0n;

  const qc = useQueryClient();

  const poolsCount = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getPoolsCount",
  });

  const shouldQueryPool = poolId !== undefined && poolsCount.isSuccess && poolId < poolsCount.data;
  const poolNotFound = poolId !== undefined && poolsCount.isSuccess && poolId >= poolsCount.data;

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

  const [dialog, setDialog] = useState({
    content1: "",
    content2: "" as string | Error,
    dialogHeader: "",
    show: false,
    onDismiss: () => {},
  });

  const locksMap = new Map<bigint, bigint>(
    locks.data?.map(({ durationSeconds, multiplierScaled }) => [durationSeconds, multiplierScaled])
  );

  const poolData = mapPoolDataAndDerive(poolDataArr.data, timestamp, precision.data);

  async function unstake(stakeId: bigint) {
    if (poolId === undefined) {
      console.error("No poolId");
      return;
    }
    try {
      const hash = await writeContract.writeContractAsync({
        abi: stakingMasterChefAbi,
        address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
        functionName: "unstake",
        args: [BigInt(poolId), stakeId],
      });
      await pubClient?.waitForTransactionReceipt({ hash });
      setDialog({
        dialogHeader: "Transaction successful",
        content1: `You have successfully unstaked from pool ${poolId}`,
        content2: hash,
        show: true,
        onDismiss: () => {
          qc.invalidateQueries({
            predicate: (q) => {
              // invalidate all queries related to the pool
              const params = q.queryKey?.[1];
              if (!params) {
                return false;
              }
              if (params?.functionName === "pools" && params?.args?.[0] === BigInt(poolId)) {
                return true;
              }
              if (params?.functionName === "getStakes" && params?.args?.[1] === BigInt(poolId)) {
                return true;
              }
              if (
                params?.functionName === "balanceOf" &&
                params?.address === process.env.REACT_APP_LMR_ADDR &&
                params?.args?.[0] === address
              ) {
                return true;
              }
              if (
                params?.functionName === "balanceOf" &&
                params?.address === process.env.REACT_APP_MOR_ADDR &&
                params?.args?.[0] === address
              ) {
                return true;
              }
              return false;
            },
          });
          setDialog({ ...dialog, show: false });
          reset();
          onUpdate();
        },
      });
    } catch (e) {
      setDialog({
        dialogHeader: "Transaction failed",
        content1: `You have failed to unstake from pool ${poolId}`,
        content2: e as Error,
        show: true,
        onDismiss: () => {
          setDialog({ ...dialog, show: false });
          reset();
        },
      });
      console.error(e);
    }
  }

  async function withdraw(stakeId: bigint) {
    if (poolId === undefined) {
      console.error("No poolId");
      return;
    }
    try {
      const hash = await writeContract.writeContractAsync({
        abi: stakingMasterChefAbi,
        address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
        functionName: "withdrawReward",
        args: [BigInt(poolId), stakeId],
      });
      const receipt = await pubClient?.waitForTransactionReceipt({ hash });
      setDialog({
        dialogHeader: "Transaction successful",
        content1: `You have successfully withdrawn your rewards from pool ${poolId}`,
        content2: hash,
        show: true,
        onDismiss: () => {
          setDialog({ ...dialog, show: false });
          reset();
          onUpdate();
        },
      });
    } catch (e) {
      setDialog({
        dialogHeader: "Transaction failed",
        content1: `You have failed to withdraw your rewards from pool ${poolId}`,
        content2: e as Error,
        show: true,
        onDismiss: () => {
          setDialog({ ...dialog, show: false });
          reset();
        },
      });
      console.error(e);
    }
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
    lmrBalance,
    morBalance,
    navigate,
    dialog,
  };
}
