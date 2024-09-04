import { useAccount, useBlock, usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { erc20Abi, stakingMasterChefAbi } from "../../blockchain/abi.ts";
import { useNavigate, useParams } from "react-router-dom";
import { useState } from "react";
import { getStakeId } from "./utils.ts";
import { useStopwatch } from "react-timer-hook";
import { mapPoolData } from "../../helpers/pool.ts";
import { decimalsLMR, decimalsMOR } from "../../lib/units.ts";
import { useQueryClient } from "@tanstack/react-query";

export function useStake(onStakeCb?: (id: bigint) => void) {
  // set initial state
  const { poolId: poolIdString } = useParams();
  const { address, chain } = useAccount();
  const poolId = Number(poolIdString);
  const navigate = useNavigate();

  const [lockIndex, setLockIndex] = useState(0);
  const [stakeAmount, _setStakeAmount] = useState("0");
  const [stakeAmountValidEnabled, setStakeAmountValidEnabled] = useState(false);
  const [stakeTxHash, setStakeTxHash] = useState<`0x${string}` | null>(null);

  const block = useBlock({
    query: { refetchInterval: false, refetchOnMount: false, refetchOnReconnect: false },
  });
  const { totalSeconds, reset } = useStopwatch({ autoStart: true });
  const timestamp = block.isSuccess ? block.data?.timestamp + BigInt(totalSeconds) : 0n;

  const qc = useQueryClient();

  function setStakeAmount(value: string) {
    _setStakeAmount(value);
    setStakeAmountValidEnabled(true);
  }

  // load asynchronous data
  const locks = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getLockDurations",
    args: [BigInt(poolId)],
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
      retry: false,
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

  const lmrBalance = useReadContract({
    abi: erc20Abi,
    address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
    functionName: "balanceOf",
    args: [address as `0x${string}`],
    query: {
      enabled: !!address,
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const pool = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "pools",
    args: [BigInt(poolId)],
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const poolData = mapPoolData(pool.data);

  // perform input validations
  const { value: stakeAmountDecimals, error: stakeAmountValidErr } = validStakeAmount(
    stakeAmount,
    lmrBalance.data,
    decimal.data,
    stakeAmountValidEnabled
  );

  const apyValue = apy(poolData, timestamp, stakeAmountDecimals, precision.data, precision.data);

  const pubClient = usePublicClient();
  const writeContract = useWriteContract();

  // define asynchronous calls
  async function onStake() {
    if (!pubClient) {
      console.error("Public client not initialized");
      return;
    }

    if (!address) {
      console.error("No address");
      return;
    }

    const tx = await writeContract.writeContractAsync({
      abi: erc20Abi,
      address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
      functionName: "approve",
      args: [process.env.REACT_APP_STAKING_ADDR as `0x${string}`, stakeAmountDecimals],
    });

    await pubClient?.waitForTransactionReceipt({
      hash: tx,
      confirmations: 1,
    });

    const tx2 = await writeContract.writeContractAsync({
      abi: [...stakingMasterChefAbi, ...erc20Abi],
      address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
      functionName: "stake",
      args: [BigInt(poolId), stakeAmountDecimals, lockIndex],
    });
    const receipt = await pubClient.waitForTransactionReceipt({
      hash: tx2,
      confirmations: 1,
    });
    setStakeTxHash(tx2);
    const stakeId = getStakeId(receipt.logs, address, BigInt(poolId));

    await qc.invalidateQueries({
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
          console.log("invalidating getStakes", params);
          return true;
        }
        if (
          params?.functionName === "balanceOf" &&
          params?.address === process.env.REACT_APP_LMR_ADDR &&
          params?.args?.[0] === address
        ) {
          return true;
        }
        return false;
      },
    });
  }

  return {
    poolId,
    poolData,
    apyValue,
    chain,
    locks,
    decimal,
    pubClient,
    navigate,
    timestamp,
    multiplier: precision,
    lockIndex,
    setLockIndex,
    onStake,
    stakeTxHash,
    lmrBalance,
    stakeAmount,
    setStakeAmount,
    writeContract,
    stakeAmountDecimals,
    stakeAmountValidErr,
  };
}

function validStakeAmount(
  amount: string,
  balance: bigint | undefined,
  decimals: number | undefined,
  enabled: boolean
): { value: bigint; error: string | null } {
  if (!enabled) {
    return { value: BigInt(0), error: "" };
  }
  if (amount === "") {
    return { value: BigInt(0), error: "Enter stake amount" };
  }
  const n = Number.parseFloat(amount);
  if (Number.isNaN(n) || !Number.isFinite(n)) {
    return { value: BigInt(0), error: "Stake amount must be a number" };
  }

  if (n <= 0) {
    return { value: BigInt(0), error: "Stake amount must be larger than 0" };
  }

  if (balance === undefined || decimals === undefined) {
    return { value: BigInt(n), error: "" };
  }

  const value = BigInt(n) * BigInt(10 ** decimals);
  if (value > balance) {
    return { value, error: "Insufficient LMR balance" };
  }

  return { value, error: "" };
}

function apy(
  poolData: ReturnType<typeof mapPoolData>,
  timestamp: bigint,
  stakeAmount: bigint,
  precision: bigint | undefined,
  yearMultiplierScaled: bigint | undefined
) {
  if (!poolData || !yearMultiplierScaled || !precision || !yearMultiplierScaled) {
    return undefined;
  }
  if (stakeAmount === 0n) {
    return 0;
  }
  const priceOfLumerinInMor = 0.02041 / 10 ** decimalsLMR / (21.8 / 10 ** decimalsMOR);

  const shares = (stakeAmount * yearMultiplierScaled) / precision;
  const rewardDebt = (shares * poolData.accRewardPerShareScaled) / precision;

  const futureTimestamp = timestamp + BigInt(365 * 24 * 60 * 60);
  const futureRewardScaled =
    (futureTimestamp - poolData.lastRewardTime) * poolData.rewardPerSecondScaled;
  const futureTotalShares = poolData.totalShares + shares;
  const futureAccRewardPerShareScaled =
    poolData.accRewardPerShareScaled + futureRewardScaled / futureTotalShares;

  const reward1yearMor = (shares * futureAccRewardPerShareScaled) / precision - rewardDebt;
  const reward1yearLmr = BigInt(Math.floor(Number(reward1yearMor) / priceOfLumerinInMor));

  const apy = Number((reward1yearLmr * 100n * 100n) / stakeAmount) / 100; // trimmed at two decimal places
  return apy;
}
