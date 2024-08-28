import { useState } from "react";
import { ContainerNarrow } from "../../components/Container.tsx";
import { Header } from "../../components/Header.tsx";
import { Separator } from "../../components/Separator.tsx";
import { LumerinIcon } from "../../icons/LumerinIcon.tsx";
import prettyMilliseconds from "pretty-ms";
import { RangeSelect } from "../../components/RangeSelect.tsx";
import { useStake } from "./useStake.ts";
import { formatDate } from "../../lib/date.ts";

interface Props {
  onStakeCb?: (id: bigint) => void;
}

export const Stake = (props: Props) => {
  const { locks, lockIndex, setLockIndex, navigate, poolId, onStake, multiplier, stakeAmount, setStakeAmount } =
    useStake(props.onStakeCb);

  const lockTitles = locks.isSuccess ? locks.data.map((l) => formatSeconds(l.durationSeconds)) : ["0", "Loading..."];

  const lockupPeriod = locks.isSuccess ? formatSeconds(locks.data[lockIndex].durationSeconds) : "...";

  const rewardMultiplier =
    locks.isSuccess && multiplier.isSuccess
      ? Number(locks.data[lockIndex].multiplierScaled) / Number(multiplier.data)
      : 0;

  const blockchainTime = BigInt(Date.now()) / 1000n;
  const endsAt = locks.isSuccess ? blockchainTime + locks.data[lockIndex].durationSeconds : 0;

  const endsAtString = formatDate(endsAt);

  return (
    <>
      <Header address="0x1234567890abcdef" />
      <main>
        <div className="lens" />
        <ContainerNarrow>
          <section className="section add-stake">
            <h1>New staking contract</h1>
            <div className="field stake-amount">
              <input
                id="stake-amount"
                type="number"
                value={stakeAmount}
                onChange={(e) => setStakeAmount(e.target.value)}
              />
              <label htmlFor="stake-amount">
                <LumerinIcon /> LMR
              </label>
            </div>
            <Separator />
            <div className="field lockup-period">
              <label htmlFor="lockup-period">Lockup period</label>
              <RangeSelect label="Lockup period" value={lockIndex} titles={lockTitles} onChange={setLockIndex} />
            </div>
            <dl className="field summary">
              <dt>APY</dt>
              <dd>4.19%</dd>
              <dt>Lockup Period</dt>
              <dd>{lockupPeriod}</dd>
              <dt>Reward multiplier</dt>
              <dd>{rewardMultiplier}x</dd>
              <dt>Lockup ends at</dt>
              <dd>{endsAtString}</dd>
            </dl>
            <div className="field buttons">
              <button className="button" type="button" onClick={() => navigate(`/pool/${poolId}`)}>
                Cancel
              </button>
              <button className="button button-primary" type="submit" onClick={onStake}>
                Stake
              </button>
            </div>
          </section>
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
