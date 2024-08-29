import { useNavigate, useParams } from "react-router-dom";
import { useStopwatch } from "react-timer-hook";
import { useAccount, useBlock, usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { stakingMasterChefAbi } from "../../blockchain/abi.ts";
import { erc20Abi } from "viem";

export function usePool(onUpdate: () => void) {
  const pubClient = usePublicClient();
  const writeContract = useWriteContract();

  const { poolId: poolIdString } = useParams();
  const poolId = poolIdString !== "" ? Number(poolIdString) : undefined;
  const navigate = useNavigate();

  const { address } = useAccount();

  const block = useBlock();
  const { totalSeconds, reset } = useStopwatch({ autoStart: true });
  const timestamp = block.isSuccess ? block.data?.timestamp + BigInt(totalSeconds) : 0n;

  const poolsCount = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getPoolsCount",
  });

  const poolDataArr = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "pools",
    args: [BigInt(poolId as number)],
    query: {
      enabled: poolId !== undefined,
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

  const locks = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getLockDurations",
    args: [BigInt(poolId as number)],
    query: {
      enabled: poolId !== undefined,
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const locksMap = new Map<bigint, bigint>(
    locks.data?.map(({ durationSeconds, multiplierScaled }) => [durationSeconds, multiplierScaled])
  );

  const poolData = poolDataArr.data
    ? {
        rewardPerSecondScaled: poolDataArr.data[0],
        lastRewardTime: poolDataArr.data[1],
        accRewardPerShareScaled: poolDataArr.data[2],
        totalShares: poolDataArr.data[3],
        startTime: poolDataArr.data[4],
        endTime: poolDataArr.data[5],
        // balanceMOR: poolBalanceMOR.data,
        // balanceLMR: poolBalanceLMR.data,
      }
    : undefined;

  let poolProgress = poolData
    ? Number(timestamp - poolData.startTime) / Number(poolData.endTime - poolData.startTime)
    : 0;

  poolProgress = 0.5;

  if (poolProgress < 0) {
    poolProgress = 0;
  }
  if (poolProgress > 1) {
    poolProgress = 1;
  }

  const poolElapsedDays = poolData ? Math.floor(Number(timestamp - poolData.startTime) / 86400) : 0;
  const poolTotalDays = poolData
    ? Math.floor(Number(poolData.endTime - poolData.startTime) / 86400)
    : 0;
  const poolRemainingSeconds = poolData ? Number(poolData.endTime - timestamp) : 0;

  const stakes = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getStakes",
    args: [address as `0x${string}`, BigInt(poolId as number)],
    query: {
      enabled: address !== undefined && poolId !== undefined,
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
      retry: true,
    },
  });

  async function unstake(stakeId: bigint) {
    if (poolId === undefined) {
      console.error("No poolId");
      return;
    }
    const hash = await writeContract.writeContractAsync({
      abi: stakingMasterChefAbi,
      address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
      functionName: "unstake",
      args: [BigInt(poolId), stakeId],
    });
    await pubClient?.waitForTransactionReceipt({ hash });
    reset();
    onUpdate();
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
    } catch (e) {
      console.error(e);
    }
    reset();
    onUpdate();
  }

  return {
    poolId,
    unstake,
    withdraw,
    timestamp,
    poolsCount,
    stakes,
    poolData,
    poolProgress,
    poolElapsedDays,
    poolTotalDays,
    poolRemainingSeconds,
    locks,
    locksMap,
    lmrBalance,
    morBalance,
    navigate,
  };
}
