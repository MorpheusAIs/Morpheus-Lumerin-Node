import { Container } from "../components/Container.tsx";
import { Header } from "../components/Header.tsx";
import { Separator } from "../components/Separator.tsx";

export const Stake = () => {
	return (
		<>
			<Header address="0x1234567890abcdef" />
			<main>
				<Container>
					<div className="lens" />
					<section className="section add-stake">
						<h1>New staking contract</h1>
						<div className="field">
							<input id="stake-amount" type="number" />
							<label htmlFor="stake-amount">Logo LMR</label>
						</div>
						<Separator />
						<div className="field">
							<label htmlFor="lockup-period">Lockup period</label>
						</div>
						<div className="summary">
							<dl>
								<dt>APY</dt>
								<dd>4.19%</dd>
								<dt>Lockup Period</dt>
								<dd>30 days</dd>
								<dt>Reward multiplier</dt>
								<dd>1.15x</dd>
								<dt>Lockup ends at</dt>
								<dd>17 Apr 2025</dd>
							</dl>
						</div>
						<div className="field">
							<button type="button">Cancel</button>
							<button className="cta" type="submit">
								Stake
							</button>
						</div>
					</section>
				</Container>
			</main>
		</>
	);
};
