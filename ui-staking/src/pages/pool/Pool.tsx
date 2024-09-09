import { Header } from "../../components/Header.tsx";
import { Link } from "react-router-dom";
import { Separator } from "../../components/Separator.tsx";
import { Container } from "../../components/Container.tsx";
import { usePool } from "./usePool.ts";
import { Chart } from "../../components/Chart.tsx";
import { formatETH, formatLMR, formatMOR, formatUnits } from "../../lib/units.ts";
import { formatDate, formatDuration } from "../../lib/date.ts";
import { Button } from "../../components/Button.tsx";
import { SpoilerToogle } from "../../components/SpoilerToogle.tsx";
import { getReward } from "../../helpers/reward.ts";
import { Spinner } from "../../icons/Spinner.tsx";
import { Dialog } from "../../components/Dialog.tsx";
import { TxProgress } from "../../components/TxProgress.tsx";
import { getDisplayErrorMessage } from "../../helpers/error.ts";

export const Pool = () => {
  const {
    poolId,
    unstake,
    precision,
    withdraw,
    timestamp,
    poolsCount,
    stakes,
    poolData,
    poolIsLoading,
    poolError,
    poolNotFound,
    locks,
    lmrBalance,
    morBalance,
    locksMap,
    navigate,
    chain,
    withdrawModal,
    unstakeModal,
    ethBalance,
  } = usePool(() => {});

  const activeStakes = stakes.data
    ?.map((stake, id) => ({ id, ...stake }))
    .filter((stake) => stake.stakeAmount > 0n);

  return (
    <>
      <Header />
      <main>
        <Container>
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
          {poolIsLoading ||
            (poolError && (
              <div className="section loading">
                {poolIsLoading && !poolNotFound && <Spinner />}
                {poolNotFound && <p className="error">Pool not found</p>}
                {poolError && <p className="error">Error: {poolError.message}</p>}
              </div>
            ))}

          {poolData && precision.isSuccess && (
            <div className="pool">
              <section className="section pool-stats">
                <h2 className="section-heading">Pool stats</h2>
                <Separator />

                <dl className="info">
                  <dt>Reward per second</dt>
                  <dd>{formatMOR(poolData.rewardPerSecondScaled / precision.data)}</dd>

                  <dt>Total shares</dt>
                  <dd>{poolData.totalShares.toString()}</dd>

                  <dt>Total staked</dt>
                  <dd>{formatLMR(poolData.totalStaked)}</dd>

                  <dt>Start</dt>
                  <dd className="shift-left">{formatDate(poolData.startTime)}</dd>

                  <dt>End</dt>
                  <dd className="shift-left">{formatDate(poolData.endTime)}</dd>

                  <dt>Duration</dt>
                  <dd>{formatDuration(poolData.endTime - poolData.startTime)}</dd>

                  <dt>Lockup periods</dt>
                  <dd>{locks.data?.map((l) => formatDuration(l.durationSeconds)).join(", ")}</dd>
                </dl>
              </section>
              <section className="section rewards-balance">
                <h2 className="section-heading">Pool rewards balance</h2>
                <Separator />
                <dl className="info">
                  <dt>Total Rewards</dt>
                  <dd>{formatMOR(poolData.totalRewards)}</dd>
                  <dt>Locked Rewards</dt>
                  <dd>{formatMOR(poolData.lockedRewards)}</dd>
                  <dt>Unlocked Rewards</dt>
                  <dd>{formatMOR(poolData.unlockedRewards)}</dd>
                </dl>
              </section>
              <section className="section wallet-balance">
                <h2 className="section-heading">Wallet balance</h2>
                <Separator />
                <ul className="info">
                  <li>
                    {formatUnits(
                      ethBalance.data?.value || 0n,
                      BigInt(ethBalance.data?.decimals || 0)
                    )}{" "}
                    {ethBalance.data?.symbol}
                  </li>
                  <li>{formatLMR(lmrBalance.data || 0n)}</li>
                  <li>{formatMOR(morBalance.data || 0n)}</li>
                  <li>
                    <Button
                      className="button button-small"
                      onClick={() => navigate(`/pool/${poolId}/stake`)}
                    >
                      Stake
                    </Button>
                  </li>
                </ul>
              </section>
              <section className="section stake-list">
                <h2 className="section-heading">My Stakes</h2>
                {stakes.isLoading && (
                  <div className="spinner-container">
                    <Spinner />
                  </div>
                )}
                {activeStakes?.length === 0 && (
                  <div className="stake-list-no-stakes">No stakes found</div>
                )}
                <ul className="stakes">
                  {poolData &&
                    activeStakes &&
                    activeStakes.map((stake) => {
                      if (stake.stakeAmount === 0n) {
                        return null;
                      }
                      const stakedAt = stake.stakedAt || 0n;
                      const lockRemainingSeconds = stake.lockEndsAt - timestamp;
                      const stakeStartTime =
                        stakedAt < poolData.startTime ? poolData.startTime : stakedAt;
                      const lockTotalSeconds = stake.lockEndsAt - stakeStartTime;
                      let lockPassedSeconds = timestamp - stakeStartTime;
                      if (lockPassedSeconds < 0) {
                        lockPassedSeconds = 0n;
                      }
                      let lockProgress = Number(lockPassedSeconds) / Number(lockTotalSeconds);
                      lockProgress = lockProgress > 1 ? 1 : lockProgress;
                      const rewardMultiplier = locksMap.get(lockTotalSeconds);

                      const rewardMultiplierString =
                        rewardMultiplier && precision.data
                          ? `${Number(rewardMultiplier) / Number(precision.data)}x`
                          : "";

                      const timeLeftString =
                        lockRemainingSeconds > 0
                          ? formatDuration(lockRemainingSeconds)
                          : "Stake unlocked";

                      return (
                        <li key={stake.id} className="stake">
                          <SpoilerToogle />
                          <ul className="unchecked">
                            <li className="amount">{formatLMR(stake.stakeAmount)}</li>
                            <li className="chart-item">
                              <Chart
                                progress={lockProgress}
                                lineWidth={18}
                                className="chart-small"
                              />
                              <span className="chart-small-text">{timeLeftString}</span>
                            </li>
                            <li className="reward">
                              {formatMOR(getReward(stake, poolData, timestamp, precision.data))}{" "}
                              earned
                            </li>
                            <li className="multiplier">{rewardMultiplierString} multiplier</li>
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
                              <p className="value">{timeLeftString}</p>
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
                              <p className="value">{rewardMultiplierString}</p>
                            </li>
                            <li>
                              <p className="title">Current Rewards</p>
                              <p className="value">
                                {formatMOR(getReward(stake, poolData, timestamp, precision.data))}
                              </p>
                            </li>
                            <li>
                              <p className="title">Share Amount</p>
                              <p className="value">{stake.shareAmount.toString()}</p>
                            </li>
                            <li>
                              <p className="title">Unlock Date</p>
                              <p className="value">{formatDate(stake.lockEndsAt)}</p>
                            </li>
                            <li className="item-button">
                              <Button
                                className="button button-small"
                                onClick={() => withdraw(BigInt(stake.id))}
                              >
                                Withdraw rewards
                              </Button>
                            </li>
                            <li className="item-button">
                              <Button
                                className="button button-small"
                                disabled={lockRemainingSeconds > 0}
                                title={
                                  lockRemainingSeconds > 0
                                    ? "Lockup period has not ended yet"
                                    : "Unstake the stake and withdraw all rewards"
                                }
                                onClick={() => unstake(BigInt(stake.id))}
                              >
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
          )}
        </Container>
      </main>

      {withdrawModal.isVisible && (
        <Dialog onDismiss={() => withdrawModal.reset()}>
          <div className="dialog-content">
            <h2>Withdrawing rewards</h2>
            <p>Withdrawing all of your staking rewards</p>
            <ul className="tx-stages">
              <li>
                <p className="stage-name">Withdraw transaction</p>
                <p className="stage-progress">
                  <TxProgress
                    isTransacting={withdrawModal.isTransacting}
                    txHash={withdrawModal.txHash}
                    error={getDisplayErrorMessage(withdrawModal.txError)}
                  />
                </p>
              </li>
            </ul>
            <button
              className="button button-small button-primary"
              type="button"
              onClick={() => {
                withdrawModal.reset();
                if (withdrawModal.isTransactionSuccess) {
                  navigate(`/pool/${poolId}`);
                }
              }}
            >
              OK
            </button>
          </div>
        </Dialog>
      )}

      {unstakeModal.isVisible && (
        <Dialog onDismiss={() => unstakeModal.reset()}>
          <div className="dialog-content">
            <h2>Unstake transaction</h2>
            <p>Withdrawing your stake and all of the collected rewards</p>
            <ul className="tx-stages">
              <li>
                <p className="stage-name">Unstaking</p>
                <p className="stage-progress">
                  <TxProgress
                    isTransacting={unstakeModal.isTransacting}
                    txHash={unstakeModal.txHash}
                    error={getDisplayErrorMessage(unstakeModal.txError)}
                  />
                </p>
              </li>
            </ul>
            <button
              className="button button-small button-primary"
              type="button"
              onClick={() => {
                unstakeModal.reset();
                if (unstakeModal.isTransactionSuccess) {
                  navigate(`/pool/${poolId}`);
                }
              }}
            >
              OK
            </button>
          </div>
        </Dialog>
      )}
    </>
  );
};
