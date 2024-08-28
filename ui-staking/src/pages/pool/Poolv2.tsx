import { Header } from "../../components/Header.tsx";
import { Link, useParams } from "react-router-dom";
import { Separator } from "../../components/Separator.tsx";
import { PieChart } from "react-minimal-pie-chart";
import { Container } from "../../components/Container.tsx";
import { usePool } from "./usePool.ts";
import { Chart } from "../../components/Chart.tsx";
import { formatLMR, formatMOR } from "../../lib/units.ts";
import { formatDate, formatDuration } from "../../lib/date.ts";
import { useState } from "react";
import { Button, ButtonSecondary } from "../../components/Button.tsx";
import { SpoilerToogle } from "../../components/SpoilerToogle.tsx";
import { getReward } from "../../reward.ts";

export const PoolV2 = () => {
  const {
    poolId,
    unstake,
    withdraw,
    timestamp,
    poolsCount,
    poolData,
    stakes,
    poolProgress,
    poolElapsedDays,
    poolTotalDays,
    locks,
    lmrBalance,
    morBalance,
    locksMap,
    navigate,
  } = usePool(() => {});

  return (
    <>
      <Header />
      <main>
        <Container>
          <div className="lens" />
          <nav className="pool-nav">
            <ul>
              {[...Array(poolsCount.data)].map((_, i) => (
                // biome-ignore lint/suspicious/noArrayIndexKey: order of items is fixed
                <li key={i}>
                  <Link className={poolId === i ? "active" : ""} to={`/pool/${i}`}>
                    Pool {i}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>

          <div className="pool">
            <section className="section pool-stats">
              <h2 className="section-heading">Pool stats</h2>
              <Separator />
              {poolData && (
                <dl className="info">
                  <dt>Reward per second</dt>
                  <dd>{formatMOR(poolData.accRewardPerShareScaled)}</dd>

                  <dt>Total shares</dt>
                  <dd>{poolData.totalShares.toString()}</dd>

                  <dt>Total staked</dt>
                  <dd>{formatLMR(3000n)}</dd>

                  <dt>Start date</dt>
                  <dd>{formatDate(poolData.startTime)}</dd>

                  <dt>End date</dt>
                  <dd>{formatDate(poolData.endTime)}</dd>

                  <dt>Lockup periods</dt>
                  <dd>{locks.data?.map((l) => formatDuration(l.durationSeconds)).join(", ")}</dd>
                </dl>
              )}
            </section>
            <section className="section rewards-balance">
              <h2 className="section-heading">Rewards balance</h2>
              <Separator />
              <dl className="info">
                <dt>Locked Rewards</dt>
                <dd>{formatMOR(0n)}</dd>
                <dt>Unlocked Rewards</dt>
                <dd>{formatMOR(0n)}</dd>
              </dl>
            </section>
            <section className="section wallet-balance">
              <h2 className="section-heading">Wallet balance</h2>
              <Separator />
              <ul className="info">
                <li>{formatLMR(lmrBalance.data || 0n)}</li>
                <li>{formatMOR(morBalance.data || 0n)}</li>
                <li>
                  <Button className="button-secondary button-small" onClick={() => navigate(`/pool/${poolId}/stake`)}>
                    Stake
                  </Button>
                </li>
              </ul>
            </section>
            <section className="section stake-list">
              {/* <Chart progress={poolProgress}>
								<dl>
									<dt>Elapsed</dt>
									<dd>
										{poolElapsedDays}/{poolTotalDays} days
									</dd>
								</dl>
							</Chart> */}

              <h2 className="section-heading">My Stakes</h2>
              <ul className="stakes">
                {poolData &&
                  stakes.isSuccess &&
                  stakes.data.map((stake, index) => {
                    const stakedAt = stake.stakedAt || 0n;
                    const lockPassedSeconds = timestamp - stakedAt;
                    const lockRemainingSeconds = stake.lockEndsAt - timestamp;
                    const lockTotalSeconds = stake.lockEndsAt - stakedAt;
                    const lockProgress = Number(lockPassedSeconds) / Number(lockTotalSeconds);
                    const lockMultiplier = locksMap.get(lockTotalSeconds);

                    return (
                      // biome-ignore lint/suspicious/noArrayIndexKey: order of items is fixed
                      <li key={index} className="stake">
                        <SpoilerToogle />
                        <ul className="unchecked">
                          <li>{formatLMR(stake.stakeAmount)}</li>
                          <li>
                            <Chart progress={lockProgress} lineWidth={18} className="chart-small" />
                            {formatDuration(lockRemainingSeconds)} left
                          </li>
                          <li>{formatMOR(getReward(stake, poolData, timestamp, BigInt(1e12)))} earned</li>
                          <li>{lockMultiplier ? `${Number(lockMultiplier) / 1e12}x` : "unknown"} multiplier</li>
                        </ul>
                        <ul className="checked">
                          <li>
                            <p className="title">Amount Staked</p>
                            <p className="value">{formatLMR(stake.stakeAmount)}</p>
                          </li>
                          <li>
                            <p className="title">Lockup Period</p>
                            <p className="value">{formatDuration(lockTotalSeconds)}</p>
                          </li>
                          <li>
                            <p className="title">Time Left</p>
                            <p className="value">{formatDuration(lockRemainingSeconds)}</p>
                          </li>
                          <li className="progress">
                            <Chart progress={lockProgress} lineWidth={23}>
                              <dl>
                                <dt>Lockup Period</dt>
                                <dd>{Math.trunc(lockProgress * 100)} %</dd>
                              </dl>
                            </Chart>
                          </li>
                          <li>
                            <p className="title">Reward Multiplier</p>
                            <p className="value">1.15x</p>
                          </li>
                          <li>
                            <p className="title">Current Rewards</p>
                            <p className="value">{formatMOR(getReward(stake, poolData, timestamp, BigInt(1e12)))}</p>
                          </li>
                          <li>
                            <p className="title">Share Amount</p>
                            <p className="value">{stake.shareAmount.toString()}</p>
                          </li>
                          <li>
                            <p className="title">Unlock Date</p>
                            <p className="value">{formatDate(stake.lockEndsAt)}</p>
                          </li>
                          <li>
                            <Button className="button-secondary button-small" onClick={() => withdraw(BigInt(index))}>
                              Withdraw rewards
                            </Button>
                          </li>
                          <li>
                            <Button className="button-secondary button-small" onClick={() => unstake(BigInt(index))}>
                              Unstake
                            </Button>
                          </li>
                        </ul>
                      </li>
                    );
                  })}
              </ul>
            </section>
          </div>
        </Container>
      </main>
    </>
  );
};
