import { useState } from "react";
import { Container, ContainerNarrow } from "../components/Container.tsx";
import { Header } from "../components/Header.tsx";
import { Separator } from "../components/Separator.tsx";
import { LumerinIcon } from "../icons/LumerinIcon.tsx";
import { Range } from "react-range";

export const Stake = () => {
	const [lockupPeriod, setLockupPeriod] = useState(0);
	return (
		<>
			<Header address="0x1234567890abcdef" />
			<main>
				<div className="lens" />
				<ContainerNarrow>
					<section className="section add-stake">
						<h1>New staking contract</h1>
						<div className="field stake-amount">
							<input id="stake-amount" type="number" />
							<label htmlFor="stake-amount">
								<LumerinIcon /> LMR
							</label>
						</div>
						<Separator />
						<div className="field lockup-period">
							<label htmlFor="lockup-period">Lockup period</label>
							<Range
								label="Lockup period"
								values={[lockupPeriod]}
								min={0}
								max={4}
								onChange={(v) => setLockupPeriod(v[0])}
								renderTrack={({ props, children }) => (
									<div {...props} className="range-track">
										{children}
									</div>
								)}
								renderMark={({ props, index }) => (
									<div {...props} key={props.key} className="range-mark">
										<div className="range-mark-label">1 day</div>
									</div>
								)}
								renderThumb={({ props }) => (
									<div {...props} className="range-thumb" />
								)}
							/>
						</div>
						<dl className="field summary">
							<dt>APY</dt>
							<dd>4.19%</dd>
							<dt>Lockup Period</dt>
							<dd>30 days</dd>
							<dt>Reward multiplier</dt>
							<dd>1.15x</dd>
							<dt>Lockup ends at</dt>
							<dd>17 Apr 2025</dd>
						</dl>
						<div className="field buttons">
							<button className="button" type="button">
								Cancel
							</button>
							<button className="button button-primary" type="submit">
								Stake
							</button>
						</div>
					</section>
				</ContainerNarrow>
			</main>
		</>
	);
};
