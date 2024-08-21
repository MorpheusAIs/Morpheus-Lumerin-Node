import { useNavigate, useParams } from "react-router-dom";
import { useStopwatch } from "react-timer-hook";
import { useAccount, useBlock, usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { stakingMasterChefAbi } from "../../blockchain/abi.ts";

export function usePool(address: `0x${string}`, onUpdate: () => void) {
  const pubClient = usePublicClient();
  const writeContract = useWriteContract();

  const { poolId: poolIdString } = useParams();
  const poolId = Number(poolIdString);
  const navigate = useNavigate();

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
    args: [BigInt(poolId)],
  });

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

  if (poolProgress < 0) {
    poolProgress = 0;
  }
  if (poolProgress > 1) {
    poolProgress = 1;
  }

  console.log(new Date(Number(timestamp) * 1000));

  const poolElapsedDays = poolData ? Math.floor(Number(timestamp - poolData.startTime) / 86400) : 0;
  const poolTotalDays = poolData
    ? Math.floor(Number(poolData.endTime - poolData.startTime) / 86400)
    : 0;

  const stakes = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getStakes",
    args: [address, BigInt(poolId)],
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
      retry: true,
    },
  });

  async function unstake(stakeId: bigint) {
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
    navigate,
  };
}
