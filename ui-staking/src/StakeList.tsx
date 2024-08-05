import React from "react";
import { usePublicClient, useReadContract, useWriteContract } from "wagmi";
import { stakingMasterChefAbi } from "./blockchain/abi.ts";
import { getReward, type UserStake, type Pool } from "./reward.ts";
import { formatLMR, formatMOR } from "./lib/units.ts";
import { formatDate } from "./lib/date.ts";
import { useStopwatch } from "react-timer-hook";

interface Props {
	userAddr: `0x${string}`;
	poolId: bigint;
	blockTimestamp: bigint;
	poolData: Pool;
	stakes: readonly UserStake[];
	precision: bigint;
	onUpdate: () => void;
}

export const StakeList = (props: Props) => {
	const pubClient = usePublicClient();
	const writeContract = useWriteContract();

	const { totalSeconds, reset } = useStopwatch({ autoStart: true });
	const timestamp = props.blockTimestamp + BigInt(totalSeconds);

	async function onUnstake(stakeId: bigint) {
		const hash = await writeContract.writeContractAsync({
			abi: stakingMasterChefAbi,
			address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
			functionName: "unstake",
			args: [props.poolId, stakeId],
		});
		await pubClient?.waitForTransactionReceipt({ hash });
		reset();
		props.onUpdate();
	}

	async function onWithdraw(stakeId: bigint) {
		try {
			const hash = await writeContract.writeContractAsync({
				abi: stakingMasterChefAbi,
				address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
				functionName: "withdrawReward",
				args: [props.poolId, stakeId],
			});
			const receipt = await pubClient?.waitForTransactionReceipt({ hash });
			console.log(receipt);
		} catch (e) {
			console.error(e);
		}
		reset();
		props.onUpdate();
	}

	return (
		<div>
			<h1>Stake List</h1>

			{props.stakes.map((stake, index) => (
				<p key={stake.lockEndsAt}>
					Stake {index}:<br />
					amount staked {formatLMR(stake.stakeAmount)}
					<br />
					staking ends at {formatDate(stake.lockEndsAt)}
					<br />
					share amount {stake.shareAmount.toString()}
					<br />
					earned reward{" "}
					{formatMOR(
						getReward(stake, props.poolData, timestamp, props.precision),
					)}
					<br />
					reward debt {formatMOR(stake.rewardDebt)}
					<br />
					<button
						type="button"
						onClick={() => onUnstake(BigInt(index))}
						disabled={props.blockTimestamp < stake.lockEndsAt}
					>
						Unstake
					</button>
					<button type="button" onClick={() => onWithdraw(BigInt(index))}>
						Withdraw Reward
					</button>
				</p>
			))}
		</div>
	);
};
