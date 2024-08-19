import { Header } from "../components/Header.tsx";
import { Link, useParams } from "react-router-dom";
import { Separator } from "../components/Separator.tsx";
import { PieChart } from "react-minimal-pie-chart";
import { Container } from "../components/Container.tsx";

const TOTAL_POOLS = 3;

export const App: React.FC = () => {
	const { poolId: poolIdString } = useParams();
	const poolId = Number(poolIdString);

	return (
		<>
			<Header address="0x1234567890abcdef" />
			<main>
				<Container>
					<div className="lens" />
					<nav className="pool-nav">
						<ul>
							{[...Array(TOTAL_POOLS)].map((_, i) => (
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
							<div className="chart">
								<PieChart
									data={[
										{ value: 30, color: "#fff" },
										{ value: 70, color: "#36C6D9" },
									]}
									totalValue={100}
									lineWidth={27}
									rounded={true}
									startAngle={-125}
								/>
								<div className="chart-text">
									<dl>
										<dt>Lockup Period</dt>
										<dd>26/30 days</dd>
									</dl>
								</div>
							</div>

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
						</section>
						<section className="section pool-stats">
							<h2 className="section-heading">Pool stats</h2>
							<Separator />
							<dl className="info">
								<dt>Reward per second</dt>
								<dd>0.525 MOR</dd>

								<dt>Total shares</dt>
								<dd>6,000,000</dd>

								<dt>Total staked</dt>
								<dd>3,000,000 LMR</dd>

								<dt>Start date</dt>
								<dd>2021-08-01 7:00</dd>

								<dt>End date</dt>
								<dd>2021-09-01 7:00</dd>

								<dt>Lockup periods</dt>
								<dd>7 days, 30 days, 180 days</dd>
							</dl>
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
