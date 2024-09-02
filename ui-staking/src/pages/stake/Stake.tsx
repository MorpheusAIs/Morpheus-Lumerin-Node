import { ContainerNarrow } from "../../components/Container.tsx";
import { Header } from "../../components/Header.tsx";
import { Separator } from "../../components/Separator.tsx";
import { LumerinIcon } from "../../icons/LumerinIcon.tsx";
import prettyMilliseconds from "pretty-ms";
import { RangeSelect } from "../../components/RangeSelect.tsx";
import { useStake } from "./useStake.ts";
import { formatDate, formatDuration } from "../../lib/date.ts";
import { isErr } from "../../lib/error.ts";
import type { stakingMasterChefAbi } from "../../blockchain/abi.ts";
import { Spinner } from "../../icons/Spinner.tsx";
import { Dialog } from "../../components/Dialog.tsx";
import { formatLMR } from "../../lib/units.ts";
import { useChainId } from "wagmi";
import type { Chain } from "viem";

interface Props {
	onStakeCb?: (id: bigint) => void;
}

export const Stake = (props: Props) => {
	const {
		locks,
		pool,
		lockIndex,
		setLockIndex,
		navigate,
		poolId,
		onStake,
		multiplier,
		stakeAmount,
		stakeAmountDecimals,
		setStakeAmount,
		stakeAmountValidErr,
		chain,
		stakeTxHash,
		writeContract,
	} = useStake(props.onStakeCb);

	const blockchainTime = BigInt(Date.now()) / 1000n;

	const isNoPoolError = isErr<typeof stakingMasterChefAbi>(
		locks.error,
		"PoolOrStakeNotExists",
	);
	const isNoLockPeriods = locks.isSuccess && locks.data.length === 0;
	const isPoolExpired = pool.isSuccess && blockchainTime > pool.data[5];

	const rewardMultiplier =
		locks.isSuccess && multiplier.isSuccess
			? Number(locks.data[lockIndex].multiplierScaled) / Number(multiplier.data)
			: 0;

	const stakeTxURL = stakeTxHash && chain ? getTxURL(stakeTxHash, chain) : null;

	return (
		<>
			<Header />
			<main>
				<div className="lens" />
				<ContainerNarrow>
					<section className="section add-stake">
						<h1>New stake</h1>
						{locks.isLoading && <Spinner className="spinner-center" />}
						{(locks.isError || isNoLockPeriods || isPoolExpired) && (
							<div className="error">
								{isNoPoolError && "Pool not found"}
								{locks.isError && !isNoPoolError && "Pool error"}
								{isNoLockPeriods && "Lock periods not set"}
								{isPoolExpired && "Pool expired"}
							</div>
						)}
						{locks.isSuccess && !isNoLockPeriods && !isPoolExpired && (
							<>
								<div className="field stake-amount">
									<div className="field-input">
										<input
											// biome-ignore lint/a11y/noAutofocus: the main focus is on this input
											autoFocus
											id="stake-amount"
											type="number"
											value={stakeAmount}
											onFocus={(e) => e.currentTarget.select()}
											onChange={(e) =>
												setStakeAmount(
													e.target.value === "" || Number(e.target.value) > 0
														? e.target.value
														: "0",
												)
											}
											onWheel={(e) => e.currentTarget.blur()}
										/>
										<label htmlFor="stake-amount">
											<LumerinIcon /> LMR
										</label>
									</div>
									<div className="field-error">{stakeAmountValidErr}</div>
								</div>
								<Separator />
								<div className="field lockup-period">
									<label htmlFor="lockup-period">Lockup period</label>
									<RangeSelect
										label="Lockup period"
										value={lockIndex}
										titles={locks.data.map((l) =>
											formatSeconds(l.durationSeconds),
										)}
										onChange={setLockIndex}
									/>
								</div>
								<dl className="field summary">
									<dt>APY</dt>
									<dd>4.19%</dd>
									<dt>Lockup Period</dt>
									<dd>
										{formatSeconds(locks.data[lockIndex].durationSeconds)}
									</dd>
									<dt>Reward multiplier</dt>
									<dd>{rewardMultiplier}x</dd>
									<dt>Lockup ends at</dt>
									<dd>
										{formatDate(
											blockchainTime + locks.data[lockIndex].durationSeconds,
										)}
									</dd>
								</dl>
								<div className="field buttons">
									<button
										className="button"
										type="button"
										onClick={() => navigate(`/pool/${poolId}`)}
									>
										Cancel
									</button>
									<button
										className="button button-primary"
										type="submit"
										onClick={onStake}
										disabled={stakeAmount === "0" || stakeAmountValidErr !== ""}
									>
										Stake
									</button>
								</div>
							</>
						)}
					</section>
					{stakeTxHash !== null && locks.isSuccess && (
						<Dialog onDismiss={() => navigate("/pool/0")}>
							<div className="dialog-content">
								<h2>Stake successful</h2>
								<p>
									You have successfully staked {formatLMR(stakeAmountDecimals)}{" "}
									with lock period of{" "}
									{formatDuration(locks.data[lockIndex].durationSeconds)}.
								</p>
								<p>
									Transaction id:{" "}
									<a
										href={stakeTxURL || undefined}
										target="_blank"
										rel="noreferrer"
									>
										{stakeTxHash}
									</a>
								</p>
								<button
									className="button button-primary"
									type="button"
									onClick={() => navigate("/pool/0")}
								>
									OK
								</button>
							</div>
						</Dialog>
					)}
					<Dialog
						// isOpen={stakeTxHash !== null}
						isOpen={writeContract.isError}
						onDismiss={() => writeContract.reset()}
					>
						<div className="dialog-content">
							<h2>Stake error</h2>
							<p>
								There was an error during staking{" "}
								{formatLMR(stakeAmountDecimals)}. Error:{" "}
								{String(writeContract?.error?.cause)}
							</p>
							<p>
								Transaction id:{" "}
								<a
									href={stakeTxURL || undefined}
									target="_blank"
									rel="noreferrer"
								>
									{stakeTxHash}
								</a>
							</p>
							<button
								className="button button-primary"
								type="button"
								onClick={() => writeContract.reset()}
							>
								OK
							</button>
						</div>
					</Dialog>
				</ContainerNarrow>
			</main>
		</>
	);
};

function formatSeconds(seconds: number | bigint): string {
	let ms: number | bigint;
	if (typeof seconds === "bigint") {
		ms = seconds * 1000n;
	} else {
		ms = seconds * 1000;
	}
	return prettyMilliseconds(ms, { verbose: true });
}

function getTxURL(txhash: `0x${string}`, chain: Chain): string | null {
	if (!chain.blockExplorers) {
		return null;
	}
	return `${chain.blockExplorers?.default.url}/tx/${txhash}`;
}
