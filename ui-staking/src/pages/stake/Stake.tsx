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
import { TxProgress } from "../../components/TxProgress.tsx";
import { getDisplayErrorMessage } from "../../helpers/error.ts";

export const Stake = () => {
  const {
    timestamp,
    locks,
    poolData,
    apyValue,
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
    txModal,
    lockDurationSeconds,
    lockEndsAt,
  } = useStake();

  const isNoPoolError = isErr<typeof stakingMasterChefAbi>(locks.error, "PoolOrStakeNotExists");
  const isNoLockPeriods = locks.isSuccess && locks.data.length === 0;
  const isPoolExpired = poolData && timestamp > poolData.endTime;

  const rewardMultiplier =
    locks.isSuccess && multiplier.isSuccess
      ? Number(locks.data[lockIndex].multiplierScaled) / Number(multiplier.data)
      : 0;

  // const stakeTxURL = getTxURL(stakeTxHash, chain);

  return (
    <>
      <Header />
      <main>
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
                          e.target.value === "" || Number(e.target.value) > 0 ? e.target.value : "0"
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
                  <div className="range-wrap">
                    <RangeSelect
                      label="Lockup period"
                      value={lockIndex}
                      titles={locks.data.map((l) => formatSeconds(l.durationSeconds))}
                      onChange={setLockIndex}
                    />
                  </div>
                </div>
                <dl className="field summary">
                  <dt>APY</dt>
                  <dd>unknown</dd>
                  <dt>Lockup Period</dt>
                  <dd>{formatSeconds(lockDurationSeconds)}</dd>
                  <dt>Reward multiplier</dt>
                  <dd>{rewardMultiplier}x</dd>
                  <dt>Lockup ends at</dt>
                  <dd>{lockEndsAt ? formatDate(lockEndsAt) : "unknown"}</dd>
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
          {txModal.isVisible && (
            <Dialog onDismiss={() => txModal.reset()}>
              <div className="dialog-content">
                <h2>Staking transaction</h2>
                <p>
                  Staking {formatLMR(stakeAmountDecimals)} with lock period of{" "}
                  {formatDuration(lockDurationSeconds)}.
                </p>
                <ul className="tx-stages">
                  <li>
                    <p className="stage-name">Approving funds</p>
                    <p className="stage-progress">
                      <TxProgress
                        isTransacting={txModal.isApproving}
                        txHash={txModal.approveTxHash}
                        error={getDisplayErrorMessage(txModal.approveError)}
                      />
                    </p>
                  </li>
                  <li>
                    <p className="stage-name">Adding stake</p>
                    <p className="stage-progress">
                      <TxProgress
                        isTransacting={txModal.isTransacting}
                        txHash={txModal.txHash}
                        error={getDisplayErrorMessage(txModal.txError)}
                      />
                    </p>
                  </li>
                </ul>
                <button
                  className="button button-small button-primary"
                  type="button"
                  onClick={() => {
                    txModal.reset();
                    if (txModal.isTransactionSuccess) {
                      navigate(`/pool/${poolId}`);
                    }
                  }}
                >
                  OK
                </button>
              </div>
            </Dialog>
          )}
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
