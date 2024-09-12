import { useAccount, useBlock, usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { erc20Abi, stakingMasterChefAbi } from "../../blockchain/abi.ts";
import { useNavigate, useParams } from "react-router-dom";
import { useState } from "react";
import { useStopwatch } from "react-timer-hook";
import { mapPoolData } from "../../helpers/pool.ts";
import { decimalsLMR, decimalsMOR } from "../../lib/units.ts";
import { useQueryClient } from "@tanstack/react-query";
import {
  filterPoolQuery,
  filterStakeQuery,
  filterUserBalanceQuery,
} from "../../helpers/invalidators.ts";
import { useTxModal } from "../../hooks/useTxModal.ts";

export function useStake() {
  // set initial state
  const { poolId: poolIdString } = useParams();
  const { address, chain } = useAccount();
  const poolId = Number(poolIdString);
  const navigate = useNavigate();

  const [lockIndex, setLockIndex] = useState(0);
  const [stakeAmount, _setStakeAmount] = useState("0");
  const [stakeAmountValidEnabled, setStakeAmountValidEnabled] = useState(false);
  const txModal = useTxModal();

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

  const lockDurationSeconds = locks.data?.[lockIndex].durationSeconds || 0n;
  const effectiveStakeStartTime =
    poolData && timestamp > poolData?.startTime ? timestamp : poolData?.startTime;
  const lockEndsAt = effectiveStakeStartTime && effectiveStakeStartTime + lockDurationSeconds;

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

    await txModal.start({
      approveCall: () =>
        writeContract.writeContractAsync({
          abi: erc20Abi,
          address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
          functionName: "approve",
          args: [process.env.REACT_APP_STAKING_ADDR as `0x${string}`, stakeAmountDecimals],
        }),
      txCall: () =>
        writeContract.writeContractAsync({
          abi: [...stakingMasterChefAbi, ...erc20Abi],
          address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
          functionName: "stake",
          args: [BigInt(poolId), stakeAmountDecimals, lockIndex],
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
        await qc.invalidateQueries({
          predicate: filterUserBalanceQuery(address),
          refetchType: "all",
        });
      },
    });
  }

  return {
    txModal,
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
    lmrBalance,
    stakeAmount,
    setStakeAmount,
    writeContract,
    stakeAmountDecimals,
    stakeAmountValidErr,
    lockDurationSeconds,
    effectiveStakeStartTime,
    lockEndsAt,
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
  const priceOfLumerinInMor =
    0.02041 / 10 ** Number(decimalsLMR) / (21.8 / 10 ** Number(decimalsMOR));

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
