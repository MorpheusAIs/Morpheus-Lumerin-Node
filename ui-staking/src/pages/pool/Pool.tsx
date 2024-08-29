import { Header } from "../../components/Header.tsx";
import { Link, useParams } from "react-router-dom";
import { Separator } from "../../components/Separator.tsx";
import { PieChart } from "react-minimal-pie-chart";
import { Container } from "../../components/Container.tsx";
import { usePool } from "./usePool.ts";
import { Chart } from "../../components/Chart.tsx";
import { formatLMR, formatMOR } from "../../lib/units.ts";
import { formatDate } from "../../lib/date.ts";

interface Props {
	address: `0x${string}`;
}

export const App = (props: Props) => {
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
		navigate,
	} = usePool(() => {});

	return (
		<>
			<Header address={props.address} />
			<main>
				<Container>
					<div className="lens" />
					<nav className="pool-nav">
						<ul>
							{[...Array(poolsCount.data)].map((_, i) => (
								// biome-ignore lint/suspicious/noArrayIndexKey: order of items is fixed
								<li key={i}>
									<Link
										className={poolId === i ? "active" : ""}
										to={`/pool/${i}`}
									>
										Pool {i}
									</Link>
								</li>
							))}
						</ul>
					</nav>

					<div className="pool">
						<section className="section pool-progress">
							<Chart progress={poolProgress}>
								<dl>
									<dt>Elapsed</dt>
									<dd>
										{poolElapsedDays}/{poolTotalDays} days
									</dd>
								</dl>
							</Chart>

							<dl className="current-reward">
								<dt>Current Reward Balance</dt>
								<Separator />
								<dd>
									3,450 <span className="currency">MOR</span>
								</dd>
							</dl>
						</section>
						<section className="section pool-cta">
							<h2 className="section-heading">Stake tokens and earn rewards</h2>
							<Separator />
							<button
								type="button"
								className="button button-primary"
								onClick={() => navigate(`/pool/${poolId}/stake`)}
							>
								Stake
							</button>
						</section>
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
									<dd>7 days, 30 days, 180 days</dd>
								</dl>
							)}
						</section>
						<section className="section stake-info">
							<h2 className="section-heading">Stake info</h2>
							<Separator />
							<dl className="info">
								<dt>Amount staked</dt>
								<dd>100 LMR</dd>

								<dt>Share amount</dt>
								<dd>1,200</dd>

								<dt>Current reward</dt>
								<dd>100 MOR</dd>

								<dt>Lockup ends</dt>
								<dd>2021-08-08 7:00</dd>
							</dl>
						</section>
					</div>
				</Container>
			</main>
		</>
	);
};
