import React, { useState } from "react";
import { usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { stakingMasterChefAbi, erc20Abi } from "./blockchain/abi.ts";
import { parseUnits } from "viem";
import { formatDuration } from "./lib/date.ts";

interface StakeAddProps {
	poolId: bigint;
	userBalance: bigint;
	precision: bigint;
	onAdd: (stakeId: number) => void;
}

export const StakeAdd = (props: StakeAddProps) => {
	const locks = useReadContract({
		abi: stakingMasterChefAbi,
		address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
		functionName: "getLockDurations",
		args: [props.poolId],
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

	const pubClient = usePublicClient();

	const approve = useWriteContract();
	const { writeContractAsync, data, isPending, isError, error } =
		useWriteContract();

	const [stakeAmount, setStakeAmount] = useState(0n);
	const [lockIndex, setLockIndex] = useState(0);

	async function onAddStake() {
		const tx = await writeContractAsync({
			abi: erc20Abi,
			address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
			functionName: "approve",
			args: [process.env.REACT_APP_STAKING_ADDR as `0x${string}`, stakeAmount],
		});

		const receipt = await pubClient?.waitForTransactionReceipt({
			hash: tx,
			confirmations: 1,
			timeout: 10000,
		});
		await writeContractAsync({
			abi: [...stakingMasterChefAbi, ...erc20Abi],
			address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
			functionName: "stake",
			args: [props.poolId, stakeAmount, lockIndex],
		});
		props.onAdd(0);
	}

	if (isError) {
		console.error(error);
	}

	if (!decimal.isSuccess || !pubClient) {
		return <p>Loading decimals...</p>;
	}

	return (
		<div>
			<h1>Add a new stake</h1>
			{isPending && <p>Adding stake...</p>}
			{isError && <p>Error: {error.message}</p>}
			<input
				type="number"
				id="stakeAmount"
				name="stakeAmount"
				onChange={(e) =>
					setStakeAmount(parseUnits(e.target.value, decimal.data))
				}
			/>

			<select
				value={lockIndex}
				onChange={(e) => setLockIndex(Number(e.target.value))}
			>
				{locks.data?.map((lock, index) => (
					<option key={lock.durationSeconds} value={index}>
						Duration {formatDuration(lock.durationSeconds).toString()},
						multiplier {Number(lock.multiplierScaled) / Number(props.precision)}
					</option>
				))}
			</select>

			<button type="submit" onClick={() => onAddStake()}>
				Add stake
			</button>
		</div>
	);
};
